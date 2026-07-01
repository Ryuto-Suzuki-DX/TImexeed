"use client";

import type { ReactNode } from "react";
import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { useState } from "react";
import { removeAccessToken } from "@/api/auth";

export type SideMenuItem = {
  label: ReactNode;
  href: string;
};

type BaseSideMenuProps = {
  title: string;
  items: SideMenuItem[];
};

export default function BaseSideMenu({ title, items }: BaseSideMenuProps) {
  const router = useRouter();
  const pathname = usePathname();

  const [isOpen, setIsOpen] = useState(false);

  const handleLogout = () => {
    removeAccessToken();
    router.push("/login");
  };

  return (
    <aside
      onMouseEnter={() => setIsOpen(true)}
      onMouseLeave={() => setIsOpen(false)}
      style={{
        position: "fixed",
        top: 0,
        left: 0,
        width: isOpen ? "240px" : "16px",
        height: "100dvh",
        backgroundColor: isOpen ? "#ffffff" : "#ea580c",
        borderRight: isOpen ? "1px solid #fed7aa" : "none",
        boxShadow: isOpen ? "4px 0 20px rgba(0, 0, 0, 0.08)" : "none",
        transition:
          "width 0.2s ease, background-color 0.2s ease, box-shadow 0.2s ease",
        zIndex: 1000,
        overflow: "hidden",
      }}
    >
      <div
        style={{
          width: "240px",
          height: "100%",
          display: "flex",
          flexDirection: "column",
          padding: "16px 12px",
          boxSizing: "border-box",
          minHeight: 0,
        }}
      >
        <div
          style={{
            marginBottom: "16px",
            flexShrink: 0,
          }}
        >
          <p
            style={{
              margin: "0 0 2px",
              fontSize: "11px",
              fontWeight: "bold",
              color: "#9a3412",
            }}
          >
            Timexeed
          </p>

          <h2
            style={{
              margin: 0,
              fontSize: "18px",
              color: "#ea580c",
            }}
          >
            {title}
          </h2>
        </div>

        <nav
          style={{
            display: "flex",
            flexDirection: "column",
            gap: "4px",
            flex: 1,
            minHeight: 0,
            overflowY: "auto",
            overflowX: "hidden",
            paddingRight: "4px",
            paddingBottom: "8px",
          }}
        >
          {items.map((item) => {
            const isActive =
              pathname === item.href ||
              pathname.startsWith(`${item.href}/`);

            return (
              <Link
                key={item.href}
                href={item.href}
                style={{
                  display: "block",
                  padding: "8px 10px",
                  borderRadius: "8px",
                  textDecoration: "none",
                  fontSize: "13px",
                  lineHeight: 1.4,
                  fontWeight: isActive ? "bold" : "normal",
                  color: isActive ? "#ffffff" : "#374151",
                  backgroundColor: isActive ? "#ea580c" : "transparent",
                  border: isActive
                    ? "1px solid #ea580c"
                    : "1px solid transparent",
                  whiteSpace: "normal",
                  wordBreak: "break-word",
                  flexShrink: 0,
                }}
              >
                {item.label}
              </Link>
            );
          })}
        </nav>

        <div
          style={{
            paddingTop: "10px",
            flexShrink: 0,
          }}
        >
          <button
            type="button"
            onClick={handleLogout}
            style={{
              width: "100%",
              padding: "9px 10px",
              borderRadius: "8px",
              border: "1px solid #fed7aa",
              backgroundColor: "#fff7ed",
              color: "#9a3412",
              fontSize: "13px",
              fontWeight: "bold",
              cursor: "pointer",
            }}
          >
            ログアウト
          </button>
        </div>
      </div>
    </aside>
  );
}
