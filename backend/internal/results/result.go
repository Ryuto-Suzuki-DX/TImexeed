package results

import "net/http"

/*
 * 共通結果
 *
 * Controller / Service / Builder / Repository から返す共通の結果型。
 *
 * 役割：
 * ・HTTPステータスコードを持つ
 * ・成功/失敗を表す
 * ・フロントへ返す code / message / details を持つ
 *
 * 方針：
 * ・code / message / details は各層で決める
 * ・results.go はそれらを受け取って Result に詰めるだけ
 * ・HTTPステータスごとに関数を分ける
 */
type Result struct {
	StatusCode int
	Data       any
	Error      bool
	Code       string
	Message    string
	Details    any
}

/*
 * 200 OK
 *
 * 通常の成功時に使う。
 */
func OK(data any, code string, message string, details any) Result {
	return Result{
		StatusCode: http.StatusOK,
		Data:       data,
		Error:      false,
		Code:       code,
		Message:    message,
		Details:    details,
	}
}

/*
 * 201 Created
 *
 * 新規作成成功時に使う。
 */
func Created(data any, code string, message string, details any) Result {
	return Result{
		StatusCode: http.StatusCreated,
		Data:       data,
		Error:      false,
		Code:       code,
		Message:    message,
		Details:    details,
	}
}

/*
 * 400 Bad Request
 *
 * リクエスト形式エラー、入力内容不正、検索条件不正などに使う。
 */
func BadRequest(code string, message string, details any) Result {
	return Result{
		StatusCode: http.StatusBadRequest,
		Data:       nil,
		Error:      true,
		Code:       code,
		Message:    message,
		Details:    details,
	}
}

/*
 * 401 Unauthorized
 *
 * 未認証、トークン不正、認証情報なしなどに使う。
 */
func Unauthorized(code string, message string, details any) Result {
	return Result{
		StatusCode: http.StatusUnauthorized,
		Data:       nil,
		Error:      true,
		Code:       code,
		Message:    message,
		Details:    details,
	}
}

/*
 * 403 Forbidden
 *
 * 権限不足の場合に使う。
 */
func Forbidden(code string, message string, details any) Result {
	return Result{
		StatusCode: http.StatusForbidden,
		Data:       nil,
		Error:      true,
		Code:       code,
		Message:    message,
		Details:    details,
	}
}

/*
 * 404 Not Found
 *
 * 対象データが存在しない場合に使う。
 */
func NotFound(code string, message string, details any) Result {
	return Result{
		StatusCode: http.StatusNotFound,
		Data:       nil,
		Error:      true,
		Code:       code,
		Message:    message,
		Details:    details,
	}
}

/*
 * 409 Conflict
 *
 * メールアドレス重複など、既存データとの競合がある場合に使う。
 */
func Conflict(code string, message string, details any) Result {
	return Result{
		StatusCode: http.StatusConflict,
		Data:       nil,
		Error:      true,
		Code:       code,
		Message:    message,
		Details:    details,
	}
}

/*
 * 500 Internal Server Error
 *
 * DBエラー、想定外エラーなどに使う。
 */
func InternalServerError(code string, message string, details any) Result {
	return Result{
		StatusCode: http.StatusInternalServerError,
		Data:       nil,
		Error:      true,
		Code:       code,
		Message:    message,
		Details:    details,
	}
}
