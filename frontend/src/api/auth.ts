import { getAccessToken } from "@/lib/auth";

const API_BASE_URL = "http://127.0.0.1:8080";

type ApiResponse<T> = {
  data: T | null;
  error: boolean;
  code: string;
  message: string;
  detail?: unknown;
};

export type LoginResponse = {
  accessToken: string;
  user: {
    id: number;
    name: string;
    email: string;
    role: string;
  };
};

export type MeResponse = {
  userId: number;
  email: string;
  role: string;
};

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

export async function fetchMe(): Promise<ApiResponse<MeResponse>> {
  const token = getAccessToken();

  const response = await fetch(`${API_BASE_URL}/auth/me`, {
    method: "GET",
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });

  return response.json();
}