package builders

import (
	"time"

	"timexeed/backend/internal/models"
	"timexeed/backend/internal/modules/admin/types"
	"timexeed/backend/internal/results"

	"gorm.io/gorm"
)

/*
 * 管理者用お知らせ自動リマインドBuilder interface
 *
 * ServiceがBuilderに求める処理だけを定義する。
 */
type NotificationReminderBuilder interface {
	BuildSearchNotificationRemindersQuery(req types.SearchNotificationRemindersRequest) (*gorm.DB, results.Result)
	BuildCreateNotificationReminderModel(req types.CreateNotificationReminderRequest) (models.NotificationReminder, results.Result)
	BuildFindNotificationReminderByIDQuery(reminderID uint, includeDeleted bool) (*gorm.DB, results.Result)
	BuildUpdateNotificationReminderModel(currentReminder models.NotificationReminder, req types.UpdateNotificationReminderRequest) (models.NotificationReminder, results.Result)
	BuildDeleteNotificationReminderModel(currentReminder models.NotificationReminder) (models.NotificationReminder, results.Result)
	BuildToggleNotificationReminderEnabledModel(currentReminder models.NotificationReminder, req types.ToggleNotificationReminderEnabledRequest) (models.NotificationReminder, results.Result)
}

/*
 * 管理者用お知らせ自動リマインドBuilder
 *
 * 役割：
 * ・Serviceから受け取った値をもとにGORMクエリを作成する
 * ・Serviceから受け取ったRequest / ModelをもとにDB保存用Modelを作成する
 * ・Builder内で発生したエラーはBuilderでcode/message/detailsを作って返す
 *
 * 注意：
 * ・DB実行はしない
 * ・Find / Save / Create はRepositoryに任せる
 */
type notificationReminderBuilder struct {
	db *gorm.DB
}

/*
 * NotificationReminderBuilder生成
 */
func NewNotificationReminderBuilder(db *gorm.DB) NotificationReminderBuilder {
	return &notificationReminderBuilder{db: db}
}

/*
 * 自動リマインド検索用クエリ作成
 *
 * 管理者画面で自動リマインド設定一覧を取得する。
 */
func (builder *notificationReminderBuilder) BuildSearchNotificationRemindersQuery(
	req types.SearchNotificationRemindersRequest,
) (*gorm.DB, results.Result) {
	if req.Limit <= 0 {
		return nil, results.BadRequest(
			"BUILD_SEARCH_NOTIFICATION_REMINDERS_QUERY_INVALID_LIMIT",
			"自動リマインド検索条件の作成に失敗しました",
			map[string]any{
				"limit": req.Limit,
			},
		)
	}

	if req.Offset < 0 {
		return nil, results.BadRequest(
			"BUILD_SEARCH_NOTIFICATION_REMINDERS_QUERY_INVALID_OFFSET",
			"自動リマインド検索条件の作成に失敗しました",
			map[string]any{
				"offset": req.Offset,
			},
		)
	}

	query := builder.db.
		Model(&models.NotificationReminder{})

	if req.Keyword != "" {
		likeKeyword := "%" + req.Keyword + "%"
		query = query.Where(
			"title LIKE ? OR message LIKE ?",
			likeKeyword,
			likeKeyword,
		)
	}

	if !req.IncludeDisabled {
		query = query.Where("is_enabled = ?", true)
	}

	if !req.IncludeDeleted {
		query = query.Where("is_deleted = ?", false)
	}

	query = query.
		Order("created_at DESC").
		Order("id DESC").
		Limit(req.Limit).
		Offset(req.Offset)

	return query, results.OK(
		nil,
		"BUILD_SEARCH_NOTIFICATION_REMINDERS_QUERY_SUCCESS",
		"",
		nil,
	)
}

/*
 * 自動リマインド作成用Model作成
 */
