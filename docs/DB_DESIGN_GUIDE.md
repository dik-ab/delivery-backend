# データベース設計ガイド（未経験者向け）

このドキュメントは、本プロジェクトのDB設計を初めて触る人が「何から手を付ければいいか」を迷わないためのガイドです。

---

## 0. 前提知識：このプロジェクトのDB環境

| 項目 | 値 |
|------|-----|
| DBMS | MySQL |
| ORM | GORM（Go言語用） |
| マイグレーション | GORMのAutoMigrate（開発時）|
| モデル定義場所 | `delivery-api/internal/model/` |
| リポジトリ（DB操作）| `delivery-api/internal/repository/` |

GORMでは、Goの構造体がそのままテーブル定義になる。SQLを直接書く場面は少ないが、SQLの基礎知識は必要。

---

## 1. テーブル設計の進め方

### Step 1: 「何を管理したいか」をリストアップする

まず、システムが管理する「もの」を洗い出す。名詞に注目する。

```
このシステムで管理するもの:
- 会社（companies）
- ユーザー（users）
- 車両（vehicles）
- 配車計画（dispatch_plans）
- 区間（trip_legs）
- マッチング（matches）
- 決済（payments）
- チャットルーム（chat_rooms）
- チャットメッセージ（chat_messages）
- 評価（reviews）
...
```

**1つの「もの」＝ 1つのテーブル** が基本。

### Step 2: 各テーブルの「カラム」を決める

そのテーブルが持つべき情報を考える。質問形式で考えるとわかりやすい:

```
「会社」について知りたいこと:
- 名前は？ → name
- 住所は？ → address
- 電話番号は？ → phone
- いつ登録された？ → created_at
```

### Step 3: テーブル間の「関係」を決める

「AはBを持つ」「AはBに属する」という関係を考える:

```
- 会社は複数のユーザーを「持つ」 → 1対多
- 会社は複数の車両を「持つ」 → 1対多
- 配車計画は1つの車両を「使う」 → 多対1
- マッチングは1つのチャットルームを「持つ」 → 1対1
```

### Step 4: 外部キー（FK）を設定する

関係を表現するために、「子テーブル」に「親テーブルのID」を持たせる:

```
users テーブルに company_id を追加 → companies.id を参照
vehicles テーブルに company_id を追加 → companies.id を参照
```

**覚え方: 「多」の側にFKを置く。** 1つの会社に複数のユーザーがいるなら、users側にcompany_idを置く。

---

## 2. GORMでのテーブル定義の書き方

### 基本テンプレート

```go
// model/company.go
package model

import "time"

type Company struct {
    ID          uint      `json:"id" gorm:"primaryKey"`
    Name        string    `json:"name" gorm:"not null"`
    DisplayName string    `json:"display_name" gorm:"not null"`
    Address     string    `json:"address"`
    Phone       string    `json:"phone"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

func (Company) TableName() string {
    return "companies"
}
```

### よく使うGORMタグ

| タグ | 意味 | 例 |
|------|------|-----|
| `gorm:"primaryKey"` | 主キー | ID |
| `gorm:"not null"` | NULL禁止 | 必須フィールド |
| `gorm:"uniqueIndex"` | ユニーク制約 | email, plate_number |
| `gorm:"default:値"` | デフォルト値 | `gorm:"default:'pending'"` |
| `gorm:"type:text"` | カラム型を指定 | 長いテキスト |
| `gorm:"type:mediumtext"` | さらに大きいテキスト | route_steps_json |
| `gorm:"foreignKey:カラム名"` | 外部キーを指定 | リレーション定義 |

### リレーションの書き方

```go
// 1対多: 会社 → ユーザー
type Company struct {
    ID    uint   `json:"id" gorm:"primaryKey"`
    Name  string `json:"name"`
    Users []User `json:"users" gorm:"foreignKey:CompanyID"` // ← 子の配列
}

type User struct {
    ID        uint    `json:"id" gorm:"primaryKey"`
    CompanyID *uint   `json:"company_id"`                          // ← FK（nilを許可するならポインタ型）
    Company   Company `json:"company" gorm:"foreignKey:CompanyID"` // ← 親への参照
    Name      string  `json:"name"`
}
```

**ポイント:**
- `*uint`（ポインタ型）にすると NULL を許可できる。荷主はcompany_idがnullの場合がある
- `binding:"-"` をつけると、APIリクエストのバリデーションから除外できる（リレーションフィールドに必要）

---

## 3. リポジトリ（DB操作）の書き方

### 基本テンプレート

```go
// repository/company.go
package repository

import (
    "errors"
    "github.com/delivery-app/delivery-api/internal/model"
    "gorm.io/gorm"
)

type CompanyRepository struct {
    db *gorm.DB
}

func NewCompanyRepository(db *gorm.DB) *CompanyRepository {
    return &CompanyRepository{db: db}
}

// 全件取得
func (r *CompanyRepository) GetAll() ([]model.Company, error) {
    var companies []model.Company
    if err := r.db.Find(&companies).Error; err != nil {
        return nil, err
    }
    return companies, nil
}

