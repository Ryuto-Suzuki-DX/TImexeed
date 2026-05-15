import { apiPost } from "@/api/client";
import type {
  CreateExternalStorageLinkRequest,
  CreateExternalStorageLinkResponse,
  DeleteExternalStorageLinkRequest,
  DeleteExternalStorageLinkResponse,
  ExternalStorageLinkDetailRequest,
  ExternalStorageLinkDetailResponse,
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

export function getExternalStorageLinkDetail(request: ExternalStorageLinkDetailRequest) {
  return apiPost<ExternalStorageLinkDetailResponse, ExternalStorageLinkDetailRequest>(
    "/admin/external-storage-links/detail",
    request
  );
}

export function createExternalStorageLink(request: CreateExternalStorageLinkRequest) {
  return apiPost<CreateExternalStorageLinkResponse, CreateExternalStorageLinkRequest>(
    "/admin/external-storage-links/create",
    request
  );
}

export function updateExternalStorageLink(request: UpdateExternalStorageLinkRequest) {
  return apiPost<UpdateExternalStorageLinkResponse, UpdateExternalStorageLinkRequest>(
    "/admin/external-storage-links/update",
    request
  );
}

export function deleteExternalStorageLink(request: DeleteExternalStorageLinkRequest) {
  return apiPost<DeleteExternalStorageLinkResponse, DeleteExternalStorageLinkRequest>(
    "/admin/external-storage-links/delete",
    request
  );
}