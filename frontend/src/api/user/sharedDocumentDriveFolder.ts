import { apiPost } from "@/api/client";
import type {
  SearchSharedDocumentDriveFoldersRequest,
  SearchSharedDocumentDriveFoldersResponse,
  SharedDocumentDriveFolderDetailRequest,
  SharedDocumentDriveFolderDetailResponse,
} from "@/types/user/sharedDocumentDriveFolder";

/*
 * 従業員 共有資料Driveフォルダ一覧取得
 *
 * POST /user/shared-document-drive-folders/search
 *
 * ログイン済みUSERが閲覧できる全ユーザー向け共有資料を取得する。
 */
export function searchSharedDocumentDriveFolders(
  request: SearchSharedDocumentDriveFoldersRequest
) {
  return apiPost<
    SearchSharedDocumentDriveFoldersResponse,
    SearchSharedDocumentDriveFoldersRequest
  >("/user/shared-document-drive-folders/search", request);
}

/*
 * 従業員 共有資料Driveフォルダ詳細取得
 *
 * POST /user/shared-document-drive-folders/detail
 *
 * ログイン済みUSERが閲覧できる全ユーザー向け共有資料の詳細を取得する。
 */
export function getSharedDocumentDriveFolderDetail(
  request: SharedDocumentDriveFolderDetailRequest
) {
  return apiPost<
    SharedDocumentDriveFolderDetailResponse,
    SharedDocumentDriveFolderDetailRequest
  >("/user/shared-document-drive-folders/detail", request);
}
