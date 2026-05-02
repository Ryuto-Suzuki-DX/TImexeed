"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import AdminSideMenu from "@/components/admin/navigation/AdminSideMenu";
import Button from "@/components/atoms/Button";
import ConfirmModal from "@/components/atoms/ConfirmModal";
import { searchDepartments, deleteDepartment } from "@/api/admin/departments";
import type { Department } from "@/types/admin/department";

export default function AdminDepartmentsPage() {
  const router = useRouter();

  const [keyword, setKeyword] = useState("");
  const [departments, setDepartments] = useState<Department[]>([]);
  const [message, setMessage] = useState("");
  const [deleteTargetDepartment, setDeleteTargetDepartment] = useState<Department | null>(null);
  const [includeDeleted, setIncludeDeleted] = useState(false);

  /*
   * 所属一覧を取得する
   * 検索ボタン押下時に使用する
   */
  const loadDepartments = async () => {
    setMessage("");

    const result = await searchDepartments(keyword, includeDeleted);

    if (result.error || !result.data) {
      setMessage(result.message);
      return;
    }

    setDepartments(result.data.departments);
  };

  /*
   * 初回表示時に所属一覧を取得する
   */
  useEffect(() => {
    searchDepartments("", false).then((result) => {
      if (result.error || !result.data) {
        setMessage(result.message);
        return;
      }

      setDepartments(result.data.departments);
    });
  }, []);

  /*
   * 新規作成画面へ移動する
   */
  const handleMoveCreatePage = () => {
    router.push("/admin/departments/new");
  };

  /*
   * 編集画面へ移動する
   */
  const handleMoveUpdatePage = (departmentId: number) => {
    router.push(`/admin/departments/update/${departmentId}`);
  };

  /*
   * 削除確認モーダルを表示する
   */
  const handleOpenDeleteModal = (departmentId: number) => {
    setMessage("");

    const targetDepartment = departments.find((department) => department.id === departmentId);

    if (!targetDepartment) {
      setMessage("削除対象の所属が見つかりません");
      return;
    }

    setDeleteTargetDepartment(targetDepartment);
  };

  /*
   * 削除確認モーダルを閉じる
   */
  const handleCloseDeleteModal = () => {
    setDeleteTargetDepartment(null);
  };

  /*
   * 削除確定
   */
  const handleConfirmDelete = async () => {
    if (!deleteTargetDepartment) {
      setMessage("削除対象の所属が見つかりません");
      return;
    }

    setMessage("");

    const result = await deleteDepartment({
      id: deleteTargetDepartment.id,
    });

    if (result.error) {
      setMessage(result.message);
      return;
    }

    setDeleteTargetDepartment(null);
    setMessage(result.message);

    await loadDepartments();
  };

  return (
    <>
      <AdminSideMenu />

      <main style={{ minHeight: "100vh", padding: "40px", fontFamily: "sans-serif", backgroundColor: "#fff7ed", color: "#333333" }}>
        <section style={{ maxWidth: "900px", margin: "40px auto", padding: "32px", borderRadius: "16px", backgroundColor: "#ffffff", boxShadow: "0 8px 24px rgba(0, 0, 0, 0.08)" }}>
          <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: "24px" }}>
            <h1 style={{ margin: 0, fontSize: "32px", color: "#ea580c" }}>所属管理</h1>

            <div style={{ width: "140px" }}>
              <Button type="button" onClick={handleMoveCreatePage}>
                新規作成
              </Button>
            </div>
          </div>

          <div style={{ display: "flex", gap: "12px", marginBottom: "24px" }}>
            <input
              value={keyword}
              onChange={(e) => setKeyword(e.target.value)}
              placeholder="所属名で検索"
              style={{
                flex: 1,
                padding: "12px",
                fontSize: "16px",
                border: "1px solid #fed7aa",
                borderRadius: "8px",
                outline: "none",
                color: "#333333",
                backgroundColor: "#ffffff",
              }}
            />

            <label style={checkboxLabelStyle}>
              <input
                type="checkbox"
                checked={includeDeleted}
                onChange={(e) => setIncludeDeleted(e.target.checked)}
              />
              削除済みを含める
            </label>

            <div style={{ width: "120px" }}>
              <Button type="button" onClick={loadDepartments}>
                検索
              </Button>
            </div>
          </div>

          {message && <p style={{ marginBottom: "16px", color: "#dc2626", fontSize: "14px" }}>{message}</p>}

          <table style={{ width: "100%", borderCollapse: "collapse", fontSize: "15px", color: "#333333" }}>
            <thead>
              <tr style={{ backgroundColor: "#fff7ed" }}>
                <th style={thStyle}>ID</th>
                <th style={thStyle}>所属名</th>
                <th style={thStyle}>状態</th>
                <th style={thStyle}>操作</th>
              </tr>
            </thead>

            <tbody>
              {departments.map((department) => {
                const canDelete = !department.isDeleted;

                return (
                  <tr key={department.id}>
                    <td style={tdStyle}>{department.id}</td>
                    <td style={tdStyle}>{department.name}</td>
                    <td style={tdStyle}>{department.isDeleted ? "削除済み" : "有効"}</td>
                    <td style={tdStyle}>
                      <div style={{ display: "flex", gap: "8px" }}>
                        <button
                          type="button"
                          onClick={() => handleMoveUpdatePage(department.id)}
                          disabled={department.isDeleted}
                          style={{
                            ...updateButtonStyle,
                            opacity: department.isDeleted ? 0.5 : 1,
                            cursor: department.isDeleted ? "not-allowed" : "pointer",
                          }}
                        >
                          編集
                        </button>

                        <button
                          type="button"
                          onClick={() => handleOpenDeleteModal(department.id)}
                          disabled={!canDelete}
                          style={{
                            ...deleteButtonStyle,
                            opacity: canDelete ? 1 : 0.5,
                            cursor: canDelete ? "pointer" : "not-allowed",
                          }}
                        >
                          削除
                        </button>
                      </div>
                    </td>
                  </tr>
                );
              })}

              {departments.length === 0 && (
                <tr>
                  <td style={emptyStyle} colSpan={4}>
                    所属が見つかりません
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </section>

        {deleteTargetDepartment && (
          <ConfirmModal
            title="削除確認"
            message="以下の所属を削除してもいいですか？"
            confirmText="削除する"
            cancelText="キャンセル"
            onConfirm={handleConfirmDelete}
            onCancel={handleCloseDeleteModal}
          >
            <p style={deleteTargetTextStyle}>
              <strong>所属名：</strong>
              {deleteTargetDepartment.name}
            </p>
          </ConfirmModal>
        )}
      </main>
    </>
  );
}

