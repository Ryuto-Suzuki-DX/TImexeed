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
    title: "給与全体設定",
    description: "通勤手当上限、在宅勤務補助、給与計算で使う会社全体の設定を管理します。",
    href: "/admin/salary/company-settings",
    statusLabel: "全体設定",
  },
  {
    title: "ユーザー給与詳細",
    description: "ユーザーごとの給与区分、基本金額、固定手当、固定控除、適用期間を管理します。",
    href: "/admin/salary/user-salary-details",
    statusLabel: "個人設定",
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
            <PageTitle title="給与管理" description="給与計算に使う全体設定とユーザーごとの給与詳細を管理します。" />

            <MessageBox variant="info">
              管理したい項目を選択してください。会社全体の給与設定と、ユーザーごとの給与詳細設定をここにまとめます。
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
