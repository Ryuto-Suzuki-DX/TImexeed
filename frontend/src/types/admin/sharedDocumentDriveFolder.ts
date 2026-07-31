/*
 * 管理者 共有資料Driveフォルダ Type
 *
 * バックエンドの admin/shared_document_drive_folder_type.go に対応する。
 *
 * 共有資料Driveフォルダ：
 * ・管理者がTimexeed上で表示名・説明を入力する
 * ・バックエンドが external_storage_links の SHARED_DOCUMENT_DRIVE_ROOT を取得する
 * ・その親フォルダ配下にGoogle Driveフォルダを作成する
 * ・作成されたDriveフォルダID/URLをDBへ保存する
 * ・権限同期ボタンで、有効な管理者・一般ユーザーへDrive権限を同期する
 *
 * 廃止したもの：
 * ・画面からの driveFolderUrlOrId 入力
 * ・共有対象ユーザー選択
 * ・/users/update API
 * ・sharedUserCount
 * ・sharedUsers
 */

/*
 * 管理者用 共有資料Driveフォルダ検索Request
 */
export type SearchSharedDocumentDriveFoldersRequest = {
  keyword: string;
  offset: number;
  limit: number;
};

/*
 * 管理者用 共有資料Driveフォルダ詳細Request
 */
export type SharedDocumentDriveFolderDetailRequest = {
  targetSharedDocumentDriveFolderId: number;
};

/*
 * 管理者用 共有資料Driveフォルダ作成Request
 */
export type CreateSharedDocumentDriveFolderRequest = {
  folderName: string;
  description: string | null;
};

/*
 * 管理者用 共有資料Driveフォルダ更新Request
 */
export type UpdateSharedDocumentDriveFolderRequest = {
  targetSharedDocumentDriveFolderId: number;
  folderName: string;
  description: string | null;
};

/*
 * 管理者用 共有資料Driveフォルダ削除Request
 */
export type DeleteSharedDocumentDriveFolderRequest = {
  targetSharedDocumentDriveFolderId: number;
};

/*
 * 管理者用 共有資料Driveフォルダ同期Request
 */
export type SyncSharedDocumentDriveFolderRequest = {
  targetSharedDocumentDriveFolderId: number;
};

/*
 * 管理者用 共有資料Driveフォルダ検索Row
 */
export type SharedDocumentDriveFolderSearchRow = {
  id: number;
  folderName: string;
  description: string | null;
  driveFolderId: string;
  folderUrl: string;
  syncedAt: string | null;
  createdAt: string;
  updatedAt: string;
};

/*
 * 管理者用 共有資料DriveフォルダResponse
 */
export type SharedDocumentDriveFolder = {
  id: number;
  folderName: string;
  description: string | null;
  driveFolderId: string;
  folderUrl: string;
  syncedAt: string | null;
  createdAt: string;
  updatedAt: string;
};

/*
 * 管理者用 共有資料Driveフォルダ検索Response
 */
export type SearchSharedDocumentDriveFoldersResponse = {
  sharedDocumentDriveFolders: SharedDocumentDriveFolderSearchRow[];
  total: number;
  offset: number;
  limit: number;
  hasMore: boolean;
};

/*
 * 管理者用 共有資料Driveフォルダ詳細Response
 */
export type SharedDocumentDriveFolderDetailResponse = {
  sharedDocumentDriveFolder: SharedDocumentDriveFolder;
};

/*
 * 管理者用 共有資料Driveフォルダ作成Response
 */
export type CreateSharedDocumentDriveFolderResponse = {
  sharedDocumentDriveFolder: SharedDocumentDriveFolder;
};

/*
 * 管理者用 共有資料Driveフォルダ更新Response
 */
export type UpdateSharedDocumentDriveFolderResponse = {
  sharedDocumentDriveFolder: SharedDocumentDriveFolder;
};

/*
 * 管理者用 共有資料Driveフォルダ削除Response
 */
export type DeleteSharedDocumentDriveFolderResponse = {
  sharedDocumentDriveFolderId: number;
};

/*
 * 共有資料Driveフォルダ権限同期失敗情報
 */
export type SharedDocumentDrivePermissionSyncFailure = {
  sharedDocumentDriveFolderId: number;
  sharedDocumentDriveFolderName: string;
  emailAddress: string;
  operation: string;
  error: string;
};

/*
 * 管理者用 共有資料Driveフォルダ同期Response
 */
export type SyncSharedDocumentDriveFolderResponse = {
  sharedDocumentDriveFolders: SharedDocumentDriveFolder[];
  syncedFolderCount: number;
  targetAdminCount: number;
  targetUserCount: number;
  permissionSuccessCount: number;
  permissionFailureCount: number;
  permissionFailures: SharedDocumentDrivePermissionSyncFailure[];
  hasPermissionFailures: boolean;
  syncedAt: string;
};
