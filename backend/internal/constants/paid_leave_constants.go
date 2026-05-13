package constants

/*
 * 〇 有給休暇 固定値
 *
 * このファイルでは、有給休暇に関する法定の固定値を管理する。
 *
 * 方針：
 * ・DBには保存しない
 * ・法律上の付与日数を固定値として持つ
 * ・将来の定年延長を考慮し、60年分の付与ルールを用意する
 *
 * 注意：
 * ・ここでは通常の労働者を対象にする
 * ・パートタイムなどの比例付与は、今後必要になったら別定義を追加する
 * ・出勤率8割判定はここでは行わない
 *   → Service側で勤怠DBを見て判定する
 */

/*
 * 有給休暇の計算対象年数
 *
 * 60年分用意する。
 * 例：
 * 	18歳入社 → 78歳まで対応
 * 	22歳入社 → 82歳まで対応
 */
const PaidLeaveMaxServiceYears = 60

/*
 * 有給休暇 年5日取得義務の日数
 *
 * 年10日以上の有給が付与された労働者について、
 * 付与日から1年以内に5日取得させる必要がある。
 */
const PaidLeaveRequiredUseDays = 5.0

/*
 * 有給休暇 出勤率判定基準
 *
 * 有給付与には、原則として全労働日の8割以上出勤が必要。
 */
const PaidLeaveAttendanceRateThreshold = 0.8

/*
 * 有給休暇 付与ルール
 *
 * AfterMonths:
 * 	雇入れ日から何か月後に付与されるか
 *
 * GrantDays:
 * 	付与される有給日数
 */
type PaidLeaveGrantRule struct {
	AfterMonths int
	GrantDays   float64
}

/*
 * 通常労働者の有給付与ルール
 *
 * 法定の付与日数：
 * 	6か月      → 10日
 * 	1年6か月   → 11日
 * 	2年6か月   → 12日
 * 	3年6か月   → 14日
 * 	4年6か月   → 16日
 * 	5年6か月   → 18日
 * 	6年6か月以降 → 20日
 *
 * 60年分保持する。
 */
var PaidLeaveGrantRules = []PaidLeaveGrantRule{
	{AfterMonths: 6, GrantDays: 10},
	{AfterMonths: 18, GrantDays: 11},
	{AfterMonths: 30, GrantDays: 12},
	{AfterMonths: 42, GrantDays: 14},
	{AfterMonths: 54, GrantDays: 16},
	{AfterMonths: 66, GrantDays: 18},
	{AfterMonths: 78, GrantDays: 20},
	{AfterMonths: 90, GrantDays: 20},
	{AfterMonths: 102, GrantDays: 20},
	{AfterMonths: 114, GrantDays: 20},
	{AfterMonths: 126, GrantDays: 20},
	{AfterMonths: 138, GrantDays: 20},
	{AfterMonths: 150, GrantDays: 20},
	{AfterMonths: 162, GrantDays: 20},
	{AfterMonths: 174, GrantDays: 20},
	{AfterMonths: 186, GrantDays: 20},
	{AfterMonths: 198, GrantDays: 20},
	{AfterMonths: 210, GrantDays: 20},
	{AfterMonths: 222, GrantDays: 20},
	{AfterMonths: 234, GrantDays: 20},
	{AfterMonths: 246, GrantDays: 20},
	{AfterMonths: 258, GrantDays: 20},
	{AfterMonths: 270, GrantDays: 20},
	{AfterMonths: 282, GrantDays: 20},
	{AfterMonths: 294, GrantDays: 20},
	{AfterMonths: 306, GrantDays: 20},
	{AfterMonths: 318, GrantDays: 20},
	{AfterMonths: 330, GrantDays: 20},
	{AfterMonths: 342, GrantDays: 20},
	{AfterMonths: 354, GrantDays: 20},
	{AfterMonths: 366, GrantDays: 20},
	{AfterMonths: 378, GrantDays: 20},
	{AfterMonths: 390, GrantDays: 20},
	{AfterMonths: 402, GrantDays: 20},
	{AfterMonths: 414, GrantDays: 20},
	{AfterMonths: 426, GrantDays: 20},
	{AfterMonths: 438, GrantDays: 20},
	{AfterMonths: 450, GrantDays: 20},
	{AfterMonths: 462, GrantDays: 20},
	{AfterMonths: 474, GrantDays: 20},
	{AfterMonths: 486, GrantDays: 20},
	{AfterMonths: 498, GrantDays: 20},
	{AfterMonths: 510, GrantDays: 20},
	{AfterMonths: 522, GrantDays: 20},
	{AfterMonths: 534, GrantDays: 20},
	{AfterMonths: 546, GrantDays: 20},
	{AfterMonths: 558, GrantDays: 20},
	{AfterMonths: 570, GrantDays: 20},
	{AfterMonths: 582, GrantDays: 20},
	{AfterMonths: 594, GrantDays: 20},
	{AfterMonths: 606, GrantDays: 20},
	{AfterMonths: 618, GrantDays: 20},
	{AfterMonths: 630, GrantDays: 20},
	{AfterMonths: 642, GrantDays: 20},
	{AfterMonths: 654, GrantDays: 20},
	{AfterMonths: 666, GrantDays: 20},
	{AfterMonths: 678, GrantDays: 20},
	{AfterMonths: 690, GrantDays: 20},
	{AfterMonths: 702, GrantDays: 20},
	{AfterMonths: 714, GrantDays: 20},
}
