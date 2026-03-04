# 🔐 Issue #008: パスワードハッシュ & JWT ユーティリティ

## 📋 概要

認証の基盤となる2つのユーティリティを作成する：
1. **パスワードハッシュ**: bcrypt を使ったパスワードの暗号化と検証
2. **JWT**: トークンの生成と検証（後のIssueでHTTP-only Cookieに入れる）

## 🎯 ゴール

- `util/password.go` でパスワードのハッシュ化・検証ができる
- `util/jwt.go` でJWTトークンの生成・検証ができる
- なぜパスワードを平文で保存してはいけないか理解する
- JWTの構造を理解する

## 📝 前提条件

- ✅ Issue #007（User Repository）が完了していること

## 📝 タスク

### 1. 必要なパッケージのインストール

```bash
# コンテナ内で実行（または go.mod に追記して docker compose up --build）
go get golang.org/x/crypto/bcrypt
go get github.com/golang-jwt/jwt/v5
```

> 💡 `go get` は Go のパッケージマネージャ。`npm install` に相当する。
> `go.mod` と `go.sum` に依存関係が記録される。

### 2. パスワードハッシュユーティリティの作成

`internal/util/password.go` を作成する。

```go
package util

import "golang.org/x/crypto/bcrypt"

// HashPassword はパスワードを bcrypt でハッシュ化する
func HashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return "", err
    }
    return string(bytes), nil
}

// CheckPassword はパスワードとハッシュを比較する
// 一致すれば nil、不一致ならエラーを返す
func CheckPassword(password, hash string) error {
    return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
```

> ⚠️ **なぜパスワードをハッシュ化するのか？**
> もしDBが流出した場合、平文パスワードが保存されていると全ユーザーが被害を受ける。
> bcryptでハッシュ化しておけば、元のパスワードを復元することは**事実上不可能**。
>
> **bcryptの仕組み（ざっくり）**:
> ```
> "mypassword123"
>   ↓ bcrypt（salt + ストレッチング）
> "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"
> ```
> - **salt**: ランダムな文字列を混ぜることで、同じパスワードでも毎回違うハッシュになる
> - **ストレッチング**: 計算を意図的に遅くして、総当たり攻撃を困難にする
> - **DefaultCost = 10**: ストレッチングの回数。数字が大きいほど安全だが遅い

### 3. JWT ユーティリティの作成

`internal/util/jwt.go` を作成する。

```go
package util

import (
    "errors"
    "time"

    "github.com/golang-jwt/jwt/v5"
)

// Claims はJWTに含めるデータを定義する
type Claims struct {
    UserID uint   `json:"user_id"`
    Email  string `json:"email"`
    Role   string `json:"role"`
    jwt.RegisteredClaims
}

// GenerateToken はユーザー情報からJWTトークンを生成する
func GenerateToken(userID uint, email, role, secret string) (string, error) {
    // トークンの有効期限を24時間に設定
    expirationTime := time.Now().Add(24 * time.Hour)

    claims := &Claims{
        UserID: userID,
        Email:  email,
        Role:   role,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(expirationTime),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
        },
    }

    // HS256アルゴリズムでトークンを生成
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    tokenString, err := token.SignedString([]byte(secret))
    if err != nil {
        return "", err
    }

    return tokenString, nil
}

// ParseToken はJWTトークンを検証してClaimsを取り出す
func ParseToken(tokenString, secret string) (*Claims, error) {
    claims := &Claims{}
    token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
        return []byte(secret), nil
    })

    if err != nil {
        return nil, err
    }

    if !token.Valid {
        return nil, errors.New("invalid token")
    }

    return claims, nil
}
```

> 💡 **JWTとは？（JSON Web Token）**
> ユーザーの認証情報を安全にやりとりするための仕組み。
>
> ```
> JWTの構造（ドットで3つに分かれている）:
>
> eyJhbGciOiJIUzI1NiJ9.eyJ1c2VyX2lkIjoxfQ.abc123
> ├── Header（暗号方式）  ├── Payload（ユーザー情報）  ├── Signature（署名）
> ```
>
> - **Header**: `{"alg": "HS256"}` → 暗号化アルゴリズム
> - **Payload**: `{"user_id": 1, "email": "...", "role": "shipper"}` → 埋め込むデータ
> - **Signature**: Header + Payload + Secret から生成。改竄を検知する
>
> **重要**: Payload はBase64でエンコードされているだけで**暗号化されていない**。
> ブラウザのDevToolsで中身を見られるので、パスワードは絶対に入れない。

