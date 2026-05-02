import { getAccessToken } from "@/lib/auth";
import type {
  ApiResponse,
  CreateDepartmentRequest,
  CreateDepartmentResponse,
  SearchDepartmentsResponse,
  GetDepartmentResponse,
  UpdateDepartmentRequest,
  UpdateDepartmentResponse,
  DeleteDepartmentRequest,
  DeleteDepartmentResponse,
} from "@/types/admin/department";

const API_BASE_URL = "http://127.0.0.1:8080";

/*
 * 管理者用所属一覧取得
 */
export async function searchDepartments(
  keyword: string,
  includeDeleted: boolean
): Promise<ApiResponse<SearchDepartmentsResponse>> {
  const token = getAccessToken();

  const query = new URLSearchParams();

  if (keyword) {
    query.set("keyword", keyword);
  }

  query.set("includeDeleted", String(includeDeleted));

  const response = await fetch(`${API_BASE_URL}/admin/departments?${query.toString()}`, {
    method: "GET",
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });

  return response.json();
}

/*
 * 管理者用所属詳細取得
 */
export async function getDepartment(id: number): Promise<ApiResponse<GetDepartmentResponse>> {
  const token = getAccessToken();

  const response = await fetch(`${API_BASE_URL}/admin/departments/${id}`, {
    method: "GET",
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });

  return response.json();
}

/*
 * 管理者用所属新規作成
 */
export async function createDepartment(
  req: CreateDepartmentRequest
): Promise<ApiResponse<CreateDepartmentResponse>> {
  const token = getAccessToken();

  const response = await fetch(`${API_BASE_URL}/admin/departments`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${token}`,
    },
    body: JSON.stringify(req),
  });

  return response.json();
}

/*
 * 管理者用所属更新
 */
export async function updateDepartment(
  req: UpdateDepartmentRequest
): Promise<ApiResponse<UpdateDepartmentResponse>> {
  const token = getAccessToken();

  const response = await fetch(`${API_BASE_URL}/admin/departments/${req.id}`, {
    method: "PUT",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${token}`,
    },
    body: JSON.stringify({
      name: req.name,
    }),
  });

  return response.json();
}

/*
 * 管理者用所属削除
 */
export async function deleteDepartment(
  req: DeleteDepartmentRequest
): Promise<ApiResponse<DeleteDepartmentResponse>> {
  const token = getAccessToken();

  const response = await fetch(`${API_BASE_URL}/admin/departments/${req.id}`, {
    method: "DELETE",
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });

  return response.json();
}