"use client";

type PageTitleProps = {
  title: string;
  description?: string;
};

export default function PageTitle({ title, description }: PageTitleProps) {
  return (
    <div style={{ marginBottom: "24px" }}>
      <h1 style={{ margin: "0 0 8px", fontSize: "32px", color: "#ea580c" }}>{title}</h1>
      {description && <p style={{ margin: 0, fontSize: "14px", color: "#666666" }}>{description}</p>}
    </div>
  );
}