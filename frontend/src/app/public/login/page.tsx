"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { login } from "@/api/auth";
import { saveAccessToken } from "@/lib/auth";
import Button from "@/components/atoms/Button";

export default function LoginPage() {
  const router = useRouter();

  const [email, setEmail] = useState("test@example.com");
  const [password, setPassword] = useState("password123");
  const [message, setMessage] = useState("");

  /*
   * ログイン処理
   * 入力されたメールアドレスとパスワードをAPIに送信する
   */
  const handleLogin = async () => {
    setMessage("");

    const result = await login(email, password);

    if (result.error || !result.data) {
      setMessage(result.message);
      return;
    }

    saveAccessToken(result.data.accessToken);

    if (result.data.user.role === "ADMIN") {
      router.push("/admin/mypage");
      return;
    }

    router.push("/user/mypage");
  };

  return (
    <main style={{ minHeight: "100vh", padding: "40px", fontFamily: "sans-serif", backgroundColor: "#fff7ed" }}>
      <section style={{ width: "360px", margin: "80px auto", padding: "32px", borderRadius: "16px", backgroundColor: "#ffffff", boxShadow: "0 8px 24px rgba(0, 0, 0, 0.08)" }}>
        <h1 style={{ fontSize: "32px", marginBottom: "24px", color: "#ea580c", textAlign: "center" }}>ログイン</h1>

        <div style={{ display: "flex", flexDirection: "column", gap: "16px" }}>
          <label style={{ display: "flex", flexDirection: "column", gap: "8px", fontSize: "14px", fontWeight: "bold", color: "#333333" }}>
            メールアドレス
            <input value={email} onChange={(e) => setEmail(e.target.value)} style={{ padding: "12px", fontSize: "16px", border: "1px solid #fed7aa", borderRadius: "8px", outline: "none" }} />
          </label>

          <label style={{ display: "flex", flexDirection: "column", gap: "8px", fontSize: "14px", fontWeight: "bold", color: "#333333" }}>
            パスワード
            <input type="password" value={password} onChange={(e) => setPassword(e.target.value)} style={{ padding: "12px", fontSize: "16px", border: "1px solid #fed7aa", borderRadius: "8px", outline: "none" }} />
          </label>

          <Button type="button" onClick={handleLogin}>
            ログイン
          </Button>

          {message && <p style={{ color: "#dc2626", fontSize: "14px", margin: "0", textAlign: "center" }}>{message}</p>}
        </div>
      </section>
    </main>
  );
}