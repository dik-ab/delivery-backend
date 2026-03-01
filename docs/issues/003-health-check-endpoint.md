# 🏥 Issue #003: Health Check エンドポイントの実装

## 📋 概要

APIサーバーとデータベースの死活監視ができる `/health` エンドポイントを実装する。
このエンドポイントは、サーバーが正常に動作しているか・DBに接続できているかを返す。

## 🎯 ゴール

- `GET /health` でAPIサーバーの状態を確認できる
- DBへの疎通確認も含める
- curl で叩いて正常なレスポンスが返ってくる

## 📝 前提条件

- ✅ Issue #002（Go + MySQL の Docker Compose 環境構築）が完了していること

## 📝 タスク

### 1. ハンドラーの作成

`internal/handler/health.go` を作成する。

```go
package handler

import (
    "net/http"

    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
)

// HealthHandler はヘルスチェック用のハンドラー
type HealthHandler struct {
    db *gorm.DB
}

// NewHealthHandler は新しい HealthHandler を生成する
func NewHealthHandler(db *gorm.DB) *HealthHandler {
    return &HealthHandler{db: db}
}

// Check はAPIサーバーとDBの状態を返す
func (h *HealthHandler) Check(c *gin.Context) {
    // DBの疎通確認
    sqlDB, err := h.db.DB()
    if err != nil {
        c.JSON(http.StatusServiceUnavailable, gin.H{
            "status":   "unhealthy",
            "database": "disconnected",
            "error":    err.Error(),
        })
        return
    }

    // Ping で実際にDBに接続できるか確認
    if err := sqlDB.Ping(); err != nil {
        c.JSON(http.StatusServiceUnavailable, gin.H{
            "status":   "unhealthy",
            "database": "unreachable",
            "error":    err.Error(),
        })
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "status":   "healthy",
        "database": "connected",
    })
}
```

> 💡 **ポイント**:
> - `gorm.DB` から `sql.DB` を取得して `Ping()` を呼ぶことで、実際にDBに通信できるかを確認している
> - ステータスコードでも健全性を判断できるようにしている（200 = OK、503 = NG）

### 2. main.go にルーティングを追加

`cmd/server/main.go` を修正して、`/health` ルートを追加する。

```go
package main

import (
    "fmt"
    "log"
    "os"

    "github.com/gin-gonic/gin"
    "gorm.io/driver/mysql"
    "gorm.io/gorm"

    "github.com/delivery-app/delivery-api/internal/handler"
)

func main() {
    dbUser := os.Getenv("DB_USER")
    dbPassword := os.Getenv("DB_PASSWORD")
    dbHost := os.Getenv("DB_HOST")
    dbPort := os.Getenv("DB_PORT")
    dbName := os.Getenv("DB_NAME")

    dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
        dbUser, dbPassword, dbHost, dbPort, dbName,
    )

    db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatal("❌ DB接続に失敗しました:", err)
    }

    log.Println("✅ DB接続成功！")

    r := gin.Default()

    // ヘルスチェック
    healthHandler := handler.NewHealthHandler(db)
    r.GET("/health", healthHandler.Check)

    r.GET("/", func(c *gin.Context) {
        c.JSON(200, gin.H{
            "message": "Hello from Delivery API!",
        })
    })

    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
    r.Run(":" + port)
}
```

### 3. ディレクトリ構成の確認

この時点でのプロジェクト構成：

```
delivery-api/
├── cmd/
│   └── server/
│       └── main.go          ← エントリーポイント
├── internal/
│   └── handler/
│       └── health.go        ← ⭐ 今回作成
├── docker-compose.yml
├── Dockerfile
├── go.mod
├── go.sum
├── .env
└── .gitignore
```

> 💡 **ポイント**: `internal/` ディレクトリ配下のパッケージは、このモジュール外からインポートできない。Goの慣習で「外部に公開しない内部実装」を置く場所として使われる。

## ✅ 動作確認

```bash
# 再ビルド & 起動
docker compose up --build

# 別ターミナルでヘルスチェック
curl http://localhost:8080/health
```

### 🟢 正常時のレスポンス

```json
{
  "status": "healthy",
  "database": "connected"
}
```

### 🔴 DB接続失敗時のレスポンス

```json
{
  "status": "unhealthy",
  "database": "unreachable",
  "error": "dial tcp: connect: connection refused"
}
```

## 🧪 追加チャレンジ（余裕があれば）

- [ ] レスポンスに `uptime`（サーバー起動からの経過時間）を追加してみよう
- [ ] レスポンスに `version` フィールドを追加してみよう（ハードコードでOK）
- [ ] `internal/handler/health_test.go` にテストを書いてみよう

## 📖 学習ポイント

| 概念 | 説明 |
|------|------|
| `gin.H{}` | `map[string]interface{}` のエイリアス。JSONレスポンスを簡単に書ける |
| `http.StatusOK` | `net/http` パッケージの定数。`200` と同じだけど可読性が高い |
| `sqlDB.Ping()` | 実際にDBにパケットを送って疎通確認する |
| `internal/` | Go のアクセス制御。このフォルダ配下は外部モジュールからアクセス不可 |
| Clean Architecture | handler → repository → model の層構造。今回は handler 層のみ |

## 🏷️ ラベル

`feature` `backend` `api` `初心者向け`
