"use client";

import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { useState } from "react";
import { removeAccessToken } from "@/api/auth";

export type SideMenuItem = {
  label: string;
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
    <aside onMouseEnter={() => setIsOpen(true)} onMouseLeave={() => setIsOpen(false)} style={{ position: "fixed", top: 0, left: 0, width: isOpen ? "240px" : "16px", height: "100vh", backgroundColor: isOpen ? "#ffffff" : "#ea580c", borderRight: isOpen ? "1px solid #fed7aa" : "none", boxShadow: isOpen ? "4px 0 20px rgba(0, 0, 0, 0.08)" : "none", transition: "width 0.2s ease, background-color 0.2s ease, box-shadow 0.2s ease", zIndex: 1000, overflow: "hidden" }}>
      <div style={{ width: "240px", height: "100%", display: "flex", flexDirection: "column", padding: "24px 16px", boxSizing: "border-box" }}>
        <div style={{ marginBottom: "32px" }}>
          <p style={{ margin: "0 0 4px", fontSize: "12px", fontWeight: "bold", color: "#9a3412" }}>Timexeed</p>
          <h2 style={{ margin: 0, fontSize: "20px", color: "#ea580c" }}>{title}</h2>
        </div>

        <nav style={{ display: "flex", flexDirection: "column", gap: "8px", flex: 1 }}>
          {items.map((item) => {
            const isActive = pathname === item.href;

            return (
              <Link key={item.href} href={item.href} style={{ display: "block", padding: "12px 14px", borderRadius: "10px", textDecoration: "none", fontSize: "14px", fontWeight: isActive ? "bold" : "normal", color: isActive ? "#ffffff" : "#374151", backgroundColor: isActive ? "#ea580c" : "transparent", border: isActive ? "1px solid #ea580c" : "1px solid transparent" }}>
                {item.label}
              </Link>
            );
          })}
        </nav>

        <button type="button" onClick={handleLogout} style={{ width: "100%", padding: "12px 14px", borderRadius: "10px", border: "1px solid #fed7aa", backgroundColor: "#fff7ed", color: "#9a3412", fontSize: "14px", fontWeight: "bold", cursor: "pointer" }}>
          ログアウト
        </button>
      </div>
    </aside>
  );
}