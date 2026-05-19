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
 * 本人に共有されている資料だけ返る。
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
 * 本人に共有されていない資料は取得できない。
 */
export function getSharedDocumentDriveFolderDetail(
  request: SharedDocumentDriveFolderDetailRequest
) {
  return apiPost<
    SharedDocumentDriveFolderDetailResponse,
    SharedDocumentDriveFolderDetailRequest
  >("/user/shared-document-drive-folders/detail", request);
}