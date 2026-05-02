package seeders

import (
	"timexeed/backend/internal/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

/*
 * ユーザー初期データ投入
 */
func SeedUsers(db *gorm.DB) error {
	users := []struct {
		Name       string
		Email      string
		Password   string
		Role       string
		Department string
	}{
		{
			Name:       "管理者",
			Email:      "admin@example.com",
			Password:   "password123",
			Role:       "ADMIN",
			Department: "管理部",
		},
		{
			Name:       "山田太郎",
			Email:      "yamada@example.com",
			Password:   "password123",
			Role:       "USER",
			Department: "開発部",
		},
		{
			Name:       "佐藤花子",
			Email:      "sato@example.com",
			Password:   "password123",
			Role:       "USER",
			Department: "営業部",
		},
	}

	for _, user := range users {
		var count int64

		if err := db.
			Model(&models.User{}).
			Where("email = ?", user.Email).
			Count(&count).Error; err != nil {
			return err
		}

		if count > 0 {
			continue
		}

		var department models.Department

		if err := db.
			Where("name = ? AND is_deleted = ?", user.Department, false).
			First(&department).Error; err != nil {
			return err
		}

		passwordHash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}

		createUser := models.User{
			Name:         user.Name,
			Email:        user.Email,
			PasswordHash: string(passwordHash),
			Role:         user.Role,
			DepartmentID: &department.ID,
			IsDeleted:    false,
		}

		if err := db.Create(&createUser).Error; err != nil {
			return err
		}
	}

	return nil
}
