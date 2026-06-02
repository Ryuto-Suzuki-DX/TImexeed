import {
  getAccessToken,
  handleUnauthorizedResponse,
} from "@/api/auth";
import { API_BASE_URL } from "@/api/config";
import type { ApiResponse } from "@/types/api";

/*
 * API通信共通処理
 *
 * ・APIのベースURLは .env.local の NEXT_PUBLIC_API_BASE_URL を使う
 * ・Content-Typeを付ける
 * ・accessTokenがあれば Authorization を付ける
 * ・401 Unauthorized の場合はログインページへ戻す
 * ・JSONを返す
 */
async function apiRequest<T>(path: string, options: RequestInit): Promise<ApiResponse<T>> {
  const token = getAccessToken();

  const headers: HeadersInit = {
    "Content-Type": "application/json",
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
    ...options.headers,
  };

  const response = await fetch(`${API_BASE_URL}${path}`, {
    ...options,
    headers,
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

/*
 * GET通信
 */
export function apiGet<T>(path: string): Promise<ApiResponse<T>> {
  return apiRequest<T>(path, {
    method: "GET",
  });
}

/*
 * POST通信
 */
export function apiPost<T, B = unknown>(path: string, body: B): Promise<ApiResponse<T>> {
  return apiRequest<T>(path, {
    method: "POST",
    body: JSON.stringify(body),
  });
}