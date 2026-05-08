"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { login, removeAccessToken, saveAccessToken } from "@/api/auth";
import Button from "@/components/atoms/Button";

export default function LoginPage() {
  const router = useRouter();

  const [email, setEmail] = useState("test@example.com");
  const [password, setPassword] = useState("password123");
  const [message, setMessage] = useState("");
  const [isLoading, setIsLoading] = useState(false);

  /*
   * ログイン処理
   *
   * 1. メールアドレスとパスワードをAPIへ送信する
   * 2. ログイン成功時はaccessTokenを保存する
   * 3. roleを見て遷移先を分ける
   *
   * ADMIN → 管理者マイページ
   * USER  → 従業員マイページ
   *
   * 想定外のroleの場合は、tokenを削除してログイン失敗扱いにする
   */
  const handleLogin = async () => {
    if (isLoading) {
      return;
    }

    setMessage("");
    setIsLoading(true);

    try {
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

      if (result.data.user.role === "USER") {
        router.push("/user/mypage");
        return;
      }

      removeAccessToken();
      setMessage("利用できない権限です。");
    } catch {
      setMessage("ログイン処理中にエラーが発生しました。");
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <main style={{ minHeight: "100vh", padding: "40px", fontFamily: "sans-serif", backgroundColor: "#fff7ed" }}>
      <section style={{ width: "360px", margin: "80px auto", padding: "32px", borderRadius: "16px", backgroundColor: "#ffffff", boxShadow: "0 8px 24px rgba(0, 0, 0, 0.08)" }}>
        <h1 style={{ fontSize: "32px", marginBottom: "24px", color: "#ea580c", textAlign: "center" }}>ログイン</h1>

        <div style={{ display: "flex", flexDirection: "column", gap: "16px" }}>
          <label style={{ display: "flex", flexDirection: "column", gap: "8px", fontSize: "14px", fontWeight: "bold", color: "#333333" }}>
            メールアドレス
            <input
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              disabled={isLoading}
              style={{ padding: "12px", fontSize: "16px", border: "1px solid #fed7aa", borderRadius: "8px", outline: "none", backgroundColor: isLoading ? "#f5f5f5" : "#ffffff" }}
            />
          </label>

          <label style={{ display: "flex", flexDirection: "column", gap: "8px", fontSize: "14px", fontWeight: "bold", color: "#333333" }}>
            パスワード
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              disabled={isLoading}
              style={{ padding: "12px", fontSize: "16px", border: "1px solid #fed7aa", borderRadius: "8px", outline: "none", backgroundColor: isLoading ? "#f5f5f5" : "#ffffff" }}
            />
          </label>

          <Button type="button" fullWidth disabled={isLoading} onClick={handleLogin}>
            {isLoading ? "ログイン中..." : "ログイン"}
          </Button>

          {message && <p style={{ color: "#dc2626", fontSize: "14px", margin: "0", textAlign: "center" }}>{message}</p>}
        </div>
      </section>
    </main>
  );
}