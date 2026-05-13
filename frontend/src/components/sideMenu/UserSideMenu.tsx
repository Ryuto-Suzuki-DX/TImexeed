"use client";

import { useEffect, useState } from "react";
import BaseSideMenu, { type SideMenuItem } from "@/components/sideMenu/BaseSideMenu";
import { countUnreadNotifications } from "@/api/user/notification";

export default function UserSideMenu() {
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

  const userMenuItems: SideMenuItem[] = [
    { label: "マイページ", href: "/user/mypage" },
    { label: "勤怠入力", href: "/user/attendance" },
    { label: "勤怠履歴", href: "/user/attendance/history" },
    { label: "設定", href: "/user/settings" },
    {
      label: unreadNotificationCount > 0 ? "お知らせ NEW!" : "お知らせ",
      href: "/user/notifications",
    },
  ];

  return <BaseSideMenu title="従業員メニュー" items={userMenuItems} />;
}