import { API_BASE_URL } from "@/api/config";
import type { ApiResponse } from "@/types/api";
import type { LoginResponse, MeResponse } from "@/types/auth";

const ACCESS_TOKEN_KEY = "accessToken";

/*
 * accessTokenを保存する
 */
export function saveAccessToken(token: string) {
  localStorage.setItem(ACCESS_TOKEN_KEY, token);
}

/*
 * accessTokenを取得する
 */
export function getAccessToken() {
  return localStorage.getItem(ACCESS_TOKEN_KEY);
}

/*
 * accessTokenを削除する
 */
export function removeAccessToken() {
  localStorage.removeItem(ACCESS_TOKEN_KEY);
}

/*
 * ログインする
 *
 * POST /auth/login
 */
export async function login(email: string, password: string): Promise<ApiResponse<LoginResponse>> {
  const response = await fetch(`${API_BASE_URL}/auth/login`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      email,
      password,
    }),
  });

  return response.json();
}

/*
 * ログイン中ユーザー情報を取得する
 *
 * GET /auth/me
 */
export async function fetchMe(): Promise<ApiResponse<MeResponse>> {
  const token = getAccessToken();

  if (!token) {
    return {
      data: null,
      error: true,
      code: "TOKEN_NOT_FOUND",
      message: "ログイン情報がありません",
    };
  }

  const response = await fetch(`${API_BASE_URL}/auth/me`, {
    method: "GET",
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });

  return response.json();
}