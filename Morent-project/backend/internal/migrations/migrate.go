package migrations

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"morent-backend/internal/models"
)

func RunMigrations(db *gorm.DB) error {
	errAutoMigrate := db.AutoMigrate(
		&models.Car{},
		&models.Comment{},
		&models.User{},
		&models.Session{},
		&models.Favorite{},
		&models.Rental{},
	)

	if errAutoMigrate != nil {
		return fmt.Errorf("ошибка выполнения миграций: %v", errAutoMigrate)
	}

	// Стоковый админ только при явном флаге (dev).
	if !strings.EqualFold(strings.TrimSpace(os.Getenv("SEED_ADMIN")), "true") {
		fmt.Println("Миграции выполнены успешно")
		return nil
	}

	adminEmail := "admin@morent.com"
	var existingAdmin models.User
	if err := db.Where("email = ?", adminEmail).First(&existingAdmin).Error; err != nil {
		// Админ не существует, создаем
		hash, errHash := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		if errHash != nil {
			return fmt.Errorf("ошибка создания хеша пароля админа: %v", errHash)
		}

		admin := models.User{
			Name:          "Admin",
			Email:         adminEmail,
			PasswordHash:  string(hash),
			AvatarURL:     "https://avatars.mds.yandex.net/i?id=18025267d7d94e6289d82fda9b36eea0_l-5256838-images-thumbs&n=13",
			Nickname:      "Admin",
			Position:      "Administrator",
			Role:          "admin",
			EmailVerified: true,
		}

		if errCreate := db.Create(&admin).Error; errCreate != nil {
			return fmt.Errorf("ошибка создания админа: %v", errCreate)
		}
		fmt.Println("✅ Стоковый админ создан: admin@morent.com / admin123")
	} else {
		updates := map[string]any{}
		if strings.ToLower(strings.TrimSpace(existingAdmin.Role)) != "admin" {
			updates["role"] = "admin"
		}
		if !existingAdmin.EmailVerified {
			updates["email_verified"] = true
		}
		if strings.TrimSpace(existingAdmin.Nickname) == "" {
			updates["nickname"] = "Admin"
		}
		if strings.TrimSpace(existingAdmin.Position) == "" {
			updates["position"] = "Administrator"
		}
		if len(updates) > 0 {
			if errUpdate := db.Model(&existingAdmin).Updates(updates).Error; errUpdate != nil {
				return fmt.Errorf("ошибка обновления стокового админа: %v", errUpdate)
			}
			fmt.Println("✅ Стоковый админ обновлен: admin@morent.com / admin123")
		}
	}

	fmt.Println("Миграции выполнены успешно")
	return nil
}
