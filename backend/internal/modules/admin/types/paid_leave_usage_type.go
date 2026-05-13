package types

import "time"

/*
 * 管理者用 有給使用日 型定義
 *
 * このファイルには、管理者用の有給使用日管理機能で使う型をまとめる。
 *
 * まとめるもの：
 * ・Request型
 * ・Response型
 *
 * 対象API：
 * ・有給使用日取得
 * ・現時点の有給残数取得
 * ・過去有給使用日追加
 * ・過去有給使用日編集
 * ・過去有給使用日削除
 *
 * 方針：
 * ・URLにIDは載せない
 * ・対象ユーザーIDは targetUserId で受け取る
 * ・対象有給使用日IDは targetPaidLeaveUsageId で受け取る
 * ・ControllerではRequest型にbindして、そのままServiceへ渡す
 * ・Responseの日付は string に変換せず time.Time / *time.Time のまま返す
 * ・表示形式 yyyy-MM-dd などはフロント側で整形する
 * ・管理者による追加は Service側で isManual = true を強制する
 */

/*
 * =========================================================
 * Request
 * =========================================================
 */

/*
 * 有給使用日取得Request
 *
 * POST /admin/paid-leave-usages/search
 *
 * body例：
 * {
 *   "targetUserId": 1,
 *   "includeDeleted": false,
 *   "offset": 0,
 *   "limit": 50
 * }
 *
 * 用途：
 * ・対象ユーザーの有給使用日一覧を取得する
 * ・管理者画面のmodal表示
 * ・管理者の有給使用日編集画面
 */
type SearchPaidLeaveUsagesRequest struct {
	// 対象ユーザーID
	TargetUserID uint `json:"targetUserId" binding:"required"`

	// 削除済みの有給使用日も含めるか
	IncludeDeleted bool `json:"includeDeleted"`

	// 取得開始位置
	// 初回は0
	// さらに表示の場合は、現在フロントに表示済みの件数を入れる
	Offset int `json:"offset"`

	// 取得件数
	// 基本は50
	// 未指定、0以下、50超えの場合はService側で補正する
	Limit int `json:"limit"`
}

/*
 * 有給残数取得Request
 *
 * POST /admin/paid-leave-usages/balance
 *
 * body例：
 * {
 *   "targetUserId": 1
 * }
 *
 * 用途：
 * ・対象ユーザーの現時点の有給残数を取得する
 *
 * 注意：
 * ・現時点の残数なので targetDate は受け取らない
 * ・基準日は Service側で time.Now() を使う
 */
type GetPaidLeaveBalanceRequest struct {
	// 対象ユーザーID
	TargetUserID uint `json:"targetUserId" binding:"required"`
}

/*
 * 過去有給使用日追加Request
 *
 * POST /admin/paid-leave-usages/create
 *
 * body例：
 * {
 *   "targetUserId": 1,
 *   "usageDate": "2026-05-01",
 *   "usageDays": 1,
 *   "memo": "システム導入前使用分"
 * }
 *
 * UsageDate は "2026-05-01" のような文字列で受け取る。
 * time.Timeで直接受けるとJSONではRFC3339形式が必要になりやすいため、
 * Service側で日付文字列をparseする。
 *
 * 注意：
 * ・管理者による追加なので isManual はフロントから受け取らない
 * ・Service側で isManual = true を強制する
 */
type CreatePaidLeaveUsageRequest struct {
	// 対象ユーザーID
	TargetUserID uint `json:"targetUserId" binding:"required"`

	// 有給使用日
	// yyyy-MM-dd形式
	UsageDate string `json:"usageDate" binding:"required"`

	// 使用日数
	// 1日 = 1.0
	// 半日 = 0.5
	UsageDays float64 `json:"usageDays" binding:"required"`

	// メモ
	Memo string `json:"memo"`
}

/*
 * 過去有給使用日編集Request
 *
 * POST /admin/paid-leave-usages/update
 *
 * body例：
 * {
 *   "targetUserId": 1,
 *   "targetPaidLeaveUsageId": 10,
 *   "usageDate": "2026-05-01",
 *   "usageDays": 0.5,
 *   "memo": "半休に修正"
 * }
 *
 * 注意：
 * ・更新対象が targetUserId の有給使用日であることをService側で確認する
 * ・基本的には isManual = true のものだけ更新対象にする
 * ・勤怠連携や有給申請連携で作られたものをここで直接編集しない
 */
