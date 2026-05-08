package utils

import (
	"strings"

	"timexeed/backend/internal/results"
)

const (
	DefaultPageLimit = 50
	MaxPageLimit     = 50
)

/*
 * ページング検索条件
 *
 * 各検索Requestの offset / limit / keyword を共通処理するための型。
 */
type PageSearchCondition struct {
	Keyword string
	Offset  int
	Limit   int
}

/*
 * ページング検索条件を正規化する
 *
 * 役割：
 * ・keywordの前後空白を削除する
 * ・offsetがマイナスならエラーにする
 * ・limitが0以下なら50件にする
 * ・limitが50件を超えるなら50件に丸める
 *
 * code / message は呼び出し元で決める。
 */
func NormalizePageSearchCondition(
	condition PageSearchCondition,
	code string,
	message string,
) (PageSearchCondition, results.Result) {
	if condition.Offset < 0 {
		return condition, results.BadRequest(
			code,
			message,
			map[string]any{
				"offset": condition.Offset,
			},
		)
	}

	condition.Keyword = strings.TrimSpace(condition.Keyword)

	if condition.Limit <= 0 {
		condition.Limit = DefaultPageLimit
	}

	if condition.Limit > MaxPageLimit {
		condition.Limit = MaxPageLimit
	}

	return condition, results.OK(
		nil,
		"NORMALIZE_PAGE_SEARCH_CONDITION_SUCCESS",
		"",
		nil,
	)
}

/*
 * さらに表示するデータがあるか判定する
 *
 * total：
 * 	検索条件に一致する総件数
 *
 * offset：
 * 	今回の取得開始位置
 *
 * fetchedCount：
 * 	今回実際に取得できた件数
 */
func HasMore(total int64, offset int, fetchedCount int) bool {
	nextOffset := offset + fetchedCount
	return int64(nextOffset) < total
}
