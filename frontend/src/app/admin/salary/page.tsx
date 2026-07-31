"use client";

import { useRouter } from "next/navigation";
import Button from "@/components/atoms/Button";
import MessageBox from "@/components/atoms/MessageBox";
import PageContainer from "@/components/atoms/PageContainer";
import PageTitle from "@/components/atoms/PageTitle";
import AdminSideMenu from "@/components/sideMenu/AdminSideMenu";
import { useRequireRole } from "@/hooks/useRequireRole";
import styles from "./page.module.css";

type SalaryMenuItem = {
  title: string;
  description: string;
  href: string;
  statusLabel: string;
};

const salaryMenuItems: SalaryMenuItem[] = [
  {
    title: "ユーザー給与・月次手当管理",
    description: "ユーザーごとの給与区分、基本金額、適用期間と、月ごとの手当を管理します。",
    href: "/admin/salary/user-salary-details",
    statusLabel: "個人設定",
  },
  {
    title: "手当種別管理",
    description: "固定残業手当、在宅手当、資格手当など、月次手当で選択する種別を追加・編集できます。",
    href: "/admin/salary/allowance-types",
    statusLabel: "給与設定",
  },
];

export default function AdminSalaryPage() {
  const router = useRouter();
  const { user, isLoading, message } = useRequireRole("ADMIN");

  if (isLoading || !user) {
    return (
      <PageContainer>
        <AdminSideMenu />

        <section className={styles.loadingCard}>
          <PageTitle title="給与管理" description="ログイン情報を確認しています。" />
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
            <PageTitle title="給与管理" description="給与計算に使う設定とユーザーごとの給与・手当を管理します。" />

            <MessageBox variant="info">
              管理したい項目を選択してください。手当種別を先に登録すると、ユーザーごとの月次手当入力で選択できます。
            </MessageBox>
          </div>

          <div className={styles.salaryGrid}>
            {salaryMenuItems.map((item) => (
              <article key={item.href} className={styles.salaryCard}>
                <div className={styles.salaryCardHeader}>
                  <div>
                    <p className={styles.salaryCardTitle}>{item.title}</p>
                    <p className={styles.salaryCardDescription}>{item.description}</p>
                  </div>

                  <span className={styles.statusBadge}>{item.statusLabel}</span>
                </div>

                <div className={styles.salaryCardFooter}>
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
