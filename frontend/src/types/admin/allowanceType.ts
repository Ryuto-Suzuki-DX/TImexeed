/*
 * 管理者 手当種別マスター
 *
 * 月次手当登録時のドロップダウン候補を管理する。
 */

export type SearchAllowanceTypesRequest = {
  keyword: string;
  includeInactive: boolean;
  includeDeleted: boolean;
  offset: number;
  limit: number;
};

export type AllowanceTypeDetailRequest = {
  allowanceTypeId: number;
};

export type CreateAllowanceTypeRequest = {
  name: string;
  displayOrder: number;
  isActive: boolean;
};

export type UpdateAllowanceTypeRequest = {
  allowanceTypeId: number;
  name: string;
  displayOrder: number;
  isActive: boolean;
};

export type DeleteAllowanceTypeRequest = {
  allowanceTypeId: number;
};

export type AllowanceTypeResponse = {
  id: number;
  name: string;
  displayOrder: number;
  isActive: boolean;
  isDeleted: boolean;
  createdAt: string;
  updatedAt: string;
  deletedAt: string | null;
};

export type SearchAllowanceTypesResponse = {
  allowanceTypes: AllowanceTypeResponse[];
  total: number;
  offset: number;
  limit: number;
  hasMore: boolean;
};

export type AllowanceTypeDetailResponse = {
  allowanceType: AllowanceTypeResponse;
};

export type CreateAllowanceTypeResponse = {
  allowanceType: AllowanceTypeResponse;
};

export type UpdateAllowanceTypeResponse = {
  allowanceType: AllowanceTypeResponse;
};

export type DeleteAllowanceTypeResponse = {
  allowanceTypeId: number;
};
