import { apiPost } from "@/api/client";
import type {
  SearchPersonalInformationDriveFoldersRequest,
  SearchPersonalInformationDriveFoldersResponse,
  SyncPersonalInformationDriveFolderRequest,
  SyncPersonalInformationDriveFolderResponse,
  ViewPersonalInformationDriveFolderRequest,
  ViewPersonalInformationDriveFolderResponse,
} from "@/types/admin/personalInformationDriveFolder";

/*
 * 個人情報Driveフォルダ検索
 *
 * POST /admin/personal-information-drive-folders/search
 */
export function searchPersonalInformationDriveFolders(
  request: SearchPersonalInformationDriveFoldersRequest
) {
  return apiPost<
    SearchPersonalInformationDriveFoldersResponse,
    SearchPersonalInformationDriveFoldersRequest
  >("/admin/personal-information-drive-folders/search", request);
}

/*
 * 個人情報Driveフォルダ作成/権限同期
 *
 * POST /admin/personal-information-drive-folders/sync
 */
export function syncPersonalInformationDriveFolder(
  request: SyncPersonalInformationDriveFolderRequest
) {
  return apiPost<
    SyncPersonalInformationDriveFolderResponse,
    SyncPersonalInformationDriveFolderRequest
  >("/admin/personal-information-drive-folders/sync", request);
}

/*
 * 個人情報Driveフォルダ表示URL取得
 *
 * POST /admin/personal-information-drive-folders/view
 */
export function viewPersonalInformationDriveFolder(
  request: ViewPersonalInformationDriveFolderRequest
) {
  return apiPost<
    ViewPersonalInformationDriveFolderResponse,
    ViewPersonalInformationDriveFolderRequest
  >("/admin/personal-information-drive-folders/view", request);
}
