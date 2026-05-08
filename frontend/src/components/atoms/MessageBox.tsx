"use client";

import type { ReactNode } from "react";

type MessageBoxVariant = "info" | "success" | "warning" | "error";

type MessageBoxProps = {
  children: ReactNode;
  variant?: MessageBoxVariant;
};

export default function MessageBox({ children, variant = "info" }: MessageBoxProps) {
  const variantStyle: Record<MessageBoxVariant, React.CSSProperties> = {
    info: { backgroundColor: "#eff6ff", border: "1px solid #bfdbfe", color: "#1d4ed8" },
    success: { backgroundColor: "#f0fdf4", border: "1px solid #bbf7d0", color: "#15803d" },
    warning: { backgroundColor: "#fffbeb", border: "1px solid #fde68a", color: "#b45309" },
    error: { backgroundColor: "#fef2f2", border: "1px solid #fecaca", color: "#b91c1c" },
  };

  return <div style={{ padding: "12px 14px", borderRadius: "12px", fontSize: "14px", fontWeight: "bold", lineHeight: 1.6, ...variantStyle[variant] }}>{children}</div>;
}