package utils

import "time"

const (
	DateLayout = "2006-01-02"
)

/*
 * yyyy-MM-dd形式の日付文字列をtime.Timeへ変換する
 *
 * Requestでは日付をstringで受け取るため、
 * Modelへ入れる前にtime.Timeへ変換する。
 */
func ParseDate(value string) (time.Time, error) {
	return time.Parse(DateLayout, value)
}

/*
 * null または 空文字を許容して日付文字列をtime.Timeポインタへ変換する
 *
 * 未設定の日付はnilとして扱う。
 */
func ParseOptionalDate(value *string) (*time.Time, error) {
	if value == nil {
		return nil, nil
	}

	if *value == "" {
		return nil, nil
	}

	parsedDate, err := time.Parse(DateLayout, *value)
	if err != nil {
		return nil, err
	}

	return &parsedDate, nil
}

/*
 * RFC3339形式の日時文字列をtime.Timeへ変換する
 *
 * 勤怠の開始日時・終了日時などで使う。
 *
 * 例：
 * 2026-05-04T09:00:00+09:00
 */
func ParseDateTime(value string) (time.Time, error) {
	return time.Parse(time.RFC3339, value)
}

/*
 * null または 空文字を許容してRFC3339形式の日時文字列をtime.Timeポインタへ変換する
 */
func ParseOptionalDateTime(value *string) (*time.Time, error) {
	if value == nil {
		return nil, nil
	}

	if *value == "" {
		return nil, nil
	}

	parsedDateTime, err := time.Parse(time.RFC3339, *value)
	if err != nil {
		return nil, err
	}

	return &parsedDateTime, nil
}
