"use client";

import { FormEvent, useState } from "react";
import styles from "./page.module.css";

export default function UserPasswordPage() {
  const [currentPassword, setCurrentPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [errorMessage, setErrorMessage] = useState("");
  const [successMessage, setSuccessMessage] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();

    setErrorMessage("");
    setSuccessMessage("");

    if (!currentPassword) {
      setErrorMessage("現在のパスワードを入力してください。");
      return;
    }

    if (!newPassword) {
      setErrorMessage("新しいパスワードを入力してください。");
      return;
    }

    if (newPassword.length < 8) {
      setErrorMessage("新しいパスワードは8文字以上で入力してください。");
      return;
    }

    if (newPassword !== confirmPassword) {
      setErrorMessage("新しいパスワードと確認用パスワードが一致しません。");
      return;
    }

    if (currentPassword === newPassword) {
      setErrorMessage("現在のパスワードと異なるパスワードを設定してください。");
      return;
    }

    setIsSubmitting(true);

    try {
      /*
       * パスワード変更APIは次の作業で接続する。
       *
       * POST /user/password/change
       */
      setSuccessMessage("入力内容を確認しました。");
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <main className={styles.page}>
      <section className={styles.card}>
        <div className={styles.header}>
          <div>
            <h1 className={styles.title}>パスワード変更</h1>
            <p className={styles.description}>
              現在のパスワードと新しいパスワードを入力してください。
            </p>
          </div>
        </div>

        <form className={styles.form} onSubmit={handleSubmit}>
          <div className={styles.formGroup}>
            <label className={styles.label} htmlFor="currentPassword">
              現在のパスワード
            </label>
            <input
              id="currentPassword"
              className={styles.input}
              type="password"
              autoComplete="current-password"
              value={currentPassword}
              onChange={(event) => setCurrentPassword(event.target.value)}
              placeholder="現在のパスワードを入力"
            />
          </div>

          <div className={styles.formGroup}>
            <label className={styles.label} htmlFor="newPassword">
              新しいパスワード
            </label>
            <input
              id="newPassword"
              className={styles.input}
              type="password"
              autoComplete="new-password"
              value={newPassword}
              onChange={(event) => setNewPassword(event.target.value)}
              placeholder="8文字以上で入力"
            />
            <p className={styles.helpText}>
              8文字以上で、現在のパスワードとは異なるものを設定してください。
            </p>
          </div>

          <div className={styles.formGroup}>
            <label className={styles.label} htmlFor="confirmPassword">
              新しいパスワード（確認）
            </label>
            <input
              id="confirmPassword"
              className={styles.input}
              type="password"
              autoComplete="new-password"
              value={confirmPassword}
              onChange={(event) => setConfirmPassword(event.target.value)}
              placeholder="新しいパスワードを再入力"
            />
          </div>

          {errorMessage && (
            <div className={styles.errorMessage}>{errorMessage}</div>
          )}

          {successMessage && (
            <div className={styles.successMessage}>{successMessage}</div>
          )}

          <div className={styles.actionArea}>
            <button
              className={styles.submitButton}
              type="submit"
              disabled={isSubmitting}
            >
              {isSubmitting ? "変更中..." : "パスワードを変更"}
            </button>
          </div>
        </form>
      </section>
    </main>
  );
}
