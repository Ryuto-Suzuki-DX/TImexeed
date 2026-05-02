"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import AdminSideMenu from "@/components/admin/navigation/AdminSideMenu";
import Button from "@/components/atoms/Button";
import { getDepartment, updateDepartment } from "@/api/admin/departments";

type ValidationDetails = {
  name?: string;
  request?: string;
};

export default function AdminDepartmentUpdatePage() {
  const router = useRouter();
  const params = useParams();

  const departmentIdText = Array.isArray(params.id) ? params.id[0] : params.id;
  const departmentId = departmentIdText ? Number(departmentIdText) : 0;

  const [name, setName] = useState("");
  const [message, setMessage] = useState("");
  const [details, setDetails] = useState<ValidationDetails>({});

  /*
   * 初回表示時に編集対象所属を取得する
   */
  useEffect(() => {
    if (!departmentId) {
      return;
    }

    getDepartment(departmentId).then((result) => {
      if (result.error || !result.data) {
        setMessage(result.message);
        return;
      }

      setName(result.data.name);
    });
  }, [departmentId]);

  /*
   * 所属更新
   */
  const handleUpdate = async () => {
    setMessage("");
    setDetails({});

    const result = await updateDepartment({
      id: departmentId,
      name,
    });

    if (result.error || !result.data) {
      setMessage(result.message);

      const resultDetails = getResultDetails(result);
      if (resultDetails && typeof resultDetails === "object") {
        setDetails(resultDetails as ValidationDetails);
      }

      return;
    }

    router.push("/admin/departments");
  };

  /*
   * 一覧画面へ戻る
   */
  const handleBack = () => {
    router.push("/admin/departments");
  };

  return (
    <>
      <AdminSideMenu />

      <main style={{ minHeight: "100vh", padding: "40px", fontFamily: "sans-serif", backgroundColor: "#fff7ed", color: "#333333" }}>
        <section style={{ maxWidth: "720px", margin: "40px auto", padding: "32px", borderRadius: "16px", backgroundColor: "#ffffff", boxShadow: "0 8px 24px rgba(0, 0, 0, 0.08)" }}>
          <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: "24px" }}>
            <h1 style={{ margin: 0, fontSize: "32px", color: "#ea580c" }}>
              所属編集
            </h1>

            <div style={{ width: "120px" }}>
              <Button type="button" onClick={handleBack}>
                戻る
              </Button>
            </div>
          </div>

          {!departmentId && (
            <p style={{ margin: "0 0 16px", color: "#dc2626", fontSize: "14px" }}>
              所属IDが正しくありません
            </p>
          )}

          {message && (
            <p style={{ margin: "0 0 16px", color: "#dc2626", fontSize: "14px" }}>
              {message}
            </p>
          )}

          <div style={{ display: "flex", flexDirection: "column", gap: "18px" }}>
            <label style={labelStyle}>
              所属名
              <input
                value={name}
                onChange={(e) => setName(e.target.value)}
                style={inputStyle}
              />
              {details.name && <span style={errorStyle}>{details.name}</span>}
            </label>

            {details.request && (
              <p style={{ margin: 0, color: "#dc2626", fontSize: "14px" }}>
                {details.request}
              </p>
            )}

            <div style={{ width: "180px", marginTop: "8px" }}>
              <Button type="button" onClick={handleUpdate}>
                更新
              </Button>
            </div>
          </div>
        </section>
      </main>
    </>
  );
}

/*
 * Result内の詳細情報を取得する
 * backendのレスポンスが details / detail のどちらでも拾えるようにする
 */
function getResultDetails(result: unknown): unknown {
  if (typeof result !== "object" || result === null) {
    return undefined;
  }

  if ("details" in result) {
    return (result as { details?: unknown }).details;
  }

  if ("detail" in result) {
    return (result as { detail?: unknown }).detail;
  }

  return undefined;
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