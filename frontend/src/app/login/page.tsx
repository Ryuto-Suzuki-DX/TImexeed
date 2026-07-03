"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { login, removeAccessToken, saveAccessToken } from "@/api/auth";
import Button from "@/components/atoms/Button";
import styles from "./page.module.css";

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
    <main className={styles.page}>
      <section className={styles.loginCard}>
        <h1 className={styles.title}>ログイン</h1>

        <div className={styles.form}>
          <label className={styles.fieldLabel}>
            <span className={styles.labelText}>メールアドレス</span>

            <input
              type="email"
              value={email}
              onChange={(event) => setEmail(event.target.value)}
              disabled={isLoading}
              className={styles.input}
              autoComplete="email"
            />
          </label>

          <label className={styles.fieldLabel}>
            <span className={styles.labelText}>パスワード</span>

            <input
              type="password"
              value={password}
              onChange={(event) => setPassword(event.target.value)}
              disabled={isLoading}
              className={styles.input}
              autoComplete="current-password"
            />
          </label>

          <div className={styles.actionArea}>
            <Button type="button" fullWidth disabled={isLoading} onClick={handleLogin}>
              {isLoading ? "ログイン中..." : "ログイン"}
            </Button>
          </div>

          {message && <p className={styles.errorMessage}>{message}</p>}
        </div>
      </section>
    </main>
  );
}
