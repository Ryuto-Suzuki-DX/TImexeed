"use client";

import type { ButtonHTMLAttributes, ReactNode } from "react";

type ButtonVariant = "primary" | "secondary" | "danger";

type ButtonProps = ButtonHTMLAttributes<HTMLButtonElement> & {
  children: ReactNode;
  variant?: ButtonVariant;
  fullWidth?: boolean;
};

export default function Button({
  children,
  variant = "primary",
  fullWidth = false,
  disabled = false,
  type = "button",
  className,
  ...props
}: ButtonProps) {
  const buttonClassName = [
    "timexeed-button",
    `timexeed-button-${variant}`,
    fullWidth ? "timexeed-button-full-width" : "",
    className ?? "",
  ]
    .filter(Boolean)
    .join(" ");

  return (
    <>
      <button
        type={type}
        disabled={disabled}
        className={buttonClassName}
        {...props}
      >
        {children}
      </button>

      <style>{`
        .timexeed-button {
          display: inline-flex;
          align-items: center;
          justify-content: center;
          width: auto;
          min-width: 0;
          min-height: 40px;
          padding: 10px 16px;
          border: 1px solid transparent;
          border-radius: 10px;
          box-sizing: border-box;
          font-family: inherit;
          font-size: 14px;
          font-weight: bold;
          line-height: 1;
          text-align: center;
          white-space: nowrap;
          cursor: pointer;
          transition:
            background-color 0.15s ease,
            color 0.15s ease,
            border-color 0.15s ease,
            opacity 0.15s ease;
        }

        .timexeed-button:disabled {
          opacity: 0.55;
          cursor: not-allowed;
        }

        .timexeed-button-primary {
          background-color: #ea580c;
          color: #ffffff;
          border-color: #ea580c;
        }

        .timexeed-button-secondary {
          background-color: #ffffff;
          color: #ea580c;
          border-color: #fdba74;
        }

        .timexeed-button-danger {
          background-color: #dc2626;
          color: #ffffff;
          border-color: #dc2626;
        }

        .timexeed-button-full-width {
          width: 100%;
        }

        @media (max-width: 768px) {
          .timexeed-button {
            min-height: 28px;
            padding: 3px 6px;
            border-radius: 7px;
            font-size: 9px;
          }
        }

        @media (max-width: 480px) {
          .timexeed-button {
            min-height: 24px;
            padding: 2px 5px;
            border-radius: 6px;
            font-size: 8px;
          }
        }
      `}</style>
    </>
  );
}
