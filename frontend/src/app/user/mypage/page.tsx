"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { fetchMe } from "@/api/auth";
import { removeAccessToken } from "@/lib/auth";
import Button from "@/components/atoms/Button";
import UserSideMenu from "@/components/user/navigation/UserSideMenu";

type UserInfo = {
  userId: number;
  email: string;
  role: string;
};

export default function UserMyPage() {
  const router = useRouter();

  const [user, setUser] = useState<UserInfo | null>(null);
  const [message, setMessage] = useState("認証確認中...");

  /*
   * ログイン中ユーザーの情報を取得する
   * 未ログイン、または一般ユーザー以外の場合はログイン画面へ戻す
   */
  useEffect(() => {
    const loadMe = async () => {
      const result = await fetchMe();

      if (result.error || !result.data) {
        setMessage(result.message);
        router.push("/login");
        return;
      }

      if (result.data.role !== "USER") {
        setMessage("一般ユーザー権限がありません");
        router.push("/login");
        return;
      }

      setUser(result.data);
      setMessage("");
    };

    loadMe();
  }, [router]);

  /*
   * ログアウト処理
   * 保存しているアクセストークンを削除してログイン画面へ戻す
   */
  const handleLogout = () => {
    removeAccessToken();
    router.push("/login");
  };

  return (
    <>
      <UserSideMenu />

      <main style={{ minHeight: "100vh", padding: "40px", fontFamily: "sans-serif", backgroundColor: "#fff7ed" }}>
        <section style={{ maxWidth: "720px", margin: "40px auto", padding: "32px", borderRadius: "16px", backgroundColor: "#ffffff", boxShadow: "0 8px 24px rgba(0, 0, 0, 0.08)" }}>
          <h1 style={{ margin: "0 0 24px", fontSize: "32px", color: "#ea580c" }}>マイページ</h1>

          {message && <p style={{ color: "#555555", fontSize: "16px" }}>{message}</p>}

          {user && (
            <div style={{ display: "flex", flexDirection: "column", gap: "20px" }}>
              <div style={{ display: "flex", flexDirection: "column", gap: "12px", padding: "20px", borderRadius: "12px", backgroundColor: "#fff7ed", border: "1px solid #fed7aa" }}>
                <p style={{ margin: 0, fontSize: "16px" }}>
                  <strong>ユーザーID：</strong>
                  {user.userId}
                </p>

                <p style={{ margin: 0, fontSize: "16px" }}>
                  <strong>メール：</strong>
                  {user.email}
                </p>

                <p style={{ margin: 0, fontSize: "16px" }}>
                  <strong>権限：</strong>
                  {user.role}
                </p>
              </div>

              <div style={{ display: "flex", gap: "12px" }}>
                <div style={{ width: "160px" }}>
                  <Button type="button" onClick={() => router.push("/user/attendance")}>
                    勤怠登録へ
                  </Button>
                </div>

                <div style={{ width: "160px" }}>
                  <Button type="button" onClick={handleLogout}>
                    ログアウト
                  </Button>
                </div>
              </div>
            </div>
          )}
        </section>
      </main>
    </>
  );
}