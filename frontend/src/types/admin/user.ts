/*
 * ユーザー
 */

export type SearchUsersRequest = {
  keyword: string;
  includeDeleted: boolean;
  offset: number;
  limit: number;
};

/*
 * 業務対象ユーザー検索Request
 *
 * 勤怠、給与、経費、有給、個人情報Driveなどで使う。
 * ADMINは対象外。
 */
export type SearchBusinessTargetUsersRequest = {
  keyword: string;
  offset: number;
  limit: number;
};

export type UserDetailRequest = {
  targetUserId: number;
};

export type CreateUserRequest = {
  name: string;
  email: string;
  password: string;
  role: string;
  departmentId: number | null;
  hireDate: string;
};

export type UpdateUserRequest = {
  targetUserId: number;
  name: string;
  email: string;
  role: string;
  departmentId: number | null;
  hireDate: string;
  retirementDate: string | null;
};

export type DeleteUserRequest = {
  targetUserId: number;
};

export type UserResponse = {
  id: number;
  name: string;
  email: string;
  role: string;
  departmentId: number | null;
  hireDate: string;
  retirementDate: string | null;
  isDeleted: boolean;
  createdAt: string;
  updatedAt: string;
  deletedAt: string | null;
};

export type SearchUsersResponse = {
  users: UserResponse[];
  total: number;
  offset: number;
  limit: number;
  hasMore: boolean;
};

/*
 * 業務対象ユーザー検索Response
 *
 * 返却形式は通常のユーザー検索と同じ。
 * ただし中身はUSERのみ。
 */
export type SearchBusinessTargetUsersResponse = {
  users: UserResponse[];
  total: number;
  offset: number;
  limit: number;
  hasMore: boolean;
};

export type UserDetailResponse = {
  user: UserResponse;
};

export type CreateUserResponse = {
  user: UserResponse;
};

export type UpdateUserResponse = {
  user: UserResponse;
};

export type DeleteUserResponse = {
  targetUserId: number;
};