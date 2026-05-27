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
 *
 * 注意：
 * ・DriveフォルダURL/IDは画面から送らない
 * ・バックエンド側で external_storage_links の SHARED_DOCUMENT_DRIVE_ROOT を参照する
 * ・その親フォルダ配下に folderName のDriveフォルダを作成する
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
 *
 * 注意：
 * ・Driveフォルダ自体の場所は更新しない
 * ・Timexeed上の表示名・説明のみ更新する
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
 *
 * 注意：
 * ・DB上は論理削除
 * ・Drive上のフォルダ自体は削除しない
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
 * 管理者 共有資料Driveフォルダ権限同期
 *
 * POST /admin/shared-document-drive-folders/sync
 *
 * targetSharedDocumentDriveFolderId = 0:
 * ・有効な共有資料Driveフォルダ全件を同期する
 *
 * targetSharedDocumentDriveFolderId > 0:
 * ・指定した共有資料Driveフォルダ1件だけ同期する
 */
export function syncSharedDocumentDriveFolder(
  request: SyncSharedDocumentDriveFolderRequest
) {
  return apiPost<
    SyncSharedDocumentDriveFolderResponse,
    SyncSharedDocumentDriveFolderRequest
  >("/admin/shared-document-drive-folders/sync", request);
}
