"use client";

import { useEffect, useState } from "react";
import BaseSideMenu, { type SideMenuItem } from "@/components/sideMenu/BaseSideMenu";
import { countUnreadNotifications } from "@/api/user/notification";
import { fetchMe } from "@/api/auth";

export default function UserSideMenu() {
  const [unreadNotificationCount, setUnreadNotificationCount] = useState(0);
  const [mustChangePassword, setMustChangePassword] = useState(false);

  useEffect(() => {
    const loadUnreadNotificationCount = async () => {
      const result = await countUnreadNotifications({});

      if (result.error || !result.data) {
        setUnreadNotificationCount(0);
        return;
      }

      setUnreadNotificationCount(result.data.unreadCount);
    };

    const loadCurrentUser = async () => {
      const result = await fetchMe();

      if (result.error || !result.data) {
        setMustChangePassword(false);
        return;
      }

      setMustChangePassword(result.data.mustChangePassword);
    };

    void loadUnreadNotificationCount();
    void loadCurrentUser();
  }, []);

  const userMenuItems: SideMenuItem[] = [
    { label: "マイページ", href: "/user/mypage" },
    { label: "勤怠入力", href: "/user/attendance" },
    { label: "経費登録", href: "/user/expenses" },
    { label: "個人情報", href: "/user/personal-information-drive-folders" },
    { label: "共有資料(FAQ)", href: "/user/shared-document-drive-folders" },
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
      href: "/user/notifications",
    },
    {
      label: mustChangePassword ? (
        <>
          パスワード変更{" "}
          <span style={{ color: "#dc2626", fontWeight: "bold" }}>！</span>
        </>
      ) : (
        "パスワード変更"
      ),
      href: "/user/password",
    },
  ];

  return <BaseSideMenu title="従業員メニュー" items={userMenuItems} />;
}
