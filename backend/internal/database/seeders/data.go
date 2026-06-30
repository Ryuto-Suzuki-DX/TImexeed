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
	Name               string
	Email              string
	Password           string
	Role               string
	DepartmentName     string
	HireDate           string
	MustChangePassword bool
}

var users = []SeedUser{
	{
		Name:               "管理者ユーザー",
		Email:              "admin@example.com",
		Password:           "password123",
		Role:               "ADMIN",
		DepartmentName:     "管理部",
		HireDate:           "2025-05-01",
		MustChangePassword: false,
	},
	{
		Name:               "一般ユーザー1",
		Email:              "user1@example.com",
		Password:           "password123",
		Role:               "USER",
		DepartmentName:     "開発部",
		HireDate:           "2025-05-01",
		MustChangePassword: true,
	},
	{
		Name:               "一般ユーザー2",
		Email:              "user2@example.com",
		Password:           "password123",
		Role:               "USER",
		DepartmentName:     "営業部",
		HireDate:           "2026-05-01",
		MustChangePassword: true,
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

/*
 * 〇 外部ストレージリンクSeed
 *
 * Google Driveなど、Timexeed外で管理するフォルダURLやファイルURLを登録する。
 */
type SeedExternalStorageLink struct {
	LinkType    string
	LinkName    string
	URL         string
	Description string
	Memo        string
}

var externalStorageLinks = []SeedExternalStorageLink{
	{
		LinkType:    "PERSONAL_INFORMATION_DRIVE_ROOT",
		LinkName:    "個人情報Drive親フォルダ",
		URL:         "https://drive.google.com/drive/folders/1mUEVTa-XoQ6IzlEyQf0uErHC-xOndMff",
		Description: "ユーザーごとの個人情報Driveフォルダを作成する親フォルダ",
		Memo:        "親フォルダはアプリ用Googleアカウントのみ共有する。ユーザー本人・管理者への共有は子フォルダ単位で同期する。",
	},
	{
		LinkType:    "EXPENSE_RECEIPT_BOX",
		LinkName:    "経費レシート格納先",
		URL:         "https://drive.google.com/drive/folders/10II_cvD7lTlmX6OvcLpz8eZT3Shyp4NU",
		Description: "経費申請でアップロードされた領収書画像を保存するGoogle Driveフォルダです。",
		Memo:        "経費申請の領収書保存先。管理者設定画面から変更可能。",
	},
	{
		LinkType:    "SHARED_DOCUMENT_DRIVE_ROOT",
		LinkName:    "共有資料Drive親フォルダ",
		URL:         "https://drive.google.com/drive/folders/1EcbkLhImlWZHxPHfeFFD7u5CFKxiv8fo",
		Description: "共有資料・FAQなど、全ユーザー向けDriveフォルダを作成する親フォルダ",
		Memo:        "この親フォルダ配下にTimexeedが共有資料Driveフォルダを自動作成する。Drive権限は管理者の同期操作で有効な一般ユーザーへ付与する。",
	},
	{
		LinkType:    "SYSTEM_LOG_DRIVE_ROOT",
		LinkName:    "システムログDrive親フォルダ",
		URL:         "https://drive.google.com/drive/folders/ここにログ保存用フォルダID",
		Description: "API操作ログやシステムログを保存するGoogle Drive親フォルダ",
		Memo:        "API操作ログはDBに保存し、日次CSVとしてこのDriveフォルダ配下へアップロードする。必要に応じてエラーログ、バッチログ、メール送信ログもこの配下で管理する。",
	},
}
