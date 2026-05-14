package services

import (
	"timexeed/backend/internal/models"
	"timexeed/backend/internal/modules/admin/builders"
	"timexeed/backend/internal/modules/admin/repositories"
	"timexeed/backend/internal/modules/admin/types"
	"timexeed/backend/internal/results"
)

/*
 * 管理者用お知らせ自動リマインドService interface
 *
 * ControllerがServiceに求める処理だけを定義する。
 */
type NotificationReminderService interface {
	SearchNotificationReminders(req types.SearchNotificationRemindersRequest) results.Result
	CreateNotificationReminder(req types.CreateNotificationReminderRequest) results.Result
	UpdateNotificationReminder(req types.UpdateNotificationReminderRequest) results.Result
	DeleteNotificationReminder(req types.DeleteNotificationReminderRequest) results.Result
	ToggleNotificationReminderEnabled(req types.ToggleNotificationReminderEnabledRequest) results.Result
}

/*
 * 管理者用お知らせ自動リマインドService
 *
 * 役割：
 * ・Controllerから受け取ったRequestをもとに処理を進める
 * ・Serviceで発生したエラーはServiceでcode/message/detailsを作る
 * ・Builderで検索クエリや保存用Modelを作成する
 * ・Builderで発生したエラーはBuilderから返されたResultをそのまま返す
 * ・RepositoryでDB処理を実行する
 * ・Repositoryで発生したエラーはRepositoryから返されたResultをそのまま返す
 * ・成功時はResponse型に変換してControllerへ返す
 *
 * 注意：
 * ・Controllerにはgin.Contextを渡さない
 * ・Serviceではc.JSONしない
 * ・DBへの直接アクセスはRepositoryに任せる
 * ・Builder/Repositoryのエラー文言をServiceで作り直さない
 */
type notificationReminderService struct {
	notificationReminderBuilder    builders.NotificationReminderBuilder
	notificationReminderRepository repositories.NotificationReminderRepository
}

/*
 * NotificationReminderService生成
 */
func NewNotificationReminderService(
	notificationReminderBuilder builders.NotificationReminderBuilder,
	notificationReminderRepository repositories.NotificationReminderRepository,
) *notificationReminderService {
	return &notificationReminderService{
		notificationReminderBuilder:    notificationReminderBuilder,
		notificationReminderRepository: notificationReminderRepository,
	}
}

/*
 * models.NotificationReminderをフロント返却用NotificationReminderResponseへ変換する
 */
func toNotificationReminderResponse(
	notificationReminder models.NotificationReminder,
) types.NotificationReminderResponse {
	return types.NotificationReminderResponse{
		ID:                    notificationReminder.ID,
		Title:                 notificationReminder.Title,
		Message:               notificationReminder.Message,
		DayOffsetFromMonthEnd: notificationReminder.DayOffsetFromMonthEnd,
		SendHour:              notificationReminder.SendHour,
		SendMinute:            notificationReminder.SendMinute,
		IsEnabled:             notificationReminder.IsEnabled,
		IsDeleted:             notificationReminder.IsDeleted,
		CreatedAt:             notificationReminder.CreatedAt,
		UpdatedAt:             notificationReminder.UpdatedAt,
		DeletedAt:             notificationReminder.DeletedAt,
	}
}

/*
 * 自動リマインド検索
 *
 * 管理者画面で自動リマインド設定一覧を取得する。
 */
