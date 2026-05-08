import type { ButtonHTMLAttributes, ReactNode } from "react";

type ButtonVariant = "primary" | "secondary" | "danger";

type ButtonProps = ButtonHTMLAttributes<HTMLButtonElement> & {
  children: ReactNode;
  variant?: ButtonVariant;
  fullWidth?: boolean;
};

/*
 * 汎用ボタン
 *
 * primary:
 *   メイン操作用
 *   例：ログイン、保存、検索、作成
 *
 * secondary:
 *   サブ操作用
 *   例：戻る、キャンセル、詳細
 *
 * danger:"use client";

import type { ButtonHTMLAttributes, ReactNode } from "react";

type ButtonVariant = "primary" | "secondary" | "danger";

type ButtonProps = ButtonHTMLAttributes<HTMLButtonElement> & {
  children: ReactNode;
  variant?: ButtonVariant;
};

export default function Button({ children, variant = "primary", disabled = false, style, ...props }: ButtonProps) {
  const variantStyle: Record<ButtonVariant, React.CSSProperties> = {
    primary: { backgroundColor: "#ea580c", color: "#ffffff", border: "1px solid #ea580c" },
    secondary: { backgroundColor: "#ffffff", color: "#ea580c", border: "1px solid #fed7aa" },
    danger: { backgroundColor: "#dc2626", color: "#ffffff", border: "1px solid #dc2626" },
  };

  return (
    <button
      disabled={disabled}
      style={{ padding: "10px 16px", borderRadius: "10px", fontSize: "14px", fontWeight: "bold", cursor: disabled ? "not-allowed" : "pointer", opacity: disabled ? 0.6 : 1, transition: "background-color 0.15s ease, opacity 0.15s ease", ...variantStyle[variant], ...style }}
      {...props}
    >
      {children}
    </button>
  );
}
 *   危険操作用
 *   例：削除
 *
 * disabled:
 *   無効化状態
 *   クリック不可、薄い表示、カーソル変更
 */
export default function Button({ children, variant = "primary", fullWidth = false, disabled = false, style, type = "button", ...props }: ButtonProps) {
  const baseStyle: React.CSSProperties = {
    width: fullWidth ? "100%" : "auto",
    padding: "12px 20px",
    borderRadius: "10px",
    border: "1px solid transparent",
    fontSize: "15px",
    fontWeight: "bold",
    lineHeight: "1",
    cursor: disabled ? "not-allowed" : "pointer",
    opacity: disabled ? 0.55 : 1,
    transition: "background-color 0.15s ease, color 0.15s ease, border-color 0.15s ease, opacity 0.15s ease",
  };

  const variantStyle: Record<ButtonVariant, React.CSSProperties> = {
    primary: {
      backgroundColor: "#ea580c",
      color: "#ffffff",
      borderColor: "#ea580c",
    },
    secondary: {
      backgroundColor: "#ffffff",
      color: "#ea580c",
      borderColor: "#fdba74",
    },
    danger: {
      backgroundColor: "#dc2626",
      color: "#ffffff",
      borderColor: "#dc2626",
    },
  };

  return (
    <button type={type} disabled={disabled} style={{ ...baseStyle, ...variantStyle[variant], ...style }} {...props}>
      {children}
    </button>
  );
}