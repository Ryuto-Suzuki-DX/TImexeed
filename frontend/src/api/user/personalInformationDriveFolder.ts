import { apiPost } from "@/api/client";
import type {
  GetMyPersonalInformationDriveFolderRequest,
  GetMyPersonalInformationDriveFolderResponse,
} from "@/types/user/personalInformationDriveFolder";

/*
 * 自分の個人情報Driveフォルダ取得
 *
 * POST /user/personal-information-drive-folders/get
 *
 * 注意：
 * ・従業員側はtargetUserIdを送らない
 * ・バックエンド側でJWTから本人userIdを取得する
 */
export function getMyPersonalInformationDriveFolder(
  request: GetMyPersonalInformationDriveFolderRequest = {}
) {
  return apiPost<
    GetMyPersonalInformationDriveFolderResponse,
    GetMyPersonalInformationDriveFolderRequest
  >("/user/personal-information-drive-folders/get", request);
}
