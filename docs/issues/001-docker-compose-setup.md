# 🐳 Issue #1: Docker Composeでの開発環境構築

## 🎯 目標

Docker Composeを使ってローカル開発環境を構築し、MySQL + Go APIサーバーが問題なく起動できる状態にする。

## 📝 概要

このissueでは、プロジェクト全体で使用する開発環境をセットアップします。Docker Composeを使うことで、複数のサービス（MySQL、APIサーバー）を簡単に管理できます。このステップを完了すれば、以降の開発がスムーズに進みます。

## ✅ 完了条件（Acceptance Criteria）

このissueが完了したと判断する条件は以下の通りです。すべてを確認してください：

1. **起動確認**: `docker compose up -d` コマンドでMySQLとAPIが起動する
2. **ヘルスチェック**: `curl http://localhost:8080/api/v1/health` で以下のレスポンスが返る
   ```json
   {"status":"healthy"}
   ```
3. **データベース**: MySQLに`deliveries`テーブルが自動作成される
4. **ログ確認**: `docker compose logs -f api` でエラーが表示されない

---

## 🔧 作業手順（Step by Step）

### 1️⃣ Dockerのインストール確認

まず、あなたのマシンにDocker・Docker Composeがインストール済みか確認しましょう。

**Dockerのバージョン確認:**
```bash
docker --version
```

**Docker Composeのバージョン確認:**
```bash
docker compose version
```

**期待される出力例:**
```
Docker version 24.0.0
Docker Compose version 2.20.0
```

