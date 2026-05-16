0. まず決めること

保存先にするGoogleアカウントをこれに統一します。

社長のGoogleアカウント

以後、全部これに合わせます。

OAuthのテストユーザー
=
OAuth認可でログインするアカウント
=
Google Drive保存先フォルダを持つアカウント
1. 社長アカウント側でDriveフォルダを作る

社長のGoogleアカウントでGoogle Driveにログイン。

フォルダを作成。

TimexeedExpenseReceipts

そのフォルダURLをコピー。

例：

https://drive.google.com/drive/folders/xxxxxxxxxxxxxxxxxxxx

このURLをあとでDBに入れます。

2. DBの保存先URLを更新

新しいフォルダURL を社長アカウント側のフォルダURLに差し替えて実行。

docker compose exec db psql -U timexeed -d timexeed_db -c "update external_storage_links set url = '新しいフォルダURL', updated_at = now() where link_type = 'EXPENSE_RECEIPT_BOX' and is_deleted = false;"

確認。

docker compose exec db psql -U timexeed -d timexeed_db -c "select id, link_type, url, is_deleted, updated_at from external_storage_links where link_type = 'EXPENSE_RECEIPT_BOX';"
3. Google Cloud Consoleでテストユーザーに社長アカウントを追加

Google Cloud Consoleで、Timexeedのプロジェクトを開く。

Google Auth Platform
↓
対象
↓
テストユーザー

または、

APIとサービス
↓
OAuth 同意画面
↓
テストユーザー

ここに 社長のGoogleアカウントのメールアドレス を追加。

これをやらないと、テスト公開状態では社長アカウントでOAuth認可できません。

4. .env のrefresh tokenを一旦削除

対象：

C:\Users\zukis\Desktop\Timexeed\backend\.env

今あるこれを一旦消す、または空にする。

GOOGLE_OAUTH_REFRESH_TOKEN=...

CLIENT_ID と CLIENT_SECRET は同じOAuthクライアントを使うならそのままでOK。

GOOGLE_OAUTH_CLIENT_ID=...
GOOGLE_OAUTH_CLIENT_SECRET=...
GOOGLE_DRIVE_EXPENSE_RECEIPT_LINK_TYPE=EXPENSE_RECEIPT_BOX
5. backend再起動

.env とコマンド実行環境を揃えるため。

docker compose down
docker compose up -d --build
6. 社長アカウントでrefresh tokenを取り直す

CMDでこれ。

docker compose exec backend go run /app/cmd/google_oauth_token/main.go

表示されたURLをブラウザで開く。

https://accounts.google.com/o/oauth2/auth?...

ブラウザでは、必ず社長のGoogleアカウントを選択。

許可後、ページは失敗してOK。
URL欄にこういうURLが出ます。

http://localhost/?state=timexeed-google-drive-token&code=4/xxxx&scope=...

このURL全体をコピーして、ターミナルのここへ貼る。

code またはリダイレクトURLを貼り付け:

成功するとこれが出ます。

GOOGLE_OAUTH_REFRESH_TOKEN=...
7. .env に社長アカウント用refresh tokenを設定

対象：

C:\Users\zukis\Desktop\Timexeed\backend\.env

こうする。

GOOGLE_OAUTH_CLIENT_ID=あなたのOAuthクライアントID
GOOGLE_OAUTH_CLIENT_SECRET=あなたのOAuthクライアントシークレット
GOOGLE_OAUTH_REFRESH_TOKEN=社長アカウントで取得したrefresh_token
GOOGLE_DRIVE_EXPENSE_RECEIPT_LINK_TYPE=EXPENSE_RECEIPT_BOX

GOOGLE_APPLICATION_CREDENTIALS は不要です。

# GOOGLE_APPLICATION_CREDENTIALS=/app/secrets/google-service-account.json
8. backend再起動
docker compose down
docker compose up -d --build
9. 単体アップロード確認

この前作った drivecheck を実行。

docker compose exec backend go run /app/cmd/drivecheck/main.go

成功ならこう出ます。

OK: id=... name=timexeed_drive_check.txt url=...

Google Drive APIはDriveへのアップロード/ダウンロードを扱えます。今の drivecheck は、OAuthで認可したアカウントの権限で、DBに入れたフォルダIDへファイル作成を試す確認です。

10. 経費画面で領収書付き登録

最後に、管理者経費画面で領収書付き登録。

成功すれば完了。

社長アカウントに切り替えるたびに必要な作業

毎回全部ではなく、基本はこの3つだけです。

1. 社長アカウントをOAuthテストユーザーに追加
2. 社長アカウントでrefresh tokenを取り直す
3. backend/.env の GOOGLE_OAUTH_REFRESH_TOKEN を差し替える

保存先フォルダも変えるなら、DBの external_storage_links.url も更新します。
