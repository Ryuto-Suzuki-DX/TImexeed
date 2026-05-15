package migrations

import (
	"timexeed/backend/internal/models"

	"gorm.io/gorm"
)

/*
 * DBマイグレーション
 *
 * テーブル構造をDBへ反映する
 */
func RunMigrations(db *gorm.DB) error {
	return db.AutoMigrate(

		// 所属
		&models.Department{},
		// ユーザー
		&models.User{},
		// 各ユーザーの給与詳細
		&models.UserSalaryDetail{},
		// 勤怠区分マスタ
		&models.AttendanceType{},
		// 勤怠/日
		&models.AttendanceDay{},
		// 休憩/日
		&models.AttendanceBreak{},
		// 通勤定期/月
		&models.MonthlyCommuterPass{},
		// 有給使用日
		&models.PaidLeaveUsage{},
		// 勤怠申請系
		&models.MonthlyAttendanceRequest{},
		// お知らせ
		&models.Notification{},
		// 自動お知らせ機能
		&models.NotificationReminder{},
		// 祝日
		&models.HolidayDate{},
		// 外部ストレージリンク
		&models.ExternalStorageLink{},
	)
}
