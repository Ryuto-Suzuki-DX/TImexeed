package results

import "net/http"

/*
 * ServiceからControllerへ返す共通結果
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
 * 成功
 */
func Success(data any, message string) Result {
	return Result{
		StatusCode: http.StatusOK,
		Data:       data,
		Error:      false,
		Code:       "SUCCESS",
		Message:    message,
	}
}

/*
 * 作成成功
 */
func Created(data any, message string) Result {
	return Result{
		StatusCode: http.StatusCreated,
		Data:       data,
		Error:      false,
		Code:       "CREATED",
		Message:    message,
	}
}

/*
 * リクエスト不正
 */
func BadRequest(code string, message string) Result {
	return Result{
		StatusCode: http.StatusBadRequest,
		Data:       nil,
		Error:      true,
		Code:       code,
		Message:    message,
	}
}

/*
 * バリデーションエラー
 */
func ValidationError(details any) Result {
	return Result{
		StatusCode: http.StatusBadRequest,
		Data:       nil,
		Error:      true,
		Code:       "VALIDATION_ERROR",
		Message:    "入力内容を確認してください",
		Details:    details,
	}
}

/*
 * データなし
 */
func NotFound(message string) Result {
	return Result{
		StatusCode: http.StatusNotFound,
		Data:       nil,
		Error:      true,
		Code:       "NOT_FOUND",
		Message:    message,
	}
}

/*
 * 重複
 */
func Conflict(message string) Result {
	return Result{
		StatusCode: http.StatusConflict,
		Data:       nil,
		Error:      true,
		Code:       "CONFLICT",
		Message:    message,
	}
}

/*
 * サーバーエラー
 */
func InternalServerError(message string) Result {
	return Result{
		StatusCode: http.StatusInternalServerError,
		Data:       nil,
		Error:      true,
		Code:       "INTERNAL_ERROR",
		Message:    message,
	}
}

/*
 * 未認証
 */
func Unauthorized(message string) Result {
	return Result{
		StatusCode: http.StatusUnauthorized,
		Data:       nil,
		Error:      true,
		Code:       "UNAUTHORIZED",
		Message:    message,
	}
}
