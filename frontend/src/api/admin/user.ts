import { apiPost } from "@/api/client";
import type {
  CreateUserRequest,
  CreateUserResponse,
  DeleteUserRequest,
  DeleteUserResponse,
  SearchUsersRequest,
  SearchUsersResponse,
  UpdateUserRequest,
  UpdateUserResponse,
  UserDetailRequest,
  UserDetailResponse,
} from "@/types/admin/user";

export function searchUsers(request: SearchUsersRequest) {
  return apiPost<SearchUsersResponse, SearchUsersRequest>("/admin/users/search", request);
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