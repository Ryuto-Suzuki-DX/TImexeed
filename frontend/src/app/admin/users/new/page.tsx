"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import AdminSideMenu from "@/components/admin/navigation/AdminSideMenu";
import Button from "@/components/atoms/Button";
import { createUser } from "@/api/admin/users";
import { searchDepartments } from "@/api/admin/departments";
import type { Department } from "@/types/admin/department";

type ValidationDetails = {
  name?: string;
  email?: string;
  password?: string;
  role?: string;
  departmentId?: string;
  request?: string;
};

export default function AdminUserNewPage() {
  const router = useRouter();

  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [role, setRole] = useState("USER");
  const [departmentId, setDepartmentId] = useState<number | null>(null);
  const [departments, setDepartments] = useState<Department[]>([]);
  const [message, setMessage] = useState("");
  const [details, setDetails] = useState<ValidationDetails>({});

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
   * ユーザー新規作成
   */
  const handleCreate = async () => {
    setMessage("");
    setDetails({});

    const result = await createUser({
      name,
      email,
      password,
      role,
      departmentId,
    });

    if (result.error || !result.data) {
      setMessage(result.message);

      if (result.details && typeof result.details === "object") {
        setDetails(result.details as ValidationDetails);
      }

      return;
    }

    router.push("/admin/users");
  };

  /*
   * 一覧画面へ戻る
   */
  const handleBack = () => {
    router.push("/admin/users");
  };

  return (
    <>
      <AdminSideMenu />

      <main style={{ minHeight: "100vh", padding: "40px", fontFamily: "sans-serif", backgroundColor: "#fff7ed", color: "#333333" }}>
        <section style={{ maxWidth: "720px", margin: "40px auto", padding: "32px", borderRadius: "16px", backgroundColor: "#ffffff", boxShadow: "0 8px 24px rgba(0, 0, 0, 0.08)" }}>
          <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: "24px" }}>
            <h1 style={{ margin: 0, fontSize: "32px", color: "#ea580c" }}>
              ユーザー新規作成
            </h1>

            <div style={{ width: "120px" }}>
              <Button type="button" onClick={handleBack}>
                戻る
              </Button>
            </div>
          </div>

          {message && (
            <p style={{ margin: "0 0 16px", color: "#dc2626", fontSize: "14px" }}>
              {message}
            </p>
          )}

          <div style={{ display: "flex", flexDirection: "column", gap: "18px" }}>
            <label style={labelStyle}>
              名前
              <input
                value={name}
                onChange={(e) => setName(e.target.value)}
                style={inputStyle}
              />
              {details.name && <span style={errorStyle}>{details.name}</span>}
            </label>

            <label style={labelStyle}>
              メールアドレス
              <input
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                style={inputStyle}
              />
              {details.email && <span style={errorStyle}>{details.email}</span>}
            </label>

            <label style={labelStyle}>
              パスワード
              <input
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                style={inputStyle}
              />
              {details.password && <span style={errorStyle}>{details.password}</span>}
            </label>

            <label style={labelStyle}>
              権限
              <select
                value={role}
                onChange={(e) => setRole(e.target.value)}
                style={inputStyle}
              >
                <option value="USER">USER</option>
                <option value="ADMIN">ADMIN</option>
              </select>
              {details.role && <span style={errorStyle}>{details.role}</span>}
            </label>

            <label style={labelStyle}>
              所属
              <select
                value={departmentId ?? ""}
                onChange={(e) => {
                  const value = e.target.value;
                  setDepartmentId(value === "" ? null : Number(value));
                }}
                style={inputStyle}
              >
                <option value="">所属を選択してください</option>
                {departments.map((department) => (
                  <option key={department.id} value={department.id}>
                    {department.name}
                  </option>
                ))}
              </select>
              {details.departmentId && <span style={errorStyle}>{details.departmentId}</span>}
            </label>

            {details.request && (
              <p style={{ margin: 0, color: "#dc2626", fontSize: "14px" }}>
                {details.request}
              </p>
            )}

            <div style={{ width: "180px", marginTop: "8px" }}>
              <Button type="button" onClick={handleCreate}>
                登録
              </Button>
            </div>
          </div>
        </section>
      </main>
    </>
  );
}

const labelStyle: React.CSSProperties = {
  display: "flex",
  flexDirection: "column",
  gap: "8px",
  fontSize: "14px",
  fontWeight: "bold",
  color: "#333333",
};

const inputStyle: React.CSSProperties = {
  padding: "12px",
  fontSize: "16px",
  border: "1px solid #fed7aa",
  borderRadius: "8px",
  outline: "none",
  color: "#333333",
  backgroundColor: "#ffffff",
};

const errorStyle: React.CSSProperties = {
  color: "#dc2626",
  fontSize: "13px",
  fontWeight: "normal",
};