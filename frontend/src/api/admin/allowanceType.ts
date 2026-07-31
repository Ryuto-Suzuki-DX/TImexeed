import { apiPost } from "@/api/client";
import type {
  AllowanceTypeDetailRequest,
  AllowanceTypeDetailResponse,
  CreateAllowanceTypeRequest,
  CreateAllowanceTypeResponse,
  DeleteAllowanceTypeRequest,
  DeleteAllowanceTypeResponse,
  SearchAllowanceTypesRequest,
  SearchAllowanceTypesResponse,
  UpdateAllowanceTypeRequest,
  UpdateAllowanceTypeResponse,
} from "@/types/admin/allowanceType";

export function searchAllowanceTypes(request: SearchAllowanceTypesRequest) {
  return apiPost<SearchAllowanceTypesResponse, SearchAllowanceTypesRequest>(
    "/admin/allowance-types/search",
    request,
  );
}

export function getAllowanceTypeDetail(request: AllowanceTypeDetailRequest) {
  return apiPost<AllowanceTypeDetailResponse, AllowanceTypeDetailRequest>(
    "/admin/allowance-types/detail",
    request,
  );
}

export function createAllowanceType(request: CreateAllowanceTypeRequest) {
  return apiPost<CreateAllowanceTypeResponse, CreateAllowanceTypeRequest>(
    "/admin/allowance-types/create",
    request,
  );
}

export function updateAllowanceType(request: UpdateAllowanceTypeRequest) {
  return apiPost<UpdateAllowanceTypeResponse, UpdateAllowanceTypeRequest>(
    "/admin/allowance-types/update",
    request,
  );
}

export function deleteAllowanceType(request: DeleteAllowanceTypeRequest) {
  return apiPost<DeleteAllowanceTypeResponse, DeleteAllowanceTypeRequest>(
    "/admin/allowance-types/delete",
    request,
  );
}
