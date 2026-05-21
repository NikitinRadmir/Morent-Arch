package main

import (
	//"user-system/app/token"
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	//"strings"

	"user-system/app/di"
	kafkamsg "user-system/app/messaging/kafka"
	"user-system/app/middleware"
	//"user-system/app/graphql/generated"
	"user-system/app/models"
	"user-system/app/repositories"
	"user-system/app/routes"
	"user-system/app/service"
	"user-system/pkg/logger"

	//"github.com/99designs/gqlgen/graphql/handler"
	"github.com/gin-gonic/gin"
	//"github.com/google/uuid"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	log := logger.New()
	slog.SetDefault(log)

	defaultClearDB := false
	if envValue := os.Getenv("CLEAR_DB_ON_START"); envValue != "" {
		parsedValue, err := strconv.ParseBool(envValue)
		if err != nil {
			log.Warn("invalid CLEAR_DB_ON_START, using default false", "value", envValue, "error", err)
		} else {
			defaultClearDB = parsedValue
		}
	}

	clearDB := flag.Bool("clear-db", defaultClearDB, "Clear database before migrations")
	migrateOnly := flag.Bool("migrate-only", false, "Run migrations only, don't start server")
	flag.Parse()

	if err := godotenv.Load(); err != nil {
		log.Info("env file not found, using environment variables")
	}

	requiredEnvVars := []string{"DB_HOST", "DB_USER", "DB_PASSWORD", "DB_NAME", "JWT_SECRET"}
	for _, envVar := range requiredEnvVars {
		if value := os.Getenv(envVar); value == "" {
			log.Error("missing required environment variable", "name", envVar)
			os.Exit(1)
		}
	}

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_SSL"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Error("database connection failed", "error", err)
		os.Exit(1)
	}
	log.Info("connected to postgres")

	log.Info("running database migrations")
	if *clearDB {
		log.Info("clearing database before migrations")
	}

	err = runMigrations(db, *clearDB, log)
	if err != nil {
		log.Error("migration failed", "error", err)
		os.Exit(1)
	}
	log.Info("database migrations completed")

	if *migrateOnly {
		log.Info("migrate-only mode, exiting")
		return
	}

	log.Info("initializing controllers")
	userController := di.InitUserController(db)
	roleController := di.InitRoleController(db)
	companyController := di.InitCompanyController(db)
	rolePermissionController := di.InitRolePermissionController(db)
	authController := di.InitAuthController(db)
	log.Info("controllers initialized")

	log.Info("initializing json-rpc services")
	rpcServer := di.InitRPCServer(db)
	log.Info("json-rpc services initialized")

	log.Info("initializing graphql services")
	//graphQLResolver := di.InitGraphQLResolver(db)
	//tokenService := token.NewTokenService(os.Getenv("JWT_SECRET"))
	//graphqlServer := handler.New(generated.NewExecutableSchema(generated.Config{Resolvers: graphQLResolver}))
	log.Info("graphql services initialized")

	router := gin.New()
	router.Use(middleware.RequestLog(log))
	router.Use(middleware.Recover(log))

	log.Info("setting up routes")
	routes.SetupRoutes(
		router,
		userController,
		roleController,
		companyController,
		rolePermissionController,
		authController,
		rpcServer,
		nil, //TODO: передать инициализацию
	)
	log.Info("routes configured")

	startKafkaConsumer(db, log)

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	log.Info("server starting",
		"addr", "http://localhost:"+port,
		"swagger", "http://localhost:"+port+"/swagger/index.html",
	)

	if err := router.Run(":" + port); err != nil {
		log.Error("server failed", "error", err)
		os.Exit(1)
	}
}

func runMigrations(db *gorm.DB, clearDB bool, log *slog.Logger) error {
	if clearDB {
		log.Info("dropping existing tables")

		db.Exec("DROP INDEX IF EXISTS idx_companies_domain CASCADE")
		db.Exec("DROP INDEX IF EXISTS uix_companies_domain CASCADE")
		db.Exec("DROP INDEX IF EXISTS companies_domain_key CASCADE")

		db.Exec("DROP TABLE IF EXISTS role_permissions CASCADE")
		db.Exec("DROP TABLE IF EXISTS user_roles CASCADE")
		db.Exec("DROP TABLE IF EXISTS users CASCADE")
		db.Exec("DROP TABLE IF EXISTS roles CASCADE")
		db.Exec("DROP TABLE IF EXISTS permissions CASCADE")
		db.Exec("DROP TABLE IF EXISTS companies CASCADE")

		log.Info("tables dropped")
	}

	modelsToMigrate := []interface{}{
		&models.Company{},
		&models.User{},
		&models.Role{},
		&models.Permission{},
		&models.UserRole{},
		&models.RolePermission{},
		&models.ProcessedEvent{},
	}

	for _, model := range modelsToMigrate {
		err := db.AutoMigrate(model)
		if err != nil {
			return fmt.Errorf("failed to migrate %T: %v", model, err)
		}
		log.Info("table migrated", "model", fmt.Sprintf("%T", model))
	}

	return nil
}

func startKafkaConsumer(db *gorm.DB, log *slog.Logger) {
	enabled := os.Getenv("KAFKA_ENABLED")
	if enabled == "false" || enabled == "0" {
		log.Info("kafka consumer disabled", "reason", "KAFKA_ENABLED=false")
		return
	}
	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		log.Info("kafka consumer disabled", "reason", "KAFKA_BROKERS empty")
		return
	}

	syncService := service.NewMorentSyncService(
		db,
		repositories.NewUserRepository(db),
		repositories.NewCompanyRepository(db),
		repositories.NewRoleRepository(db),
		repositories.NewProcessedEventRepository(db),
		os.Getenv("MORENT_COMPANY_NAME"),
	)

	consumer := kafkamsg.NewConsumer(
		brokers,
		os.Getenv("KAFKA_TOPIC_USERS"),
		os.Getenv("KAFKA_CONSUMER_GROUP"),
		syncService,
	)

	go func() {
		if err := consumer.Run(context.Background()); err != nil {
			log.Error("kafka consumer stopped", "error", err)
		}
	}()
}
