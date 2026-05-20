package main

import (
	//"user-system/app/token"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	//"strings"

	"user-system/app/di"
	kafkamsg "user-system/app/messaging/kafka"
	//"user-system/app/graphql/generated"
	"user-system/app/models"
	"user-system/app/repositories"
	"user-system/app/routes"
	"user-system/app/service"

	//"github.com/99designs/gqlgen/graphql/handler"
	"github.com/gin-gonic/gin"
	//"github.com/google/uuid"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	defaultClearDB := false
	if envValue := os.Getenv("CLEAR_DB_ON_START"); envValue != "" {
		parsedValue, err := strconv.ParseBool(envValue)
		if err != nil {
			log.Printf("Invalid CLEAR_DB_ON_START value %q, using default false", envValue)
		} else {
			defaultClearDB = parsedValue
		}
	}

	clearDB := flag.Bool("clear-db", defaultClearDB, "Clear database before migrations")
	migrateOnly := flag.Bool("migrate-only", false, "Run migrations only, don't start server")
	flag.Parse()

	if err := godotenv.Load(); err != nil {
		log.Println(" .env not found, using docker environment")
	}

	requiredEnvVars := []string{"DB_HOST", "DB_USER", "DB_PASSWORD", "DB_NAME", "JWT_SECRET"}
	for _, envVar := range requiredEnvVars {
		if value := os.Getenv(envVar); value == "" {
			log.Fatalf("Missing required environment variable: %s", envVar)
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
		log.Fatalf("DB connection failed: %v", err)
	}
	log.Println(" Connected to Postgres")

	log.Println(" Running database migrations...")
	if *clearDB {
		log.Println(" Clearing database...")
	}

	err = runMigrations(db, *clearDB)
	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
	log.Println(" Database migrations completed")

	if *migrateOnly {
		log.Println(" Migration completed successfully (--migrate-only flag set)")
		return
	}

	log.Println(" Initializing controllers...")
	userController := di.InitUserController(db)
	roleController := di.InitRoleController(db)
	companyController := di.InitCompanyController(db)
	rolePermissionController := di.InitRolePermissionController(db)
	authController := di.InitAuthController(db)
	log.Println(" Controllers initialized successfully")

	// ========== НАЧАЛО: JSON-RPC инициализация ==========
	log.Println(" Initializing JSON-RPC services...")
	rpcServer := di.InitRPCServer(db)
	log.Println(" JSON-RPC services initialized successfully")
	// ========== КОНЕЦ: JSON-RPC инициализация ==========

	// ========== НАЧАЛО: GraphQL инициализация через Wire ==========
	log.Println(" Initializing GraphQL services...")
	//graphQLResolver := di.InitGraphQLResolver(db)
	//tokenService := token.NewTokenService(os.Getenv("JWT_SECRET"))
	//graphqlServer := handler.New(generated.NewExecutableSchema(generated.Config{Resolvers: graphQLResolver}))
	log.Println(" GraphQL services initialized successfully")
	// ========== КОНЕЦ: GraphQL инициализация ==========

	router := gin.Default()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// ========== GraphQL Middleware ==========
	//router.Use(func(c *gin.Context) {
	//	if c.Request.URL.Path == "/query" || c.Request.URL.Path == "/graphql" {
	//		authHeader := c.GetHeader("Authorization")
	//		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
	//			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	//			claims, err := tokenService.ValidateToken(tokenStr)
	//			if err == nil {
	//				userID, _ := uuid.Parse(claims.UserId)
	//				c.Set("user_id", userID)
	//			}
	//		}
	//	}
	//	c.Next()
	//})
	// ========== КОНЕЦ: GraphQL Middleware ==========

	log.Println(" Setting up routes...")
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
	log.Println(" Routes setup completed")

	startKafkaConsumer(db)

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf(" Server running at http://localhost:%s", port)
	log.Println(" Swagger UI: http://localhost:" + port + "/swagger/index.html")
	log.Println(" Available endpoints:")
	log.Println("  POST   /api/auth/register")
	log.Println("  POST   /api/auth/login")
	log.Println("  GET    /health")
	log.Println("  POST   /api/companies (public)")

	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func runMigrations(db *gorm.DB, clearDB bool) error {
	if clearDB {
		log.Println(" Clearing existing tables...")

		db.Exec("DROP INDEX IF EXISTS idx_companies_domain CASCADE")
		db.Exec("DROP INDEX IF EXISTS uix_companies_domain CASCADE")
		db.Exec("DROP INDEX IF EXISTS companies_domain_key CASCADE")

		db.Exec("DROP TABLE IF EXISTS role_permissions CASCADE")
		db.Exec("DROP TABLE IF EXISTS user_roles CASCADE")
		db.Exec("DROP TABLE IF EXISTS users CASCADE")
		db.Exec("DROP TABLE IF EXISTS roles CASCADE")
		db.Exec("DROP TABLE IF EXISTS permissions CASCADE")
		db.Exec("DROP TABLE IF EXISTS companies CASCADE")

		log.Println(" All tables dropped")
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
		log.Printf("  ✓ Table created: %T", model)
	}

	return nil
}

func startKafkaConsumer(db *gorm.DB) {
	enabled := os.Getenv("KAFKA_ENABLED")
	if enabled == "false" || enabled == "0" {
		log.Println(" Kafka consumer disabled (KAFKA_ENABLED=false)")
		return
	}
	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		log.Println(" Kafka consumer disabled (KAFKA_BROKERS is empty)")
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
			log.Printf(" Kafka consumer stopped: %v", err)
		}
	}()
}
