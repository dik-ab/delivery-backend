# 📝 Issue #009: ユーザー登録 API（POST /api/v1/auth/register）

## 📋 概要

ユーザーが新しいアカウントを作成できる登録エンドポイントを実装する。
これまで作った Model → Repository → Util を初めて**組み合わせて** 1つの API を完成させる。

## 🎯 ゴール

- `POST /api/v1/auth/register` でユーザー登録ができる
- パスワードが bcrypt でハッシュ化されてDBに保存される
- 登録成功時、JWT が **HTTP-only Cookie** にセットされる
- 同じメールアドレスで2回登録するとエラーになる

## 📝 前提条件

- ✅ Issue #008（パスワードハッシュ & JWT ユーティリティ）が完了していること

## 📝 タスク

### 1. リクエスト & レスポンスの定義を理解する

```
リクエスト (POST /api/v1/auth/register):
{
    "email": "yamada@example.com",
    "password": "mypassword123",
    "name": "山田太郎",
    "role": "shipper",
    "company": "山田運送",
    "phone": "090-1234-5678"
}

レスポンス (201 Created):
{
    "user": {
        "id": 1,
        "email": "yamada@example.com",
        "name": "山田太郎",
        "role": "shipper",
        "company": "山田運送",
        "phone": "090-1234-5678",
        "created_at": "2025-01-01T00:00:00Z",
        "updated_at": "2025-01-01T00:00:00Z"
    }
}

+ Set-Cookie: auth_token=eyJhbGci...; Path=/; HttpOnly; Max-Age=86400
```

> 💡 **注目ポイント**:
> - リクエストの `password` はレスポンスに**含まれない**（`json:"-"` のおかげ）
> - JWT は**レスポンスボディではなくCookieに入る**

### 2. Auth Handler の作成

`internal/handler/auth_handler.go` を作成する。

```go
package handler

import (
    "net/http"
    "os"

    "github.com/delivery-app/delivery-api/internal/model"
    "github.com/delivery-app/delivery-api/internal/repository"
    "github.com/delivery-app/delivery-api/internal/util"
    "github.com/gin-gonic/gin"
)

// AuthHandler は認証関連のリクエストを処理する
type AuthHandler struct {
    userRepo *repository.UserRepository
}

// NewAuthHandler は新しい AuthHandler を生成する
func NewAuthHandler(userRepo *repository.UserRepository) *AuthHandler {
    return &AuthHandler{
        userRepo: userRepo,
    }
}

// RegisterRequest はユーザー登録リクエストの型
type RegisterRequest struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=6"`
    Name     string `json:"name" binding:"required"`
    Role     string `json:"role" binding:"required,oneof=shipper transport_company"`
    Company  string `json:"company"`
    Phone    string `json:"phone"`
}

// Register はユーザー登録を処理する
func (h *AuthHandler) Register(c *gin.Context) {
    // 1. リクエストボディをパースする
    var req RegisterRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // 2. メールアドレスの重複チェック
    existingUser, err := h.userRepo.GetByEmail(req.Email)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "データベースエラー"})
        return
    }
    if existingUser != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "このメールアドレスは既に登録されています"})
        return
    }

    // 3. パスワードをハッシュ化する
    hashedPassword, err := util.HashPassword(req.Password)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "パスワードのハッシュ化に失敗しました"})
        return
    }

    // 4. ユーザーを作成する
    user := &model.User{
        Email:        req.Email,
        PasswordHash: hashedPassword,
        Name:         req.Name,
        Role:         req.Role,
        Company:      req.Company,
        Phone:        req.Phone,
    }

    if err := h.userRepo.Create(user); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "ユーザーの作成に失敗しました"})
        return
    }

    // 5. JWTトークンを生成する
    jwtSecret := os.Getenv("JWT_SECRET")
    token, err := util.GenerateToken(user.ID, user.Email, user.Role, jwtSecret)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "トークンの生成に失敗しました"})
        return
    }

    // 6. HTTP-only Cookie にトークンをセットする
    c.SetCookie(
        "auth_token",  // Cookie名
        token,         // 値（JWTトークン）
        86400,         // 有効期限（秒）= 24時間
        "/",           // パス（全てのパスで有効）
        "",            // ドメイン（空 = 現在のドメイン）
        false,         // Secure（HTTPS限定）※ 開発中は false、本番は true
        true,          // HttpOnly ⭐ JavaScriptからアクセス不可
    )

    // 7. レスポンスを返す（トークンはCookieに入っているのでボディには含めない）
    c.JSON(http.StatusCreated, gin.H{
        "user": user,
    })
}
```

> 💡 **HTTP-only Cookie とは？**
> ```
> 通常のCookie:
>   ブラウザのJS → document.cookie でアクセスできる → XSS攻撃で盗まれるリスク
>
> HTTP-only Cookie:
>   ブラウザのJS → document.cookie でアクセスできない！
>   ブラウザが自動的にリクエストヘッダに付けてくれる → XSS攻撃に強い
> ```
>
> **`c.SetCookie` の引数解説**:
> | 引数 | 値 | 意味 |
> |------|-----|------|
> | name | `"auth_token"` | Cookieの名前 |
> | value | token | JWTトークン文字列 |
> | maxAge | `86400` | 有効期限（24時間 = 86400秒）|
> | path | `"/"` | 全パスで有効 |
> | domain | `""` | 現在のドメイン |
> | secure | `false` | 開発中はHTTPでもOK（本番では `true` にする）|
> | httpOnly | `true` | ⭐ **JSからアクセス不可** |

> 💡 **`binding` タグ解説**:
> | タグ | 意味 |
> |------|------|
> | `binding:"required"` | 必須フィールド。空だと 400 エラー |
> | `binding:"required,email"` | 必須 + メール形式のバリデーション |
> | `binding:"required,min=6"` | 必須 + 最低6文字 |
> | `binding:"required,oneof=shipper transport_company"` | 必須 + 指定値のどれか |

### 3. main.go にルーティングを追加

`cmd/server/main.go` を修正する。

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
    "github.com/delivery-app/delivery-api/internal/model"
    "github.com/delivery-app/delivery-api/internal/repository"  // ← 追加
)

func main() {
    // ... （既存のDB接続 & AutoMigrate コード）

    // ⭐ Repository を生成
    userRepo := repository.NewUserRepository(db)

    // ⭐ Handler を生成
    authHandler := handler.NewAuthHandler(userRepo)

    r := gin.Default()

    // ヘルスチェック（既存）
    healthHandler := handler.NewHealthHandler(db)
    r.GET("/health", healthHandler.Check)

    // ⭐ 認証ルートを追加
    auth := r.Group("/api/v1/auth")
    {
        auth.POST("/register", authHandler.Register)
    }

    // ... （既存のサーバー起動コード）
}
```

