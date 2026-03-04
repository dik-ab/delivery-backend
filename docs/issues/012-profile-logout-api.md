# 👤 Issue #012: プロフィール取得 & ログアウト API

## 📋 概要

認証済みユーザーが自分のプロフィール情報を取得するエンドポイントと、
ログアウト（Cookie 削除）するエンドポイントを実装する。
これが認証システムの最後のピースになる。

## 🎯 ゴール

- `GET /api/v1/users/me` で自分のプロフィールを取得できる
- `POST /api/v1/auth/logout` でログアウト（Cookie 削除）できる
- 認証の一連のフロー（登録 → ログイン → プロフィール取得 → ログアウト）が完成する

## 📝 前提条件

- ✅ Issue #011（認証ミドルウェア）が完了していること

## 📝 タスク

### 1. User Handler の作成

`internal/handler/user_handler.go` を新しく作成する。

```go
package handler

import (
    "net/http"

    "github.com/delivery-app/delivery-api/internal/repository"
    "github.com/gin-gonic/gin"
)

// UserHandler はユーザー関連のリクエストを処理する
type UserHandler struct {
    userRepo *repository.UserRepository
}

// NewUserHandler は新しい UserHandler を生成する
func NewUserHandler(userRepo *repository.UserRepository) *UserHandler {
    return &UserHandler{
        userRepo: userRepo,
    }
}

// GetMe は認証済みユーザーの自分のプロフィールを返す
func (h *UserHandler) GetMe(c *gin.Context) {
    // 1. ミドルウェアがセットした user_id を取得する
    userID, exists := c.Get("user_id")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "認証情報が見つかりません"})
        return
    }

    // 2. user_id を uint に型変換する
    id, ok := userID.(uint)
    if !ok {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "ユーザーIDの変換に失敗しました"})
        return
    }

    // 3. DB からユーザーを取得する
    user, err := h.userRepo.GetByID(id)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "データベースエラー"})
        return
    }
    if user == nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "ユーザーが見つかりません"})
        return
    }

    // 4. レスポンスを返す
    c.JSON(http.StatusOK, gin.H{
        "user": user,
    })
}
```

> 💡 **なぜ `auth_handler.go` と分けるのか？**
> - `auth_handler.go` → **認証の処理**（登録、ログイン、ログアウト）
> - `user_handler.go` → **ユーザー情報の処理**（プロフィール取得、更新など）
>
> 責務が違うので別ファイルにする。将来 `user_handler.go` には「プロフィール更新」なども追加できる。

> 💡 **`c.Get()` の型変換パターン**:
> ```go
> // c.Set("user_id", claims.UserID)  ← ミドルウェアでセットした値
> // c.Get() は interface{} 型で返ってくるので、型アサーション（type assertion）が必要
>
> value, exists := c.Get("user_id")  // value は interface{} 型
> id, ok := value.(uint)              // uint 型に変換。失敗したら ok = false
> ```

### 2. Auth Handler にログアウトメソッドを追加

`internal/handler/auth_handler.go` に以下を **追記** する。

```go
// Logout はログアウト（Cookie を削除）する
func (h *AuthHandler) Logout(c *gin.Context) {
    // Cookie を削除する（MaxAge を -1 にすると即削除される）
    c.SetCookie(
        "auth_token",
        "",      // 値を空にする
        -1,      // MaxAge = -1 で即削除
        "/",
        "",
        false,
        true,
    )

    c.JSON(http.StatusOK, gin.H{
        "message": "ログアウトしました",
    })
}
```

> 💡 **Cookie の削除方法**:
> HTTP には「Cookie を削除する」という命令はない。
> 代わりに、同じ名前の Cookie を `MaxAge = -1`（または過去の有効期限）でセットすると、
> ブラウザが自動的にその Cookie を削除する。
>
> ```
> Set-Cookie: auth_token=; Path=/; HttpOnly; Max-Age=-1
>                    ↑                        ↑
>                  値が空                   即削除
> ```

### 3. main.go にルーティングを追加

`cmd/server/main.go` を修正する。

```go
func main() {
    // ... （既存のコード）

    // Repository & Handler の生成
    userRepo := repository.NewUserRepository(db)
    authHandler := handler.NewAuthHandler(userRepo)
    userHandler := handler.NewUserHandler(userRepo)  // ← 追加

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

    // 認証が必要なルート
    api := r.Group("/api/v1")
    api.Use(middleware.AuthMiddleware())
    {
        // ユーザー
        api.GET("/users/me", userHandler.GetMe)       // ← 追加

        // 認証（ログアウトは認証が必要）
        api.POST("/auth/logout", authHandler.Logout)   // ← 追加
    }

    // ...
}
```

> 💡 **ログアウトが認証必要グループにある理由**:
> ログアウトするには「今ログインしている」必要がある。
> 未認証の人がログアウトしようとしても意味がないので、認証グループに置く。

### 4. ディレクトリ構成の確認（認証システム完成版）

