# 🛡️ Issue #011: 認証ミドルウェア（Cookie ベース）

## 📋 概要

認証が必要なAPIエンドポイントを保護するミドルウェアを実装する。
リクエストの Cookie から JWT を読み取り、検証して、ユーザー情報を後続のハンドラーに渡す。

## 🎯 ゴール

- `middleware/auth.go` で Cookie から JWT を読み取り・検証できる
- 認証が必要なルートに適用して、未認証ユーザーを弾ける
- 認証済みユーザーの情報（ID, Email, Role）をハンドラーで使えるようにする
- ミドルウェアという概念を理解する

## 📝 前提条件

- ✅ Issue #010（ユーザーログイン API）が完了していること

## 📝 タスク

### 1. ミドルウェアとは？

```
ミドルウェアなし:
リクエスト → Handler（処理して返す）

ミドルウェアあり:
リクエスト → Middleware（認証チェック）→ Handler（処理して返す）
                 ↓
            認証NGなら 401 を返して Handler には到達しない
```

> 💡 **リアルワールドの例え**:
> ミドルウェア = 建物の入口にいるセキュリティガード
> - ID カード（Cookie）を持っている → 通してもらえる（Handler に到達）
> - ID カードがない or 偽物 → 入口で止められる（401 Unauthorized）

### 2. Auth Middleware の作成

`internal/middleware/auth.go` を作成する。

```go
package middleware

import (
    "net/http"
    "os"

    "github.com/delivery-app/delivery-api/internal/util"
    "github.com/gin-gonic/gin"
)

// AuthMiddleware はCookieからJWTを読み取り、認証を検証するミドルウェア
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 1. Cookie からトークンを取得する
        tokenString, err := c.Cookie("auth_token")
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "認証が必要です"})
            c.Abort()  // ⭐ 後続の処理を実行しない
            return
        }

        // 2. トークンを検証する
        jwtSecret := os.Getenv("JWT_SECRET")
        claims, err := util.ParseToken(tokenString, jwtSecret)
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "トークンが無効です"})
            c.Abort()
            return
        }

        // 3. ユーザー情報をコンテキストにセットする
        c.Set("user_id", claims.UserID)
        c.Set("email", claims.Email)
        c.Set("role", claims.Role)

        // 4. 後続のハンドラーに処理を渡す
        c.Next()
    }
}
```

> 💡 **`c.Abort()` vs `return`**:
> - `return` だけだと、Ginのミドルウェアチェーンの**次のミドルウェア**が実行される可能性がある
> - `c.Abort()` を呼ぶと、**全ての後続処理をスキップ**する
> - 認証失敗時は必ず `c.Abort()` + `return` のセットで使う

> 💡 **`c.Set()` / `c.Get()` パターン**:
> ```go
> // ミドルウェアでセット
> c.Set("user_id", claims.UserID)
>
> // ハンドラーで取り出す
> userID, exists := c.Get("user_id")
> if !exists {
>     // user_id がセットされていない（ありえないが念のため）
> }
> ```
> `gin.Context` はリクエストスコープのキーバリューストア。
> ミドルウェアとハンドラー間でデータを受け渡しできる。

### 3. ロールチェックミドルウェア（オプション）

特定のロールだけがアクセスできるミドルウェアも作成する。

```go
// RoleMiddleware は指定されたロールのみアクセスを許可するミドルウェア
func RoleMiddleware(allowedRoles ...string) gin.HandlerFunc {
    return func(c *gin.Context) {
        role, exists := c.Get("role")
        if !exists {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "認証が必要です"})
            c.Abort()
            return
        }

        // ロールが許可リストに含まれているかチェック
        roleStr, ok := role.(string)
        if !ok {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "ロール情報の取得に失敗しました"})
            c.Abort()
            return
        }

        for _, allowed := range allowedRoles {
            if roleStr == allowed {
                c.Next()
                return
            }
        }

        c.JSON(http.StatusForbidden, gin.H{"error": "この操作を行う権限がありません"})
        c.Abort()
    }
}
```

> 💡 **401 vs 403 の違い**:
> | コード | 意味 | 例 |
> |--------|------|-----|
> | 401 Unauthorized | 認証されていない（誰だかわからない）| ログインしてない |
> | 403 Forbidden | 認証済みだが権限がない | ログイン済みだが管理者じゃない |

### 4. main.go にミドルウェアを適用する