type UpdatePaidLeaveUsageRequest struct {
	// 対象ユーザーID
	TargetUserID uint `json:"targetUserId" binding:"required"`

	// 更新対象の有給使用日ID
	TargetPaidLeaveUsageID uint `json:"targetPaidLeaveUsageId" binding:"required"`

	// 有給使用日
	// yyyy-MM-dd形式
	UsageDate string `json:"usageDate" binding:"required"`

	// 使用日数
	// 1日 = 1.0
	// 半日 = 0.5
	UsageDays float64 `json:"usageDays" binding:"required"`

	// メモ
	Memo string `json:"memo"`
}

/*
 * 過去有給使用日削除Request
 *
 * POST /admin/paid-leave-usages/delete
 *
 * body例：
 * {
 *   "targetUserId": 1,
 *   "targetPaidLeaveUsageId": 10
 * }
 *
 * 注意：
 * ・物理削除ではなく論理削除する
 * ・更新対象が targetUserId の有給使用日であることをService側で確認する
 * ・基本的には isManual = true のものだけ削除対象にする
 */
type DeletePaidLeaveUsageRequest struct {
	// 対象ユーザーID
	TargetUserID uint `json:"targetUserId" binding:"required"`

	// 削除対象の有給使用日ID
	TargetPaidLeaveUsageID uint `json:"targetPaidLeaveUsageId" binding:"required"`
}

/*
 * =========================================================
 * Response
 * =========================================================
 */

/*
 * 有給使用日1件分のResponse
 *
 * 日付は time.Time / *time.Time のまま返す。
 * フロント側で表示時に yyyy-MM-dd などへ整形する。
 */
type PaidLeaveUsageResponse struct {
	ID uint `json:"id"`

	UserID uint `json:"userId"`

	UsageDate time.Time `json:"usageDate"`

	UsageDays float64 `json:"usageDays"`

	IsManual bool `json:"isManual"`

	Memo string `json:"memo"`

	IsDeleted bool       `json:"isDeleted"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	DeletedAt *time.Time `json:"deletedAt"`
}

/*
 * 有給使用日検索Response
 *
 * hasMore：
 * ・さらに表示するデータがある場合は true
 * ・すべて取得済みの場合は false
 *
 * フロント側の使い方：
 * ・hasMore が true なら「さらに表示」ボタンを表示する
 * ・hasMore が false なら「さらに表示」ボタンを非表示にする
 */
type SearchPaidLeaveUsagesResponse struct {
	PaidLeaveUsages []PaidLeaveUsageResponse `json:"paidLeaveUsages"`
	Total           int64                    `json:"total"`
	Offset          int                      `json:"offset"`
	Limit           int                      `json:"limit"`
	HasMore         bool                     `json:"hasMore"`
}

/*
 * 有給残数Response
 *
 * 現時点の有給残数を返す。
 *
 * TotalGrantedDays：
 * ・雇入れ日を基準に、現時点までに法定付与された合計日数
 *
 * UsedDays：
 * ・有給使用日テーブルに登録されている使用日数合計
 * ・isDeleted = false のものだけ集計する
 *
 * RemainingDays：
 * ・TotalGrantedDays - UsedDays
 *
 * NextGrantDate：
 * ・次回付与予定日
 * ・60年分のルールを超えた場合など、算出できない場合は nil
 *
 * NextGrantDays：
 * ・次回付与予定日数
 *
 * RequiredUseDays：
 * ・年5日取得義務の日数
 *
 * RequiredUseDeadline：
 * ・年5日取得義務の期限
 * ・現時点で義務判定対象がない場合は nil
 *
 * RequiredUseRemainingDays：
 * ・年5日取得義務に対して、あと何日取得が必要か
 */
type PaidLeaveBalanceResponse struct {
	TargetUserID uint `json:"targetUserId"`

	TotalGrantedDays float64 `json:"totalGrantedDays"`
	UsedDays         float64 `json:"usedDays"`
	RemainingDays    float64 `json:"remainingDays"`

	NextGrantDate *time.Time `json:"nextGrantDate"`
	NextGrantDays float64    `json:"nextGrantDays"`

	RequiredUseDays          float64    `json:"requiredUseDays"`
	RequiredUseDeadline      *time.Time `json:"requiredUseDeadline"`
	RequiredUseRemainingDays float64    `json:"requiredUseRemainingDays"`
}

/*
 * 有給使用日作成Response
 */
type CreatePaidLeaveUsageResponse struct {
	PaidLeaveUsage PaidLeaveUsageResponse `json:"paidLeaveUsage"`
}

/*
 * 有給使用日更新Response
 */
type UpdatePaidLeaveUsageResponse struct {
	PaidLeaveUsage PaidLeaveUsageResponse `json:"paidLeaveUsage"`
}

/*
 * 有給使用日削除Response
 */
type DeletePaidLeaveUsageResponse struct {
	TargetUserID           uint `json:"targetUserId"`
	TargetPaidLeaveUsageID uint `json:"targetPaidLeaveUsageId"`
}
