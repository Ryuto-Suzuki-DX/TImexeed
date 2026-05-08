"use client";

import type { InputHTMLAttributes } from "react";

type InputProps = InputHTMLAttributes<HTMLInputElement> & {
  label?: string;
  errorMessage?: string;
};

export default function Input({ label, errorMessage, style, ...props }: InputProps) {
  return (
    <label style={{ display: "flex", flexDirection: "column", gap: "6px", width: "100%" }}>
      {label && <span style={{ fontSize: "13px", fontWeight: "bold", color: "#374151" }}>{label}</span>}

      <input
        style={{ width: "100%", padding: "10px 12px", borderRadius: "10px", border: errorMessage ? "1px solid #dc2626" : "1px solid #d1d5db", fontSize: "14px", color: "#111827", backgroundColor: "#ffffff", boxSizing: "border-box", outline: "none", ...style }}
        {...props}
      />

      {errorMessage && <span style={{ fontSize: "12px", color: "#dc2626" }}>{errorMessage}</span>}
    </label>
  );
}