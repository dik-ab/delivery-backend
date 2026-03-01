# 🐳 Issue #002: Go + MySQL の Docker Compose 環境構築

## 📋 概要

Docker Composeを使って、Go（Gin）のAPIサーバーとMySQLデータベースが連携して動作する開発環境を構築する。

## 🎯 ゴール

- `docker compose up` で Go API + MySQL が起動する
- Go API が MySQL に接続できている状態になる
- ブラウザ or curl で `http://localhost:8080` にアクセスできる

## 📝 タスク

### 1. プロジェクトの初期化

```bash
mkdir delivery-api && cd delivery-api
go mod init github.com/delivery-app/delivery-api
```

### 2. 必要なパッケージのインストール

```bash
go get github.com/gin-gonic/gin
go get gorm.io/gorm
go get gorm.io/driver/mysql
```

### 3. エントリーポイントの作成

`cmd/server/main.go` を作成する。

```go
package main

import (
    "fmt"
    "log"
    "os"

    "github.com/gin-gonic/gin"
    "gorm.io/driver/mysql"
    "gorm.io/gorm"
)

func main() {
    // 環境変数からDB接続情報を取得
    dbUser := os.Getenv("DB_USER")
    dbPassword := os.Getenv("DB_PASSWORD")
    dbHost := os.Getenv("DB_HOST")
    dbPort := os.Getenv("DB_PORT")
    dbName := os.Getenv("DB_NAME")

    // MySQL接続
    dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
        dbUser, dbPassword, dbHost, dbPort, dbName,
    )

    db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatal("❌ DB接続に失敗しました:", err)
    }

    log.Println("✅ DB接続成功！")

    // Ginルーター作成
    r := gin.Default()

    r.GET("/", func(c *gin.Context) {
        c.JSON(200, gin.H{
            "message": "Hello from Delivery API!",
        })
    })

    // ポート8080で起動
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }

    _ = db // 後のIssueで使う
    r.Run(":" + port)
}
```

### 4. Dockerfile の作成

プロジェクトルートに `Dockerfile` を作成する。

```dockerfile
# ビルドステージ
FROM golang:1.21-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod ./
RUN go mod tidy && go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /delivery-api ./cmd/server

# 実行ステージ
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

COPY --from=builder /delivery-api .

EXPOSE 8080

CMD ["./delivery-api"]
```

> 💡 **ポイント**: マルチステージビルドを使うことで、最終イメージのサイズを大幅に削減できる。ビルドに必要なGoツールチェーンは実行ステージには含まれない。

### 5. docker-compose.yml の作成

```yaml
services:
  db:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: rootpassword
      MYSQL_DATABASE: delivery_db
      MYSQL_USER: delivery_user
      MYSQL_PASSWORD: delivery_pass
    ports:
      - "3306:3306"
    volumes:
      - mysql_data:/var/lib/mysql
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost"]
      interval: 10s
      timeout: 5s
      retries: 5

  api:
    build: .
    ports:
      - "8080:8080"
    environment:
      DB_USER: delivery_user
      DB_PASSWORD: delivery_pass
      DB_HOST: db
      DB_PORT: "3306"
      DB_NAME: delivery_db
      PORT: "8080"
    depends_on:
      db:
        condition: service_healthy

volumes:
  mysql_data:
```

> 💡 **ポイント**: `depends_on` に `condition: service_healthy` を指定することで、MySQLが完全に起動してからAPIサーバーが起動する。これがないとDB接続エラーになることがある。

### 6. .gitignore の作成

```
# 環境変数
.env

# バイナリ
delivery-api
*.exe

# IDE
.idea/
.vscode/
*.swp

# OS
.DS_Store
```

### 7. .env ファイルの作成（ローカル開発用）

```
DB_USER=delivery_user
DB_PASSWORD=delivery_pass
DB_HOST=db
DB_PORT=3306
DB_NAME=delivery_db
PORT=8080
```

## ✅ 動作確認

```bash
# ビルド & 起動
docker compose up --build

# 別ターミナルで確認
curl http://localhost:8080/
# => {"message":"Hello from Delivery API!"}
```

## 🔍 トラブルシューティング

| 症状 | 原因 | 対処法 |
|------|------|--------|
| `DB接続に失敗しました` | MySQLがまだ起動してない | `docker compose down` → `docker compose up --build` |
| `git` not found | alpine に git がない | Dockerfile に `RUN apk add --no-cache git` があるか確認 |
| ポート3306が使えない | 他のMySQLが動いてる | `docker ps` で確認して停止 or ポート変更 |
| Docker build がキャッシュされる | 古いイメージが残ってる | `docker compose build --no-cache` |

## 📚 参考リンク

- [Docker Compose 公式ドキュメント](https://docs.docker.com/compose/)
- [Gin Web Framework](https://gin-gonic.com/docs/)
- [GORM ドキュメント](https://gorm.io/ja_JP/docs/)

## 🏷️ ラベル

`setup` `docker` `backend` `初心者向け`