func (service *notificationReminderService) SearchNotificationReminders(
	req types.SearchNotificationRemindersRequest,
) results.Result {
	if req.Limit <= 0 {
		req.Limit = 10
	}

	if req.Offset < 0 {
		return results.BadRequest(
			"SEARCH_NOTIFICATION_REMINDERS_INVALID_OFFSET",
			"自動リマインド検索の開始位置が正しくありません",
			map[string]any{
				"offset": req.Offset,
			},
		)
	}

	// hasMore判定用に1件多く取得する
	searchLimit := req.Limit + 1
	req.Limit = searchLimit

	// Builderで自動リマインド検索用クエリを作成する
	query, buildResult := service.notificationReminderBuilder.BuildSearchNotificationRemindersQuery(req)
	if buildResult.Error {
		return buildResult
	}

	// Repositoryで自動リマインド一覧を取得する
	notificationReminders, findResult := service.notificationReminderRepository.FindNotificationReminders(query)
	if findResult.Error {
		return findResult
	}

	hasMore := false
	displayLimit := searchLimit - 1

	if len(notificationReminders) > displayLimit {
		hasMore = true
		notificationReminders = notificationReminders[:displayLimit]
	}

	// DBモデルをフロント返却用Responseへ変換する
	reminderResponses := make([]types.NotificationReminderResponse, 0, len(notificationReminders))
	for _, notificationReminder := range notificationReminders {
		reminderResponses = append(reminderResponses, toNotificationReminderResponse(notificationReminder))
	}

	return results.OK(
		types.SearchNotificationRemindersResponse{
			Reminders: reminderResponses,
			HasMore:   hasMore,
		},
		"SEARCH_NOTIFICATION_REMINDERS_SUCCESS",
		"自動リマインド一覧を取得しました",
		nil,
	)
}

/*
 * 自動リマインド作成
 */
func (service *notificationReminderService) CreateNotificationReminder(
	req types.CreateNotificationReminderRequest,
) results.Result {
	if req.Title == "" {
		return results.BadRequest(
			"CREATE_NOTIFICATION_REMINDER_EMPTY_TITLE",
			"自動リマインドタイトルを入力してください",
			nil,
		)
	}

	if req.Message == "" {
		return results.BadRequest(
			"CREATE_NOTIFICATION_REMINDER_EMPTY_MESSAGE",
			"自動リマインド本文を入力してください",
			nil,
		)
	}

	// Builderで自動リマインド作成用Modelを作る
	notificationReminder, buildResult := service.notificationReminderBuilder.BuildCreateNotificationReminderModel(req)
	if buildResult.Error {
		return buildResult
	}

	// Repositoryで自動リマインドを作成する
	createdReminder, createResult := service.notificationReminderRepository.CreateNotificationReminder(notificationReminder)
	if createResult.Error {
		return createResult
	}

	return results.Created(
		types.CreateNotificationReminderResponse{
			Reminder: toNotificationReminderResponse(createdReminder),
		},
		"CREATE_NOTIFICATION_REMINDER_SUCCESS",
		"自動リマインドを作成しました",
		nil,
	)
}

/*
 * 自動リマインド更新
 */
func (service *notificationReminderService) UpdateNotificationReminder(
	req types.UpdateNotificationReminderRequest,
) results.Result {
	if req.ReminderID == 0 {
		return results.BadRequest(
			"UPDATE_NOTIFICATION_REMINDER_INVALID_REMINDER_ID",
			"自動リマインドIDが正しくありません",
			map[string]any{
				"reminderId": req.ReminderID,
			},
		)
	}

	if req.Title == "" {
		return results.BadRequest(
			"UPDATE_NOTIFICATION_REMINDER_EMPTY_TITLE",
			"自動リマインドタイトルを入力してください",
			nil,
		)
	}

	if req.Message == "" {
		return results.BadRequest(
			"UPDATE_NOTIFICATION_REMINDER_EMPTY_MESSAGE",
			"自動リマインド本文を入力してください",
			nil,
		)
	}

	// Builderで対象自動リマインド取得用クエリを作成する
	findQuery, buildFindResult := service.notificationReminderBuilder.BuildFindNotificationReminderByIDQuery(req.ReminderID, false)
	if buildFindResult.Error {
		return buildFindResult
	}

	// Repositoryで対象自動リマインドを取得する
	currentReminder, findResult := service.notificationReminderRepository.FindNotificationReminder(findQuery)
	if findResult.Error {
		return findResult
	}

	// Builderで更新用Modelを作る
	updateReminder, buildUpdateResult := service.notificationReminderBuilder.BuildUpdateNotificationReminderModel(currentReminder, req)
	if buildUpdateResult.Error {
		return buildUpdateResult
	}

	// Repositoryで自動リマインドを保存する
	savedReminder, saveResult := service.notificationReminderRepository.SaveNotificationReminder(updateReminder)
	if saveResult.Error {
		return saveResult
	}

	return results.OK(
		types.UpdateNotificationReminderResponse{
			Reminder: toNotificationReminderResponse(savedReminder),
		},
		"UPDATE_NOTIFICATION_REMINDER_SUCCESS",
		"自動リマインドを更新しました",
		nil,
	)
}

