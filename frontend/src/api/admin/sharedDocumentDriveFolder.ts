import { apiPost } from "@/api/client";
import type {
  CreateSharedDocumentDriveFolderRequest,
  CreateSharedDocumentDriveFolderResponse,
  DeleteSharedDocumentDriveFolderRequest,
  DeleteSharedDocumentDriveFolderResponse,
  SearchSharedDocumentDriveFoldersRequest,
  SearchSharedDocumentDriveFoldersResponse,
  SharedDocumentDriveFolderDetailRequest,
  SharedDocumentDriveFolderDetailResponse,
  SyncSharedDocumentDriveFolderRequest,
  SyncSharedDocumentDriveFolderResponse,
  UpdateSharedDocumentDriveFolderRequest,
  UpdateSharedDocumentDriveFolderResponse,
  UpdateSharedDocumentDriveFolderUsersRequest,
  UpdateSharedDocumentDriveFolderUsersResponse,
} from "@/types/admin/sharedDocumentDriveFolder";

/*
 * 管理者 共有資料Driveフォルダ一覧取得
 *
 * POST /admin/shared-document-drive-folders/search
 */
export function searchSharedDocumentDriveFolders(
  request: SearchSharedDocumentDriveFoldersRequest
) {
  return apiPost<
    SearchSharedDocumentDriveFoldersResponse,
    SearchSharedDocumentDriveFoldersRequest
  >("/admin/shared-document-drive-folders/search", request);
}

/*
 * 管理者 共有資料Driveフォルダ詳細取得
 *
 * POST /admin/shared-document-drive-folders/detail
 */
export function getSharedDocumentDriveFolderDetail(
  request: SharedDocumentDriveFolderDetailRequest
) {
  return apiPost<
    SharedDocumentDriveFolderDetailResponse,
    SharedDocumentDriveFolderDetailRequest
  >("/admin/shared-document-drive-folders/detail", request);
}

/*
 * 管理者 共有資料Driveフォルダ作成
 *
 * POST /admin/shared-document-drive-folders/create
 */
export function createSharedDocumentDriveFolder(
  request: CreateSharedDocumentDriveFolderRequest
) {
  return apiPost<
    CreateSharedDocumentDriveFolderResponse,
    CreateSharedDocumentDriveFolderRequest
  >("/admin/shared-document-drive-folders/create", request);
}

/*
 * 管理者 共有資料Driveフォルダ更新
 *
 * POST /admin/shared-document-drive-folders/update
 */
export function updateSharedDocumentDriveFolder(
  request: UpdateSharedDocumentDriveFolderRequest
) {
  return apiPost<
    UpdateSharedDocumentDriveFolderResponse,
    UpdateSharedDocumentDriveFolderRequest
  >("/admin/shared-document-drive-folders/update", request);
}

/*
 * 管理者 共有資料Driveフォルダ削除
 *
 * POST /admin/shared-document-drive-folders/delete
 */
export function deleteSharedDocumentDriveFolder(
  request: DeleteSharedDocumentDriveFolderRequest
) {
  return apiPost<
    DeleteSharedDocumentDriveFolderResponse,
    DeleteSharedDocumentDriveFolderRequest
  >("/admin/shared-document-drive-folders/delete", request);
}

/*
 * 管理者 共有資料Driveフォルダ共有ユーザー更新
 *
 * POST /admin/shared-document-drive-folders/users/update
 *
 * 通常時：
 * ・targetUserIds に共有対象ユーザーIDの最終状態を渡す
 *
 * 全員追加：
 * ・shareAllUsers: true
 * ・targetUserIds: []
 *
 * 全員削除：
 * ・shareAllUsers: false
 * ・targetUserIds: []
 */
export function updateSharedDocumentDriveFolderUsers(
  request: UpdateSharedDocumentDriveFolderUsersRequest
) {
  return apiPost<
    UpdateSharedDocumentDriveFolderUsersResponse,
    UpdateSharedDocumentDriveFolderUsersRequest
  >("/admin/shared-document-drive-folders/users/update", request);
}

/*
 * 管理者 共有資料Driveフォルダ権限同期
 *
 * POST /admin/shared-document-drive-folders/sync
 */
export function syncSharedDocumentDriveFolder(
  request: SyncSharedDocumentDriveFolderRequest
) {
  return apiPost<
    SyncSharedDocumentDriveFolderResponse,
    SyncSharedDocumentDriveFolderRequest
  >("/admin/shared-document-drive-folders/sync", request);
}