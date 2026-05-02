"use client";

import UserSideMenu from "@/components/user/navigation/UserSideMenu";

export default function UserAttendancePage() {
  return (
    <div style={{ minHeight: "100vh", backgroundColor: "#f5f6f8" }}>
      <UserSideMenu />

      <main style={{ padding: "24px" }}>
        <h1 style={{ margin: 0, fontSize: "24px", color: "#333333" }}>
          勤怠登録
        </h1>

        <p style={{ marginTop: "12px", color: "#666666" }}>
          勤怠ページ作成中です。
        </p>
      </main>
    </div>
  );
}