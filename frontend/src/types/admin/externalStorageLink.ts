/*
 * 外部ストレージリンク
 *
 * Google Driveなど、外部ストレージ上のフォルダURL / ファイルURLを管理する。
 *
 * 注意：
 * ・管理者が任意にリンクを新規作成/削除する運用にはしない
 * ・固定されたリンク種別/リンク名に対して、URL/説明/管理メモだけを更新する
 */

export type SearchExternalStorageLinksRequest = {
  keyword: string;
  linkType: string;
  includeDeleted: boolean;
  offset: number;
  limit: number;
};

export type UpdateExternalStorageLinkRequest = {
  externalStorageLinkId: number;
  url: string;
  description: string | null;
  memo: string | null;
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

export type UpdateExternalStorageLinkResponse = {
  externalStorageLink: ExternalStorageLinkResponse;
};