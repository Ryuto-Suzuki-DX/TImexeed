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
		&models.Department{},
		&models.User{},
		&models.UserSalaryDetail{},
		&models.AllowanceType{},
		&models.MonthlyAllowance{},
		&models.SharedDocumentDriveFolder{},
		&models.AttendanceType{},
		&models.AttendanceDay{},
		&models.AttendanceTransportExpense{},
		&models.AttendanceBreak{},
		&models.MonthlyCommuterPass{},
		&models.PaidLeaveUsage{},
		&models.MonthlyAttendanceRequest{},
		&models.Notification{},
		&models.NotificationReminder{},
		&models.HolidayDate{},
		&models.ExternalStorageLink{},
		&models.ApiOperationLog{},
		&models.PersonalInformationDriveFolder{},
		&models.Expense{},
		&models.AttendanceRealtimeEvent{},
	)
}
