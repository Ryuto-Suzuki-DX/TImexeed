"use client";

import BaseSideMenu, { type SideMenuItem } from "@/components/sideMenu/BaseSideMenu";

const userMenuItems: SideMenuItem[] = [
  { label: "マイページ", href: "/user/mypage" },
  { label: "勤怠入力", href: "/user/attendance" },
  { label: "勤怠履歴", href: "/user/attendance/history" },
  { label: "設定", href: "/user/settings" },
];

export default function UserSideMenu() {
  return <BaseSideMenu title="従業員メニュー" items={userMenuItems} />;
}