// IDで1件取得
func (r *CompanyRepository) GetByID(id uint) (*model.Company, error) {
    var company model.Company
    if err := r.db.First(&company, id).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, errors.New("company not found")
        }
        return nil, err
    }
    return &company, nil
}

// 作成
func (r *CompanyRepository) Create(company *model.Company) error {
    return r.db.Create(company).Error
}

// 更新（特定フィールドだけ更新する場合）
func (r *CompanyRepository) UpdateStatus(id uint, status string) error {
    return r.db.Model(&model.Company{}).Where("id = ?", id).
        Update("status", status).Error
}
```

### Preload（リレーションを一緒に取得する）

```go
// 配車計画を取得するとき、車両とドライバーの情報も一緒に取りたい
func (r *DispatchPlanRepository) GetByID(id uint) (*model.DispatchPlan, error) {
    var plan model.DispatchPlan
    err := r.db.
        Preload("Vehicle").       // ← 車両情報も取得
        Preload("Driver").        // ← ドライバー情報も取得
        Preload("TripLegs").      // ← 区間一覧も取得
        First(&plan, id).Error
    if err != nil {
        return nil, err
    }
    return &plan, nil
}
```

**Preloadを忘れるとどうなる？**
→ APIレスポンスで `vehicle: null`, `driver: null` になる。フロントエンドで `plan.vehicle.plate_number` がundefinedになってバグる。これは実際にこのプロジェクトで起きたバグ。

---

## 4. よくあるミスと対策

### ミス1: 更新時に全フィールドを上書きしてしまう

```go
// ❌ ダメな例: 構造体ごとUpdate → 意図しないフィールドも上書きされる
func (r *Repo) Update(id uint, match *model.Match) error {
    return r.db.Model(&model.Match{}).Where("id = ?", id).Updates(match).Error
}

// ✅ 良い例: 更新するフィールドだけ明示する
func (r *Repo) UpdateStatus(id uint, status string) error {
    return r.db.Model(&model.Match{}).Where("id = ?", id).
        Update("status", status).Error
}
```

### ミス2: トランザクションを使わない

マッチング承認時は「matchのステータス更新」と「チャットルーム作成」を同時にやる必要がある。片方だけ成功して片方が失敗するとデータが壊れる。

```go
// ✅ トランザクションで一貫性を保証
func (h *MatchHandler) ApproveMatch(c *gin.Context) {
    err := h.db.Transaction(func(tx *gorm.DB) error {
        // 1. マッチングステータスを更新
        if err := tx.Model(&model.Match{}).Where("id = ?", id).
            Update("status", "approved").Error; err != nil {
            return err // ← エラーが返るとロールバック
        }

        // 2. チャットルームを作成
        chatRoom := &model.ChatRoom{MatchID: id, Status: "active"}
        if err := tx.Create(chatRoom).Error; err != nil {
            return err // ← これもエラーならロールバック
        }

        return nil // ← nilを返すとコミット
    })
}
```

### ミス3: N+1問題

```go
// ❌ ダメな例: ループ内でDBクエリ → 100件あれば100回クエリが走る
plans, _ := repo.GetAll()
for _, plan := range plans {
    vehicle, _ := vehicleRepo.GetByID(plan.VehicleID) // ← 毎回DBアクセス
}

// ✅ 良い例: Preloadで一発取得
func (r *Repo) GetAll() ([]model.DispatchPlan, error) {
    var plans []model.DispatchPlan
    err := r.db.Preload("Vehicle").Preload("Driver").Find(&plans).Error
    return plans, err
}
```

### ミス4: インデックスを貼り忘れる

検索で使うカラムにはインデックスが必要。貼らないとデータが増えたときに検索が遅くなる。

```go
// モデルにインデックスを定義
type TripLeg struct {
    ID              uint   `gorm:"primaryKey"`
    DispatchPlanID  uint   `gorm:"index"`                           // ← 単体インデックス
    Visibility      string `gorm:"index:idx_vis_cargo"`             // ← 複合インデックス
    CargoStatus     string `gorm:"index:idx_vis_cargo"`             // ← 同じ名前で複合
    DepartureAt     time.Time `gorm:"index"`
    Status          string    `gorm:"index"`
}
```

**インデックスが必要なカラムの目安:**
- WHERE句で使うカラム（status, visibility, cargo_status）
- 外部キー（company_id, dispatch_plan_id, trip_leg_id）
- ORDER BY で使うカラム（created_at, departure_at）
- JOINのキーになるカラム

---

## 5. マイグレーションの手順

### 新テーブルを追加する場合

```
1. model/ に新しい構造体を定義
2. main.go（またはdb初期化処理）の AutoMigrate に追加
   db.AutoMigrate(&model.Company{}, &model.DispatchPlan{}, ...)
