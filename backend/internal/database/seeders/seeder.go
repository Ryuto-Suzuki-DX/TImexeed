package seeders

import "gorm.io/gorm"

/*
 * 初期データ投入
 *
 * 初回起動時に必要なマスタデータを登録する
 */
func RunSeeders(db *gorm.DB) error {
	if err := SeedDepartments(db); err != nil {
		return err
	}

	if err := SeedUsers(db); err != nil {
		return err
	}

	return nil
}
