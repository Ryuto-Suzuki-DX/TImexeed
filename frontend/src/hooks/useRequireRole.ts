"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { fetchMe } from "@/api/auth";

type Role = "ADMIN" | "USER";

type LoginUser = {
  userId: number;
  name: string;
  email: string;
  role: Role;
};

type UseRequireRoleResult = {
  user: LoginUser | null;
  isLoading: boolean;
  message: string;
};

/*
 * 〇 指定したロールだけページ表示を許可する共通hook
 *
 * 管理者ページ → useRequireRole("ADMIN")
 * 従業員ページ → useRequireRole("USER")
 */
export function useRequireRole(requiredRole: Role): UseRequireRoleResult {
  const router = useRouter();

  const [user, setUser] = useState<LoginUser | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [message, setMessage] = useState("認証確認中...");

  useEffect(() => {
    const loadMe = async () => {
      setIsLoading(true);
      setMessage("認証確認中...");

      const result = await fetchMe();

      if (result.error || !result.data) {
        setUser(null);
        setMessage(result.message || "ログイン情報を確認できませんでした。");
        setIsLoading(false);
        router.push("/login");
        return;
      }

      if (result.data.role !== requiredRole) {
        setUser(null);
        setMessage("このページを表示する権限がありません。");
        setIsLoading(false);
        router.push("/login");
        return;
      }

      setUser({
        userId: result.data.userId,
        name: result.data.name,
        email: result.data.email,
        role: result.data.role,
      });

      setMessage("");
      setIsLoading(false);
    };

    loadMe();
  }, [requiredRole, router]);

  return {
    user,
    isLoading,
    message,
  };
}