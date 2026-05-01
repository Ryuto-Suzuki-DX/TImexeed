"use client";

import { useState } from "react";

export default function Home() {
  const [message, setMessage] = useState("");

  const checkBackend = async () => {
    const baseUrl = "http://127.0.0.1:8080";

    const response = await fetch(`${baseUrl}/health`);
    const data = await response.json();

    setMessage(data.message);
  };

  const checkDatabase = async () => {
    const baseUrl = "http://127.0.0.1:8080";

    const response = await fetch(`${baseUrl}/db-health`);
    const data = await response.json();

    setMessage(data.message);
  };

  return (
    <main style={{ minHeight: "100vh", padding: "40px", fontFamily: "sans-serif" }}>
      <h1 style={{ fontSize: "32px", marginBottom: "24px" }}>Timexeed</h1>

      <div style={{ display: "flex", gap: "12px", marginBottom: "24px" }}>
        <button onClick={checkBackend} style={{ padding: "12px 20px", cursor: "pointer" }}>
          Backend確認
        </button>

        <button onClick={checkDatabase} style={{ padding: "12px 20px", cursor: "pointer" }}>
          DB確認
        </button>
      </div>

      <p style={{ fontSize: "20px" }}>{message}</p>
    </main>
  );
}