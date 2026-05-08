"use client";

import type { ReactNode } from "react";

type PageContainerProps = {
  children: ReactNode;
};

export default function PageContainer({ children }: PageContainerProps) {
  return (
    <main style={{ minHeight: "100vh", padding: "40px", fontFamily: "sans-serif", backgroundColor: "#fff7ed", boxSizing: "border-box" }}>
      {children}
    </main>
  );
}