func (builder *notificationReminderBuilder) BuildCreateNotificationReminderModel(
	req types.CreateNotificationReminderRequest,
) (models.NotificationReminder, results.Result) {
	if req.Title == "" {
		return models.NotificationReminder{}, results.BadRequest(
			"BUILD_CREATE_NOTIFICATION_REMINDER_MODEL_EMPTY_TITLE",
			"自動リマインド作成データの作成に失敗しました",
			nil,
		)
	}

	if req.Message == "" {
		return models.NotificationReminder{}, results.BadRequest(
			"BUILD_CREATE_NOTIFICATION_REMINDER_MODEL_EMPTY_MESSAGE",
			"自動リマインド作成データの作成に失敗しました",
			nil,
		)
	}

	if req.DayOffsetFromMonthEnd < 0 {
		return models.NotificationReminder{}, results.BadRequest(
			"BUILD_CREATE_NOTIFICATION_REMINDER_MODEL_INVALID_DAY_OFFSET_FROM_MONTH_END",
			"自動リマインド作成データの作成に失敗しました",
			map[string]any{
				"dayOffsetFromMonthEnd": req.DayOffsetFromMonthEnd,
			},
		)
	}

	if req.SendHour < 0 || req.SendHour > 23 {
		return models.NotificationReminder{}, results.BadRequest(
			"BUILD_CREATE_NOTIFICATION_REMINDER_MODEL_INVALID_SEND_HOUR",
			"自動リマインド作成データの作成に失敗しました",
			map[string]any{
				"sendHour": req.SendHour,
			},
		)
	}

	if req.SendMinute < 0 || req.SendMinute > 59 {
		return models.NotificationReminder{}, results.BadRequest(
			"BUILD_CREATE_NOTIFICATION_REMINDER_MODEL_INVALID_SEND_MINUTE",
			"自動リマインド作成データの作成に失敗しました",
			map[string]any{
				"sendMinute": req.SendMinute,
			},
		)
	}

	reminder := models.NotificationReminder{
		Title:                 req.Title,
		Message:               req.Message,
		DayOffsetFromMonthEnd: req.DayOffsetFromMonthEnd,
		SendHour:              req.SendHour,
		SendMinute:            req.SendMinute,
		IsEnabled:             true,
		IsDeleted:             false,
	}

	return reminder, results.OK(
		nil,
		"BUILD_CREATE_NOTIFICATION_REMINDER_MODEL_SUCCESS",
		"",
		nil,
	)
}

/*
 * 自動リマインドID指定取得用クエリ作成
 */
func (builder *notificationReminderBuilder) BuildFindNotificationReminderByIDQuery(
	reminderID uint,
	includeDeleted bool,
) (*gorm.DB, results.Result) {
	if reminderID == 0 {
		return nil, results.BadRequest(
			"BUILD_FIND_NOTIFICATION_REMINDER_BY_ID_QUERY_INVALID_REMINDER_ID",
			"自動リマインド取得条件の作成に失敗しました",
			map[string]any{
				"reminderId": reminderID,
			},
		)
	}

	query := builder.db.
		Model(&models.NotificationReminder{}).
		Where("id = ?", reminderID)

	if !includeDeleted {
		query = query.Where("is_deleted = ?", false)
	}

	return query, results.OK(
		nil,
		"BUILD_FIND_NOTIFICATION_REMINDER_BY_ID_QUERY_SUCCESS",
		"",
		nil,
	)
}

/*
 * 自動リマインド更新用Model作成
 */
