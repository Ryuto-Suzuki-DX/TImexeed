package seeders

import (
	"timexeed/backend/internal/models"

	"gorm.io/gorm"
)

/*
 * 所属マスタ初期データ投入
 */
func SeedDepartments(db *gorm.DB) error {
	departments := []models.Department{
		{Name: "管理部"},
		{Name: "開発部"},
		{Name: "営業部"},
	}

	for _, department := range departments {
		var count int64

		if err := db.
			Model(&models.Department{}).
			Where("name = ?", department.Name).
			Count(&count).Error; err != nil {
			return err
		}

		if count > 0 {
			continue
		}

		if err := db.Create(&department).Error; err != nil {
			return err
		}
	}

	return nil
}
