import { apiPost } from "@/api/client";
import type {
  CreateMonthlyAllowanceRequest,
  CreateMonthlyAllowanceResponse,
  DeleteMonthlyAllowanceRequest,
  DeleteMonthlyAllowanceResponse,
  MonthlyAllowanceDetailRequest,
  MonthlyAllowanceDetailResponse,
  SearchMonthlyAllowancesRequest,
  SearchMonthlyAllowancesResponse,
  UpdateMonthlyAllowanceRequest,
  UpdateMonthlyAllowanceResponse,
} from "@/types/admin/monthlyAllowance";

export function searchMonthlyAllowances(request: SearchMonthlyAllowancesRequest) {
  return apiPost<SearchMonthlyAllowancesResponse, SearchMonthlyAllowancesRequest>(
    "/admin/monthly-allowances/search",
    request,
  );
}

export function getMonthlyAllowanceDetail(request: MonthlyAllowanceDetailRequest) {
  return apiPost<MonthlyAllowanceDetailResponse, MonthlyAllowanceDetailRequest>(
    "/admin/monthly-allowances/detail",
    request,
  );
}

export function createMonthlyAllowance(request: CreateMonthlyAllowanceRequest) {
  return apiPost<CreateMonthlyAllowanceResponse, CreateMonthlyAllowanceRequest>(
    "/admin/monthly-allowances/create",
    request,
  );
}

export function updateMonthlyAllowance(request: UpdateMonthlyAllowanceRequest) {
  return apiPost<UpdateMonthlyAllowanceResponse, UpdateMonthlyAllowanceRequest>(
    "/admin/monthly-allowances/update",
    request,
  );
}

export function deleteMonthlyAllowance(request: DeleteMonthlyAllowanceRequest) {
  return apiPost<DeleteMonthlyAllowanceResponse, DeleteMonthlyAllowanceRequest>(
    "/admin/monthly-allowances/delete",
    request,
  );
}
