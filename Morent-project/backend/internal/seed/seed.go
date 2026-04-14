package seed

import (
	"fmt"
	"morent-backend/internal/models"
	"gorm.io/gorm"
)

func RunSeed(db *gorm.DB) error {
	if errTruncateComments := db.Exec("TRUNCATE TABLE comments CASCADE").Error; errTruncateComments != nil {
		return fmt.Errorf("ошибка очистки comments: %v", errTruncateComments)
	}
	if errTruncateCars := db.Exec("TRUNCATE TABLE cars CASCADE").Error; errTruncateCars != nil {
		return fmt.Errorf("ошибка очистки cars: %v", errTruncateCars)
	}
	_ = db.Exec("ALTER SEQUENCE cars_id_seq RESTART WITH 1").Error
	_ = db.Exec("ALTER SEQUENCE comments_id_seq RESTART WITH 1").Error

	cars := []models.Car{
		{
			Name:        "Skoda Octavia RS",
			Type:        "Sport",
			Capacity:    5,
			Price:       50.00,
			Description: "Powerful sport sedan with excellent acceleration",
			ImgSrc:      "/images/cars/OctaviaRsBlack.png",
			Fuel:        7.5,
			Transmission: "Manual",
		},
		{
			Name:        "Skoda Kodiaq",
			Type:        "SUV",
			Capacity:    7,
			Price:       65.00,
			Description: "Spacious SUV perfect for family trips",
			ImgSrc:      "/images/cars/KodiaqBlue.png",
			Fuel:        8.2,
			Transmission: "Automatic",
		},
		{
			Name:        "Skoda Fabia",
			Type:        "Sedan",
			Capacity:    5,
			Price:       35.00,
			Description: "Compact and economical city car",
			ImgSrc:      "/images/cars/FabiaBlack.png",
			Fuel:        5.5,
			Transmission: "Manual",
		},
		{
			Name:        "Skoda Enyaq",
			Type:        "Electric",
			Capacity:    5,
			Price:       75.00,
			Description: "Modern electric SUV with advanced technology",
			ImgSrc:      "/images/cars/EnyaqBlack.png",
			Fuel:        0, // Electric
			Transmission: "Automatic",
		},
	}

	for _, car := range cars {
		if errCreateCar := db.Create(&car).Error; errCreateCar != nil {
			return fmt.Errorf("ошибка создания машины %s: %v", car.Name, errCreateCar)
		}
	}

	comments := []models.Comment{
		{
			CarID:       1,
			Name:        "Alex Johnson",
			Post:        "Car Enthusiast",
			Photo:       "/images/reviewers/man.jpg",
			Date:        "2 days ago",
			Description: "Excellent car! Very comfortable and fast. Highly recommend!",
		},
		{
			CarID:       1,
			Name:        "Sarah Miller",
			Post:        "Travel Blogger",
			Photo:       "/images/reviewers/woman.jpg",
			Date:        "5 days ago",
			Description: "Perfect for long trips. The car handles beautifully on highways.",
		},
		{
			CarID:       2,
			Name:        "Mike Brown",
			Post:        "Family Man",
			Photo:       "/images/reviewers/man.jpg",
			Date:        "1 week ago",
			Description: "Great SUV for our family of 5. Plenty of space and very comfortable.",
		},
	}

	for _, comment := range comments {
		if errCreateComment := db.Create(&comment).Error; errCreateComment != nil {
			return fmt.Errorf("ошибка создания комментария: %v", errCreateComment)
		}
	}

	fmt.Println("✅ Seed данные загружены")
	return nil
}
