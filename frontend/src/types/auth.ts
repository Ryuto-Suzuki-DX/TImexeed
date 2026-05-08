/*
 * 認証系で使用する型
 */

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
  name: string;
  email: string;
  role: string;
};