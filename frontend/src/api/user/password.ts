import { API_BASE_URL } from "@/api/config";
import {
  getAccessToken,
  handleUnauthorizedResponse,
  redirectToLogin,
} from "@/api/auth";
import type { ApiResponse } from "@/types/api";
import type {
  ChangePasswordRequest,
  ChangePasswordResponse,
} from "@/types/user/password";

/*
 * パスワード変更
 *
 * POST /user/password/change
 */
export async function changePassword(
  request: ChangePasswordRequest,
): Promise<ApiResponse<ChangePasswordResponse>> {
  const token = getAccessToken();

  if (!token) {
    redirectToLogin();

    return {
      data: null,
      error: true,
      code: "TOKEN_NOT_FOUND",
      message: "ログイン情報がありません",
    };
  }

  const response = await fetch(`${API_BASE_URL}/user/password/change`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${token}`,
    },
    body: JSON.stringify(request),
  });

  if (handleUnauthorizedResponse(response)) {
    return {
      data: null,
      error: true,
      code: "UNAUTHORIZED",
      message: "ログイン期限が切れました。再ログインしてください。",
    };
  }

  return response.json();
}
