import { getAccessToken } from "@/lib/auth";
import type {
  ApiResponse,
  CreateUserRequest,
  CreateUserResponse,
  SearchUsersResponse,
  GetUserResponse,
  UpdateUserRequest,
  UpdateUserResponse,
  DeleteUserRequest,
  DeleteUserResponse,
} from "@/types/admin/user";

const API_BASE_URL = "http://127.0.0.1:8080";

/*
 * 管理者用ユーザー一覧取得　(削除済み込み　or　抜き)
 */
export async function searchUsers(keyword: string, includeDeleted: boolean): Promise<ApiResponse<SearchUsersResponse>> {
  const token = getAccessToken();

  const query = new URLSearchParams();

  if (keyword) {
    query.set("keyword", keyword);
  }

  query.set("includeDeleted", String(includeDeleted));

  const response = await fetch(`${API_BASE_URL}/admin/users?${query.toString()}`, {
    method: "GET",
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });

  return response.json();
}

/*
 * 管理者用ユーザー詳細取得
 */
export async function getUser(id: number): Promise<ApiResponse<GetUserResponse>> {
  const token = getAccessToken();

  const response = await fetch(`${API_BASE_URL}/admin/users/${id}`, {
    method: "GET",
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });

  return response.json();
}

/*
 * 管理者用ユーザー新規作成
 */
export async function createUser(req: CreateUserRequest): Promise<ApiResponse<CreateUserResponse>> {
  const token = getAccessToken();

  const response = await fetch(`${API_BASE_URL}/admin/users`, {
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
 * 管理者用ユーザー更新
 */
export async function updateUser(req: UpdateUserRequest): Promise<ApiResponse<UpdateUserResponse>> {
  const token = getAccessToken();

  const response = await fetch(`${API_BASE_URL}/admin/users/${req.id}`, {
    method: "PUT",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${token}`,
    },
    body: JSON.stringify({
      name: req.name,
      email: req.email,
      password: req.password,
      role: req.role,
      departmentId: req.departmentId,
    }),
  });

  return response.json();
}

/*
 * 管理者用ユーザー削除
 */
export async function deleteUser(req: DeleteUserRequest): Promise<ApiResponse<DeleteUserResponse>> {
  const token = getAccessToken();

  const response = await fetch(`${API_BASE_URL}/admin/users/${req.id}`, {
    method: "DELETE",
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });

  return response.json();
}