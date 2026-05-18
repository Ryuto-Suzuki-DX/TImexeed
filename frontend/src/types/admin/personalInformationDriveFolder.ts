/*
 * 管理者 個人情報Driveフォルダ Type
 *
 * バックエンドAPIのRequest/Response型に対応する。
 *
 * 方針：
 * ・管理者は全ユーザー分の個人情報Driveフォルダを検索できる
 * ・管理者は対象ユーザーのフォルダを作成/権限同期できる
 * ・管理者は対象ユーザーのフォルダURLを取得できる
 * ・URLにIDは載せず、targetUserId は request body で送る
 */

export type PersonalInformationDriveFolderSearchRow = {
  userId: number;
  userName: string;
  userEmail: string;
  userRole: string;
  departmentId: number | null;

  personalInformationDriveFolderId: number | null;
  externalStorageLinkId: number | null;
  folderName: string | null;
  driveFolderId: string | null;
  folderUrl: string | null;
  syncedAt: string | null;
  folderCreatedAt: string | null;
  folderUpdatedAt: string | null;
  folderRegistered: boolean;
};

export type PersonalInformationDriveFolder = {
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

export type SearchPersonalInformationDriveFoldersRequest = {
  keyword: string;
  offset: number;
  limit: number;
};

export type SearchPersonalInformationDriveFoldersResponse = {
  personalInformationDriveFolders: PersonalInformationDriveFolderSearchRow[];
  total: number;
  offset: number;
  limit: number;
  hasMore: boolean;
};

export type SyncPersonalInformationDriveFolderRequest = {
  targetUserId: number;
};

export type SyncPersonalInformationDriveFolderResponse = {
  personalInformationDriveFolder: PersonalInformationDriveFolder;
};

export type ViewPersonalInformationDriveFolderRequest = {
  targetUserId: number;
};

export type ViewPersonalInformationDriveFolderResponse = {
  personalInformationDriveFolder: PersonalInformationDriveFolder;
};
