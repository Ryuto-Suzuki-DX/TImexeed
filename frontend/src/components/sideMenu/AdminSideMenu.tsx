"use client";

import BaseSideMenu, { type SideMenuItem } from "@/components/sideMenu/BaseSideMenu";

const adminMenuItems: SideMenuItem[] = [
  { label: "マイページ", href: "/admin/mypage" },
  { label: "ユーザー管理", href: "/admin/users" },
  { label: "所属管理", href: "/admin/departments" },
  { label: "勤怠管理", href: "/admin/attendance" },
  { label: "給与管理", href: "/admin/salary" },
  { label: "設定", href: "/admin/settings" },
];

export default function AdminSideMenu() {
  return <BaseSideMenu title="管理者メニュー" items={adminMenuItems} />;
}