const thStyle: React.CSSProperties = {
  padding: "12px",
  borderBottom: "1px solid #fed7aa",
  textAlign: "left",
  color: "#9a3412",
};

const tdStyle: React.CSSProperties = {
  padding: "12px",
  borderBottom: "1px solid #ffedd5",
  color: "#333333",
};

const emptyStyle: React.CSSProperties = {
  padding: "24px",
  borderBottom: "1px solid #ffedd5",
  textAlign: "center",
  color: "#666666",
};

const checkboxLabelStyle: React.CSSProperties = {
  display: "flex",
  alignItems: "center",
  gap: "6px",
  whiteSpace: "nowrap",
  fontSize: "14px",
  color: "#333333",
};

const updateButtonStyle: React.CSSProperties = {
  padding: "8px 12px",
  border: "none",
  borderRadius: "6px",
  backgroundColor: "#f97316",
  color: "#ffffff",
  fontSize: "14px",
  fontWeight: "bold",
};

const deleteButtonStyle: React.CSSProperties = {
  padding: "8px 12px",
  border: "none",
  borderRadius: "6px",
  backgroundColor: "#dc2626",
  color: "#ffffff",
  fontSize: "14px",
  fontWeight: "bold",
};

const deleteTargetTextStyle: React.CSSProperties = {
  margin: "0 0 8px",
  fontSize: "15px",
  color: "#333333",
};