[1mdiff --git a/frontend/src/app/admin/monthly-attendance-summary-exports/page.tsx b/frontend/src/app/admin/monthly-attendance-summary-exports/page.tsx[m
[1mindex c168b93..d106781 100644[m
[1m--- a/frontend/src/app/admin/monthly-attendance-summary-exports/page.tsx[m
[1m+++ b/frontend/src/app/admin/monthly-attendance-summary-exports/page.tsx[m
[36m@@ -30,6 +30,31 @@[m [mtype Department = {[m
   name: string;[m
 };[m
 [m
[32m+[m
[32m+[m[32mtype SearchBusinessTargetUsersResponse = {[m
[32m+[m[32m  users: BusinessTargetUser[];[m
[32m+[m[32m  total?: number;[m
[32m+[m[32m  offset?: number;[m
[32m+[m[32m  limit?: number;[m
[32m+[m[32m  hasMore?: boolean;[m
[32m+[m[32m};[m
[32m+[m
[32m+[m[32mtype SearchDepartmentsResponse = {[m
[32m+[m[32m  departments: Department[];[m
[32m+[m[32m  total?: number;[m
[32m+[m[32m  offset?: number;[m
[32m+[m[32m  limit?: number;[m
[32m+[m[32m  hasMore?: boolean;[m
[32m+[m[32m};[m
[32m+[m
[32m+[m[32mtype ApiResponse<TData> = {[m
[32m+[m[32m  data: TData | null;[m
[32m+[m[32m  error: boolean;[m
[32m+[m[32m  code: string;[m
[32m+[m[32m  message: string;[m
[32m+[m[32m  details?: unknown;[m
[32m+[m[32m};[m
[32m+[m
 type ExportFormState = {[m
   targetMonth: string;[m
   targetType: ExportTargetType;[m
[36m@@ -144,31 +169,37 @@[m [mexport default function AdminMonthlyAttendanceSummaryExportsPage() {[m
        * 既存の共通API関数がある場合は、このfetch部分を[m
        * searchDepartments(...) の呼び出しに置き換えてよい。[m
        */[m
[31m-      const response = await fetch("/api/admin/departments/search", {[m
[31m-        method: "POST",[m
[31m-        headers: {[m
[31m-          "Content-Type": "application/json",[m
[32m+[m[32m      const response = await fetch([m
[32m+[m[32m        buildApiUrl("/admin/departments/search"),[m
[32m+[m[32m        {[m
[32m+[m[32m          method: "POST",[m
[32m+[m[32m          headers: {[m
[32m+[m[32m            ...buildAuthHeaders(),[m
[32m+[m[32m            "Content-Type": "application/json",[m
[32m+[m[32m          },[m
[32m+[m[32m          body: JSON.stringify({[m
[32m+[m[32m            keyword: "",[m
[32m+[m[32m            includeDeleted: false,[m
[32m+[m[32m            offset: 0,[m
[32m+[m[32m            limit: 50,[m
[32m+[m[32m          }),[m
         },[m
[31m-        body: JSON.stringify({[m
[31m-          keyword: "",[m
[31m-          includeDeleted: false,[m
[31m-          offset: 0,[m
[31m-          limit: 50,[m
[31m-        }),[m
[31m-      });[m
[32m+[m[32m      );[m
 [m
[31m-      if (!response.ok) {[m
[31m-        throw new Error("所属一覧の取得に失敗しました。");[m
[31m-      }[m
[32m+[m[32m      const payload =[m
[32m+[m[32m        (await response.json()) as ApiResponse<SearchDepartmentsResponse>;[m
 [m
[31m-      const json = await response.json();[m
[32m+[m[32m      if (!response.ok || payload.error) {[m
[32m+[m[32m        throw new Error([m
[32m+[m[32m          payload.message || "所属一覧の取得に失敗しました。",[m
[32m+[m[32m        );[m
[32m+[m[32m      }[m
 [m
[31m-      const departmentList =[m
[31m-        json?.data?.departments ??[m
[31m-        json?.departments ??[m
[31m-        [];[m
[32m+[m[32m      if (!payload.data) {[m
[32m+[m[32m        throw new Error("所属一覧の取得結果が空です。");[m
[32m+[m[32m      }[m
 [m
[31m-      setDepartments(departmentList);[m
[32m+[m[32m      setDepartments(payload.data.departments);[m
     } catch (error) {[m
       setPageMessage({[m
         variant: "error",[m
[36m@@ -197,10 +228,11 @@[m [mexport default function AdminMonthlyAttendanceSummaryExportsPage() {[m
        * searchBusinessTargetUsers(...) の呼び出しに置き換えてよい。[m
        */[m
       const response = await fetch([m
[31m-        "/api/admin/users/search-business-targets",[m
[32m+[m[32m        buildApiUrl("/admin/users/search-business-targets"),[m
         {[m
           method: "POST",[m
           headers: {[m
[32m+[m[32m            ...buildAuthHeaders(),[m
             "Content-Type": "application/json",[m
           },[m
           body: JSON.stringify({[m
[36m@@ -211,16 +243,20 @@[m [mexport default function AdminMonthlyAttendanceSummaryExportsPage() {[m
         },[m
       );[m
 [m
[31m-      if (!response.ok) {[m
[31m-        throw new Error("ユーザー検索に失敗しました。");[m
[32m+[m[32m      const payload =[m
[32m+[m[32m        (await response.json()) as ApiResponse<SearchBusinessTargetUsersResponse>;[m
[32m+[m
[32m+[m[32m      if (!response.ok || payload.error) {[m
[32m+[m[32m        throw new Error([m
[32m+[m[32m          payload.message || "ユーザー検索に失敗しました。",[m
[32m+[m[32m        );[m
       }[m
 [m
[31m-      const json = await response.json();[m
[32m+[m[32m      if (!payload.data) {[m
[32m+[m[32m        throw new Error("ユーザー検索の取得結果が空です。");[m
[32m+[m[32m      }[m
 [m
[31m-      const users =[m
[31m-        json?.data?.users ??[m
[31m-        json?.users ??[m
[31m-        [];[m
[32m+[m[32m      const users = payload.data.users;[m
 [m
       setBusinessTargetUsers(users);[m
       setExportForm((current) => ({[m
[36m@@ -861,4 +897,39 @@[m [mfunction formatMonthPickerLabel([m
   return `${year}年${month}月`;[m
 }[m
 [m
[32m+[m[32mfunction buildApiUrl(path: string) {[m
[32m+[m[32m  const baseUrl =[m
[32m+[m[32m    process.env.NEXT_PUBLIC_API_BASE_URL ??[m
[32m+[m[32m    "http://localhost:8080";[m
[32m+[m
[32m+[m[32m  const normalizedBaseUrl = baseUrl.endsWith("/")[m
[32m+[m[32m    ? baseUrl.slice(0, -1)[m
[32m+[m[32m    : baseUrl;[m
[32m+[m
[32m+[m[32m  const normalizedPath = path.startsWith("/")[m
[32m+[m[32m    ? path[m
[32m+[m[32m    : `/${path}`;[m
[32m+[m
[32m+[m[32m  return `${normalizedBaseUrl}${normalizedPath}`;[m
[32m+[m[32m}[m
[32m+[m
[32m+[m[32mfunction buildAuthHeaders(): HeadersInit {[m
[32m+[m[32m  const token = getAccessToken();[m
[32m+[m
[32m+[m[32m  if (!token) {[m
[32m+[m[32m    return {};[m
[32m+[m[32m  }[m
[32m+[m
[32m+[m[32m  return {[m
[32m+[m[32m    Authorization: `Bearer ${token}`,[m
[32m+[m[32m  };[m
[32m+[m[32m}[m
[32m+[m
[32m+[m[32mfunction getAccessToken() {[m
[32m+[m[32m  if (typeof window === "undefined") {[m
[32m+[m[32m    return null;[m
[32m+[m[32m  }[m
[32m+[m
[32m+[m[32m  return window.localStorage.getItem("accessToken");[m
[32m+[m[32m}[m
 [m