func (builder *notificationReminderBuilder) BuildUpdateNotificationReminderModel(
	currentReminder models.NotificationReminder,
	req types.UpdateNotificationReminderRequest,
) (models.NotificationReminder, results.Result) {
	if currentReminder.ID == 0 {
		return models.NotificationReminder{}, results.BadRequest(
			"BUILD_UPDATE_NOTIFICATION_REMINDER_MODEL_EMPTY_CURRENT_REMINDER",
			"自動リマインド更新データの作成に失敗しました",
			nil,
		)
	}

	if req.Title == "" {
		return models.NotificationReminder{}, results.BadRequest(
			"BUILD_UPDATE_NOTIFICATION_REMINDER_MODEL_EMPTY_TITLE",
			"自動リマインド更新データの作成に失敗しました",
			nil,
		)
	}

	if req.Message == "" {
		return models.NotificationReminder{}, results.BadRequest(
			"BUILD_UPDATE_NOTIFICATION_REMINDER_MODEL_EMPTY_MESSAGE",
			"自動リマインド更新データの作成に失敗しました",
			nil,
		)
	}

	if req.DayOffsetFromMonthEnd < 0 {
		return models.NotificationReminder{}, results.BadRequest(
			"BUILD_UPDATE_NOTIFICATION_REMINDER_MODEL_INVALID_DAY_OFFSET_FROM_MONTH_END",
			"自動リマインド更新データの作成に失敗しました",
			map[string]any{
				"dayOffsetFromMonthEnd": req.DayOffsetFromMonthEnd,
			},
		)
	}

	if req.SendHour < 0 || req.SendHour > 23 {
		return models.NotificationReminder{}, results.BadRequest(
			"BUILD_UPDATE_NOTIFICATION_REMINDER_MODEL_INVALID_SEND_HOUR",
			"自動リマインド更新データの作成に失敗しました",
			map[string]any{
				"sendHour": req.SendHour,
			},
		)
	}

	if req.SendMinute < 0 || req.SendMinute > 59 {
		return models.NotificationReminder{}, results.BadRequest(
			"BUILD_UPDATE_NOTIFICATION_REMINDER_MODEL_INVALID_SEND_MINUTE",
			"自動リマインド更新データの作成に失敗しました",
			map[string]any{
				"sendMinute": req.SendMinute,
			},
		)
	}

	currentReminder.Title = req.Title
	currentReminder.Message = req.Message
	currentReminder.DayOffsetFromMonthEnd = req.DayOffsetFromMonthEnd
	currentReminder.SendHour = req.SendHour
	currentReminder.SendMinute = req.SendMinute
	currentReminder.IsEnabled = req.IsEnabled

	return currentReminder, results.OK(
		nil,
		"BUILD_UPDATE_NOTIFICATION_REMINDER_MODEL_SUCCESS",
		"",
		nil,
	)
}

/*
 * 自動リマインド削除用Model作成
 *
 * 論理削除する。
 */
func (builder *notificationReminderBuilder) BuildDeleteNotificationReminderModel(
	currentReminder models.NotificationReminder,
) (models.NotificationReminder, results.Result) {
	if currentReminder.ID == 0 {
		return models.NotificationReminder{}, results.BadRequest(
			"BUILD_DELETE_NOTIFICATION_REMINDER_MODEL_EMPTY_CURRENT_REMINDER",
			"自動リマインド削除データの作成に失敗しました",
			nil,
		)
	}

	if currentReminder.IsDeleted {
		return currentReminder, results.OK(
			nil,
			"BUILD_DELETE_NOTIFICATION_REMINDER_MODEL_ALREADY_DELETED",
			"",
			nil,
		)
	}

	now := time.Now()

	currentReminder.IsDeleted = true
	currentReminder.IsEnabled = false
	currentReminder.DeletedAt = &now

	return currentReminder, results.OK(
		nil,
		"BUILD_DELETE_NOTIFICATION_REMINDER_MODEL_SUCCESS",
		"",
		nil,
	)
}

/*
 * 自動リマインド有効/無効切替用Model作成
 */
func (builder *notificationReminderBuilder) BuildToggleNotificationReminderEnabledModel(
	currentReminder models.NotificationReminder,
	req types.ToggleNotificationReminderEnabledRequest,
) (models.NotificationReminder, results.Result) {
	if currentReminder.ID == 0 {
		return models.NotificationReminder{}, results.BadRequest(
			"BUILD_TOGGLE_NOTIFICATION_REMINDER_ENABLED_MODEL_EMPTY_CURRENT_REMINDER",
			"自動リマインド有効状態更新データの作成に失敗しました",
			nil,
		)
	}

	currentReminder.IsEnabled = req.IsEnabled

	return currentReminder, results.OK(
		nil,
		"BUILD_TOGGLE_NOTIFICATION_REMINDER_ENABLED_MODEL_SUCCESS",
		"",
		nil,
	)
}