/*
 * 自動リマインド削除
 *
 * 論理削除する。
 */
func (service *notificationReminderService) DeleteNotificationReminder(
	req types.DeleteNotificationReminderRequest,
) results.Result {
	if req.ReminderID == 0 {
		return results.BadRequest(
			"DELETE_NOTIFICATION_REMINDER_INVALID_REMINDER_ID",
			"自動リマインドIDが正しくありません",
			map[string]any{
				"reminderId": req.ReminderID,
			},
		)
	}

	// Builderで対象自動リマインド取得用クエリを作成する
	findQuery, buildFindResult := service.notificationReminderBuilder.BuildFindNotificationReminderByIDQuery(req.ReminderID, false)
	if buildFindResult.Error {
		return buildFindResult
	}

	// Repositoryで対象自動リマインドを取得する
	currentReminder, findResult := service.notificationReminderRepository.FindNotificationReminder(findQuery)
	if findResult.Error {
		return findResult
	}

	// Builderで削除用Modelを作る
	deleteReminder, buildDeleteResult := service.notificationReminderBuilder.BuildDeleteNotificationReminderModel(currentReminder)
	if buildDeleteResult.Error {
		return buildDeleteResult
	}

	// Repositoryで自動リマインドを保存する
	savedReminder, saveResult := service.notificationReminderRepository.SaveNotificationReminder(deleteReminder)
	if saveResult.Error {
		return saveResult
	}

	return results.OK(
		types.DeleteNotificationReminderResponse{
			Reminder: toNotificationReminderResponse(savedReminder),
		},
		"DELETE_NOTIFICATION_REMINDER_SUCCESS",
		"自動リマインドを削除しました",
		nil,
	)
}

/*
 * 自動リマインド有効/無効切替
 */
func (service *notificationReminderService) ToggleNotificationReminderEnabled(
	req types.ToggleNotificationReminderEnabledRequest,
) results.Result {
	if req.ReminderID == 0 {
		return results.BadRequest(
			"TOGGLE_NOTIFICATION_REMINDER_ENABLED_INVALID_REMINDER_ID",
			"自動リマインドIDが正しくありません",
			map[string]any{
				"reminderId": req.ReminderID,
			},
		)
	}

	// Builderで対象自動リマインド取得用クエリを作成する
	findQuery, buildFindResult := service.notificationReminderBuilder.BuildFindNotificationReminderByIDQuery(req.ReminderID, false)
	if buildFindResult.Error {
		return buildFindResult
	}

	// Repositoryで対象自動リマインドを取得する
	currentReminder, findResult := service.notificationReminderRepository.FindNotificationReminder(findQuery)
	if findResult.Error {
		return findResult
	}

	// Builderで有効/無効切替用Modelを作る
	toggleReminder, buildToggleResult := service.notificationReminderBuilder.BuildToggleNotificationReminderEnabledModel(currentReminder, req)
	if buildToggleResult.Error {
		return buildToggleResult
	}

	// Repositoryで自動リマインドを保存する
	savedReminder, saveResult := service.notificationReminderRepository.SaveNotificationReminder(toggleReminder)
	if saveResult.Error {
		return saveResult
	}

	return results.OK(
		types.ToggleNotificationReminderEnabledResponse{
			Reminder: toNotificationReminderResponse(savedReminder),
		},
		"TOGGLE_NOTIFICATION_REMINDER_ENABLED_SUCCESS",
		"自動リマインドの有効状態を更新しました",
		nil,
	)
}
