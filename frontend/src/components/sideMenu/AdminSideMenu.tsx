"use client";

import { useEffect, useState } from "react";
import BaseSideMenu, { type SideMenuItem } from "@/components/sideMenu/BaseSideMenu";
import { countUnreadNotifications } from "@/api/admin/notification";

export default function AdminSideMenu() {
  const [unreadNotificationCount, setUnreadNotificationCount] = useState(0);

  useEffect(() => {
    const loadUnreadNotificationCount = async () => {
      const result = await countUnreadNotifications({});

      if (result.error || !result.data) {
        setUnreadNotificationCount(0);
        return;
      }

      setUnreadNotificationCount(result.data.unreadCount);
    };

    void loadUnreadNotificationCount();
  }, []);

  const adminMenuItems: SideMenuItem[] = [
    { label: "出退勤リアルタイム", href: "/admin/attendance-realtime-events" },
    { label: "マイページ", href: "/admin/mypage" },
    { label: "ユーザー管理", href: "/admin/users" },
    { label: "所属管理", href: "/admin/departments" },
    { label: "勤怠管理", href: "/admin/attendance" },
    { label: "月次勤怠申請管理", href: "/admin/monthly-attendance-requests" },
    { label: "有給確認", href: "/admin/paid-leave-check" },
    { label: "給与管理", href: "/admin/salary" },
    { label: "経費登録", href: "/admin/expenses" },
    { label: "月次集計CSV出力", href: "/admin/monthly-attendance-summary-exports" },
    { label: "個人情報", href: "/admin/personal-information-drive-folders" },
    {
      label:
        unreadNotificationCount > 0 ? (
          <>
            お知らせ{" "}
            <span style={{ color: "#dc2626", fontWeight: "bold" }}>NEW!</span>
          </>
        ) : (
          "お知らせ"
        ),
      href: "/admin/notifications",
    },
    { label: "自動リマインド", href: "/admin/notification-reminders" },
    { label: "共有資料(FAQ)", href: "/admin/shared-document-drive-folders" },
    { label: "設定", href: "/admin/settings" },
  ];

  return <BaseSideMenu title="管理者メニュー" items={adminMenuItems} />;
}