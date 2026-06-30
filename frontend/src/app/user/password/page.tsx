"use client";

import { FormEvent, useState } from "react";
import { useRouter } from "next/navigation";
import Button from "@/components/atoms/Button";
import MessageBox from "@/components/atoms/MessageBox";
import PageContainer from "@/components/atoms/PageContainer";
import PageTitle from "@/components/atoms/PageTitle";
import UserSideMenu from "@/components/sideMenu/UserSideMenu";
import { useRequireRole } from "@/hooks/useRequireRole";
import { changePassword } from "@/api/user/password";
import styles from "./page.module.css";

type PageMessageVariant = "info" | "success" | "warning" | "error";

export default function UserPasswordPage() {
  const router = useRouter();
  const { user, isLoading, message } = useRequireRole("USER");

  const [currentPassword, setCurrentPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");

  const [pageMessage, setPageMessage] = useState(
    "現在のパスワードと新しいパスワードを入力してください。",
  );
  const [pageMessageVariant, setPageMessageVariant] =
    useState<PageMessageVariant>("info");

  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();

    if (!currentPassword) {
      setPageMessage("現在のパスワードを入力してください。");
      setPageMessageVariant("error");
      return;
    }

    if (!newPassword) {
      setPageMessage("新しいパスワードを入力してください。");
      setPageMessageVariant("error");
      return;
    }

    if (newPassword.length < 8) {
      setPageMessage("新しいパスワードは8文字以上で入力してください。");
      setPageMessageVariant("error");
      return;
    }

    if (newPassword !== confirmPassword) {
      setPageMessage(
        "新しいパスワードと確認用パスワードが一致しません。",
      );
      setPageMessageVariant("error");
      return;
    }

    if (currentPassword === newPassword) {
      setPageMessage(
        "現在のパスワードと異なるパスワードを設定してください。",
      );
      setPageMessageVariant("error");
      return;
    }

    setIsSubmitting(true);
    setPageMessage("パスワードを変更しています。");
    setPageMessageVariant("info");

    const result = await changePassword({
      currentPassword,
      newPassword,
    });

    if (result.error || !result.data) {
      setPageMessage(result.message || "パスワードの変更に失敗しました。");
      setPageMessageVariant("error");
      setIsSubmitting(false);
      return;
    }

    setCurrentPassword("");
    setNewPassword("");
    setConfirmPassword("");

    setPageMessage("パスワードを変更しました。");
    setPageMessageVariant("success");
    setIsSubmitting(false);

    router.push("/user/mypage");
    router.refresh();
  };

  if (isLoading || !user) {
    return (
      <PageContainer>
        <UserSideMenu />

        <section className={styles.loadingCard}>
          <PageTitle
            title="パスワード変更"
            description="ログイン情報を確認しています。"
          />
          <MessageBox variant="info">{message}</MessageBox>
        </section>
      </PageContainer>
    );
  }

  return (
    <PageContainer>
      <UserSideMenu />

      <div className={styles.pageWrap}>
        <section className={styles.pageCard}>
          <div className={styles.header}>
            <PageTitle
              title="パスワード変更"
              description="現在のパスワードを確認し、新しいパスワードへ変更します。"
            />
          </div>

          <div className={styles.messageArea}>
            <MessageBox variant={pageMessageVariant}>
              {pageMessage}
            </MessageBox>
          </div>

          <form className={styles.form} onSubmit={handleSubmit}>
            <section className={styles.formSection}>
              <div className={styles.sectionHeader}>
                <div>
                  <h2 className={styles.sectionTitle}>パスワード情報</h2>
                  <p className={styles.sectionDescription}>
                    新しいパスワードは8文字以上で設定してください。
                  </p>
                </div>
              </div>

              <div className={styles.formGrid}>
                <label className={styles.formLabel}>
                  <span className={styles.labelText}>
                    現在のパスワード
                  </span>
                  <input
                    type="password"
                    value={currentPassword}
                    onChange={(event) =>
                      setCurrentPassword(event.target.value)
                    }
                    className={styles.textInput}
                    autoComplete="current-password"
                    placeholder="現在のパスワード"
                    disabled={isSubmitting}
                  />
                </label>

                <label className={styles.formLabel}>
                  <span className={styles.labelText}>
                    新しいパスワード
                  </span>
                  <input
                    type="password"
                    value={newPassword}
                    onChange={(event) =>
                      setNewPassword(event.target.value)
                    }
                    className={styles.textInput}
                    autoComplete="new-password"
                    placeholder="8文字以上"
                    disabled={isSubmitting}
                  />
                </label>

                <label className={styles.formLabel}>
                  <span className={styles.labelText}>
                    新しいパスワード（確認）
                  </span>
                  <input
                    type="password"
                    value={confirmPassword}
                    onChange={(event) =>
                      setConfirmPassword(event.target.value)
                    }
                    className={styles.textInput}
                    autoComplete="new-password"
                    placeholder="新しいパスワードを再入力"
                    disabled={isSubmitting}
                  />
                </label>
              </div>
            </section>

            <div className={styles.actionArea}>
              <Button
                type="submit"
                variant="primary"
                disabled={isSubmitting}
              >
                {isSubmitting ? "変更中..." : "パスワードを変更"}
              </Button>
            </div>
          </form>
        </section>
      </div>
    </PageContainer>
  );
}
