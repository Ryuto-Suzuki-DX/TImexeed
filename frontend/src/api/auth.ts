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
  if (typeof window === "undefined") {
    return null;
  }

  return localStorage.getItem(ACCESS_TOKEN_KEY);
}

/*
 * accessTokenを削除する
 */
export function removeAccessToken() {
  if (typeof window === "undefined") {
    return;
  }

  localStorage.removeItem(ACCESS_TOKEN_KEY);
}

/*
 * ログインページへ遷移する
 *
 * 認証切れ時に使う。
 * 現在のURLを redirect パラメータに残しておく。
 */
export function redirectToLogin() {
  if (typeof window === "undefined") {
    return;
  }

  removeAccessToken();

  const currentPath = window.location.pathname + window.location.search;

  if (window.location.pathname === "/login") {
    return;
  }

  window.location.href = `/login?redirect=${encodeURIComponent(currentPath)}`;
}

/*
 * 401 Unauthorized を共通処理する
 *
 * true  : 401だったのでログイン画面へ遷移した
 * false : 401ではない
 */
export function handleUnauthorizedResponse(response: Response) {
  if (response.status !== 401) {
    return false;
  }

  redirectToLogin();
  return true;
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
    redirectToLogin();

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