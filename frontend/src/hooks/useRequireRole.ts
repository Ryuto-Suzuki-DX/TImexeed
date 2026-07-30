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
    let isMounted = true;

    const loadMe = async () => {
      setIsLoading(true);
      setMessage("認証確認中...");

      try {
        const result = await fetchMe();

        if (!isMounted) {
          return;
        }

        /*
         * 401やトークン未保存の場合は、fetchMe側で
         * ログインページへの遷移処理が行われる。
         */
        if (
          result.code === "UNAUTHORIZED" ||
          result.code === "TOKEN_NOT_FOUND"
        ) {
          setUser(null);
          setMessage(
            result.message ||
              "ログイン期限が切れました。再ログインしてください。",
          );
          setIsLoading(false);
          return;
        }

        /*
         * 一時的な通信エラーやサーバーエラーでは、
         * ログイン画面へ強制遷移させない。
         */
        if (result.error || !result.data) {
          setUser(null);
          setMessage(
            result.message ||
              "ログイン情報を確認できませんでした。再度お試しください。",
          );
          setIsLoading(false);
          return;
        }

        /*
         * ログイン中のロールと表示対象ページが異なる場合は、
         * ログアウトさせず、正しいロールのトップ画面へ戻す。
         */
        if (result.data.role !== requiredRole) {
          setUser(null);
          setMessage("このページを表示する権限がありません。");
          setIsLoading(false);

          if (result.data.role === "ADMIN") {
            router.replace("/admin");
          } else {
            router.replace("/user");
          }

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
      } catch {
        if (!isMounted) {
          return;
        }

        /*
         * fetch自体が失敗した場合もログアウト扱いにしない。
         */
        setUser(null);
        setMessage(
          "通信に失敗しました。接続状況を確認して再度お試しください。",
        );
        setIsLoading(false);
      }
    };

    loadMe();

    return () => {
      isMounted = false;
    };
  }, [requiredRole, router]);

  return {
    user,
    isLoading,
    message,
  };
}
