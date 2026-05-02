/*
 * 管理者用ユーザー型
 */

// 単体
export type User = {
  id: number;
  name: string;
  email: string;
  role: string;
  departmentId: number | null;
  departmentName: string;
  isDeleted: boolean;
};

// 検索リクエスト
export type SearchUsersRequest = {
  keyword: string;
  includeDeleted: boolean;
};

// 検索レスポンス
export type SearchUsersResponse = {
  users: User[];
};

// 新規作成リクエスト
export type CreateUserRequest = {
  name: string;
  email: string;
  password: string;
  role: string;
  departmentId: number | null;
};

export type CreateUserResponse = User;

// APIレスポンス
export type ApiResponse<T> = {
  data: T | null;
  error: boolean;
  code: string;
  message: string;
  details?: unknown;
};

// 取得
export type GetUserResponse = User;

// 更新
export type UpdateUserRequest = {
  id: number;
  name: string;
  email: string;
  password: string;
  role: string;
  departmentId: number | null;
};

export type UpdateUserResponse = User;

// 削除
export type DeleteUserRequest = {
  id: number;
};

export type DeleteUserResponse = null;