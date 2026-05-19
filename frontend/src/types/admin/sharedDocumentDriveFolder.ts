/*
 * 管理者 共有資料Driveフォルダ Type
 *
 * バックエンドの shared_document_drive_folder_type.go に対応する。
 *
 * 共有資料Driveフォルダ：
 * ・管理者がGoogle Drive上で作成済みのフォルダをTimexeedに登録する
 * ・共有対象ユーザーを追加/削除できる
 * ・Drive権限同期で、管理者全員 + 共有対象ユーザーへ writer 権限を付与する
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
 *
 * driveFolderUrlOrId:
 * ・Google DriveフォルダURL
 * ・またはフォルダID
 */
export type CreateSharedDocumentDriveFolderRequest = {
  folderName: string;
  description: string | null;
  driveFolderUrlOrId: string;
};

/*
 * 管理者用 共有資料Driveフォルダ更新Request
 */
export type UpdateSharedDocumentDriveFolderRequest = {
  targetSharedDocumentDriveFolderId: number;
  folderName: string;
  description: string | null;
  driveFolderUrlOrId: string;
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
 * 管理者用 共有資料Driveフォルダ共有ユーザー更新Request
 *
 * 通常時：
 * ・targetUserIds を共有対象の最終状態として扱う
 *
 * shareAllUsers = true の場合：
 * ・targetUserIds は空配列でよい
 * ・バックエンド側で有効なUSER全員を共有対象にする
 *
 * targetUserIds = [] かつ shareAllUsers = false の場合：
 * ・共有対象ユーザーを全削除する
 */
export type UpdateSharedDocumentDriveFolderUsersRequest = {
  targetSharedDocumentDriveFolderId: number;
  targetUserIds: number[];
  shareAllUsers: boolean;
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

  sharedUserCount: number;

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
 * 管理者用 共有資料Driveフォルダ共有ユーザーResponse
 */
export type SharedDocumentDriveFolderUser = {
  id: number;

  sharedDocumentDriveFolderId: number;

  userId: number;
  userName: string;
  userEmail: string;
  userRole: string;

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
  sharedUsers: SharedDocumentDriveFolderUser[];
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
 * 管理者用 共有資料Driveフォルダ同期Response
 */
export type SyncSharedDocumentDriveFolderResponse = {
  sharedDocumentDriveFolder: SharedDocumentDriveFolder;
  sharedUsers: SharedDocumentDriveFolderUser[];
};

/*
 * 管理者用 共有資料Driveフォルダ共有ユーザー更新Response
 */
export type UpdateSharedDocumentDriveFolderUsersResponse = {
  sharedDocumentDriveFolderId: number;
  sharedUsers: SharedDocumentDriveFolderUser[];
};