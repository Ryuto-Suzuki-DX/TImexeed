"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";

/*
 * 管理者用サイドメニュー
 * 画面左端にカーソルを当てるとメニューを表示する
 */
export default function AdminSideMenu() {
  const router = useRouter();
  const [isOpen, setIsOpen] = useState(false);

  const menuItems = [
    { label: "マイページ", path: "/admin/mypage" },
    { label: "ユーザー管理", path: "/admin/users" },
    { label: "所属管理", path: "/admin/departments"},
    { label: "勤怠管理", path: "/admin/attendance" },
    { label: "給与管理", path: "/admin/salary" },
    { label: "Google Drive生成", path: "/admin/drive-files" },
    { label: "設定", path: "/admin/settings" },
  ];

  const handleMove = (path: string) => {
    router.push(path);
  };

  return (
    <>
      {/* 左端の反応エリア */}
      <div
        onMouseEnter={() => setIsOpen(true)}
        style={{ position: "fixed", top: 0, left: 0, width: "16px", height: "100vh", zIndex: 1000 }}
      />

      {/* サイドメニュー本体 */}
      <aside
        onMouseEnter={() => setIsOpen(true)}
        onMouseLeave={() => setIsOpen(false)}
        style={{
          position: "fixed",
          top: 0,
          left: isOpen ? 0 : "-240px",
          width: "240px",
          height: "100vh",
          backgroundColor: "#ffffff",
          borderRight: "1px solid #fed7aa",
          boxShadow: "4px 0 16px rgba(0, 0, 0, 0.08)",
          transition: "left 0.2s ease",
          zIndex: 1001,
          padding: "24px 16px",
          boxSizing: "border-box",
        }}
      >
        <h2 style={{ margin: "0 0 24px", fontSize: "20px", color: "#ea580c" }}>
          管理者メニュー
        </h2>

        <nav style={{ display: "flex", flexDirection: "column", gap: "12px" }}>
          {menuItems.map((item) => (
            <button
              key={item.path}
              type="button"
              onClick={() => handleMove(item.path)}
              style={{
                width: "100%",
                padding: "12px",
                border: "none",
                borderRadius: "8px",
                backgroundColor: "#fff7ed",
                color: "#333333",
                fontSize: "15px",
                fontWeight: "bold",
                textAlign: "left",
                cursor: "pointer",
              }}
            >
              {item.label}
            </button>
          ))}
        </nav>
      </aside>
    </>
  );
}