> 💡 **`r.Group("/api/v1/auth")` とは？**
> ルートをグループ化する機能。`/api/v1/auth` を共通のプレフィックスにして、
> 配下に `/register`, `/login` などを追加していく。
>
> ```
> /api/v1/auth/register  ← 今回
> /api/v1/auth/login     ← 次のIssue
> /api/v1/auth/logout    ← その次
> ```

### 4. docker-compose.yml に JWT_SECRET を追加

```yaml
services:
  api:
    build: .
    ports:
      - "8080:8080"
    environment:
      - DB_USER=root
      - DB_PASSWORD=password
      - DB_HOST=db
      - DB_PORT=3306
      - DB_NAME=delivery_db
      - JWT_SECRET=${JWT_SECRET}  # ← 追加（.env から読み込む）
    depends_on:
      - db
```

### 5. ディレクトリ構成の確認

```
delivery-api/
├── cmd/
│   └── server/
│       └── main.go              ← ルーティング追加
├── internal/
│   ├── handler/
│   │   ├── health.go
│   │   └── auth_handler.go      ← ⭐ 今回作成
│   ├── model/
│   │   └── user.go
│   ├── repository/
│   │   └── user_repository.go
│   └── util/
│       ├── password.go
│       └── jwt.go
├── .env                         ← JWT_SECRET
├── docker-compose.yml           ← JWT_SECRET 追加
├── Dockerfile
├── go.mod
└── go.sum
```

## ✅ 動作確認

```bash
# 再ビルド & 起動
docker compose up --build
```

### テスト1: 正常な登録

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "yamada@example.com",
    "password": "mypassword123",
    "name": "山田太郎",
    "role": "shipper",
    "company": "山田運送",
    "phone": "090-1234-5678"
  }' \
  -v
```

### 🟢 期待されるレスポンス

```
< HTTP/1.1 201 Created
< Set-Cookie: auth_token=eyJhbGciOiJIUzI1NiI...; Path=/; HttpOnly; Max-Age=86400

{
  "user": {
    "id": 1,
    "email": "yamada@example.com",
    "name": "山田太郎",
    "role": "shipper",
    "company": "山田運送",
    "phone": "090-1234-5678",
    "created_at": "2025-01-01T00:00:00Z",
    "updated_at": "2025-01-01T00:00:00Z"
  }
}
```

> ⭐ `-v`（verbose）オプションで `Set-Cookie` ヘッダが表示される！
> `auth_token=eyJ...` という JWT と `HttpOnly` が含まれていることを確認しよう。

### テスト2: 同じメールで重複登録（エラー）

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "yamada@example.com",
    "password": "password456",
    "name": "山田次郎",
    "role": "shipper"
  }'
```

### 🔴 期待されるレスポンス

```json
{
  "error": "このメールアドレスは既に登録されています"
}
```

### テスト3: バリデーションエラー

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "not-an-email",
    "password": "12345",
    "name": ""
  }'
```

### 🔴 期待されるレスポンス

バリデーションエラーが返ってくる（email形式不正、password短すぎ、name未入力）。

## 🧪 追加チャレンジ（余裕があれば）

- [ ] curlの `-c cookies.txt` オプションで Cookie をファイルに保存してみよう
- [ ] DBに保存された `password_hash` が平文でないことを MySQL で確認してみよう
- [ ] `role` に `"admin"` を指定して登録しようとしたらどうなる？

## 📖 学習ポイント

| 概念 | 説明 |
|------|------|
| HTTP-only Cookie | JavaScriptからアクセスできないCookie。XSS攻撃対策として有効 |
| `c.SetCookie()` | Gin でレスポンスにCookieをセットするメソッド |
| `c.ShouldBindJSON()` | リクエストボディのJSONをGoの構造体にマッピングする |
| `binding` タグ | バリデーションルールを struct タグで宣言的に定義する |
| `r.Group()` | 共通プレフィックスでルートをグループ化する |
| Handler → Repository → Model | 各層の責務を分離するアーキテクチャ |

## 🏷️ ラベル

`feature` `backend` `api` `auth` `初心者向け`