3. サーバー起動 → GORMが自動でCREATE TABLEを実行
```

### 既存テーブルにカラムを追加する場合

```
1. model/ の構造体にフィールドを追加
2. サーバー起動 → GORMが自動でALTER TABLE ADD COLUMNを実行
```

### 既存カラムを削除・リネームする場合

**⚠️ GORMのAutoMigrateはカラム削除・リネームをやってくれない。手動SQLが必要。**

```go
// マイグレーション用の関数を作る
func MigrateV2(db *gorm.DB) error {
    // カラムリネーム
    if err := db.Exec("ALTER TABLE matches CHANGE shipper_id requester_id INT UNSIGNED").Error; err != nil {
        return err
    }
    // カラム追加
    if err := db.Exec("ALTER TABLE matches ADD COLUMN requester_company_id INT UNSIGNED").Error; err != nil {
        return err
    }
    return nil
}
```

---

## 6. 設計チェックリスト

新しいテーブルを作るとき、このチェックリストを確認する:

### テーブル定義
- [ ] 主キー（ID）があるか
- [ ] created_at / updated_at があるか
- [ ] 外部キーにはFK制約をつけたか
- [ ] NULLを許可すべきカラムはポインタ型（`*uint`, `*string`）にしたか
- [ ] NOT NULLにすべきカラムに `gorm:"not null"` をつけたか
- [ ] ユニークにすべきカラムに `gorm:"uniqueIndex"` をつけたか
- [ ] デフォルト値が必要なカラムに `gorm:"default:値"` をつけたか
- [ ] ステータスカラムの取りうる値をコメントに書いたか
- [ ] TableName() メソッドを定義したか

### リレーション
- [ ] 1対多の「多」側にFKカラムがあるか
- [ ] 親テーブルへの参照（`Company Company`のようなフィールド）があるか
- [ ] `binding:"-"` をリレーションフィールドにつけたか
- [ ] 必要なPreloadをリポジトリに書いたか

### インデックス
- [ ] 外部キーにインデックスを貼ったか
- [ ] 検索条件で使うカラムにインデックスを貼ったか
- [ ] 複合検索には複合インデックスを使ったか

### リポジトリ
- [ ] GetAll, GetByID, Create の基本CRUDがあるか
- [ ] 更新は特定フィールドだけ更新しているか（構造体ごとUpdateしていないか）
- [ ] 複数テーブルの同時更新にトランザクションを使っているか
- [ ] N+1問題を避けるためにPreloadを使っているか

### API
- [ ] ハンドラーで権限チェック（このユーザーがこのリソースを操作していいか）をしているか
- [ ] エラー時に適切なHTTPステータスコードを返しているか（400, 401, 403, 404, 500）
- [ ] レスポンスにPreload済みのリレーションデータが含まれているか

---

## 7. このプロジェクト固有の注意点

### 段階的情報開示
APIレスポンスで会社情報を返すとき、マッチングの段階に応じてフィルタリングする必要がある。「全部返してフロントで隠す」は NG（APIを直接叩かれたら見えてしまう）。**必ずバックエンド側でフィルタリングする。**

### cargo_status と visibility の組み合わせ
マッチング検索の対象になるのは `cargo_status = 'empty_seeking'` かつ `visibility IN ('public', 'company_only')` の区間のみ。この条件をリポジトリの検索メソッドに必ず入れる。

### ステータス遷移の厳密化
マッチングのステータスを更新するとき、「今のステータスから遷移可能か」をチェックする:

```go
// ✅ 遷移可能チェックの例
func canTransition(current, next string) bool {
    allowed := map[string][]string{
        "pending":         {"approved", "rejected", "cancelled"},
        "approved":        {"payment_pending", "cancelled"},
        "payment_pending": {"completed", "cancelled"},
    }
    for _, s := range allowed[current] {
        if s == next {
            return true
        }
    }
    return false
}
```

これがないと、`pending` から直接 `completed` にできてしまう（決済スキップ）。

---

## 8. 作業の進め方（おすすめ順序）

```
1. model/ に構造体を定義する
   → テーブル設計書（table-design.html）を見ながらカラムを定義

2. AutoMigrate に追加してサーバー起動
   → テーブルが作られることを確認（MySQLクライアントで SHOW TABLES;）

3. repository/ にCRUD関数を書く
   → まずは GetAll, GetByID, Create の3つから

4. handler/ にAPIハンドラーを書く
   → まずは GET（取得）から。POST（作成）は次

5. router/ にルーティングを追加
   → protected.GET("/companies", companyHandler.GetAll)

6. Postmanやcurlでテスト
   → レスポンスが期待通りか確認

7. フロントエンドのAPI clientとhookを追加
   → api/companies.js, hooks/useCompanies.js
```

**1テーブルずつ、このサイクルを回す。** 一気に全テーブルを作ろうとしない。

---

## 9. 参考リンク

- [GORM公式ドキュメント（日本語）](https://gorm.io/ja_JP/docs/)
- [GORMモデル定義](https://gorm.io/ja_JP/docs/models.html)
- [GORMリレーション](https://gorm.io/ja_JP/docs/has_many.html)
- [GORMマイグレーション](https://gorm.io/ja_JP/docs/migration.html)
- [Gin Web Framework](https://gin-gonic.com/docs/)