```go
import (
    // ... 既存のimport
    "github.com/delivery-app/delivery-api/internal/middleware"  // ← 追加
)

func main() {
    // ... （既存のコード）

    r := gin.Default()

    // ヘルスチェック（認証不要）
    healthHandler := handler.NewHealthHandler(db)
    r.GET("/health", healthHandler.Check)

    // 認証ルート（認証不要）
    auth := r.Group("/api/v1/auth")
    {
        auth.POST("/register", authHandler.Register)
        auth.POST("/login", authHandler.Login)
    }

    // ⭐ 認証が必要なルート
    api := r.Group("/api/v1")
    api.Use(middleware.AuthMiddleware())  // ← このグループ配下は全て認証必須
    {
        // ここに認証が必要なエンドポイントを追加していく（次のIssueで）
    }

    // ...
}
```

> 💡 **ルートグループの設計**:
> ```
> /health                    ← 認証不要（死活監視）
> /api/v1/auth/register      ← 認証不要（まだログインしてない）
> /api/v1/auth/login         ← 認証不要（まだログインしてない）
> /api/v1/users/me           ← 認証必要 ⭐
> /api/v1/auth/logout        ← 認証必要 ⭐
> /api/v1/trips              ← 認証必要 ⭐（将来）
> ```

### 5. ディレクトリ構成の確認

```
delivery-api/
├── cmd/
│   └── server/
│       └── main.go              ← ミドルウェア適用
├── internal/
│   ├── handler/
│   │   ├── health.go
│   │   └── auth_handler.go
│   ├── middleware/
│   │   └── auth.go              ← ⭐ 今回作成
│   ├── model/
│   │   └── user.go
│   ├── repository/
│   │   └── user_repository.go
│   └── util/
│       ├── password.go
│       └── jwt.go
├── .env
├── docker-compose.yml
├── Dockerfile
├── go.mod
└── go.sum
```

## ✅ 動作確認

```bash
docker compose up --build
```

### テスト1: Cookie なしで保護されたルートにアクセス

```bash
curl http://localhost:8080/api/v1/users/me
```

### 🔴 期待されるレスポンス（まだルート自体はないので 404 になるかもしれない）

次の Issue でプロフィールエンドポイントを追加した後に確認する。
今の段階では **ビルドが通ること** を確認できればOK。

### テスト2: ミドルウェアの動作を一時的に確認する方法

`main.go` にテスト用のエンドポイントを一時的に追加してみる：

```go
api := r.Group("/api/v1")
api.Use(middleware.AuthMiddleware())
{
    // テスト用（確認後に削除してOK）
    api.GET("/test-auth", func(c *gin.Context) {
        userID, _ := c.Get("user_id")
        email, _ := c.Get("email")
        role, _ := c.Get("role")
        c.JSON(200, gin.H{
            "message": "認証成功！",
            "user_id": userID,
            "email":   email,
            "role":    role,
        })
    })
}
```

```bash
# まずログインして Cookie を保存
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "yamada@example.com", "password": "mypassword123"}' \
  -c cookies.txt

# Cookie を使って保護されたルートにアクセス
curl http://localhost:8080/api/v1/test-auth -b cookies.txt
```

### 🟢 期待されるレスポンス

```json
{
  "message": "認証成功！",
  "user_id": 1,
  "email": "yamada@example.com",
  "role": "shipper"
}
```

## 🧪 追加チャレンジ（余裕があれば）

- [ ] 有効期限切れのトークンを手動で Cookie にセットして、弾かれることを確認してみよう
- [ ] `RoleMiddleware("admin")` を適用したルートを作って、shipper でアクセスしたら 403 になることを確認してみよう
- [ ] ミドルウェアの実行順序を確認するために、`log.Println()` をミドルウェアの先頭に追加してみよう

## 📖 学習ポイント

| 概念 | 説明 |
|------|------|
| ミドルウェア | リクエスト処理のパイプラインに挟む関数。認証、ログ、CORS など |
| `gin.HandlerFunc` | Gin のハンドラー関数型。ミドルウェアも同じ型 |
| `c.Abort()` | ミドルウェアチェーンを中断する |
| `c.Set()` / `c.Get()` | リクエストスコープのデータ受け渡し |
| `c.Cookie()` | リクエストの Cookie を読み取る |
| `api.Use()` | ルートグループ全体にミドルウェアを適用する |
| 401 vs 403 | 認証エラー vs 認可（権限）エラー |

## 🏷️ ラベル

`feature` `backend` `middleware` `auth` `初心者向け`
