type ConfirmModalProps = {
  title: string;
  message: string;
  children?: React.ReactNode;
  confirmText?: string;
  cancelText?: string;
  onConfirm: () => void;
  onCancel: () => void;
};

export default function ConfirmModal({
  title,
  message,
  children,
  confirmText = "実行する",
  cancelText = "キャンセル",
  onConfirm,
  onCancel,
}: ConfirmModalProps) {
  return (
    <div style={modalOverlayStyle}>
      <div style={modalContentStyle}>
        <h2 style={{ margin: "0 0 16px", fontSize: "22px", color: "#ea580c" }}>{title}</h2>

        <p style={{ margin: "0 0 8px", fontSize: "16px", color: "#333333" }}>{message}</p>

        {children && <div style={contentBoxStyle}>{children}</div>}

        <div style={{ display: "flex", justifyContent: "flex-end", gap: "12px", marginTop: "24px" }}>
          <button type="button" onClick={onCancel} style={cancelButtonStyle}>
            {cancelText}
          </button>

          <button type="button" onClick={onConfirm} style={confirmButtonStyle}>
            {confirmText}
          </button>
        </div>
      </div>
    </div>
  );
}

const modalOverlayStyle: React.CSSProperties = {
  position: "fixed",
  top: 0,
  left: 0,
  width: "100vw",
  height: "100vh",
  backgroundColor: "rgba(0, 0, 0, 0.45)",
  display: "flex",
  alignItems: "center",
  justifyContent: "center",
  zIndex: 2000,
};

const modalContentStyle: React.CSSProperties = {
  width: "420px",
  padding: "28px",
  borderRadius: "16px",
  backgroundColor: "#ffffff",
  boxShadow: "0 12px 32px rgba(0, 0, 0, 0.2)",
};

const contentBoxStyle: React.CSSProperties = {
  marginTop: "16px",
  padding: "16px",
  borderRadius: "12px",
  border: "1px solid #fed7aa",
  backgroundColor: "#fff7ed",
};

const cancelButtonStyle: React.CSSProperties = {
  padding: "10px 16px",
  border: "1px solid #fed7aa",
  borderRadius: "8px",
  backgroundColor: "#ffffff",
  color: "#333333",
  fontSize: "14px",
  fontWeight: "bold",
  cursor: "pointer",
};

const confirmButtonStyle: React.CSSProperties = {
  padding: "10px 16px",
  border: "none",
  borderRadius: "8px",
  backgroundColor: "#dc2626",
  color: "#ffffff",
  fontSize: "14px",
  fontWeight: "bold",
  cursor: "pointer",
};