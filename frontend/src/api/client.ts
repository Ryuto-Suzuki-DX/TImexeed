import { getAccessToken } from "@/api/auth";
import { API_BASE_URL } from "@/api/config";
import type { ApiResponse } from "@/types/api";

/*
 * API通信共通処理
 *
 * ・APIのベースURLは .env.local の NEXT_PUBLIC_API_BASE_URL を使う
 * ・Content-Typeを付ける
 * ・accessTokenがあれば Authorization を付ける
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