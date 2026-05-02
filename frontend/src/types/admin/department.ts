/*
 * 管理者用所属型定義
 */

// 共通APIレスポンス型
export type ApiResponse<T> = {
  data: T;
  error: boolean;
  code: string;
  message: string;
  detail: unknown;
};

// 所属情報
export type Department = {
  id: number;
  name: string;
  isDeleted: boolean;
};

/*
 * 所属一覧取得
 */
export type SearchDepartmentsResponse = {
  departments: Department[];
};

/*
 * 所属詳細取得
 */
export type GetDepartmentResponse = Department;

/*
 * 所属新規作成
 */
export type CreateDepartmentRequest = {
  name: string;
};

export type CreateDepartmentResponse = Department;

/*
 * 所属更新
 */
export type UpdateDepartmentRequest = {
  id: number;
  name: string;
};

export type UpdateDepartmentResponse = Department;

/*
 * 所属削除
 */
export type DeleteDepartmentRequest = {
  id: number;
};

export type DeleteDepartmentResponse = null;