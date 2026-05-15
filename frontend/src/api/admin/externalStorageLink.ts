import { apiPost } from "@/api/client";
import type {
  SearchExternalStorageLinksRequest,
  SearchExternalStorageLinksResponse,
  UpdateExternalStorageLinkRequest,
  UpdateExternalStorageLinkResponse,
} from "@/types/admin/externalStorageLink";

export function searchExternalStorageLinks(request: SearchExternalStorageLinksRequest) {
  return apiPost<SearchExternalStorageLinksResponse, SearchExternalStorageLinksRequest>(
    "/admin/external-storage-links/search",
    request
  );
}

export function updateExternalStorageLink(request: UpdateExternalStorageLinkRequest) {
  return apiPost<UpdateExternalStorageLinkResponse, UpdateExternalStorageLinkRequest>(
    "/admin/external-storage-links/update",
    request
  );
}