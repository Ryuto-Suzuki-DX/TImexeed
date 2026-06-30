/*
 * パスワード変更で使用する型
 */

export type ChangePasswordRequest = {
  currentPassword: string;
  newPassword: string;
};

export type ChangePasswordResponse = {
  mustChangePassword: boolean;
};
