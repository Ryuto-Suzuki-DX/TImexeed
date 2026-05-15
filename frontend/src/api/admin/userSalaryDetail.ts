import { apiPost } from "@/api/client";
import type {
  CreateUserSalaryDetailRequest,
  CreateUserSalaryDetailResponse,
  DeleteUserSalaryDetailRequest,
  DeleteUserSalaryDetailResponse,
  GetUserSalaryDetailRequest,
  GetUserSalaryDetailResponse,
  SearchUserSalaryDetailsRequest,
  SearchUserSalaryDetailsResponse,
  UpdateUserSalaryDetailRequest,
  UpdateUserSalaryDetailResponse,
} from "@/types/admin/userSalaryDetail";

export function searchUserSalaryDetails(request: SearchUserSalaryDetailsRequest) {
  return apiPost<SearchUserSalaryDetailsResponse, SearchUserSalaryDetailsRequest>(
    "/admin/user-salary-details/search",
    request
  );
}

export function getUserSalaryDetail(request: GetUserSalaryDetailRequest) {
  return apiPost<GetUserSalaryDetailResponse, GetUserSalaryDetailRequest>(
    "/admin/user-salary-details/get",
    request
  );
}

export function createUserSalaryDetail(request: CreateUserSalaryDetailRequest) {
  return apiPost<CreateUserSalaryDetailResponse, CreateUserSalaryDetailRequest>(
    "/admin/user-salary-details/create",
    request
  );
}

export function updateUserSalaryDetail(request: UpdateUserSalaryDetailRequest) {
  return apiPost<UpdateUserSalaryDetailResponse, UpdateUserSalaryDetailRequest>(
    "/admin/user-salary-details/update",
    request
  );
}

export function deleteUserSalaryDetail(request: DeleteUserSalaryDetailRequest) {
  return apiPost<DeleteUserSalaryDetailResponse, DeleteUserSalaryDetailRequest>(
    "/admin/user-salary-details/delete",
    request
  );
}