```
delivery-api/
├── cmd/
│   └── server/
│       └── main.go                  ← ルーティング完成
├── internal/
│   ├── handler/
│   │   ├── health.go                ← Issue #003
│   │   ├── auth_handler.go          ← Issue #009, #010, #012
│   │   └── user_handler.go          ← ⭐ 今回作成
│   ├── middleware/
│   │   └── auth.go                  ← Issue #011
│   ├── model/
│   │   └── user.go                  ← Issue #006
│   ├── repository/
│   │   └── user_repository.go       ← Issue #007
│   └── util/
│       ├── password.go              ← Issue #008
│       └── jwt.go                   ← Issue #008
├── .env
├── docker-compose.yml
├── Dockerfile
├── go.mod
└── go.sum
```

### 5. 認証フロー全体像

```
┌────────────── 認証システム全体像 ──────────────┐
│                                                │
│  ① 登録: POST /api/v1/auth/register           │
│     → ユーザー作成 + Cookie セット              │
│                                                │
│  ② ログイン: POST /api/v1/auth/login           │
│     → パスワード検証 + Cookie セット            │
│                                                │
│  ③ プロフィール: GET /api/v1/users/me          │
│     → Cookie から認証 → ユーザー情報返却        │
│                                                │
│  ④ ログアウト: POST /api/v1/auth/logout        │
│     → Cookie 削除                              │
│                                                │
│  ┌──────────┐   Cookie    ┌──────────┐        │
│  │ ブラウザ  │ ─────────→ │   API    │        │
│  │          │ ←───────── │          │        │
│  └──────────┘  Set-Cookie └──────────┘        │
│                                                │
│  Cookie は自動で送信される                      │
│  → フロントエンドで token を管理する必要なし！   │
└────────────────────────────────────────────────┘
```

## ✅ 動作確認

```bash
docker compose up --build
```

### テスト: 認証フロー全体を通しで実行

```bash
# ① 登録
echo "=== 1. ユーザー登録 ==="
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123",
    "name": "テストユーザー",
    "role": "shipper"
  }' \
  -c cookies.txt \
  -s | jq .

# ② プロフィール取得（登録時の Cookie を使う）
echo "=== 2. プロフィール取得 ==="
curl http://localhost:8080/api/v1/users/me \
  -b cookies.txt \
  -s | jq .

# ③ ログアウト
echo "=== 3. ログアウト ==="
curl -X POST http://localhost:8080/api/v1/auth/logout \
  -b cookies.txt \
  -c cookies.txt \
  -s | jq .

# ④ ログアウト後にプロフィール取得（失敗するはず）
echo "=== 4. ログアウト後のアクセス ==="
curl http://localhost:8080/api/v1/users/me \
  -b cookies.txt \
  -s | jq .

# ⑤ ログイン
echo "=== 5. ログイン ==="
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }' \
  -c cookies.txt \
  -s | jq .

# ⑥ 再度プロフィール取得（成功するはず）
echo "=== 6. ログイン後のプロフィール取得 ==="
curl http://localhost:8080/api/v1/users/me \
  -b cookies.txt \
  -s | jq .
```

### 🟢 期待される結果

| ステップ | 結果 |
|----------|------|
| 1. 登録 | 201 Created + ユーザー情報 |
| 2. プロフィール取得 | 200 OK + ユーザー情報 |
| 3. ログアウト | 200 OK + "ログアウトしました" |
| 4. ログアウト後アクセス | 401 Unauthorized + "認証が必要です" |
| 5. ログイン | 200 OK + ユーザー情報 |
| 6. 再度プロフィール取得 | 200 OK + ユーザー情報 |

## 🧪 追加チャレンジ（余裕があれば）

- [ ] `PUT /api/v1/users/me` でプロフィール更新ができるようにしてみよう
- [ ] プロフィール更新では `email` と `password` は変更できないようにバリデーションを追加してみよう
- [ ] 上の全テストをシェルスクリプトにまとめてみよう（`test_auth.sh`）
- [ ] `jq` がインストールされていない場合に備えて、`python3 -m json.tool` で代用する方法も調べてみよう

## 📖 学習ポイント

| 概念 | 説明 |
|------|------|
| Cookie 削除 | `MaxAge = -1` でブラウザに Cookie を削除させる |
| Handler の分離 | 認証系（auth）と リソース系（user, trip など）は別ファイルにする |
| 型アサーション | `value.(uint)` で `interface{}` を具体的な型に変換する |
| 認証フロー | 登録 → ログイン → API利用 → ログアウト の一連の流れ |
| 責務の分離 | auth_handler = 認証処理、user_handler = ユーザー情報処理 |

## 🎉 認証システム完成！

ここまでで認証の基本システムが完成した。
Issue #006〜#012 を通じて、以下のことを学んだ：

```
Model（データ定義）
  ↓
Repository（DB操作）
  ↓
Util（共通処理: パスワード、JWT）
  ↓
Handler（リクエスト処理）
  ↓
Middleware（認証チェック）
  ↓
Router（URLとHandlerの対応付け）
```

次のステップでは、この認証システムを使って **Trip（運行情報）** の CRUD API を実装していく！

## 🏷️ ラベル

`feature` `backend` `api` `auth` `初心者向け`
