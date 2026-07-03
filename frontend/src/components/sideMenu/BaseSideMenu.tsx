"use client";

import type { ReactNode } from "react";
import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import { removeAccessToken } from "@/api/auth";
import styles from "./BaseSideMenu.module.css";

export type SideMenuItem = {
  label: ReactNode;
  href: string;
};

type BaseSideMenuProps = {
  title: string;
  items: SideMenuItem[];
};

export default function BaseSideMenu({
  title,
  items,
}: BaseSideMenuProps) {
  const router = useRouter();
  const pathname = usePathname();

  const [isOpen, setIsOpen] = useState(false);
  const [isTouchDevice, setIsTouchDevice] = useState(false);

  useEffect(() => {
    const mediaQuery = window.matchMedia(
      "(hover: none), (pointer: coarse)",
    );

    const updateDeviceType = () => {
      setIsTouchDevice(mediaQuery.matches);
    };

    updateDeviceType();

    mediaQuery.addEventListener("change", updateDeviceType);

    return () => {
      mediaQuery.removeEventListener("change", updateDeviceType);
    };
  }, []);

  const handleLogout = () => {
    removeAccessToken();
    router.push("/login");
  };

  const handleMouseEnter = () => {
    if (!isTouchDevice) {
      setIsOpen(true);
    }
  };

  const handleMouseLeave = () => {
    if (!isTouchDevice) {
      setIsOpen(false);
    }
  };

  const handleToggleMenu = () => {
    setIsOpen((current) => !current);
  };

  const handleCloseMenu = () => {
    setIsOpen(false);
  };

  const handleMenuLinkClick = () => {
    if (isTouchDevice) {
      setIsOpen(false);
    }
  };

  return (
    <>
      {isTouchDevice && (
        <button
          type="button"
          className={`${styles.toggleButton} ${
            isOpen ? styles.toggleButtonOpen : ""
          }`}
          aria-label={isOpen ? "メニューを閉じる" : "メニューを開く"}
          aria-controls="base-side-menu"
          onClick={handleToggleMenu}
        >
          {isOpen ? "×" : "☰"}
        </button>
      )}

      {isTouchDevice && isOpen && (
        <button
          type="button"
          className={styles.overlay}
          aria-label="メニューを閉じる"
          onClick={handleCloseMenu}
        />
      )}

      <aside
        id="base-side-menu"
        className={`${styles.sideMenu} ${
          isOpen ? styles.sideMenuOpen : ""
        }`}
        onMouseEnter={handleMouseEnter}
        onMouseLeave={handleMouseLeave}
      >
        <div className={styles.sideMenuInner}>
          <div className={styles.header}>
            <p className={styles.brand}>Timexeed</p>
            <h2 className={styles.title}>{title}</h2>
          </div>

          <nav
            className={styles.navigation}
            aria-label={`${title}メニュー`}
          >
            {items.map((item) => {
              const isActive =
                pathname === item.href ||
                pathname.startsWith(`${item.href}/`);

              return (
                <Link
                  key={item.href}
                  href={item.href}
                  className={`${styles.menuLink} ${
                    isActive ? styles.menuLinkActive : ""
                  }`}
                  onClick={handleMenuLinkClick}
                >
                  {item.label}
                </Link>
              );
            })}
          </nav>

          <div className={styles.logoutArea}>
            <button
              type="button"
              className={styles.logoutButton}
              onClick={handleLogout}
            >
              ログアウト
            </button>
          </div>
        </div>
      </aside>
    </>
  );
}