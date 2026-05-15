"use client";

import { useRouter } from "next/navigation";
import Button from "@/components/atoms/Button";
import MessageBox from "@/components/atoms/MessageBox";
import PageContainer from "@/components/atoms/PageContainer";
import PageTitle from "@/components/atoms/PageTitle";
import AdminSideMenu from "@/components/sideMenu/AdminSideMenu";
import { useRequireRole } from "@/hooks/useRequireRole";
import styles from "./page.module.css";

type SettingMenuItem = {
  title: string;
  description: string;
  href: string;
  statusLabel: string;
};

const settingMenuItems: SettingMenuItem[] = [
  {
    title: "有給過去使用分管理",
    description: "システム導入前の有給使用日を、管理者が追加・更新・削除できます。",
    href: "/admin/settings/paid_leave_usages",
    statusLabel: "管理者向け",
  },
  {
    title: "祝日CSV管理",
    description: "国民の祝日CSVを取り込み、登録済み祝日を対象年月ごとに確認できます。",
    href: "/admin/settings/holiday_dates",
    statusLabel: "勤怠設定",
  },
  {
    title: "外部ストレージリンク管理",
    description: "Google Driveなどの外部ストレージにあるフォルダURLやファイルURLを管理できます。",
    href: "/admin/settings/external_storage_links",
    statusLabel: "共通設定",
  },
];

export default function AdminSettingsPage() {
  const router = useRouter();
  const { user, isLoading, message } = useRequireRole("ADMIN");

  if (isLoading || !user) {
    return (
      <PageContainer>
        <AdminSideMenu />

        <section className={styles.loadingCard}>
          <PageTitle title="設定" description="ログイン情報を確認しています。" />
          <MessageBox variant="info">{message}</MessageBox>
        </section>
      </PageContainer>
    );
  }

  return (
    <PageContainer>
      <AdminSideMenu />

      <div className={styles.pageWrap}>
        <section className={styles.pageCard}>
          <div className={styles.headerArea}>
            <PageTitle title="設定" description="Timexeedで使用する各種管理設定を選択できます。" />

            <MessageBox variant="info">
              設定したい項目を選択してください。有給管理、勤怠関連、給与関連などの管理機能をここにまとめます。
            </MessageBox>
          </div>

          <div className={styles.settingGrid}>
            {settingMenuItems.map((item) => (
              <article key={item.href} className={styles.settingCard}>
                <div className={styles.settingCardHeader}>
                  <div>
                    <p className={styles.settingCardTitle}>{item.title}</p>
                    <p className={styles.settingCardDescription}>{item.description}</p>
                  </div>

                  <span className={styles.statusBadge}>{item.statusLabel}</span>
                </div>

                <div className={styles.settingCardFooter}>
                  <Button type="button" variant="primary" onClick={() => router.push(item.href)}>
                    開く
                  </Button>
                </div>
              </article>
            ))}
          </div>
        </section>
      </div>
    </PageContainer>
  );
}