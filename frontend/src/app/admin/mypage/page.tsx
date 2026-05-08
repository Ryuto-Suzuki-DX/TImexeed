"use client";

import { useRouter } from "next/navigation";
import { removeAccessToken } from "@/api/auth";
import Button from "@/components/atoms/Button";
import { useRequireRole } from "@/hooks/useRequireRole";
import AdminSideMenu from "@/components/sideMenu/AdminSideMenu";
import styles from "./page.module.css";

export default function AdminMyPage() {
  const router = useRouter();

  const { user, isLoading, message } = useRequireRole("ADMIN");

  const handleLogout = () => {
    removeAccessToken();
    router.push("/login");
  };

  return (
    <main className={styles.page}>
      <AdminSideMenu />

      <section className={styles.card}>
        <div className={styles.header}>
          <div>
            <h1 className={styles.title}>管理者マイページ</h1>
            <p className={styles.description}>ログイン中の管理者情報を表示しています。</p>
          </div>

          <Button type="button" variant="secondary" onClick={handleLogout}>
            ログアウト
          </Button>
        </div>

        {isLoading && <p className={styles.loadingText}>{message}</p>}

        {!isLoading && user && (
          <div className={styles.infoList}>
            <div className={styles.infoBox}>
              <p className={styles.infoLabel}>名前</p>
              <p className={styles.infoValue}>{user.name}</p>
            </div>

            <div className={styles.infoBox}>
              <p className={styles.infoLabel}>ロール</p>
              <p className={styles.infoValue}>{user.role}</p>
            </div>

            <div className={styles.infoBox}>
              <p className={styles.infoLabel}>メールアドレス</p>
              <p className={styles.infoValue}>{user.email}</p>
            </div>
          </div>
        )}
      </section>
    </main>
  );
}