/*
 * 外部ストレージリンク
 *
 * Google Driveなど、外部ストレージ上のフォルダURL / ファイルURLを管理する。
 */

export type SearchExternalStorageLinksRequest = {
  keyword: string;
  linkType: string;
  includeDeleted: boolean;
  offset: number;
  limit: number;
};

export type ExternalStorageLinkDetailRequest = {
  externalStorageLinkId: number;
};

export type CreateExternalStorageLinkRequest = {
  linkType: string;
  linkName: string;
  url: string;
  description: string | null;
  memo: string | null;
};

export type UpdateExternalStorageLinkRequest = {
  externalStorageLinkId: number;
  linkType: string;
  linkName: string;
  url: string;
  description: string | null;
  memo: string | null;
};

export type DeleteExternalStorageLinkRequest = {
  externalStorageLinkId: number;
};

export type ExternalStorageLinkResponse = {
  id: number;
  linkType: string;
  linkName: string;
  url: string;
  description: string | null;
  memo: string | null;
  isDeleted: boolean;
  createdAt: string;
  updatedAt: string;
  deletedAt: string | null;
};

export type SearchExternalStorageLinksResponse = {
  externalStorageLinks: ExternalStorageLinkResponse[];
  total: number;
  offset: number;
  limit: number;
  hasMore: boolean;
};

export type ExternalStorageLinkDetailResponse = {
  externalStorageLink: ExternalStorageLinkResponse;
};

export type CreateExternalStorageLinkResponse = {
  externalStorageLink: ExternalStorageLinkResponse;
};

export type UpdateExternalStorageLinkResponse = {
  externalStorageLink: ExternalStorageLinkResponse;
};

export type DeleteExternalStorageLinkResponse = {
  externalStorageLinkId: number;
};