❌ コマンドが見つからないエラーが出た場合は、[公式サイト](https://www.docker.com/products/docker-desktop)からインストールしてください。

---

### 2️⃣ .envファイルの作成

プロジェクトのルートディレクトリに `.env` ファイルを作成します。このファイルには環境変数が記されており、Docker Composeやアプリケーションが読み込みます。

**ファイルの場所:**
```
delivery-api/
├── .env          ← ここに作成
├── docker-compose.yml
├── Dockerfile
└── ...
```

**`.env` の内容:**
```bash
# MySQL設定
MYSQL_ROOT_PASSWORD=rootpassword
MYSQL_DATABASE=deliveries_db
MYSQL_USER=delivery_user
MYSQL_PASSWORD=delivery_password

# API設定
API_PORT=8080
API_ENV=development
DATABASE_URL=delivery_user:delivery_password@tcp(db:3306)/deliveries_db?parseTime=true

# ログ設定
LOG_LEVEL=debug
```

⚠️ **注意**: この `.env` は開発環本番環境では使用しないでください。`.env` を `.gitignore` に追加して、Gitにコミットしないようにしてください。

---

### 3️⃣ docker-compose.ymlの理解

`docker-compose.yml` ファイルはすでにプロジェクトに存在しています。各行がどのような意味を持つのか、詳しく説明します。

**ファイルの場所:**
```
delivery-api/
├── docker-compose.yml  ← このファイル
├── Dockerfile
└── ...
```

**基本的な構造:**
```yaml
version: '3.8'                    # Docker Composeのバージョン

services:                          # 管理するサービスを定義
  db:                              # サービス名：MySQL
    image: mysql:8.0               # 使用するMySQLのバージョン
    container_name: delivery-db    # コンテナ名（docker ps で表示される）
    environment:                   # 環境変数を設定
      MYSQL_ROOT_PASSWORD: ${MYSQL_ROOT_PASSWORD}
      MYSQL_DATABASE: ${MYSQL_DATABASE}
      MYSQL_USER: ${MYSQL_USER}
      MYSQL_PASSWORD: ${MYSQL_PASSWORD}
    ports:
      - "3306:3306"                # ホストの3306をコンテナの3306にマッピング
    volumes:
      - db_data:/var/lib/mysql     # MySQLのデータを永続化
      - ./init.sql:/docker-entrypoint-initdb.d/init.sql
                                   # 初期化スクリプトを実行
    networks:
      - delivery-network           # ネットワークを指定
    healthcheck:                   # ヘルスチェック設定
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost"]
      interval: 10s                # 10秒ごとにチェック
      timeout: 5s                  # タイムアウト時間
      retries: 5                   # 失敗時の再試行回数

  api:                             # サービス名：Go API
    build:
      context: .                   # Dockerfileがあるディレクトリ
      dockerfile: Dockerfile       # 使用するDockerfile名
    container_name: delivery-api
    environment:
      - DATABASE_URL=${DATABASE_URL}
      - API_PORT=${API_PORT}
      - LOG_LEVEL=${LOG_LEVEL}
    ports:
      - "8080:8080"                # ホストの8080をコンテナの8080にマッピング
    depends_on:
      db:                          # db が起動したあとに起動する
        condition: service_healthy # db のヘルスチェックが成功するまで待つ
    networks:
      - delivery-network
    volumes:
      - .:/app                     # ホストとコンテナのコードを同期（開発用）
    command: go run ./cmd/main.go  # コンテナ起動時に実行するコマンド

networks:
  delivery-network:                # 複数のサービス間で通信するためのネットワーク

volumes:
  db_data:                         # MySQLのデータを永続化するボリューム
```

**重要なポイント:**
- **ports**: `ホストのポート:コンテナのポート` という形式
- **environment**: アプリケーションが読み込む環境変数
- **depends_on**: サービスの起動順序を制御
- **networks**: 異なるコンテナ間での通信を可能にする
- **volumes**: データの永続化（コンテナ再起動時もデータが消えない）

---

### 4️⃣ Dockerfileの理解

Go アプリケーションをDocker化するための `Dockerfile` を詳しく解説します。

**Dockerfileの場所:**
```
delivery-api/
├── Dockerfile     ← このファイル
├── go.mod
├── go.sum
└── cmd/
    └── main.go
```

**ファイルの内容（マルチステージビルド）:**
```dockerfile
# ステージ 1: ビルドステージ
# 大きなビルド環境を使ってGoバイナリをコンパイルします

FROM golang:1.21-alpine AS builder
                                   # golang:1.21のイメージを使用
                                   # AS builder で「builder」という名前をつける

WORKDIR /app                       # コンテナ内の作業ディレクトリを /app に設定

COPY go.mod go.sum ./              # go.mod と go.sum をコピー（依存関係ファイル）

RUN go mod download                # 依存パッケージをダウンロード

COPY . .                           # すべてのGoコードをコピー

RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/main.go
                                   # Goアプリケーションをコンパイル
                                   # CGO_ENABLED=0: C言語の機能を無効化（バイナリを小さく）
                                   # GOOS=linux: Linux用のバイナリを生成

# ステージ 2: 実行ステージ
# ビルド済みのバイナリだけを含む、小さなイメージを作成します

FROM alpine:latest                 # alpine（最小限のLinuxイメージ）を使用

RUN apk add --no-cache ca-certificates
                                   # SSL証明書をインストール（HTTPS通信用）

WORKDIR /app

COPY --from=builder /app/main .    # ステージ1でビルドした main バイナリをコピー

EXPOSE 8080                        # ポート8080を公開（ドキュメント用）

CMD ["./main"]                     # コンテナ起動時に ./main を実行
```

**マルチステージビルドのメリット:**
- 🎯 **イメージサイズを大幅削減**: ビルド環境は含まず、バイナリだけを最終イメージに含める
- 🚀 **コンテナ起動が高速**: 余分な内容がないため、起動が早い
- 🔒 **セキュリティ向上**: ビルド時のツール（コンパイラなど）は本番イメージに含まない

---

### 5️⃣ docker compose build の実行

`docker-compose.yml` と `Dockerfile` が準備できたら、イメージをビルドします。

**コマンド実行:**
```bash
docker compose build
```

**期待される出力:**
```
[+] Building 45.3s (15/15) FINISHED                       docker:default
 => [internal] load build definition from Dockerfile                      0.0s
 => [internal] load .dockerignore                                         0.0s
 => ...
 => [builder 4/6] RUN go mod download                                    12.5s
 => [builder 5/6] COPY . .                                                0.1s
 => [builder 6/6] RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/m  8.2s
 => [stage-1 3/3] COPY --from=builder /app/main .                         0.1s
 => exporting to image                                                     0.3s
 => => writing image sha256:abc123...                                      0.1s
 => => naming to docker.io/library/delivery-api:latest                    0.0s
```

⏱️ **初回ビルドは時間がかかります** （依存パッケージのダウンロード、コンパイルなど）。気長に待ちましょう。

---

### 6️⃣ docker compose up -d の実行

イメージがビルドできたら、コンテナを起動します。

**コマンド実行:**
```bash
docker compose up -d
```

**フラグの説明:**
- `-d`: デタッチモード（バックグラウンドで実行）

**期待される出力:**
```
[+] Running 2/2
 ✓ Network delivery-network   Created
 ✓ Container delivery-db      Started
 ✓ Container delivery-api     Started
```

**コンテナが起動したか確認:**
```bash
docker compose ps
```

**期待される出力:**
```
NAME            IMAGE            STATUS          PORTS
delivery-db     mysql:8.0        Up 2 minutes    0.0.0.0:3306->3306/tcp
delivery-api    delivery-api     Up 1 minute     0.0.0.0:8080->8080/tcp
```

✅ 両方のコンテナが「Up」状態になっていれば成功です。

---

### 7️⃣ ログの確認方法

コンテナが起動したら、ログを確認してエラーがないか見てみましょう。

**APIサーバーのログをリアルタイム表示:**
```bash
docker compose logs -f api
```

**フラグの説明:**
- `-f`: フォローモード（リアルタイムで追跡）

**期待されるログ:**
```
delivery-api  | 2026-03-01 10:15:30 Attempting to connect to database...
delivery-api  | 2026-03-01 10:15:32 Connected to database successfully
delivery-api  | 2026-03-01 10:15:32 Server is running on :8080
```

**特定のサービスのログ確認:**
```bash
docker compose logs db      # MySQLのログを表示
docker compose logs api     # APIのログを表示
```

**最後の100行だけ表示:**
```bash
docker compose logs --tail=100 api
```

❌ エラーログが表示されている場合は、[💡 ヒント](#-ヒント-よくあるエラーと対処法)セクションを参照してください。

---

### 8️⃣ ヘルスチェック確認

APIサーバーのヘルスチェックエンドポイントが正しく動作しているか確認します。

**コマンド実行:**
```bash
curl http://localhost:8080/api/v1/health
```

**期待される出力:**
```json
{"status":"healthy"}
```

**より詳しく確認（レスポンスコード + ボディ）:**
```bash
curl -v http://localhost:8080/api/v1/health
```

**期待される出力の一部:**
```
HTTP/1.1 200 OK
Content-Type: application/json
...

{"status":"healthy"}
```

✅ HTTPステータスが `200` でレスポンスが返れば成功です。

---

### 9️⃣ MySQLへの接続確認

MySQLコンテナが正しく動作しているか、データベースに接続して確認します。

**MySQLシェルに接続:**
```bash
docker compose exec db mysql -u delivery_user -p
```

プロンプトで パスワードを聞かれるので、`.env` ファイルで設定した `MYSQL_PASSWORD` を入力してください。

**例:**
```
$ docker compose exec db mysql -u delivery_user -p
Enter password: delivery_password
```

**MySQLに接続できたら、データベースと テーブル を確認:**
```sql
SHOW DATABASES;
```

**期待される出力:**
```
+--------------------+
| Database           |
+--------------------+
| information_schema |
| deliveries_db      |
+--------------------+
```

**テーブルを確認:**
```sql
USE deliveries_db;
SHOW TABLES;
```

**期待される出力:**
```
+----------------------+
| Tables_in_deliveries |
+----------------------+
| deliveries           |
+----------------------+
```

✅ `deliveries` テーブルが存在すれば成功です。

**MySQLシェルから出る:**
```sql
exit
```

---

## 💡 ヒント: よくあるエラーと対処法

### エラー 1️⃣: ポートが既に使用されている

**エラーメッセージ:**
```
Error response from daemon: ... Bind for 0.0.0.0:3306 failed: port is already allocated
```

**原因:** ポート 3306 または 8080 が他のプロセスで使用されている

**解決方法:**

**方法A: 既存のコンテナを停止**
```bash
docker compose down
```

**方法B: ポートを変更**
`docker-compose.yml` の ports セクションを編集：
```yaml
db:
  ports:
    - "3307:3306"        # ホストの3307をコンテナの3306にマッピング
```

**方法C: どのプロセスがポートを使っているか確認**
```bash
# macOS / Linux
lsof -i :3306
lsof -i :8080

# Windows (PowerShell)
netstat -ano | findstr :3306
netstat -ano | findstr :8080
```

---

### エラー 2️⃣: go mod download エラー

**エラーメッセージ:**
```
error: failed to resolve module not found ...
```

**原因:** Go モジュールのダウンロード失敗。ネットワークやプロキシの問題の可能性

**解決方法:**

**方法A: キャッシュをクリアして再ビルド**
```bash
docker compose down
docker system prune -a
docker compose build --no-cache
```

**方法B: go.mod と go.sum を確認**
プロジェクトのルートに `go.mod` ファイルが存在し、内容が正しいか確認してください。

**方法C: ネットワーク設定を確認**
プロキシの背後にいる場合は、Docker にプロキシを設定する必要があるかもしれません。

---

### エラー 3️⃣: MySQL接続待ちの問題

**エラーメッセージ:**
```
Error 1045 (28000): Access denied for user 'delivery_user'@'localhost'
```

または

```
Error 2003 (HY000): Can't connect to MySQL server on 'db' (111)
```

**原因:** APIサーバーがMySQL起動完了を待たずに接続しようとしている

**解決方法:**

**方法A: docker-compose.yml の depends_on を確認**
`api` サービスに以下が設定されているか確認：
```yaml
depends_on:
  db:
    condition: service_healthy
```

**方法B: MySQLのヘルスチェック設定を確認**
`db` サービスに healthcheck が設定されているか確認：
```yaml
healthcheck:
  test: ["CMD", "mysqladmin", "ping", "-h", "localhost"]
  interval: 10s
  timeout: 5s
  retries: 5
```

**方法C: 手動で待機してからAPIを起動**
```bash
# MySQLが完全に起動するまで待つ
docker compose up -d db
sleep 20

# その後、APIを起動
docker compose up -d api
```

---

### エラー 4️⃣: ビルドエラー（Dockerfile の問題）

**エラーメッセージ:**
```
COPY failed: file not found in build context
```

**原因:** Dockerfile で参照しているファイルが見つからない

**解決方法:**
1. ビルドコンテキストを確認（Dockerfile の `COPY` コマンドが正しいか）
2. 必要なファイルがプロジェクトに存在するか確認
3. `docker compose build --no-cache` で再ビルド

---

### エラー 5️⃣: コンテナがすぐに終了する

**症状:** `docker compose ps` で `Exited (1) 5 seconds ago` のように表示される

**原因:** アプリケーション実行時のエラー

**解決方法:**
```bash
# ログを詳しく確認
docker compose logs api

# コンテナを停止した状態で実行してログを見る
docker compose run --rm api
```

---

## 📦 参考コマンド一覧

よく使うDocker Composeコマンドを一覧にしました。

| コマンド | 説明 |
|---------|------|
| `docker compose build` | イメージをビルド |
| `docker compose up -d` | バックグラウンドで起動 |
| `docker compose up` | フォアグラウンドで起動（ログを見たいときに便利） |
| `docker compose down` | コンテナを停止して削除 |
| `docker compose ps` | 実行中のコンテナを表示 |
| `docker compose logs api` | APIのログを表示 |
| `docker compose logs -f api` | APIのログをリアルタイムで追跡 |
| `docker compose logs --tail=50 api` | 最後の50行を表示 |
| `docker compose exec db mysql -u delivery_user -p` | MySQLに接続 |
| `docker compose exec api sh` | APIコンテナのシェルに入る |
| `docker compose exec api go run ./cmd/main.go` | APIサーバーを再起動 |
| `docker compose restart api` | APIコンテナを再起動 |
| `docker compose stop` | コンテナを一時停止 |
| `docker compose start` | 一時停止したコンテナを再開 |
| `docker system prune -a` | 未使用のイメージ・ボリュームを削除 |

---

## 🔥 注意事項

### セキュリティについて

⚠️ このセットアップは**開発環境専用**です。以下の点に注意してください：

- **`.env` ファイルをGitにコミットしないこと**
  ```bash
  # .gitignore に以下を追加
  .env
  ```

- **初期パスワードを変更すること**
  本番環境では、`.env` のパスワードを強力なパスワードに変更してください。

- **本番環境では Docker Secrets や環境変数管理ツールを使用すること**
  AWS Secrets Manager、HashiCorp Vault など、セキュアな管理方法を導入してください。

### パフォーマンスについて

- **`volumes` マウント時のパフォーマンス低下**
  macOS・Windows で Docker Desktop を使用している場合、ボリュームマウントが遅い可能性があります。その場合は `volumes` の設定を調整してください。

### ストレージについて

- **MySQLのデータは永続化されます**
  `docker compose down` でコンテナを削除してもデータは残ります。完全にリセットする場合は：
  ```bash
  docker compose down -v
  ```

---

## 🚀 次のステップ

このissueが完了したら、以下のステップに進みます：

- Issue #2: APIサーバーの実装（Gin フレームワークの基本）
- Issue #3: データベーススキーマの設計（GORM の使用）
- Issue #4: エンドポイントの実装（REST API）

このドキュメントで不明な点やエラーが発生した場合は、遠慮なく質問してください！

Happy coding! 🎉
