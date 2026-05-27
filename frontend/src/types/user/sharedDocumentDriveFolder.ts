/*
 * 従業員 共有資料Driveフォルダ Type
 *
 * バックエンドの user/shared_document_drive_folder_type.go に対応する。
 *
 * 注意：
 * ・ユーザー側は全ユーザー向け共有資料を見るだけ
 * ・作成/更新/削除/同期はできない
 * ・targetUserId は持たない
 * ・本人IDは送らない
 * ・Drive内部IDはユーザー側へ返さない
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
 */
export type SharedDocumentDriveFolderDetailRequest = {
  targetSharedDocumentDriveFolderId: number;
};

/*
 * 従業員用 共有資料Driveフォルダ検索Row
 */
export type SharedDocumentDriveFolderSearchRow = {
  id: number;

  folderName: string;
  description: string | null;
  folderUrl: string;
  syncedAt: string | null;

  createdAt: string;
  updatedAt: string;
};

/*
 * 従業員用 共有資料Driveフォルダ
 */
export type SharedDocumentDriveFolder = {
  id: number;

  folderName: string;
  description: string | null;
  folderUrl: string;
  syncedAt: string | null;

  createdAt: string;
  updatedAt: string;
};

/*
 * 従業員用 共有資料Driveフォルダ検索Response
 */
export type SearchSharedDocumentDriveFoldersResponse = {
  sharedDocumentDriveFolders: SharedDocumentDriveFolderSearchRow[];
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
