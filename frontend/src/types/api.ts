/*
 * 全API共通のレスポンス型
 */

export type ApiResponse<T> = {
  data: T | null;
  error: boolean;
  code: string;
  message: string;
  detail?: unknown;
};

