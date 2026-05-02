"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import AdminSideMenu from "@/components/admin/navigation/AdminSideMenu";
import Button from "@/components/atoms/Button";
import ConfirmModal from "@/components/atoms/ConfirmModal";
import { fetchMe } from "@/api/auth";
import { searchUsers, deleteUser } from "@/api/admin/users";
import type { User } from "@/types/admin/user";

export default function AdminUsersPage() {
  const router = useRouter();

  const [keyword, setKeyword] = useState("");
  const [users, setUsers] = useState<User[]>([]);
  const [message, setMessage] = useState("");
  const [deleteTargetUser, setDeleteTargetUser] = useState<User | null>(null);
  const [includeDeleted, setIncludeDeleted] = useState(false);
  const [loginUserId, setLoginUserId] = useState<number | null>(null);

  /*
   * ユーザー一覧を取得する
   * 検索ボタン押下時に使用する
   */
  const loadUsers = async () => {
    setMessage("");

    const result = await searchUsers(keyword, includeDeleted);

    if (result.error || !result.data) {
      setMessage(result.message);
      return;
    }

    setUsers(result.data.users);
  };

  /*
   * 初回表示時にユーザー一覧を取得する
   */
  useEffect(() => {
    fetchMe().then((result) => {
      if (!result.error && result.data) {
        setLoginUserId(result.data.userId);
      }
    });

    searchUsers("", false).then((result) => {
      if (result.error || !result.data) {
        setMessage(result.message);
        return;
      }

      setUsers(result.data.users);
    });
  }, []);

  /*
   * 新規作成画面へ移動する
   */
  const handleMoveCreatePage = () => {
    router.push("/admin/users/new");
  };

  /*
   * 編集画面へ移動する
   */
  const handleMoveUpdatePage = (userId: number) => {
    router.push(`/admin/users/update/${userId}`);
  };

  /*
   * 削除確認モーダルを表示する
   */
  const handleOpenDeleteModal = (userId: number) => {
    setMessage("");

    const targetUser = users.find((user) => user.id === userId);

    if (!targetUser) {
      setMessage("削除対象のユーザーが見つかりません");
      return;
    }

    setDeleteTargetUser(targetUser);
  };

  /*
   * 削除確認モーダルを閉じる
   */
  const handleCloseDeleteModal = () => {
    setDeleteTargetUser(null);
  };

  /*
   * 削除確定
   */
  const handleConfirmDelete = async () => {
    if (!deleteTargetUser) {
      setMessage("削除対象のユーザーが見つかりません");
      return;
    }

    setMessage("");

    const result = await deleteUser({
      id: deleteTargetUser.id,
    });

    if (result.error) {
      setMessage(result.message);
      return;
    }

    setDeleteTargetUser(null);
    setMessage(result.message);

    await loadUsers();
  };

  return (
    <>
      <AdminSideMenu />

      <main style={{ minHeight: "100vh", padding: "40px", fontFamily: "sans-serif", backgroundColor: "#fff7ed", color: "#333333" }}>
        <section style={{ maxWidth: "1100px", margin: "40px auto", padding: "32px", borderRadius: "16px", backgroundColor: "#ffffff", boxShadow: "0 8px 24px rgba(0, 0, 0, 0.08)" }}>
          <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: "24px" }}>
            <h1 style={{ margin: 0, fontSize: "32px", color: "#ea580c" }}>ユーザー管理</h1>

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
              placeholder="名前・メールで検索"
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
              <Button type="button" onClick={loadUsers}>
                検索
              </Button>
            </div>
          </div>

          {message && <p style={{ marginBottom: "16px", color: "#dc2626", fontSize: "14px" }}>{message}</p>}

          <table style={{ width: "100%", borderCollapse: "collapse", fontSize: "15px", color: "#333333" }}>
            <thead>
              <tr style={{ backgroundColor: "#fff7ed" }}>
                <th style={thStyle}>ID</th>
                <th style={thStyle}>名前</th>
                <th style={thStyle}>メール</th>
                <th style={thStyle}>権限</th>
                <th style={thStyle}>所属</th>
                <th style={thStyle}>状態</th>
                <th style={thStyle}>操作</th>
              </tr>
            </thead>

            <tbody>
              {users.map((user) => {
                const isSelf = loginUserId === user.id;
                const canDelete = !user.isDeleted && !isSelf;

                return (
                  <tr key={user.id}>
                    <td style={tdStyle}>{user.id}</td>
                    <td style={tdStyle}>{user.name}</td>
                    <td style={tdStyle}>{user.email}</td>
                    <td style={tdStyle}>{user.role}</td>
                    <td style={tdStyle}>{user.departmentName || "未設定"}</td>
                    <td style={tdStyle}>{user.isDeleted ? "削除済み" : "有効"}</td>
                    <td style={tdStyle}>
                      <div style={{ display: "flex", gap: "8px" }}>
                        <button
                          type="button"
                          onClick={() => handleMoveUpdatePage(user.id)}
                          style={updateButtonStyle}
                        >
                          編集
                        </button>

                        <button
                          type="button"
                          onClick={() => handleOpenDeleteModal(user.id)}
                          disabled={!canDelete}
                          title={isSelf ? "自分自身は削除できません" : ""}
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

              {users.length === 0 && (
                <tr>
                  <td style={emptyStyle} colSpan={7}>
                    ユーザーが見つかりません
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </section>

        {deleteTargetUser && (
          <ConfirmModal
            title="削除確認"
            message="以下のユーザーを削除してもいいですか？"
            confirmText="削除する"
            cancelText="キャンセル"
            onConfirm={handleConfirmDelete}
            onCancel={handleCloseDeleteModal}
          >
            <p style={deleteTargetTextStyle}>
              <strong>名前：</strong>
              {deleteTargetUser.name}
            </p>
            <p style={deleteTargetTextStyle}>
              <strong>メール：</strong>
              {deleteTargetUser.email}
            </p>
            <p style={deleteTargetTextStyle}>
              <strong>権限：</strong>
              {deleteTargetUser.role}
            </p>
            <p style={deleteTargetTextStyle}>
              <strong>所属：</strong>
              {deleteTargetUser.departmentName || "未設定"}
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
  cursor: "pointer",
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