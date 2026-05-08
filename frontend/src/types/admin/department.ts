/*
 * 所属
 */

export type SearchDepartmentsRequest = {
  keyword: string;
  includeDeleted: boolean;
  offset: number;
  limit: number;
};

export type DepartmentDetailRequest = {
  departmentId: number;
};

export type CreateDepartmentRequest = {
  name: string;
};

export type UpdateDepartmentRequest = {
  departmentId: number;
  name: string;
};

export type DeleteDepartmentRequest = {
  departmentId: number;
};

export type DepartmentResponse = {
  id: number;
  name: string;
  isDeleted: boolean;
  createdAt: string;
  updatedAt: string;
  deletedAt: string | null;
};

export type SearchDepartmentsResponse = {
  departments: DepartmentResponse[];
  total: number;
  offset: number;
  limit: number;
  hasMore: boolean;
};

export type DepartmentDetailResponse = {
  department: DepartmentResponse;
};

export type CreateDepartmentResponse = {
  department: DepartmentResponse;
};

export type UpdateDepartmentResponse = {
  department: DepartmentResponse;
};

export type DeleteDepartmentResponse = {
  departmentId: number;
};