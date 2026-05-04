package seeders

import "gorm.io/gorm"

/*
 * 初期データ投入
 *
 * 初回起動時に必要なマスタデータを登録する
 */
func RunSeeders(db *gorm.DB) error {
	// 所属
	if err := SeedDepartments(db); err != nil {
		return err
	}

	// ユーザー
	if err := SeedUsers(db); err != nil {
		return err
	}

	// 有給数
	if err := SeedPaidLeaveGrants(db); err != nil {
		return err
	}

	// 有給調整
	if err := SeedPaidLeaveAdjustments(db); err != nil {
		return err
	}

	// 勤怠
	if err := SeedAttendanceRecords(db); err != nil {
		return err
	}

	// 休憩
	if err := SeedAttendanceBreaks(db); err != nil {
		return err
	}

	// 各日交通費
	if err := SeedAttendanceTransportations(db); err != nil {
		return err
	}

	// 定期
	if err := SeedCommuterPasses(db); err != nil {
		return err
	}

	return nil
}
