package seeders

import (
	"errors"

	"timexeed/backend/internal/models"
	"timexeed/backend/internal/utils"

	"gorm.io/gorm"
)

/*
 * DB初期データ投入
 *
 * 順番：
 * 1. 所属
 * 2. 勤怠区分マスタ
 * 3. ユーザー
 * 4. 外部ストレージリンク
 *
 * ユーザーは所属IDを持つため、所属を先に作成する。
 * 外部ストレージリンクは他テーブルに依存しないため最後に作成する。
 */
func RunSeeders(db *gorm.DB) error {
	for _, department := range departments {
		if err := createDepartmentIfNotExists(db, department); err != nil {
			return err
		}
	}

	for _, attendanceType := range attendanceTypes {
		if err := createAttendanceTypeIfNotExists(db, attendanceType); err != nil {
			return err
		}
	}

	for _, user := range users {
		if err := createUserIfNotExists(db, user); err != nil {
			return err
		}
	}

	for _, externalStorageLink := range externalStorageLinks {
		if err := createExternalStorageLinkIfNotExists(db, externalStorageLink); err != nil {
			return err
		}
	}

	return nil
}

/*
 * 所属名が存在しない場合のみ所属を作成する
 */
func createDepartmentIfNotExists(db *gorm.DB, seed SeedDepartment) error {
	var existing models.Department

	err := db.Where("name = ?", seed.Name).First(&existing).Error
	if err == nil {
		return nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	department := models.Department{
		Name:      seed.Name,
		IsDeleted: false,
	}

	return db.Create(&department).Error
}

/*
 * メールアドレスが存在しない場合のみユーザーを作成する
 */
func createUserIfNotExists(db *gorm.DB, seed SeedUser) error {
	var existing models.User

	err := db.Where("email = ?", seed.Email).First(&existing).Error
	if err == nil {
		return nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	passwordHash, err := utils.HashPassword(seed.Password)
	if err != nil {
		return err
	}

	hireDate, err := utils.ParseDate(seed.HireDate)
	if err != nil {
		return err
	}

	departmentID, err := findDepartmentIDByName(db, seed.DepartmentName)
	if err != nil {
		return err
	}

	user := models.User{
		Name:               seed.Name,
		Email:              seed.Email,
		PasswordHash:       passwordHash,
		Role:               seed.Role,
		DepartmentID:       departmentID,
		HireDate:           hireDate,
		MustChangePassword: seed.MustChangePassword,
		IsDeleted:          false,
	}

	return db.Create(&user).Error
}

/*
 * 所属名から所属IDを取得する
 *
 * DepartmentID は nullable だが、
 * SeedUser に DepartmentName が設定されている場合は紐づける。
 */
func findDepartmentIDByName(db *gorm.DB, departmentName string) (*uint, error) {
	if departmentName == "" {
		return nil, nil
	}

	var department models.Department

	err := db.Where("name = ? AND is_deleted = ?", departmentName, false).First(&department).Error
	if err != nil {
		return nil, err
	}

	return &department.ID, nil
}

/*
 * 勤怠区分コードが存在しない場合のみ勤怠区分を作成する
 */
func createAttendanceTypeIfNotExists(db *gorm.DB, seed SeedAttendanceType) error {
	var existing models.AttendanceType

	err := db.Where("code = ?", seed.Code).First(&existing).Error
	if err == nil {
		return nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	attendanceType := models.AttendanceType{
		Code:                 seed.Code,
		Name:                 seed.Name,
		Category:             seed.Category,
		IsWorked:             seed.IsWorked,
		RequiresRequest:      seed.RequiresRequest,
		SyncPlanActual:       seed.SyncPlanActual,
		AllowActualTimeInput: seed.AllowActualTimeInput,
		AllowBreakInput:      seed.AllowBreakInput,
		AllowTransportInput:  seed.AllowTransportInput,
		AllowLateFlag:        seed.AllowLateFlag,
		AllowEarlyLeaveFlag:  seed.AllowEarlyLeaveFlag,
		AllowAbsenceFlag:     seed.AllowAbsenceFlag,
		AllowSickLeaveFlag:   seed.AllowSickLeaveFlag,
		DisplayOrder:         seed.DisplayOrder,
		IsActive:             seed.IsActive,
		IsDeleted:            false,
	}

	return db.Create(&attendanceType).Error
}

/*
 * 外部ストレージリンク種別が存在しない場合のみ作成する
 *
 * 既に存在する場合は、Seed定義の内容で更新する。
 * 論理削除済みの場合も IsDeleted=false に戻す。
 */
func createExternalStorageLinkIfNotExists(db *gorm.DB, seed SeedExternalStorageLink) error {
	description := seed.Description
	memo := seed.Memo

	var existing models.ExternalStorageLink

	err := db.Where("link_type = ?", seed.LinkType).First(&existing).Error
	if err == nil {
		existing.LinkName = seed.LinkName
		existing.URL = seed.URL
		existing.Description = &description
		existing.Memo = &memo
		existing.IsDeleted = false

		return db.Save(&existing).Error
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	externalStorageLink := models.ExternalStorageLink{
		LinkType:    seed.LinkType,
		LinkName:    seed.LinkName,
		URL:         seed.URL,
		Description: &description,
		Memo:        &memo,
		IsDeleted:   false,
	}

	return db.Create(&externalStorageLink).Error
}
