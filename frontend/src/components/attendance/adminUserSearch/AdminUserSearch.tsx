"use client";

import { useState } from "react";
import Button from "@/components/atoms/Button";
import Input from "@/components/atoms/Input";
import { searchUsers } from "@/api/admin/user";
import type { UserResponse } from "@/types/admin/user";
import styles from "./AdminUserSearch.module.css";

const SEARCH_LIMIT = 10;

type AdminUserSearchProps = {
  selectedUser: UserResponse | null;
  disabled: boolean;
  onSelectUser: (user: UserResponse) => void;
};

export default function AdminUserSearch({
  selectedUser,
  disabled,
  onSelectUser,
}: AdminUserSearchProps) {
  const [keyword, setKeyword] = useState("");
  const [users, setUsers] = useState<UserResponse[]>([]);
  const [offset, setOffset] = useState(0);
  const [hasMore, setHasMore] = useState(false);
  const [message, setMessage] = useState("対象ユーザーを検索してください。");
  const [isSearching, setIsSearching] = useState(false);

  const handleSearch = async () => {
    if (disabled || isSearching) {
      return;
    }

    setIsSearching(true);
    setMessage("ユーザーを検索しています。");

    const result = await searchUsers({
      keyword,
      includeDeleted: false,
      offset: 0,
      limit: SEARCH_LIMIT,
    });

    setIsSearching(false);

    if (result.error || !result.data) {
      setUsers([]);
      setOffset(0);
      setHasMore(false);
      setMessage(result.message || "ユーザー検索に失敗しました。");
      return;
    }

    const data = result.data;

    setUsers(data.users);
    setOffset(data.offset + data.users.length);
    setHasMore(data.hasMore);

    if (data.users.length === 0) {
      setMessage("該当するユーザーが見つかりませんでした。");
      return;
    }

    setMessage("ユーザーを選択してください。");
  };

  const handleLoadMore = async () => {
    if (disabled || isSearching || !hasMore) {
      return;
    }

    setIsSearching(true);
    setMessage("追加でユーザーを取得しています。");

    const result = await searchUsers({
      keyword,
      includeDeleted: false,
      offset,
      limit: SEARCH_LIMIT,
    });

    setIsSearching(false);

    if (result.error || !result.data) {
      setMessage(result.message || "ユーザーの追加取得に失敗しました。");
      return;
    }

    const data = result.data;

    setUsers((current) => [...current, ...data.users]);
    setOffset(data.offset + data.users.length);
    setHasMore(data.hasMore);
    setMessage("ユーザーを選択してください。");
  };

  const handleSelectUser = (user: UserResponse) => {
    if (disabled || user.isDeleted) {
      return;
    }

    onSelectUser(user);
    setMessage(`${user.name} さんを選択しました。`);
  };

  return (
    <section className={styles.searchSection}>
      <div className={styles.sectionHeader}>
        <div>
          <h2 className={styles.sectionTitle}>対象ユーザー検索</h2>
          <p className={styles.sectionDescription}>
            勤怠を確認・編集するユーザーを検索して選択します。
          </p>
        </div>
      </div>

      <div className={styles.searchControl}>
        <Input
          label="ユーザー検索"
          placeholder="名前・メールアドレス"
          value={keyword}
          onChange={(event) => setKeyword(event.target.value)}
          disabled={disabled || isSearching}
        />

        <Button
          type="button"
          variant="secondary"
          onClick={handleSearch}
          disabled={disabled || isSearching}
        >
          検索
        </Button>
      </div>

      <p className={styles.searchMessage}>{isSearching ? "検索中..." : message}</p>

      {selectedUser && (
        <div className={styles.selectedUserBox}>
          <p className={styles.selectedUserLabel}>選択中ユーザー</p>
          <p className={styles.selectedUserName}>
            {selectedUser.name}
            <span className={styles.selectedUserSubText}>ID: {selectedUser.id}</span>
          </p>
          <p className={styles.selectedUserEmail}>{selectedUser.email}</p>
        </div>
      )}

      {users.length > 0 && (
        <div className={styles.resultList}>
          {users.map((user) => {
            const isSelected = selectedUser?.id === user.id;

            return (
              <button
                key={user.id}
                type="button"
                className={`${styles.resultItem} ${isSelected ? styles.resultItemSelected : ""}`}
                onClick={() => handleSelectUser(user)}
                disabled={disabled || isSearching || user.isDeleted}
              >
                <span className={styles.resultMain}>
                  <span className={styles.resultName}>{user.name}</span>
                  <span className={styles.resultMeta}>ID: {user.id}</span>
                </span>

                <span className={styles.resultEmail}>{user.email}</span>

                <span className={styles.resultRole}>{user.role}</span>
              </button>
            );
          })}

          {hasMore && (
            <div className={styles.loadMoreArea}>
              <Button
                type="button"
                variant="secondary"
                onClick={handleLoadMore}
                disabled={disabled || isSearching}
              >
                さらに表示
              </Button>
            </div>
          )}
        </div>
      )}
    </section>
  );
}