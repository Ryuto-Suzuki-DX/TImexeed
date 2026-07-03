"use client";

import type { InputHTMLAttributes } from "react";

type InputProps = Omit<
  InputHTMLAttributes<HTMLInputElement>,
  "style" | "aria-invalid"
> & {
  label?: string;
  errorMessage?: string;
};

export default function Input({
  label,
  errorMessage,
  className,
  disabled,
  ...props
}: InputProps) {
  const inputClassName = [
    "timexeed-input",
    errorMessage ? "timexeed-input-error" : "",
    className ?? "",
  ]
    .filter(Boolean)
    .join(" ");

  return (
    <>
      <label className="timexeed-input-field">
        {label && <span className="timexeed-input-label">{label}</span>}

        <input
          className={inputClassName}
          disabled={disabled}
          {...props}
        />

        {errorMessage && (
          <span className="timexeed-input-error-message" role="alert">
            {errorMessage}
          </span>
        )}
      </label>

      <style jsx global>{`
        .timexeed-input-field {
          display: flex;
          flex-direction: column;
          gap: 6px;
          width: 100%;
          min-width: 0;
        }

        .timexeed-input-label {
          color: #374151;
          font-size: 13px;
          font-weight: bold;
          line-height: 1.4;
        }

        .timexeed-input {
          width: 100%;
          min-width: 0;
          min-height: 40px;
          padding: 8px 10px;
          border: 1px solid #d1d5db;
          border-radius: 10px;
          outline: none;
          background-color: #ffffff;
          color: #111827;
          box-sizing: border-box;
          font-family: inherit;
          font-size: 14px;
          line-height: 1.4;
          transition:
            border-color 0.15s ease,
            box-shadow 0.15s ease,
            background-color 0.15s ease;
        }

        .timexeed-input:focus {
          border-color: #fb923c;
          box-shadow: 0 0 0 3px rgba(251, 146, 60, 0.18);
        }

        .timexeed-input:disabled {
          background-color: #f3f4f6;
          color: #9ca3af;
          cursor: not-allowed;
        }

        .timexeed-input-error {
          border-color: #dc2626;
        }

        .timexeed-input-error:focus {
          border-color: #dc2626;
          box-shadow: 0 0 0 3px rgba(220, 38, 38, 0.14);
        }

        .timexeed-input-error-message {
          color: #dc2626;
          font-size: 12px;
          line-height: 1.4;
        }

        @media (max-width: 768px) {
          .timexeed-input-field {
            gap: 4px;
          }

          .timexeed-input-label {
            font-size: 10px;
          }

          .timexeed-input {
            min-height: 30px;
            padding: 4px 6px;
            border-radius: 7px;
            font-size: 10px;
          }

          .timexeed-input-error-message {
            font-size: 9px;
          }
        }

        @media (max-width: 480px) {
          .timexeed-input-field {
            gap: 3px;
          }

          .timexeed-input-label {
            font-size: 9px;
          }

          .timexeed-input {
            min-height: 28px;
            padding: 3px 5px;
            border-radius: 6px;
            font-size: 9px;
          }

          .timexeed-input-error-message {
            font-size: 8px;
          }
        }
      `}</style>
    </>
  );
}
