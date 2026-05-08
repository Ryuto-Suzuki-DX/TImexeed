package seeders

/*
 * 〇 所属Seed
 */
type SeedDepartment struct {
	Name string
}

var departments = []SeedDepartment{
	{
		Name: "管理部",
	},
	{
		Name: "開発部",
	},
	{
		Name: "営業部",
	},
}

/*
 * 〇 ユーザーSeed
 */
type SeedUser struct {
	Name           string
	Email          string
	Password       string
	Role           string
	DepartmentName string
	HireDate       string
}

var users = []SeedUser{
	{
		Name:           "管理者ユーザー",
		Email:          "admin@example.com",
		Password:       "password123",
		Role:           "ADMIN",
		DepartmentName: "管理部",
		HireDate:       "2026-05-01",
	},
	{
		Name:           "一般ユーザー1",
		Email:          "user1@example.com",
		Password:       "password123",
		Role:           "USER",
		DepartmentName: "開発部",
		HireDate:       "2026-05-01",
	},
	{
		Name:           "一般ユーザー2",
		Email:          "user2@example.com",
		Password:       "password123",
		Role:           "USER",
		DepartmentName: "営業部",
		HireDate:       "2026-05-01",
	},
}

/*
 * 〇 勤怠区分マスタSeed
 *
 * 勤怠入力画面のプルダウンと入力制御に使う。
 */
type SeedAttendanceType struct {
	Code                 string
	Name                 string
	Category             string
	IsWorked             bool
	RequiresRequest      bool
	SyncPlanActual       bool
	AllowActualTimeInput bool
	AllowBreakInput      bool
	AllowTransportInput  bool
	AllowLateFlag        bool
	AllowEarlyLeaveFlag  bool
	AllowAbsenceFlag     bool
	AllowSickLeaveFlag   bool
	DisplayOrder         int
	IsActive             bool
}

var attendanceTypes = []SeedAttendanceType{
	{
		Code:                 "WORK",
		Name:                 "通常勤務",
		Category:             "WORK",
		IsWorked:             true,
		RequiresRequest:      false,
		SyncPlanActual:       false,
		AllowActualTimeInput: true,
		AllowBreakInput:      true,
		AllowTransportInput:  true,
		AllowLateFlag:        true,
		AllowEarlyLeaveFlag:  true,
		AllowAbsenceFlag:     true,
		AllowSickLeaveFlag:   true,
		DisplayOrder:         1,
		IsActive:             true,
	},
	{
		Code:                 "NIGHT_WORK",
		Name:                 "夜勤",
		Category:             "WORK",
		IsWorked:             true,
		RequiresRequest:      false,
		SyncPlanActual:       false,
		AllowActualTimeInput: true,
		AllowBreakInput:      true,
		AllowTransportInput:  true,
		AllowLateFlag:        true,
		AllowEarlyLeaveFlag:  true,
		AllowAbsenceFlag:     true,
		AllowSickLeaveFlag:   true,
		DisplayOrder:         2,
		IsActive:             true,
	},
	{
		Code:                 "HOLIDAY",
		Name:                 "休日",
		Category:             "HOLIDAY",
		IsWorked:             false,
		RequiresRequest:      false,
		SyncPlanActual:       true,
		AllowActualTimeInput: false,
		AllowBreakInput:      false,
		AllowTransportInput:  false,
		AllowLateFlag:        false,
		AllowEarlyLeaveFlag:  false,
		AllowAbsenceFlag:     false,
		AllowSickLeaveFlag:   false,
		DisplayOrder:         3,
		IsActive:             true,
	},
	{
		Code:                 "PAID_LEAVE",
		Name:                 "有給",
		Category:             "LEAVE",
		IsWorked:             false,
		RequiresRequest:      true,
		SyncPlanActual:       true,
		AllowActualTimeInput: false,
		AllowBreakInput:      false,
		AllowTransportInput:  false,
		AllowLateFlag:        false,
		AllowEarlyLeaveFlag:  false,
		AllowAbsenceFlag:     false,
		AllowSickLeaveFlag:   false,
		DisplayOrder:         4,
		IsActive:             true,
	},
	{
		Code:                 "SPECIAL_LEAVE",
		Name:                 "特別休暇",
		Category:             "LEAVE",
		IsWorked:             false,
		RequiresRequest:      true,
		SyncPlanActual:       true,
		AllowActualTimeInput: false,
		AllowBreakInput:      false,
		AllowTransportInput:  false,
		AllowLateFlag:        false,
		AllowEarlyLeaveFlag:  false,
		AllowAbsenceFlag:     false,
		AllowSickLeaveFlag:   false,
		DisplayOrder:         5,
		IsActive:             true,
	},
	{
		Code:                 "ABSENCE",
		Name:                 "欠勤",
		Category:             "ABSENCE",
		IsWorked:             false,
		RequiresRequest:      false,
		SyncPlanActual:       true,
		AllowActualTimeInput: false,
		AllowBreakInput:      false,
		AllowTransportInput:  false,
		AllowLateFlag:        false,
		AllowEarlyLeaveFlag:  false,
		AllowAbsenceFlag:     false,
		AllowSickLeaveFlag:   false,
		DisplayOrder:         6,
		IsActive:             true,
	},
	{
		Code:                 "SICK_LEAVE",
		Name:                 "病欠",
		Category:             "ABSENCE",
		IsWorked:             false,
		RequiresRequest:      false,
		SyncPlanActual:       true,
		AllowActualTimeInput: false,
		AllowBreakInput:      false,
		AllowTransportInput:  false,
		AllowLateFlag:        false,
		AllowEarlyLeaveFlag:  false,
		AllowAbsenceFlag:     false,
		AllowSickLeaveFlag:   false,
		DisplayOrder:         7,
		IsActive:             true,
	},
	{
		Code:                 "CAREGIVER_LEAVE",
		Name:                 "介護休業",
		Category:             "LEAVE",
		IsWorked:             false,
		RequiresRequest:      true,
		SyncPlanActual:       true,
		AllowActualTimeInput: false,
		AllowBreakInput:      false,
		AllowTransportInput:  false,
		AllowLateFlag:        false,
		AllowEarlyLeaveFlag:  false,
		AllowAbsenceFlag:     false,
		AllowSickLeaveFlag:   false,
		DisplayOrder:         8,
		IsActive:             true,
	},
	{
		Code:                 "CHILDCARE_LEAVE",
		Name:                 "育児休業",
		Category:             "LEAVE",
		IsWorked:             false,
		RequiresRequest:      true,
		SyncPlanActual:       true,
		AllowActualTimeInput: false,
		AllowBreakInput:      false,
		AllowTransportInput:  false,
		AllowLateFlag:        false,
		AllowEarlyLeaveFlag:  false,
		AllowAbsenceFlag:     false,
		AllowSickLeaveFlag:   false,
		DisplayOrder:         9,
		IsActive:             true,
	},
	{
		Code:                 "SUSPENSION",
		Name:                 "休職",
		Category:             "SUSPENSION",
		IsWorked:             false,
		RequiresRequest:      true,
		SyncPlanActual:       true,
		AllowActualTimeInput: false,
		AllowBreakInput:      false,
		AllowTransportInput:  false,
		AllowLateFlag:        false,
		AllowEarlyLeaveFlag:  false,
		AllowAbsenceFlag:     false,
		AllowSickLeaveFlag:   false,
		DisplayOrder:         10,
		IsActive:             true,
	},
}
