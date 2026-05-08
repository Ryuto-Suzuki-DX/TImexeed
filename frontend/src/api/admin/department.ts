import { apiPost } from "@/api/client";
import type {
  CreateDepartmentRequest,
  CreateDepartmentResponse,
  DeleteDepartmentRequest,
  DeleteDepartmentResponse,
  DepartmentDetailRequest,
  DepartmentDetailResponse,
  SearchDepartmentsRequest,
  SearchDepartmentsResponse,
  UpdateDepartmentRequest,
  UpdateDepartmentResponse,
} from "@/types/admin/department";

export function searchDepartments(request: SearchDepartmentsRequest) {
  return apiPost<SearchDepartmentsResponse, SearchDepartmentsRequest>("/admin/departments/search", request);
}

export function getDepartmentDetail(request: DepartmentDetailRequest) {
  return apiPost<DepartmentDetailResponse, DepartmentDetailRequest>("/admin/departments/detail", request);
}

export function createDepartment(request: CreateDepartmentRequest) {
  return apiPost<CreateDepartmentResponse, CreateDepartmentRequest>("/admin/departments/create", request);
}

export function updateDepartment(request: UpdateDepartmentRequest) {
  return apiPost<UpdateDepartmentResponse, UpdateDepartmentRequest>("/admin/departments/update", request);
}

export function deleteDepartment(request: DeleteDepartmentRequest) {
  return apiPost<DeleteDepartmentResponse, DeleteDepartmentRequest>("/admin/departments/delete", request);
}