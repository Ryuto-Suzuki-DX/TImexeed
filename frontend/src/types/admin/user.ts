/*
 * ユーザー
 */

export type SearchUsersRequest = {
  keyword: string;
  includeDeleted: boolean;
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