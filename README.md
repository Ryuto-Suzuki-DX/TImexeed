Timexeed 開発環境・作成ファイルまとめ
現在の構成
Timexeed/
├─ docker-compose.yml
├─ frontend/
│  ├─ package.json
│  ├─ package-lock.json
│  ├─ .env.local
│  └─ src/
│     ├─ api/
│     │  └─ auth.ts
│     ├─ lib/
│     │  └─ auth.ts
│     └─ app/
│        ├─ page.tsx
│        ├─ login/
│        │  └─ page.tsx
│        └─ mypage/
│           └─ page.tsx
└─ backend/
   ├─ go.mod
   ├─ go.sum
   ├─ main.go
   ├─ .air.toml
   └─ internal/
      ├─ auth/
      │  └─ jwt.go
      ├─ database/
      │  └─ database.go
      ├─ handlers/
      │  ├─ auth_handler.go
      │  └─ health_handler.go
      ├─ middlewares/
      │  └─ auth_middleware.go
      ├─ models/
      │  └─ user.go
      └─ routes/
         └─ routes.go



作成ファイルの役割
ルート直下
docker-compose.yml

PostgreSQLをDockerで起動するための設定ファイル。
DB名、ユーザー名、パスワード、ポート、永続化ボリュームを定義する。

backend
backend/go.mod

Goプロジェクトのモジュール名と依存ライブラリを管理する。

backend/go.sum

依存ライブラリの整合性を管理する。
基本的に手動編集しない。

backend/.air.toml

Airの設定ファイル。
Goファイル変更時に自動ビルド・自動再起動する。

backend/main.go

バックエンドの起動ファイル。
DB接続、CORS設定、ルーティング登録、サーバー起動を行う。

backend/internal/database/database.go

PostgreSQLへの接続処理を管理する。
GORMの初期化とAutoMigrateもここで行う。

backend/internal/models/user.go

Userモデルを定義する。
users テーブルの元になるファイル。

backend/internal/routes/routes.go

APIのURLとhandlerを紐づける。
/health、/auth/login などのルートを定義する。

backend/internal/handlers/health_handler.go

ヘルスチェック用APIを処理する。
/health と /db-health を担当する。

backend/internal/handlers/auth_handler.go

認証系APIを処理する。
ユーザー登録、ログイン、ログイン中ユーザー取得を担当する。

backend/internal/auth/jwt.go

JWTの発行と検証を担当する。
ログイン成功時のaccessToken発行に使う。

backend/internal/middlewares/auth_middleware.go

JWT認証ミドルウェア。
Authorization: Bearer トークン を検証し、認証済みAPIを保護する。

frontend
frontend/.env.local

フロント用の環境変数ファイル。
APIの接続先などを管理する。

frontend/src/lib/auth.ts

accessTokenをlocalStorageに保存・取得・削除する。

frontend/src/api/auth.ts

認証APIを呼び出すファイル。
ログインAPIと認証確認APIを担当する。

frontend/src/app/page.tsx

トップページ。
開発初期のAPI疎通確認用ページ。

frontend/src/app/login/page.tsx

ログイン画面。
ログイン成功後、accessTokenを保存して /mypage に遷移する。

frontend/src/app/mypage/page.tsx

ログイン後のマイページ。
/auth/me を呼び出してログイン中ユーザー情報を表示する。






作業手順
1. Next.js作成
cd C:\Users\zukis\Desktop\Timexeed
npx create-next-app@latest frontend
2. Gin backend作成
mkdir backend
cd backend
go mod init timexeed/backend
go get github.com/gin-gonic/gin
3. Air導入
go install github.com/air-verse/air@latest

.air.toml を作成し、Goのホットリロードを有効化。

4. PostgreSQL起動

ルート直下に docker-compose.yml を作成。

cd C:\Users\zukis\Desktop\Timexeed
docker compose up -d

今回はホスト側ポートを 15432 に設定。

5. GinからPostgreSQL接続
cd backend
go get gorm.io/gorm
go get gorm.io/driver/postgres

database.go を作成し、GORMでDB接続。

6. backend構成を分離

以下のように分離。

database → DB接続
models → DBモデル
routes → ルーティング
handlers → API処理
auth → JWT処理
middlewares → 認証ミドルウェア
7. Userモデル作成

internal/models/user.go を作成。
AutoMigrate で users テーブルを作成。

8. 認証API作成

作成したAPI。

POST /auth/register
POST /auth/login
GET  /auth/me
9. JWT対応
go get github.com/golang-jwt/jwt/v5

ログイン成功時にaccessTokenを返すようにした。

10. フロントAuth作成

作成した画面・処理。

/login  → ログイン画面
/mypage → ログイン後ページ

localStorage にaccessTokenを保存し、/auth/me で認証確認する。

起動手順
DB起動
cd C:\Users\zukis\Desktop\Timexeed
docker compose up -d
backend起動
cd C:\Users\zukis\Desktop\Timexeed\backend
air
frontend起動
cd C:\Users\zukis\Desktop\Timexeed\frontend
npm run dev