### 4. .env に JWT_SECRET を追加

`.env` ファイルに追加する。

```env
# 既存の設定
DB_USER=root
DB_PASSWORD=password
DB_HOST=db
DB_PORT=3306
DB_NAME=delivery_db

# ⭐ 追加
JWT_SECRET=your-super-secret-key-change-this-in-production
```

> ⚠️ **本番環境では必ずランダムで十分に長い文字列にすること！**
> 開発用に短い文字列を使うのは OK だが、本番では `openssl rand -hex 32` などで生成する。

### 5. ディレクトリ構成の確認

```
delivery-api/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── handler/
│   │   └── health.go
│   ├── model/
│   │   └── user.go
│   ├── repository/
│   │   └── user_repository.go
│   └── util/
│       ├── password.go          ← ⭐ 今回作成
│       └── jwt.go               ← ⭐ 今回作成
├── .env                         ← JWT_SECRET 追加
├── docker-compose.yml
├── Dockerfile
├── go.mod
└── go.sum
```

## ✅ 動作確認

ビルドが通ることを確認する。

```bash
docker compose up --build
```

> 💡 このIssueでもAPIエンドポイントは追加しない。
> ユーティリティ関数は、次のIssue（#009 ユーザー登録）で初めて使われる。
> **部品を先に作って、次のステップで組み合わせる** ← これが大事。

## 🧪 追加チャレンジ（余裕があれば）

- [ ] `main.go` で一時的にテストコードを書いて、ハッシュ化 → 検証の流れを確認してみよう
- [ ] JWTトークンを生成して、[jwt.io](https://jwt.io/) に貼り付けて中身を見てみよう
- [ ] `GenerateToken` の有効期限を 1時間（`1 * time.Hour`）に変えてみよう
- [ ] `Claims` に `Company` フィールドを追加してみよう（どこに影響が出る？）

## 📖 学習ポイント

| 概念 | 説明 |
|------|------|
| bcrypt | パスワードのハッシュ化に特化したアルゴリズム。salt + ストレッチング内蔵 |
| JWT | JSON Web Token。認証情報をトークンとして表現する標準規格（RFC 7519）|
| HS256 | HMAC-SHA256。共有秘密鍵でトークンに署名する方式 |
| Claims | JWTの Payload に含めるデータ。独自フィールドを追加できる |
| 環境変数 | パスワードやシークレットキーをコードに直書きせず、`.env` から読み込む |

## 🗺️ この Issue の位置づけ（JWT × HTTP-only Cookie 全体像）

この Issue (#008) では JWT の「生成」と「検証」の**部品だけ**を作る。
実際に HTTP-only Cookie として使うのは次の Issue 以降。

```
#008 (今回)  → JWT生成 & パスワードハッシュの「部品」を作る
                 ↓ 部品を使う
#009         → ユーザー登録時に c.SetCookie() で HTTP-only Cookie にJWTをセット
#010         → ログイン時にも同じく Cookie にセット
#011         → ミドルウェアで c.Cookie() を使い Cookie からJWTを読み取って認証
#012         → ログアウト時に MaxAge=-1 で Cookie を削除
```

**なぜ localStorage ではなく HTTP-only Cookie？**

```
localStorage に JWT を保存する場合:
  → JavaScript で読み書きできる（localStorage.getItem("token")）
  → XSS（クロスサイトスクリプティング）攻撃で JS が注入されると、トークンを盗まれる

HTTP-only Cookie に JWT を保存する場合:
  → JavaScript からアクセスできない（document.cookie でも見えない）
  → ブラウザが自動的にリクエストに Cookie を付けてくれる
  → XSS攻撃を受けてもトークンを盗まれない
```

ポートフォリオレベルでも HTTP-only Cookie で実装できると**セキュリティへの理解をアピールできる**。

## 🏷️ ラベル

`feature` `backend` `security` `auth` `初心者向け`
