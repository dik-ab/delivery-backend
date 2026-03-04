# 🗄️ Issue #007: User Repository（データベース操作層）

## 📋 概要

User モデルに対するデータベース操作（CRUD）をまとめた Repository 層を作成する。
Handler から直接 DB を触るのではなく、Repository を間に挟むことで、コードの責務を分離する。

## 🎯 ゴール

- `repository/user_repository.go` に User の CRUD 操作を実装する
- Handler → Repository → DB という呼び出しの流れを理解する
- Repository パターンの基本を学ぶ

## 📝 前提条件

- ✅ Issue #006（User Model & DBマイグレーション）が完了していること

## 📝 タスク

### 1. User Repository の作成

`internal/repository/user_repository.go` を作成する。

```go
package repository

import (
    "errors"

    "github.com/delivery-app/delivery-api/internal/model"
    "gorm.io/gorm"
)

// UserRepository はユーザーのDB操作を担当する
type UserRepository struct {
    db *gorm.DB
}

// NewUserRepository は新しい UserRepository を生成する
func NewUserRepository(db *gorm.DB) *UserRepository {
    return &UserRepository{db: db}
}

// Create は新しいユーザーをDBに保存する
func (r *UserRepository) Create(user *model.User) error {
    return r.db.Create(user).Error
}

// GetByEmail はメールアドレスでユーザーを検索する
func (r *UserRepository) GetByEmail(email string) (*model.User, error) {
    var user model.User
    if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, nil  // 見つからない場合は nil を返す（エラーではない）
        }
        return nil, err
    }
    return &user, nil
}

// GetByID はIDでユーザーを検索する
func (r *UserRepository) GetByID(id uint) (*model.User, error) {
    var user model.User
    if err := r.db.First(&user, id).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, nil
        }
        return nil, err
    }
    return &user, nil
}
```

> 💡 **Repositoryパターンとは？**
> データベース操作を専用の層に切り出すデザインパターン。
> メリット：
> - **Handlerがシンプルになる**：DBの細かい操作を知らなくていい
> - **テストしやすい**：RepositoryをモックすればDBなしでHandlerをテストできる
> - **DB変更に強い**：MySQLからPostgreSQLに変えてもRepositoryだけ修正すればOK

> 💡 **`GetByEmail` で nil を返す理由**:
> 「ユーザーが見つからない」のはエラーではなく「結果がなかった」だけ。
> Registration時に「このメールアドレスは使えるか？」のチェックでは nil が返ればOK、
> ユーザーが返ってきたら「既に登録済み」と判断する。

### 2. 各メソッドの解説

| メソッド | 用途 | SQLイメージ |
|----------|------|-------------|
| `Create` | ユーザーを新規作成する | `INSERT INTO users (email, name, ...) VALUES (?, ?, ...)` |
| `GetByEmail` | メールで1件検索する | `SELECT * FROM users WHERE email = ? LIMIT 1` |
| `GetByID` | IDで1件検索する | `SELECT * FROM users WHERE id = ? LIMIT 1` |

> 💡 **GORMのメソッド対応表**:
> | GORM | SQL |
> |------|-----|
> | `db.Create(&user)` | INSERT |
> | `db.First(&user, id)` | SELECT ... WHERE id = ? LIMIT 1 |
> | `db.Where("email = ?", email).First(&user)` | SELECT ... WHERE email = ? LIMIT 1 |
> | `db.Find(&users)` | SELECT * |

### 3. ディレクトリ構成の確認

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
│   └── repository/
│       └── user_repository.go   ← ⭐ 今回作成
├── docker-compose.yml
├── Dockerfile
├── go.mod
└── go.sum
```

## ✅ 動作確認

この Issue 単体では API エンドポイントを追加しないので、ビルドが通ることを確認する。

```bash
# コンテナ内でビルドが通るか確認
docker compose up --build
```

エラーなく起動すれば OK。

> 💡 **なぜエンドポイントを追加しないの？**
> 1つの Issue で「Repository作成」と「APIエンドポイント追加」を両方やると、
> 変更範囲が大きくなりすぎてレビューしにくい。
> **1 Issue = 1つの責務** を意識しよう。

## 🧪 追加チャレンジ（余裕があれば）

- [ ] `Update` メソッドを追加してみよう（`db.Model(&model.User{}).Where("id = ?", id).Updates(user)`)
- [ ] `Delete` メソッドを追加してみよう（`db.Delete(&model.User{}, id)`）
- [ ] `GetByEmail` で `errors.Is(err, gorm.ErrRecordNotFound)` の代わりに `errors.New("user not found")` を返すパターンも考えてみよう

## 📖 学習ポイント

| 概念 | 説明 |
|------|------|
| Repository パターン | DB操作を専用の構造体に切り出す設計 |
| コンストラクタ関数 | `NewUserRepository(db)` のように、構造体を生成する関数を用意するGoの慣習 |
| `gorm.ErrRecordNotFound` | GORMが「レコードが見つからない」時に返すエラー |
| `errors.Is()` | エラーの種類を判定する Go 1.13+ の標準関数 |
| ポインタ返り値 | `*model.User` は「見つからなければ nil」を表現するために使う |

## 🏷️ ラベル

`feature` `backend` `database` `repository` `初心者向け`
