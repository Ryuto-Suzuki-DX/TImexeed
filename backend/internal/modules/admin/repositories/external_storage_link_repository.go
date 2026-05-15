package repositories

import (
	"errors"

	"timexeed/backend/internal/models"
	"timexeed/backend/internal/results"

	"gorm.io/gorm"
)

/*
 * 管理者用外部ストレージリンクRepository interface
 */
type ExternalStorageLinkRepository interface {
	FindExternalStorageLinks(query *gorm.DB) ([]models.ExternalStorageLink, results.Result)
	FindExternalStorageLink(query *gorm.DB) (models.ExternalStorageLink, results.Result)
	CountExternalStorageLinks(query *gorm.DB) (int64, results.Result)
	CreateExternalStorageLink(externalStorageLink models.ExternalStorageLink) (models.ExternalStorageLink, results.Result)
	SaveExternalStorageLink(externalStorageLink models.ExternalStorageLink) (models.ExternalStorageLink, results.Result)
}

/*
 * 管理者用外部ストレージリンクRepository
 */
type externalStorageLinkRepository struct {
	db *gorm.DB
}

/*
 * ExternalStorageLinkRepository生成
 */
func NewExternalStorageLinkRepository(db *gorm.DB) ExternalStorageLinkRepository {
	return &externalStorageLinkRepository{
		db: db,
	}
}

/*
 * 一覧取得
 */
func (repository *externalStorageLinkRepository) FindExternalStorageLinks(query *gorm.DB) ([]models.ExternalStorageLink, results.Result) {
	var externalStorageLinks []models.ExternalStorageLink

	if err := query.Find(&externalStorageLinks).Error; err != nil {
		return nil, results.InternalServerError(
			"FIND_EXTERNAL_STORAGE_LINKS_FAILED",
			"外部ストレージリンク一覧の取得に失敗しました",
			err.Error(),
		)
	}

	return externalStorageLinks, results.OK(
		nil,
		"FIND_EXTERNAL_STORAGE_LINKS_SUCCESS",
		"外部ストレージリンク一覧を取得しました",
		nil,
	)
}

/*
 * 1件取得
 */
func (repository *externalStorageLinkRepository) FindExternalStorageLink(query *gorm.DB) (models.ExternalStorageLink, results.Result) {
	var externalStorageLink models.ExternalStorageLink

	if err := query.First(&externalStorageLink).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.ExternalStorageLink{}, results.NotFound(
				"EXTERNAL_STORAGE_LINK_NOT_FOUND",
				"外部ストレージリンクが見つかりません",
				err.Error(),
			)
		}

		return models.ExternalStorageLink{}, results.InternalServerError(
			"FIND_EXTERNAL_STORAGE_LINK_FAILED",
			"外部ストレージリンクの取得に失敗しました",
			err.Error(),
		)
	}

	return externalStorageLink, results.OK(
		nil,
		"FIND_EXTERNAL_STORAGE_LINK_SUCCESS",
		"外部ストレージリンクを取得しました",
		nil,
	)
}

/*
 * 件数取得
 */
func (repository *externalStorageLinkRepository) CountExternalStorageLinks(query *gorm.DB) (int64, results.Result) {
	var count int64

	if err := query.Count(&count).Error; err != nil {
		return 0, results.InternalServerError(
			"COUNT_EXTERNAL_STORAGE_LINKS_FAILED",
			"外部ストレージリンク件数の取得に失敗しました",
			err.Error(),
		)
	}

	return count, results.OK(
		nil,
		"COUNT_EXTERNAL_STORAGE_LINKS_SUCCESS",
		"外部ストレージリンク件数を取得しました",
		nil,
	)
}

/*
 * 新規作成
 */
func (repository *externalStorageLinkRepository) CreateExternalStorageLink(externalStorageLink models.ExternalStorageLink) (models.ExternalStorageLink, results.Result) {
	if err := repository.db.Create(&externalStorageLink).Error; err != nil {
		return models.ExternalStorageLink{}, results.InternalServerError(
			"CREATE_EXTERNAL_STORAGE_LINK_FAILED",
			"外部ストレージリンクの作成に失敗しました",
			err.Error(),
		)
	}

	return externalStorageLink, results.OK(
		nil,
		"CREATE_EXTERNAL_STORAGE_LINK_SUCCESS",
		"外部ストレージリンクを作成しました",
		nil,
	)
}

/*
 * 保存
 */
func (repository *externalStorageLinkRepository) SaveExternalStorageLink(externalStorageLink models.ExternalStorageLink) (models.ExternalStorageLink, results.Result) {
	if err := repository.db.Save(&externalStorageLink).Error; err != nil {
		return models.ExternalStorageLink{}, results.InternalServerError(
			"SAVE_EXTERNAL_STORAGE_LINK_FAILED",
			"外部ストレージリンクの保存に失敗しました",
			err.Error(),
		)
	}

	return externalStorageLink, results.OK(
		nil,
		"SAVE_EXTERNAL_STORAGE_LINK_SUCCESS",
		"外部ストレージリンクを保存しました",
		nil,
	)
}
