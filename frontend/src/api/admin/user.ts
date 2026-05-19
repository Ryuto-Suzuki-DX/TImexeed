import { apiPost } from "@/api/client";
import type {
  CreateUserRequest,
  CreateUserResponse,
  DeleteUserRequest,
  DeleteUserResponse,
  SearchBusinessTargetUsersRequest,
  SearchBusinessTargetUsersResponse,
  SearchUsersRequest,
  SearchUsersResponse,
  UpdateUserRequest,
  UpdateUserResponse,
  UserDetailRequest,
  UserDetailResponse,
} from "@/types/admin/user";

/*
 * ユーザー管理用検索
 *
 * ADMIN / USER の両方を検索対象にする。
 */
export function searchUsers(request: SearchUsersRequest) {
  return apiPost<SearchUsersResponse, SearchUsersRequest>("/admin/users/search", request);
}

/*
 * 業務対象ユーザー検索
 *
 * 勤怠、給与、経費、有給、個人情報Driveなどで使う。
 * ADMINは検索結果に含めない。
 */
export function searchBusinessTargetUsers(request: SearchBusinessTargetUsersRequest) {
  return apiPost<SearchBusinessTargetUsersResponse, SearchBusinessTargetUsersRequest>(
    "/admin/users/search-business-targets",
    request
  );
}

export function getUserDetail(request: UserDetailRequest) {
  return apiPost<UserDetailResponse, UserDetailRequest>("/admin/users/detail", request);
}

export function createUser(request: CreateUserRequest) {
  return apiPost<CreateUserResponse, CreateUserRequest>("/admin/users/create", request);
}

export function updateUser(request: UpdateUserRequest) {
  return apiPost<UpdateUserResponse, UpdateUserRequest>("/admin/users/update", request);
}

export function deleteUser(request: DeleteUserRequest) {
  return apiPost<DeleteUserResponse, DeleteUserRequest>("/admin/users/delete", request);
}