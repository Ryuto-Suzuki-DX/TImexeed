/*
 * 共通ボタン部品
 * ログイン・登録・保存など、画面内の主要ボタンとして使う
 */
type ButtonProps = {
  // ボタンの中に表示する文字や要素
  children: React.ReactNode;

  // button / submit / reset を指定できる
  type?: "button" | "submit" | "reset";

  // true の場合、押せない状態にする
  disabled?: boolean;

  // クリック時の処理
  onClick?: () => void;
};

export default function Button({
  children,
  type = "button",
  disabled = false,
  onClick,
}: ButtonProps) {
  return (
    <button
      type={type}
      disabled={disabled}
      onClick={onClick}
      style={{
        width: "100%",
        padding: "12px 16px",
        border: "none",
        borderRadius: "8px",
        backgroundColor: disabled ? "#f3c19c" : "#f97316",
        color: "#ffffff",
        fontSize: "16px",
        fontWeight: "bold",
        cursor: disabled ? "not-allowed" : "pointer",
        transition: "0.2s",
      }}
    >
      {children}
    </button>
  );
}