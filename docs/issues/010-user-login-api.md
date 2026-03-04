# 🔑 Issue #010: ユーザーログイン API（POST /api/v1/auth/login）

## 📋 概要

登録済みユーザーがメールアドレスとパスワードでログインできるエンドポイントを実装する。
ログイン成功時、JWT を HTTP-only Cookie にセットする。

## 🎯 ゴール

- `POST /api/v1/auth/login` でログインできる
- 正しいメール & パスワードで JWT Cookie がセットされる
- 間違ったメールまたはパスワードではエラーが返る
- エラーメッセージから「メールが間違い」か「パスワードが間違い」かを推測できないようにする

## 📝 前提条件

- ✅ Issue #009（ユーザー登録 API）が完了していること
- ✅ テストユーザーが1人以上登録されていること

## 📝 タスク

### 1. Auth Handler にログインメソッドを追加

`internal/handler/auth_handler.go` に以下を **追記** する。

```go
// LoginRequest はログインリクエストの型
type LoginRequest struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required"`
}

// Login はユーザーログインを処理する
func (h *AuthHandler) Login(c *gin.Context) {
    // 1. リクエストボディをパースする
    var req LoginRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // 2. メールアドレスでユーザーを検索する
    user, err := h.userRepo.GetByEmail(req.Email)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "データベースエラー"})
        return
    }
    if user == nil {
        // ⭐ 「メールが見つからない」とは言わない（セキュリティ上の理由）
        c.JSON(http.StatusUnauthorized, gin.H{"error": "メールアドレスまたはパスワードが正しくありません"})
        return
    }

    // 3. パスワードを検証する
    if err := util.CheckPassword(req.Password, user.PasswordHash); err != nil {
        // ⭐ 「パスワードが違う」とは言わない（セキュリティ上の理由）
        c.JSON(http.StatusUnauthorized, gin.H{"error": "メールアドレスまたはパスワードが正しくありません"})
        return
    }

    // 4. JWTトークンを生成する
    jwtSecret := os.Getenv("JWT_SECRET")
    token, err := util.GenerateToken(user.ID, user.Email, user.Role, jwtSecret)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "トークンの生成に失敗しました"})
        return
    }

    // 5. HTTP-only Cookie にトークンをセットする
    c.SetCookie(
        "auth_token",
        token,
        86400,   // 24時間
        "/",
        "",
        false,   // 開発中は false
        true,    // HttpOnly
    )

    // 6. レスポンスを返す
    c.JSON(http.StatusOK, gin.H{
        "user": user,
    })
}
```

> ⚠️ **なぜエラーメッセージを曖昧にするのか？**
>
> ```
> ❌ 悪い例:
> "このメールアドレスは登録されていません"  ← 攻撃者:「このメールは存在しない」
> "パスワードが間違っています"             ← 攻撃者:「メールは合ってる、パスワードだけ変えよう」
>
> ✅ 良い例:
> "メールアドレスまたはパスワードが正しくありません" ← 攻撃者:「どっちが間違いかわからない」
> ```
>
> これを **タイミング攻撃対策** と組み合わせるとさらに安全だが、ポートフォリオではメッセージ統一だけでOK。

### 2. main.go にルートを追加

`cmd/server/main.go` の auth グループに1行追加する。

```go
auth := r.Group("/api/v1/auth")
{
    auth.POST("/register", authHandler.Register)
    auth.POST("/login", authHandler.Login)        // ← 追加
}
```

### 3. 処理の流れを理解する

```
ブラウザ/curl
    │
    │ POST /api/v1/auth/login
    │ { "email": "...", "password": "..." }
    │
    ▼
┌─────────────────────────────────────┐
│  auth_handler.go (Login)            │
│                                     │
│  1. リクエストをパース               │
│  2. userRepo.GetByEmail() でDB検索   │──→ repository/user_repository.go
│  3. util.CheckPassword() で検証      │──→ util/password.go
│  4. util.GenerateToken() でJWT生成   │──→ util/jwt.go
│  5. c.SetCookie() でCookieにセット   │
│  6. レスポンスを返す                 │
└─────────────────────────────────────┘
    │
    ▼
ブラウザ: Set-Cookie ヘッダを受け取り、以降のリクエストに自動でCookieを付ける
```

## ✅ 動作確認

```bash
docker compose up --build
```

### テスト1: 正常なログイン

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "yamada@example.com",
    "password": "mypassword123"
  }' \
  -v
```

### 🟢 期待されるレスポンス

```
< HTTP/1.1 200 OK
< Set-Cookie: auth_token=eyJhbGciOiJIUzI1NiI...; Path=/; HttpOnly; Max-Age=86400

{
  "user": {
    "id": 1,
    "email": "yamada@example.com",
    "name": "山田太郎",
    "role": "shipper",
    ...
  }
}
```

### テスト2: 間違ったパスワード

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "yamada@example.com",
    "password": "wrongpassword"
  }'
```

### 🔴 期待されるレスポンス

```json
{
  "error": "メールアドレスまたはパスワードが正しくありません"
}
```

### テスト3: 存在しないメールアドレス

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "nonexistent@example.com",
    "password": "mypassword123"
  }'
```

### 🔴 期待されるレスポンス（テスト2と同じメッセージ！）

```json
{
  "error": "メールアドレスまたはパスワードが正しくありません"
}
```

### テスト4: Cookie を保存して使ってみる

```bash
# Cookie をファイルに保存
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "yamada@example.com", "password": "mypassword123"}' \
  -c cookies.txt

# 保存した Cookie で保護されたエンドポイントにアクセス（次のIssueで使う）
curl http://localhost:8080/api/v1/users/me -b cookies.txt
```

> 💡 `-c cookies.txt` で Cookie を保存、`-b cookies.txt` で Cookie を送信する。
> ブラウザでは自動的にやってくれるが、curl では手動で指定する必要がある。

## 🧪 追加チャレンジ（余裕があれば）

- [ ] ブラウザのDevToolsで Cookie が `HttpOnly` になっていることを確認してみよう
- [ ] Cookie の `Max-Age` を5分（300）に変えて、5分後にアクセスしたらどうなるか確認してみよう
- [ ] `Register` と `Login` で Cookie セット部分が重複している → ヘルパー関数に切り出してみよう

## 📖 学習ポイント

| 概念 | 説明 |
|------|------|
| 認証（Authentication） | 「あなたは誰か？」を確認すること。メール + パスワードで本人確認 |
| 曖昧なエラーメッセージ | 攻撃者にヒントを与えないために、メール/パスワードどちらが間違いか明かさない |
| 401 Unauthorized | 「認証されていない」を示すHTTPステータスコード |
| Cookie の自動送信 | ブラウザはCookieを自動的にリクエストヘッダに含めてくれる（fetch の `credentials: 'include'` が必要）|
| `-c` / `-b` オプション | curl で Cookie を保存/送信するオプション |

## 🏷️ ラベル

`feature` `backend` `api` `auth` `初心者向け`
