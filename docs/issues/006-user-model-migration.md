# 👤 Issue #006: User Model & データベースマイグレーション

## 📋 概要

ユーザー認証の土台となる `User` モデルを定義し、GORMの `AutoMigrate` を使って MySQL にテーブルを自動作成する。

## 🎯 ゴール

- `model/user.go` に User 構造体を定義する
- `main.go` で `AutoMigrate` を実行し、`users` テーブルが自動作成される
- Docker Compose 起動後、MySQL にテーブルが存在することを確認できる

## 📝 前提条件

- ✅ Issue #003（Health Check エンドポイント）が完了していること

## 📝 タスク

### 1. User モデルの作成

`internal/model/user.go` を作成する。

```go
package model

import "time"

// User はプラットフォームのユーザーを表す
type User struct {
    ID           uint      `json:"id" gorm:"primaryKey"`
    Email        string    `json:"email" gorm:"type:varchar(255);uniqueIndex" binding:"required"`
    PasswordHash string    `json:"-"`
    Name         string    `json:"name" binding:"required"`
    Role         string    `json:"role" gorm:"default:shipper"`
    Company      string    `json:"company"`
    Phone        string    `json:"phone"`
    CreatedAt    time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`
}
```

> 💡 **structタグ解説**:
> | タグ | 意味 |
> |------|------|
> | `json:"id"` | JSONのキー名を指定。APIレスポンスに使われる |
> | `json:"-"` | JSONに**含めない**。パスワードハッシュは絶対にレスポンスに出さない |
> | `gorm:"primaryKey"` | このフィールドをPRIMARY KEYにする |
> | `gorm:"type:varchar(255);uniqueIndex"` | DBの型を明示指定 + ユニーク制約 |
> | `gorm:"default:shipper"` | INSERTで値が空の時のデフォルト値 |
> | `binding:"required"` | Ginのバリデーション。リクエストに必須 |

> 💡 **なぜ `Password` ではなく `PasswordHash` なの？**
> DB に保存するのは**平文のパスワードではなく、bcrypt でハッシュ化した後の文字列**。
> フィールド名で「これはハッシュ済みの値ですよ」と明示しておくことが大事。
>
> ```
> ❌ Password string     ← 「これ平文？ハッシュ済み？」と後から混乱する
> ✅ PasswordHash string ← 「ハッシュ化済みの値が入ってるんだな」と一目でわかる
> ```
>
> 実際の流れ:
> 1. ユーザーが `"mypassword123"` を送ってくる
> 2. サーバー側で bcrypt にかけて `"$2a$10$N9qo8uLOi..."` に変換
> 3. その **ハッシュ化された文字列** を `PasswordHash` に保存
>
> DB には元のパスワードは一切保存されない。万が一 DB が流出しても、元のパスワードは復元できない。
> ハッシュ化の詳細は Issue #008 で学ぶ。

> ⚠️ **なぜ `gorm:"type:varchar(255)"` が必要？**
> GORMはGoの `string` をMySQLの `longtext` にマッピングする。
> `longtext` にはUNIQUE制約を付けられないため、明示的に `varchar(255)` を指定する。
> これを忘れると `Error 1170: BLOB/TEXT column used in key specification without a key length` というエラーになる。

### 2. main.go に AutoMigrate を追加

`cmd/server/main.go` に以下を追加する。

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
    "github.com/delivery-app/delivery-api/internal/model"  // ← 追加
)

func main() {
    // ... （既存のDB接続コード）

    db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatal("❌ DB接続に失敗しました:", err)
    }

    log.Println("✅ DB接続成功！")

    // ⭐ AutoMigrate を追加
    if err := db.AutoMigrate(&model.User{}); err != nil {
        log.Fatal("❌ マイグレーション失敗:", err)
    }
    log.Println("✅ マイグレーション完了！")

    // ... （既存のルーティングコード）
}
```

> 💡 **AutoMigrateとは？**
> GORMの機能で、Go の構造体定義を元にDBテーブルを**自動で作成・更新**する。
> テーブルがなければ CREATE TABLE し、カラムが増えていれば ALTER TABLE する。
> ただし**カラムの削除はしない**（安全のため）。

### 3. ディレクトリ構成の確認

この時点でのプロジェクト構成：

```
delivery-api/
├── cmd/
│   └── server/
│       └── main.go              ← AutoMigrate 追加
├── internal/
│   ├── handler/
│   │   └── health.go
│   └── model/
│       └── user.go              ← ⭐ 今回作成
├── docker-compose.yml
├── Dockerfile
├── go.mod
├── go.sum
├── .env
└── .gitignore
```

## ✅ 動作確認

```bash
# 再ビルド & 起動
docker compose up --build
```

起動ログに以下が表示されればOK：

```
✅ DB接続成功！
✅ マイグレーション完了！
```

さらに、MySQLに接続してテーブルを確認する：

```bash
# MySQLコンテナに入る
docker compose exec db mysql -u root -p delivery_db

# テーブル一覧
SHOW TABLES;

# users テーブルの構造を確認
DESCRIBE users;
```

### 🟢 期待される出力

```
+---------------+--------------+------+-----+---------+----------------+
| Field         | Type         | Null | Key | Default | Extra          |
+---------------+--------------+------+-----+---------+----------------+
| id            | bigint unsigned | NO  | PRI | NULL   | auto_increment |
| email         | varchar(255) | YES  | UNI | NULL    |                |
| password_hash | longtext     | YES  |     | NULL    |                |
| name          | longtext     | YES  |     | NULL    |                |
| role          | longtext     | YES  |     | shipper |                |
| company       | longtext     | YES  |     | NULL    |                |
| phone         | longtext     | YES  |     | NULL    |                |
| created_at    | datetime(3)  | YES  |     | NULL    |                |
| updated_at    | datetime(3)  | YES  |     | NULL    |                |
+---------------+--------------+------+-----+---------+----------------+
```

## 🧪 追加チャレンジ（余裕があれば）

- [ ] `TableName()` メソッドを定義して、テーブル名を明示的に指定してみよう
- [ ] `phone` フィールドにも `gorm:"type:varchar(20)"` を指定してみよう
- [ ] `CreatedAt` が自動的に現在時刻になることを確認してみよう

## 📖 学習ポイント

| 概念 | 説明 |
|------|------|
| GORMモデル | Go の構造体がDBテーブルの定義になる |
| AutoMigrate | 構造体からテーブルを自動作成する仕組み |
| struct tag | `gorm:""` や `json:""` でDBマッピングやJSON変換を制御 |
| `json:"-"` | パスワードなど秘密情報をAPIレスポンスから除外するための仕組み |
| varchar vs longtext | MySQLの型。ユニーク制約には `varchar` を使う必要がある |

## 🏷️ ラベル

`feature` `backend` `database` `model` `初心者向け`
