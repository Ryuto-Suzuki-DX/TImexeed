/*
 * 従業員 個人情報Driveフォルダ Type
 *
 * バックエンドAPIのRequest/Response型に対応する。
 *
 * 方針：
 * ・従業員側は検索しない
 * ・targetUserId は送らない
 * ・バックエンド側でJWTから本人userIdを取得する
 * ・自分の個人情報DriveフォルダURLだけ取得する
 */

export type MyPersonalInformationDriveFolder = {
  id: number;
  userId: number;
  userName: string;
  userEmail: string;
  externalStorageLinkId: number;
  folderName: string;
  driveFolderId: string;
  folderUrl: string;
  syncedAt: string | null;
  createdAt: string;
  updatedAt: string;
};

export type GetMyPersonalInformationDriveFolderRequest = Record<string, never>;

export type GetMyPersonalInformationDriveFolderResponse = {
  personalInformationDriveFolder: MyPersonalInformationDriveFolder;
};
