/*
 * 従業員 共有資料Driveフォルダ Type
 *
 * バックエンドの user/shared_document_drive_folder_type.go に対応する。
 *
 * 注意：
 * ・ユーザー側は共有されている資料を見るだけ
 * ・作成/更新/削除/同期はできない
 * ・targetUserId は持たない
 * ・本人IDはバックエンドがJWTから取得する
 */

/*
 * 従業員用 共有資料Driveフォルダ検索Request
 */
export type SearchSharedDocumentDriveFoldersRequest = {
  keyword: string;
  offset: number;
  limit: number;
};

/*
 * 従業員用 共有資料Driveフォルダ詳細Request
 *
 * 本人に共有されていない資料はバックエンドで取得不可。
 */
export type SharedDocumentDriveFolderDetailRequest = {
  targetSharedDocumentDriveFolderId: number;
};

/*
 * 従業員用 共有資料DriveフォルダRow
 */
export type SharedDocumentDriveFolderRow = {
  id: number;

  folderName: string;
  description: string | null;
  driveFolderId: string;
  folderUrl: string;
  syncedAt: string | null;

  sharedAt: string;
  updatedAt: string;
};

/*
 * 従業員用 共有資料Driveフォルダ
 */
export type SharedDocumentDriveFolder = {
  id: number;

  folderName: string;
  description: string | null;
  driveFolderId: string;
  folderUrl: string;
  syncedAt: string | null;

  sharedAt: string;
  updatedAt: string;
};

/*
 * 従業員用 共有資料Driveフォルダ検索Response
 */
export type SearchSharedDocumentDriveFoldersResponse = {
  sharedDocumentDriveFolders: SharedDocumentDriveFolderRow[];
  total: number;
  offset: number;
  limit: number;
  hasMore: boolean;
};

/*
 * 従業員用 共有資料Driveフォルダ詳細Response
 */
export type SharedDocumentDriveFolderDetailResponse = {
  sharedDocumentDriveFolder: SharedDocumentDriveFolder;
};