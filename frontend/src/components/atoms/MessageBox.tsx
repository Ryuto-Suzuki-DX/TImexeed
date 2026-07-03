"use client";

import type { ReactNode } from "react";

type MessageBoxVariant = "info" | "success" | "warning" | "error";

type MessageBoxProps = {
  children: ReactNode;
  variant?: MessageBoxVariant;
  className?: string;
};

export default function MessageBox({
  children,
  variant = "info",
  className,
}: MessageBoxProps) {
  const messageBoxClassName = [
    "timexeed-message-box",
    `timexeed-message-box-${variant}`,
    className ?? "",
  ]
    .filter(Boolean)
    .join(" ");

  return (
    <>
      <div className={messageBoxClassName}>{children}</div>

      <style jsx global>{`
        .timexeed-message-box {
          width: 100%;
          min-width: 0;
          padding: 12px 14px;
          border-radius: 12px;
          box-sizing: border-box;
          font-size: 14px;
          font-weight: bold;
          line-height: 1.6;
          overflow-wrap: anywhere;
        }

        .timexeed-message-box-info {
          background-color: #eff6ff;
          border: 1px solid #bfdbfe;
          color: #1d4ed8;
        }

        .timexeed-message-box-success {
          background-color: #f0fdf4;
          border: 1px solid #bbf7d0;
          color: #15803d;
        }

        .timexeed-message-box-warning {
          background-color: #fffbeb;
          border: 1px solid #fde68a;
          color: #b45309;
        }

        .timexeed-message-box-error {
          background-color: #fef2f2;
          border: 1px solid #fecaca;
          color: #b91c1c;
        }

        @media (max-width: 768px) {
          .timexeed-message-box {
            padding: 7px 8px;
            border-radius: 8px;
            font-size: 10px;
            line-height: 1.4;
          }
        }

        @media (max-width: 480px) {
          .timexeed-message-box {
            padding: 6px 7px;
            border-radius: 7px;
            font-size: 9px;
            line-height: 1.35;
          }
        }
      `}</style>
    </>
  );
}
