# GitHub Issues: データベース再設計（初学者向け）

> **前提**: Issue #004 で `users` テーブルが作成済み。それ以外のテーブルは全て一から作成する。
> 各 Issue は上から順番に着手すること（依存関係順）。

---

## Issue #005: `companies` テーブル作成 & `users` に `company_id` を追加

### 📋 概要
運送会社を管理する `companies` テーブルを新しく作り、`users` テーブルに「どの会社に所属しているか」を表す `company_id` カラムを追加する。

### 🎯 ゴール
- `model/company.go` に Company 構造体を定義する
- `model/user.go` に `CompanyID` フィールドとリレーションを追加する
- `main.go` の `AutoMigrate` に `Company` を追加し、`companies` テーブルが自動作成される
- `users` テーブルに `company_id` カラムが追加される

### 📝 前提条件
- ✅ Issue #004（User モデル / usersテーブル作成）が完了していること

### 📝 タスク

#### 1. Company モデルの作成

`internal/model/company.go` を新規作成する。

```go
package model

import "time"

// Company は運送会社を表す
type Company struct {
    ID              uint      `json:"id" gorm:"primaryKey"`
    Name            string    `json:"name" gorm:"type:varchar(255)" binding:"required"`
    DisplayName     string    `json:"display_name" gorm:"type:varchar(255)"`
    Address         string    `json:"address"`
    HeadquartersLat float64   `json:"headquarters_lat"`
    HeadquartersLng float64   `json:"headquarters_lng"`
    Phone           string    `json:"phone" gorm:"type:varchar(20)"`
    Email           string    `json:"email" gorm:"type:varchar(255)"`
    LogoURL         string    `json:"logo_url"`
    RatingAvg       float64   `json:"rating_avg" gorm:"default:0"`
    TotalDeals      int       `json:"total_deals" gorm:"default:0"`
    Plan            string    `json:"plan" gorm:"type:varchar(20);default:free"`
    CreatedAt       time.Time `json:"created_at"`
    UpdatedAt       time.Time `json:"updated_at"`
}

// TableName specifies the table name for GORM
func (Company) TableName() string {
    return "companies"
}
```

> 💡 **各フィールドの意味**:
> | フィールド | 意味 | なぜ必要？ |
> |-----------|------|-----------|
> | `Name` | 正式な会社名（例: 株式会社ABC運送） | マッチング承認後に表示される |
> | `DisplayName` | 匿名表示名（例: 神奈川エリアA社） | 検索時は正式名称を隠すため |
> | `HeadquartersLat/Lng` | 本社の緯度経度 | 近くの運送会社を検索するため |
> | `RatingAvg` | 平均評価（1〜5） | 信頼度の可視化 |
> | `TotalDeals` | 取引件数 | 実績の可視化 |
> | `Plan` | 契約プラン（free/standard/premium） | 機能制限の判定に使う |

> ⚠️ **`DisplayName` と `Name` を分ける理由**
> このサービスでは「段階的情報開示」を行う。
> 検索段階では `DisplayName`（匿名）しか見えず、決済完了後に初めて `Name`（正式名称）が開示される。
> これにより「会社名でググって直電」されるのを防ぐ。

#### 2. User モデルに CompanyID を追加

`internal/model/user.go` を以下のように修正する。

```go
package model

import "time"

// User はプラットフォームのユーザーを表す
type User struct {
    ID           uint      `json:"id" gorm:"primaryKey"`
    CompanyID    *uint     `json:"company_id"`
    Company      Company   `json:"company" gorm:"foreignKey:CompanyID" binding:"-"`
    Email        string    `json:"email" gorm:"type:varchar(255);uniqueIndex" binding:"required"`
    PasswordHash string    `json:"-"`
    Name         string    `json:"name" binding:"required"`
    Role         string    `json:"role" gorm:"type:varchar(20);default:shipper"`
    Phone        string    `json:"phone" gorm:"type:varchar(20)"`
    IsActive     bool      `json:"is_active" gorm:"default:true"`
    CreatedAt    time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`
}
```

> 💡 **変更ポイント解説**:
> | 変更 | Before | After | なぜ？ |
> |------|--------|-------|--------|
> | `CompanyID` 追加 | なし | `*uint`（ポインタ） | 荷主は会社に所属しない場合があるので `NULL` を許容する |
> | `Company` リレーション追加 | なし | `gorm:"foreignKey:CompanyID"` | `Preload("Company")` でJOIN取得できるようにする |
> | `Role` のデフォルト変更 | `driver` | `shipper` | 最初に登録するユーザーは荷主が多い想定 |
> | `IsActive` 追加 | なし | `gorm:"default:true"` | アカウント無効化に対応 |
> | 旧 `Company` (string) 削除 | `string` | 削除 | companiesテーブルで管理するようになったため不要 |

> ⚠️ **なぜ `*uint`（ポインタ）なの？**
> Go では `uint` のゼロ値は `0` だが、GORM は `0` を「値がある」と認識する。
> ポインタ `*uint` にすると、値がない時に `nil` になり、DBには `NULL` が入る。
> 荷主（shipper）のように会社に所属しないユーザーは `company_id = NULL` にしたいので、ポインタにする。

> 💡 **Role の種類**:
> | Role | 説明 | 主な操作 |
> |------|------|---------|
> | `owner` | 運送会社オーナー | 会社設定、配車計画作成・承認 |
> | `dispatcher` | 配車担当者 | 配車計画作成・編集 |
> | `driver` | ドライバー | 出発/到着報告 |
> | `shipper` | 荷主 | 便検索、マッチングリクエスト、決済 |
> | `admin` | システム管理者 | 全データ閲覧 |

#### 3. main.go に AutoMigrate を追加

`cmd/server/main.go` の AutoMigrate に `&model.Company{}` を**先に**追加する。

```go
if err := db.AutoMigrate(
    &model.Company{},  // ← 先に作る（usersがFK参照するため）
    &model.User{},
); err != nil {
    log.Fatal("❌ マイグレーション失敗:", err)
}
```

> ⚠️ **AutoMigrate の順番が重要！**
> `users.company_id` は `companies.id` を参照する外部キー（FK）。
> FK参照先のテーブルが先に存在しないとエラーになる場合がある。
> そのため `Company` を `User` より **前** に書く。

#### 4. ディレクトリ構成の確認

```
delivery-api/
├── internal/
│   └── model/
│       ├── company.go       ← ⭐ 今回作成
│       └── user.go          ← 修正（CompanyID追加）
└── cmd/
    └── server/
        └── main.go          ← AutoMigrate修正
```

### ✅ 動作確認

```bash
# 再ビルド & 起動
docker compose up --build
```

MySQLに接続して確認:

```bash
docker compose exec db mysql -u root -p delivery_db

-- companies テーブルの確認
DESCRIBE companies;

-- users テーブルに company_id が追加されたか確認
DESCRIBE users;

-- 外部キーの確認
SHOW CREATE TABLE users;
```

### 🟢 期待される出力（companies）

```
+------------------+--------------+------+-----+---------+----------------+
| Field            | Type         | Null | Key | Default | Extra          |
+------------------+--------------+------+-----+---------+----------------+
| id               | bigint unsigned | NO | PRI | NULL   | auto_increment |
| name             | varchar(255) | YES  |     | NULL    |                |
| display_name     | varchar(255) | YES  |     | NULL    |                |
| address          | longtext     | YES  |     | NULL    |                |
| headquarters_lat | double       | YES  |     | NULL    |                |
| headquarters_lng | double       | YES  |     | NULL    |                |
| phone            | varchar(20)  | YES  |     | NULL    |                |
| email            | varchar(255) | YES  |     | NULL    |                |
| logo_url         | longtext     | YES  |     | NULL    |                |
| rating_avg       | double       | YES  |     | 0       |                |
| total_deals      | bigint       | YES  |     | 0       |                |
| plan             | varchar(20)  | YES  |     | free    |                |
| created_at       | datetime(3)  | YES  |     | NULL    |                |
| updated_at       | datetime(3)  | YES  |     | NULL    |                |
+------------------+--------------+------+-----+---------+----------------+
```

### 🧪 追加チャレンジ（余裕があれば）
- [ ] `companies.name` にも `uniqueIndex` を付けてみよう（同じ会社名の重複を防ぐ）
- [ ] `repository/company.go` を作成し、`GetByID` と `Create` を実装してみよう
- [ ] `User` を取得する時に `Preload("Company")` で会社情報も一緒に取得してみよう

### 📖 学習ポイント

| 概念 | 説明 |
|------|------|
| 外部キー (FK) | あるテーブルのカラムが別テーブルの `id` を参照する仕組み |
| nullable FK | `*uint`（ポインタ）にすることで NULL を許容する |
| `Preload` | GORMでリレーション先を一緒に取得する機能 |
| AutoMigrate の順番 | FK参照先テーブルを先にマイグレーションする必要がある |
| 段階的情報開示 | サービスの段階に応じて表示する情報を変える設計パターン |

### 🏷️ ラベル
`feature` `backend` `database` `model` `初心者向け`

---

## Issue #006: `vehicles` テーブル作成（会社所有の車両管理）

### 📋 概要
運送会社が所有する車両を管理する `vehicles` テーブルを作成する。車両は**個人ではなく会社に紐づく**。

### 🎯 ゴール
- `model/vehicle.go` に Vehicle 構造体を定義する
- `main.go` の `AutoMigrate` に追加し、`vehicles` テーブルが自動作成される
- `plate_number`（車番）にユニーク制約が付いていることを確認する

### 📝 前提条件
- ✅ Issue #005（companies テーブル作成）が完了していること

### 📝 タスク

#### 1. Vehicle モデルの作成

`internal/model/vehicle.go` を新規作成する。

```go
package model

import "time"

// Vehicle は運送会社が所有する車両を表す
type Vehicle struct {
    ID          uint      `json:"id" gorm:"primaryKey"`
    CompanyID   uint      `json:"company_id"`
    Company     Company   `json:"company" gorm:"foreignKey:CompanyID" binding:"-"`
    PlateNumber string    `json:"plate_number" gorm:"type:varchar(20);uniqueIndex" binding:"required"`
    VehicleType string    `json:"vehicle_type" gorm:"type:varchar(20)" binding:"required"`
    MaxWeight   float64   `json:"max_weight"`
    Status      string    `json:"status" gorm:"type:varchar(20);default:active"`
    Note        string    `json:"note"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

// TableName specifies the table name for GORM
func (Vehicle) TableName() string {
    return "vehicles"
}
```

> 💡 **各フィールドの意味**:
> | フィールド | 例 | 説明 |
> |-----------|-----|------|
> | `CompanyID` | `1` | 所有する会社のID。**個人ではなく会社に紐づく** |
> | `PlateNumber` | `横浜 100 あ 1234` | 車番。日本中でユニーク |
> | `VehicleType` | `軽トラ` / `2t` / `4t` / `10t` / `大型` | 車種 |
> | `MaxWeight` | `10000` | 最大積載量（kg） |
> | `Status` | `active` | 車両の状態 |

> 💡 **Status の種類**:
> | Status | 意味 |
> |--------|------|
> | `active` | 稼働中（配車可能） |
> | `maintenance` | 整備中（一時的に使えない） |
> | `retired` | 廃車/売却済み（もう使わない） |

> ⚠️ **なぜ `CompanyID` は `uint` で `*uint` ではない？**
> Issue #005 の `User.CompanyID` は荷主がNULLになり得たのでポインタにした。
> しかし `Vehicle` は必ず会社に所属する（所有者のいない車両はありえない）。
> だから `uint`（NOT NULL）にして、必ず会社IDが入ることを保証する。

> ⚠️ **`uniqueIndex` の意味**
> `PlateNumber` に `uniqueIndex` を付けると、同じ車番のレコードを2つ作ろうとした時に
> `Error 1062: Duplicate entry` というエラーが出て拒否される。
> 車番は現実世界で一意なので、DBでもユニーク制約を付けるのが正しい。

#### 2. main.go に AutoMigrate を追加

```go
if err := db.AutoMigrate(
    &model.Company{},
    &model.User{},
    &model.Vehicle{},  // ← 追加
); err != nil {
    log.Fatal("❌ マイグレーション失敗:", err)
}
```

#### 3. ディレクトリ構成の確認

```
delivery-api/
├── internal/
│   └── model/
│       ├── company.go
│       ├── user.go
│       └── vehicle.go      ← ⭐ 今回作成
└── cmd/
    └── server/
        └── main.go         ← AutoMigrate修正
```

### ✅ 動作確認

```bash
docker compose up --build
docker compose exec db mysql -u root -p delivery_db

DESCRIBE vehicles;

-- ユニーク制約の確認
SHOW INDEX FROM vehicles WHERE Key_name != 'PRIMARY';
```

### 🟢 期待される出力

```
+--------------+-----------------+------+-----+---------+----------------+
| Field        | Type            | Null | Key | Default | Extra          |
+--------------+-----------------+------+-----+---------+----------------+
| id           | bigint unsigned | NO   | PRI | NULL    | auto_increment |
| company_id   | bigint unsigned | YES  | MUL | NULL    |                |
| plate_number | varchar(20)     | YES  | UNI | NULL    |                |
| vehicle_type | varchar(20)     | YES  |     | NULL    |                |
| max_weight   | double          | YES  |     | NULL    |                |
| status       | varchar(20)     | YES  |     | active  |                |
| note         | longtext        | YES  |     | NULL    |                |
| created_at   | datetime(3)     | YES  |     | NULL    |                |
| updated_at   | datetime(3)     | YES  |     | NULL    |                |
+--------------+-----------------+------+-----+---------+----------------+
```

### 🧪 追加チャレンジ（余裕があれば）
- [ ] `Company` モデルに `Vehicles []Vehicle` のリレーション（HasMany）を追加してみよう
- [ ] `repository/vehicle.go` を作成し、`GetByCompanyID` を実装してみよう
- [ ] `Status` を定数（`const`）で定義してみよう

### 📖 学習ポイント

| 概念 | 説明 |
|------|------|
| NOT NULL FK vs nullable FK | 必ず親が存在するなら `uint`、NULLもありなら `*uint` |
| uniqueIndex | 同じ値のレコードを2つ作れなくする制約 |
| HasMany リレーション | 1つの会社が複数の車両を持つ（1:N）関係 |
| Status パターン | 有限の状態をstringで管理する設計パターン |

### 🏷️ ラベル
`feature` `backend` `database` `model` `初心者向け`

---

## Issue #007: `dispatch_plans` テーブル作成（配車計画）

### 📋 概要
「1台のトラック × 1日」を表す **配車計画テーブル** を作成する。ソロモード（自社配車管理）の核心となるテーブル。

### 🎯 ゴール
- `model/dispatch_plan.go` に DispatchPlan 構造体を定義する
- `main.go` の `AutoMigrate` に追加し、`dispatch_plans` テーブルが自動作成される
- `(company_id, plan_date)` の複合インデックスが存在することを確認する

### 📝 前提条件
- ✅ Issue #005（companies テーブル作成）が完了していること
- ✅ Issue #006（vehicles テーブル作成）が完了していること

### 📝 タスク

#### 1. DispatchPlan モデルの作成

`internal/model/dispatch_plan.go` を新規作成する。

```go
package model

import "time"

// DispatchPlan status constants
const (
    DispatchPlanStatusPlanned    = "planned"
    DispatchPlanStatusInProgress = "in_progress"
    DispatchPlanStatusCompleted  = "completed"
    DispatchPlanStatusCancelled  = "cancelled"
)

// DispatchPlan は「1台のトラック × 1日」の配車計画を表す
type DispatchPlan struct {
    ID        uint      `json:"id" gorm:"primaryKey"`
    CompanyID uint      `json:"company_id"`
    Company   Company   `json:"company" gorm:"foreignKey:CompanyID" binding:"-"`
    VehicleID uint      `json:"vehicle_id"`
    Vehicle   Vehicle   `json:"vehicle" gorm:"foreignKey:VehicleID" binding:"-"`
    DriverID  uint      `json:"driver_id"`
    Driver    User      `json:"driver" gorm:"foreignKey:DriverID" binding:"-"`
    PlanDate  time.Time `json:"plan_date" gorm:"type:date;index:idx_dispatch_plans_company_date"`
    Status    string    `json:"status" gorm:"type:varchar(20);default:planned;index:idx_dispatch_plans_status"`
    Note      string    `json:"note"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

// TableName specifies the table name for GORM
func (DispatchPlan) TableName() string {
    return "dispatch_plans"
}
```

> 💡 **配車計画のイメージ**:
> ```
> ┌────────────────────────────────────────────────────┐
> │ 配車計画 #1                                        │
> │ 会社: ABC運送                                      │
> │ 車両: 横浜 100 あ 1234 (10t)                       │
> │ 運転手: 田中太郎                                    │
> │ 日付: 2026-03-12                                   │
> │ 状態: in_progress                                  │
> │                                                    │
> │ → この計画の中に「区間（trip_legs）」が入る        │
> │   区間1: 横浜→小田原（荷物あり）                   │
> │   区間2: 小田原→横浜（空車・募集中）               │
> └────────────────────────────────────────────────────┘
> ```

> 💡 **新しい概念: 定数（const）の定義**
> ```go
> const (
>     DispatchPlanStatusPlanned = "planned"
>     // ...
> )
> ```
> ステータスの文字列をコード中にベタ書きすると、タイポ（`"planed"` 等）でバグる。
> `const` で定義しておけば、コンパイル時にチェックされるので安全。
> 使う時は `if plan.Status == DispatchPlanStatusPlanned` と書く。

> 💡 **GORMのインデックス指定**:
> ```go
> PlanDate time.Time `gorm:"type:date;index:idx_dispatch_plans_company_date"`
> ```
> `index:idx_xxx` を付けると、GORMが AutoMigrate 時にインデックスを自動作成する。
> **同じインデックス名を複数フィールドに付けると複合インデックスになる。**
>
> ただし今回は `CompanyID` はGORMのFK自動インデックスに任せ、
> `PlanDate` と `Status` にそれぞれ単独インデックスを付ける方式でもOK。

> ⚠️ **`PlanDate` の型が `time.Time` なのに `gorm:"type:date"` を指定する理由**
> Go の `time.Time` はGORMでは `datetime(3)` にマッピングされるのがデフォルト。
> しかし配車計画の日付には時刻は不要（「2026年3月12日」だけでいい）なので、
> `type:date` を明示して MySQL の `DATE` 型にする。

#### 2. main.go に AutoMigrate を追加

```go
if err := db.AutoMigrate(
    &model.Company{},
    &model.User{},
    &model.Vehicle{},
    &model.DispatchPlan{},  // ← 追加
); err != nil {
    log.Fatal("❌ マイグレーション失敗:", err)
}
```

#### 3. ディレクトリ構成の確認

```
delivery-api/
├── internal/
│   └── model/
│       ├── company.go
│       ├── dispatch_plan.go   ← ⭐ 今回作成
│       ├── user.go
│       └── vehicle.go
└── cmd/
    └── server/
        └── main.go
```

### ✅ 動作確認

```bash
docker compose up --build
docker compose exec db mysql -u root -p delivery_db

DESCRIBE dispatch_plans;

-- インデックスの確認
SHOW INDEX FROM dispatch_plans;
```

### 🟢 期待される出力

```
+------------+-----------------+------+-----+---------+----------------+
| Field      | Type            | Null | Key | Default | Extra          |
+------------+-----------------+------+-----+---------+----------------+
| id         | bigint unsigned | NO   | PRI | NULL    | auto_increment |
| company_id | bigint unsigned | YES  | MUL | NULL    |                |
| vehicle_id | bigint unsigned | YES  | MUL | NULL    |                |
| driver_id  | bigint unsigned | YES  | MUL | NULL    |                |
| plan_date  | date            | YES  | MUL | NULL    |                |
| status     | varchar(20)     | YES  | MUL | planned |                |
| note       | longtext        | YES  |     | NULL    |                |
| created_at | datetime(3)     | YES  |     | NULL    |                |
| updated_at | datetime(3)     | YES  |     | NULL    |                |
+------------+-----------------+------+-----+---------+----------------+
```

### 🧪 追加チャレンジ（余裕があれば）
- [ ] `CompanyID` と `PlanDate` の複合ユニーク制約を `VehicleID` も含めて作ってみよう（同じ車両は同じ日に1つの計画しか作れない）
- [ ] `repository/dispatch_plan.go` を作成し、`GetByCompanyAndDate(companyID uint, date time.Time)` を実装してみよう
- [ ] テストデータを INSERT して、`WHERE company_id = 1 AND plan_date = '2026-03-12'` で検索してみよう

### 📖 学習ポイント

| 概念 | 説明 |
|------|------|
| 複合インデックス | 複数カラムを組み合わせたインデックス。検索パターンに合わせて設計する |
| `type:date` | MySQL の DATE 型。時刻を含まない日付のみの型 |
| `const` によるステータス管理 | 文字列のタイポを防ぎ、IDEの補完が効くようにする |
| 1:N:N の階層構造 | Company → DispatchPlan → (次のIssueで) TripLeg という入れ子関係 |

### 🏷️ ラベル
`feature` `backend` `database` `model` `初心者向け`

---

## Issue #008: `trip_legs` テーブル作成（配車区間）

### 📋 概要
配車計画を構成する各区間を管理する `trip_legs` テーブルを作成する。1つの配車計画は複数の区間を持つ（例: 往路・帰路）。**マッチングの対象はこの区間単位**。

### 🎯 ゴール
- `model/trip_leg.go` に TripLeg 構造体を定義する
- `main.go` の `AutoMigrate` に追加し、`trip_legs` テーブルが自動作成される
- 「空車・募集中」の区間だけを検索できるインデックスが存在することを確認する

### 📝 前提条件
- ✅ Issue #007（dispatch_plans テーブル作成）が完了していること

### 📝 タスク

#### 1. TripLeg モデルの作成

`internal/model/trip_leg.go` を新規作成する。

```go
package model

import "time"

// TripLeg cargo status constants
const (
    CargoStatusLoaded       = "loaded"        // 荷物あり（自社案件）
    CargoStatusEmptySeeking = "empty_seeking"  // 空車・荷物募集中
    CargoStatusEmptyPrivate = "empty_private"  // 空車・募集しない
)

// TripLeg visibility constants
const (
    VisibilityPrivate     = "private"      // 自社のみ閲覧可
    VisibilityCompanyOnly = "company_only"  // 運送会社のみに公開
    VisibilityPublic      = "public"        // 全ユーザーに公開
)

// TripLeg status constants
const (
    TripLegStatusScheduled  = "scheduled"
    TripLegStatusInTransit  = "in_transit"
    TripLegStatusCompleted  = "completed"
    TripLegStatusCancelled  = "cancelled"
)

// TripLeg leg type constants
const (
    LegTypeOutbound = "outbound"  // 往路
    LegTypeReturn   = "return"    // 帰路
)

// TripLeg は配車計画を構成する1つの区間を表す
type TripLeg struct {
    ID               uint      `json:"id" gorm:"primaryKey"`
    DispatchPlanID   uint      `json:"dispatch_plan_id" gorm:"index:idx_trip_legs_dispatch_plan"`
    DispatchPlan     DispatchPlan `json:"dispatch_plan" gorm:"foreignKey:DispatchPlanID" binding:"-"`
    LegOrder         int       `json:"leg_order"`
    OriginAddress    string    `json:"origin_address" binding:"required"`
    OriginLat        float64   `json:"origin_lat"`
    OriginLng        float64   `json:"origin_lng"`
    DestAddress      string    `json:"dest_address" binding:"required"`
    DestLat          float64   `json:"dest_lat"`
    DestLng          float64   `json:"dest_lng"`
    DepartureAt      time.Time `json:"departure_at" gorm:"index:idx_trip_legs_departure"`
    ArrivalAt        time.Time `json:"arrival_at"`
    LegType          string    `json:"leg_type" gorm:"type:varchar(20);default:outbound"`
    CargoStatus      string    `json:"cargo_status" gorm:"type:varchar(20);default:loaded;index:idx_trip_legs_visibility_cargo"`
    CargoDescription string    `json:"cargo_description"`
    CargoWeight      float64   `json:"cargo_weight"`
    AvailableWeight  float64   `json:"available_weight"`
    Price            int       `json:"price"`
    Visibility       string    `json:"visibility" gorm:"type:varchar(20);default:private;index:idx_trip_legs_visibility_cargo"`
    RoutePolyline    string    `json:"route_polyline" gorm:"type:text"`
    RouteDurationSec int       `json:"route_duration_sec" gorm:"default:0"`
    RouteStepsJSON   string    `json:"route_steps_json" gorm:"type:mediumtext"`
    DelayMinutes     int       `json:"delay_minutes" gorm:"default:0"`
    Status           string    `json:"status" gorm:"type:varchar(20);default:scheduled;index:idx_trip_legs_status"`
    CreatedAt        time.Time `json:"created_at"`
    UpdatedAt        time.Time `json:"updated_at"`
}

// TableName specifies the table name for GORM
func (TripLeg) TableName() string {
    return "trip_legs"
}
```

> 💡 **配車計画と区間の関係**:
> ```
> 配車計画: トラック③ (10t) / 田中太郎 / 2026-03-12
> │
> ├── 区間1（leg_order: 1）
> │   出発: 横浜 → 到着: 小田原
> │   時間: 9:00 → 15:00
> │   荷物: loaded（あり）
> │   公開: private（非公開）
> │
> └── 区間2（leg_order: 2）
>     出発: 小田原 → 到着: 横浜
>     時間: 16:00 → 21:00
>     荷物: empty_seeking（空車・募集中）  ← これがマッチング対象！
>     公開: public（全ユーザーに公開）
> ```

> 💡 **CargoStatus の意味**:
> | 値 | 意味 | マッチング対象？ |
> |----|------|:---:|
> | `loaded` | 荷物あり（自社の仕事） | ❌ |
> | `empty_seeking` | 空車で荷物を募集中 | ✅ |
> | `empty_private` | 空車だけど募集しない | ❌ |

> 💡 **Visibility の意味**:
> | 値 | 誰に見えるか | ユースケース |
> |----|-------------|-------------|
> | `private` | 自社の人だけ | ソロモードでの社内管理 |
> | `company_only` | 他の運送会社にも見える | 運送会社間のマッチング |
> | `public` | 荷主含む全員に見える | 荷主からの依頼受付 |

> 💡 **マッチング対象になる条件**:
> ```
> WHERE cargo_status = 'empty_seeking'
>   AND visibility IN ('public', 'company_only')
>   AND status = 'scheduled'
> ```
> この検索が速く動くように `(visibility, cargo_status)` の複合インデックスを付けている。

> ⚠️ **なぜ `RouteStepsJSON` は `mediumtext`？**
> Google Routes API が返すルートのステップ情報はJSONで格納する。
> ステップ数が多いと数百KB〜1MBになることがあるため、
> `text`（最大64KB）ではなく `mediumtext`（最大16MB）を指定する。

#### 2. DispatchPlan に TripLegs リレーションを追加

`internal/model/dispatch_plan.go` に以下を追加:

```go
type DispatchPlan struct {
    // ... 既存フィールド ...
    TripLegs  []TripLeg `json:"trip_legs" gorm:"foreignKey:DispatchPlanID" binding:"-"`
    // ... 既存フィールド ...
}
```

> 💡 **HasMany リレーション**
> `TripLegs []TripLeg` を追加すると、`db.Preload("TripLegs").Find(&plan)` で
> 配車計画と一緒に全区間を一括取得できるようになる。

#### 3. main.go に AutoMigrate を追加

```go
if err := db.AutoMigrate(
    &model.Company{},
    &model.User{},
    &model.Vehicle{},
    &model.DispatchPlan{},
    &model.TripLeg{},  // ← 追加
); err != nil {
    log.Fatal("❌ マイグレーション失敗:", err)
}
```

#### 4. ディレクトリ構成の確認

```
delivery-api/
├── internal/
│   └── model/
│       ├── company.go
│       ├── dispatch_plan.go   ← 修正（TripLegs追加）
│       ├── trip_leg.go        ← ⭐ 今回作成
│       ├── user.go
│       └── vehicle.go
└── cmd/
    └── server/
        └── main.go
```

### ✅ 動作確認

```bash
docker compose up --build
docker compose exec db mysql -u root -p delivery_db

DESCRIBE trip_legs;

-- インデックスの確認
SHOW INDEX FROM trip_legs;
```

### 🟢 期待される出力（主要カラムのみ抜粋）

```
+--------------------+-----------------+------+-----+-----------+----------------+
| Field              | Type            | Null | Key | Default   | Extra          |
+--------------------+-----------------+------+-----+-----------+----------------+
| id                 | bigint unsigned | NO   | PRI | NULL      | auto_increment |
| dispatch_plan_id   | bigint unsigned | YES  | MUL | NULL      |                |
| leg_order          | bigint          | YES  |     | NULL      |                |
| origin_address     | longtext        | YES  |     | NULL      |                |
| origin_lat         | double          | YES  |     | NULL      |                |
| origin_lng         | double          | YES  |     | NULL      |                |
| dest_address       | longtext        | YES  |     | NULL      |                |
| dest_lat           | double          | YES  |     | NULL      |                |
| dest_lng           | double          | YES  |     | NULL      |                |
| departure_at       | datetime(3)     | YES  | MUL | NULL      |                |
| arrival_at         | datetime(3)     | YES  |     | NULL      |                |
| leg_type           | varchar(20)     | YES  |     | outbound  |                |
| cargo_status       | varchar(20)     | YES  | MUL | loaded    |                |
| cargo_description  | longtext        | YES  |     | NULL      |                |
| cargo_weight       | double          | YES  |     | NULL      |                |
| available_weight   | double          | YES  |     | NULL      |                |
| price              | bigint          | YES  |     | NULL      |                |
| visibility         | varchar(20)     | YES  | MUL | private   |                |
| route_polyline     | text            | YES  |     | NULL      |                |
| route_duration_sec | bigint          | YES  |     | 0         |                |
| route_steps_json   | mediumtext      | YES  |     | NULL      |                |
| delay_minutes      | bigint          | YES  |     | 0         |                |
| status             | varchar(20)     | YES  | MUL | scheduled |                |
| created_at         | datetime(3)     | YES  |     | NULL      |                |
| updated_at         | datetime(3)     | YES  |     | NULL      |                |
+--------------------+-----------------+------+-----+-----------+----------------+
```

### 🧪 追加チャレンジ（余裕があれば）
- [ ] `(dispatch_plan_id, leg_order)` のユニーク制約を追加してみよう（同じ計画内で同じ順序番号が重複しないように）
- [ ] テストデータを INSERT して、`WHERE cargo_status = 'empty_seeking' AND visibility = 'public'` で検索してみよう
- [ ] `EXPLAIN` を使ってインデックスが使われているか確認してみよう

### 📖 学習ポイント

| 概念 | 説明 |
|------|------|
| 複合インデックス | `(visibility, cargo_status)` のように複数カラムを1つのインデックスにまとめる |
| mediumtext | MySQLの大きなテキスト型（最大16MB）。JSONデータの格納に使う |
| HasMany | 1つの配車計画が複数の区間を持つ（1:N）関係 |
| const によるEnum風管理 | Go にはEnum型がないため、`const` で有限の値を管理する |
| EXPLAIN | SQLの実行計画を確認するコマンド。インデックスが使われているか確認できる |

### 🏷️ ラベル
`feature` `backend` `database` `model` `初心者向け`

---

## Issue #009: `matches` テーブル作成（マッチング）

### 📋 概要
荷主や他の運送会社が「空車区間に荷物を載せたい」とリクエストするためのマッチングテーブルを作成する。マッチング対象は **区間（trip_leg）単位**。

### 🎯 ゴール
- `model/match.go` に Match 構造体を定義する
- `main.go` の `AutoMigrate` に追加し、`matches` テーブルが自動作成される
- ステータス遷移に `payment_pending` が含まれていることを確認する

### 📝 前提条件
- ✅ Issue #008（trip_legs テーブル作成）が完了していること

### 📝 タスク

#### 1. Match モデルの作成

`internal/model/match.go` を新規作成する。

```go
package model

import "time"

// Match status constants
const (
    MatchStatusPending        = "pending"
    MatchStatusApproved       = "approved"
    MatchStatusPaymentPending = "payment_pending"
    MatchStatusCompleted      = "completed"
    MatchStatusRejected       = "rejected"
    MatchStatusCancelled      = "cancelled"
)

// Match request type constants
const (
    RequestTypeShipperToCompany  = "shipper_to_company"
    RequestTypeCompanyToCompany  = "company_to_company"
)

// Match は荷主/運送会社からの区間マッチングリクエストを表す
type Match struct {
    ID                 uint      `json:"id" gorm:"primaryKey"`
    TripLegID          uint      `json:"trip_leg_id" gorm:"index:idx_matches_trip_leg"`
    TripLeg            TripLeg   `json:"trip_leg" gorm:"foreignKey:TripLegID" binding:"-"`
    RequesterID        uint      `json:"requester_id" gorm:"index:idx_matches_requester"`
    Requester          User      `json:"requester" gorm:"foreignKey:RequesterID" binding:"-"`
    RequesterCompanyID *uint     `json:"requester_company_id"`
    RequesterCompany   Company   `json:"requester_company" gorm:"foreignKey:RequesterCompanyID" binding:"-"`
    CargoWeight        float64   `json:"cargo_weight"`
    CargoDescription   string    `json:"cargo_description"`
    Message            string    `json:"message"`
    Status             string    `json:"status" gorm:"type:varchar(20);default:pending;index:idx_matches_status"`
    RejectReason       string    `json:"reject_reason"`
    RequestType        string    `json:"request_type" gorm:"type:varchar(30);default:shipper_to_company"`
    CreatedAt          time.Time `json:"created_at"`
    UpdatedAt          time.Time `json:"updated_at"`
}

// TableName specifies the table name for GORM
func (Match) TableName() string {
    return "matches"
}
```

> 💡 **ステータス遷移図**:
> ```
>                    ┌─── rejected（拒否）
>                    │
> pending ──→ approved ──→ payment_pending ──→ completed
> (リクエスト)  (承認)       (決済待ち)         (完了)
>    │
>    └─── cancelled（取消 ※どの段階からでも可能）
> ```
>
> **重要: `approved` から直接 `completed` にはできない！**
> 必ず `payment_pending`（決済待ち）を経由する。
> これが「決済しないと完了にできない」仕組みの鍵。

> 💡 **`RequesterID` と `RequesterCompanyID` の違い**:
> | フィールド | 型 | 意味 |
> |-----------|-----|------|
> | `RequesterID` | `uint`（NOT NULL） | リクエストを送った**人**のID |
> | `RequesterCompanyID` | `*uint`（nullable） | リクエストを送った人の**会社**のID |
>
> 荷主（shipper）は会社に所属していない場合があるので `RequesterCompanyID` は nullable。
> 運送会社間マッチングの場合は両方に値が入る。

> ⚠️ **なぜ `TripID` ではなく `TripLegID`？**
> 旧設計では便（trip）全体に対してマッチングしていた。
> 新設計では**区間（trip_leg）単位**でマッチングする。
> 例: 往路（荷物あり）は対象外、帰路（空車）だけがマッチング対象。

#### 2. TripLeg に Matches リレーションを追加

`internal/model/trip_leg.go` に以下を追加:

```go
type TripLeg struct {
    // ... 既存フィールド ...
    Matches []Match `json:"matches" gorm:"foreignKey:TripLegID" binding:"-"`
    // ... 既存フィールド ...
}
```

#### 3. main.go に AutoMigrate を追加

```go
if err := db.AutoMigrate(
    &model.Company{},
    &model.User{},
    &model.Vehicle{},
    &model.DispatchPlan{},
    &model.TripLeg{},
    &model.Match{},  // ← 追加
); err != nil {
    log.Fatal("❌ マイグレーション失敗:", err)
}
```

### ✅ 動作確認

```bash
docker compose up --build
docker compose exec db mysql -u root -p delivery_db

DESCRIBE matches;
SHOW INDEX FROM matches;
```

### 🟢 期待される出力

```
+----------------------+-----------------+------+-----+---------------------+----------------+
| Field                | Type            | Null | Key | Default             | Extra          |
+----------------------+-----------------+------+-----+---------------------+----------------+
| id                   | bigint unsigned | NO   | PRI | NULL                | auto_increment |
| trip_leg_id          | bigint unsigned | YES  | MUL | NULL                |                |
| requester_id         | bigint unsigned | YES  | MUL | NULL                |                |
| requester_company_id | bigint unsigned | YES  | MUL | NULL                |                |
| cargo_weight         | double          | YES  |     | NULL                |                |
| cargo_description    | longtext        | YES  |     | NULL                |                |
| message              | longtext        | YES  |     | NULL                |                |
| status               | varchar(20)     | YES  | MUL | pending             |                |
| reject_reason        | longtext        | YES  |     | NULL                |                |
| request_type         | varchar(30)     | YES  |     | shipper_to_company  |                |
| created_at           | datetime(3)     | YES  |     | NULL                |                |
| updated_at           | datetime(3)     | YES  |     | NULL                |                |
+----------------------+-----------------+------+-----+---------------------+----------------+
```

### 🧪 追加チャレンジ（余裕があれば）
- [ ] `ApproveMatch` の処理で `Status` を `approved` に変更する時、現在の `Status` が `pending` であることをチェックするバリデーションを書いてみよう
- [ ] `repository/match.go` を作成し、`GetByTripLegID` と `GetByRequesterID` を実装してみよう
- [ ] `CompleteMatch` で `Status` が `payment_pending` でない場合にエラーを返すロジックを書いてみよう

### 📖 学習ポイント

| 概念 | 説明 |
|------|------|
| ステータスマシン | 状態遷移が決まった順序で進む設計パターン |
| `payment_pending` | 決済なしで完了させない仕組みの実装 |
| nullable FK | 会社に所属しないユーザー（荷主）に対応 |
| `request_type` | 同じテーブルで荷主→会社と会社→会社を区別する |

### 🏷️ ラベル
`feature` `backend` `database` `model` `初心者向け`

---

## Issue #010: `payments` テーブル作成（決済管理）

### 📋 概要
マッチングに対する決済記録を管理する `payments` テーブルを作成する。Stripe Payment Intent API と連携する前提。

### 🎯 ゴール
- `model/payment.go` に Payment 構造体を定義する
- `main.go` の `AutoMigrate` に追加し、`payments` テーブルが自動作成される

### 📝 前提条件
- ✅ Issue #009（matches テーブル作成）が完了していること

### 📝 タスク

#### 1. Payment モデルの作成

`internal/model/payment.go` を新規作成する。

```go
package model

import "time"

// Payment status constants
const (
    PaymentStatusPending   = "pending"
    PaymentStatusSucceeded = "succeeded"
    PaymentStatusFailed    = "failed"
    PaymentStatusRefunded  = "refunded"
)

// Payment はマッチングに対する決済記録を表す
type Payment struct {
    ID              uint       `json:"id" gorm:"primaryKey"`
    MatchID         uint       `json:"match_id" gorm:"index:idx_payments_match"`
    Match           Match      `json:"match" gorm:"foreignKey:MatchID" binding:"-"`
    PayerID         uint       `json:"payer_id"`
    Payer           User       `json:"payer" gorm:"foreignKey:PayerID" binding:"-"`
    Amount          int        `json:"amount"`
    Currency        string     `json:"currency" gorm:"type:varchar(10);default:jpy"`
    StripePaymentID string     `json:"stripe_payment_id" gorm:"type:varchar(255)"`
    Status          string     `json:"status" gorm:"type:varchar(20);default:pending"`
    PaidAt          *time.Time `json:"paid_at"`
    CreatedAt       time.Time  `json:"created_at"`
    UpdatedAt       time.Time  `json:"updated_at"`
}

// TableName specifies the table name for GORM
func (Payment) TableName() string {
    return "payments"
}
```

> 💡 **Stripe との連携イメージ**:
> ```
> 1. フロントエンド: 「決済する」ボタンをクリック
> 2. バックエンド:   Stripe API に PaymentIntent を作成
> 3. バックエンド:   payments レコードを status="pending" で INSERT
> 4. フロント:       Stripe.js のカード入力フォームで決済
> 5. Stripe:         Webhook で payment_intent.succeeded を通知
> 6. バックエンド:   payments.status を "succeeded" に UPDATE
> 7. バックエンド:   matches.status を "completed" に UPDATE
> ```

> 💡 **`PaidAt` が `*time.Time`（ポインタ）の理由**
> 決済が完了するまで `PaidAt` は値がない（NULL）。
> Go の `time.Time` のゼロ値は `0001-01-01` になってしまうので、
> ポインタにして `nil`（= DBの `NULL`）を表現する。

> 💡 **`Amount` が `int`（`float64` ではない）理由**
> お金の計算に浮動小数点（`float64`）を使うと丸め誤差が発生する。
> 例: `0.1 + 0.2 = 0.30000000000000004`
> Stripe も金額を**整数（最小通貨単位）**で扱う。日本円なら1円単位のint。

#### 2. main.go に AutoMigrate を追加

```go
if err := db.AutoMigrate(
    &model.Company{},
    &model.User{},
    &model.Vehicle{},
    &model.DispatchPlan{},
    &model.TripLeg{},
    &model.Match{},
    &model.Payment{},  // ← 追加
); err != nil {
    log.Fatal("❌ マイグレーション失敗:", err)
}
```

### ✅ 動作確認

```bash
docker compose up --build
docker compose exec db mysql -u root -p delivery_db

DESCRIBE payments;
```

### 🟢 期待される出力

```
+-------------------+-----------------+------+-----+---------+----------------+
| Field             | Type            | Null | Key | Default | Extra          |
+-------------------+-----------------+------+-----+---------+----------------+
| id                | bigint unsigned | NO   | PRI | NULL    | auto_increment |
| match_id          | bigint unsigned | YES  | MUL | NULL    |                |
| payer_id          | bigint unsigned | YES  | MUL | NULL    |                |
| amount            | bigint          | YES  |     | NULL    |                |
| currency          | varchar(10)     | YES  |     | jpy     |                |
| stripe_payment_id | varchar(255)    | YES  |     | NULL    |                |
| status            | varchar(20)     | YES  |     | pending |                |
| paid_at           | datetime(3)     | YES  |     | NULL    |                |
| created_at        | datetime(3)     | YES  |     | NULL    |                |
| updated_at        | datetime(3)     | YES  |     | NULL    |                |
+-------------------+-----------------+------+-----+---------+----------------+
```

### 🧪 追加チャレンジ（余裕があれば）
- [ ] `Match` モデルに `Payments []Payment` のリレーションを追加してみよう
- [ ] 決済成功時にマッチングのステータスも `completed` に更新するトランザクション処理を考えてみよう
- [ ] `StripePaymentID` にユニーク制約を付けてみよう（同じ決済IDの重複を防ぐ）

### 📖 学習ポイント

| 概念 | 説明 |
|------|------|
| Webhook | 外部サービス（Stripe）がイベント発生時にこちらのAPIを呼ぶ仕組み |
| `*time.Time` | 値がない場合に `nil`（NULL）を使うパターン |
| 金額は整数で扱う | 浮動小数点の丸め誤差を防ぐための業界標準 |
| PaymentIntent | Stripeの決済フロー。「支払い意図」を作成→確認→完了 の3ステップ |

### 🏷️ ラベル
`feature` `backend` `database` `model` `初心者向け`

---

## Issue #011: `chat_rooms` / `chat_messages` テーブル作成（チャット機能）

### 📋 概要
マッチング承認後に自動でチャットルームを作成し、メッセージのやり取りを可能にするための2つのテーブルを作成する。

### 🎯 ゴール
- `model/chat_room.go` に ChatRoom 構造体を定義する
- `model/chat_message.go` に ChatMessage 構造体を定義する
- `main.go` の `AutoMigrate` に追加し、両テーブルが自動作成される
- `chat_rooms.match_id` にユニーク制約が付いていることを確認する

### 📝 前提条件
- ✅ Issue #009（matches テーブル作成）が完了していること

### 📝 タスク

#### 1. ChatRoom モデルの作成

`internal/model/chat_room.go` を新規作成する。

```go
package model

import "time"

// ChatRoom はマッチングに紐づくチャットルームを表す（1マッチング = 1ルーム）
type ChatRoom struct {
    ID           uint          `json:"id" gorm:"primaryKey"`
    MatchID      uint          `json:"match_id" gorm:"uniqueIndex"`
    Match        Match         `json:"match" gorm:"foreignKey:MatchID" binding:"-"`
    Status       string        `json:"status" gorm:"type:varchar(20);default:active"`
    ChatMessages []ChatMessage `json:"chat_messages" gorm:"foreignKey:ChatRoomID" binding:"-"`
    CreatedAt    time.Time     `json:"created_at"`
    UpdatedAt    time.Time     `json:"updated_at"`
}

// TableName specifies the table name for GORM
func (ChatRoom) TableName() string {
    return "chat_rooms"
}
```

> 💡 **なぜ `MatchID` に `uniqueIndex`？**
> 1つのマッチングに対してチャットルームは**1つだけ**。
> `uniqueIndex` を付けることで、同じ `match_id` で2つ目のルームを作ろうとするとエラーになる。
> これは「1:1 リレーション」をDB側で保証する方法。

#### 2. ChatMessage モデルの作成

`internal/model/chat_message.go` を新規作成する。

```go
package model

import "time"

// ChatMessage status constants
const (
    ChatMessageStatusSent     = "sent"      // 正常に送信された
    ChatMessageStatusFiltered = "filtered"   // フィルタリングでブロックされた
    ChatMessageStatusWarned   = "warned"     // 警告付きで送信された
)

// ChatMessage はチャットルーム内の1つのメッセージを表す
type ChatMessage struct {
    ID             uint      `json:"id" gorm:"primaryKey"`
    ChatRoomID     uint      `json:"chat_room_id" gorm:"index:idx_chat_messages_room_created"`
    ChatRoom       ChatRoom  `json:"chat_room" gorm:"foreignKey:ChatRoomID" binding:"-"`
    SenderID       uint      `json:"sender_id"`
    Sender         User      `json:"sender" gorm:"foreignKey:SenderID" binding:"-"`
    Content        string    `json:"content" gorm:"type:text"`
    Status         string    `json:"status" gorm:"type:varchar(20);default:sent"`
    FilteredReason string    `json:"filtered_reason"`
    CreatedAt      time.Time `json:"created_at" gorm:"index:idx_chat_messages_room_created"`
}

// TableName specifies the table name for GORM
func (ChatMessage) TableName() string {
    return "chat_messages"
}
```

> 💡 **メッセージのステータス**:
> | Status | 意味 | どうなる？ |
> |--------|------|-----------|
> | `sent` | 正常に送信された | 相手に表示される |
> | `filtered` | フィルタリングでブロック | 相手に表示されない。送信者に警告 |
> | `warned` | 警告付きで送信された | 相手に表示されるが、違反として記録される |

> 💡 **`(chat_room_id, created_at)` の複合インデックス**
> チャット画面では「あるルームのメッセージを時系列で取得」する。
> ```sql
> SELECT * FROM chat_messages
> WHERE chat_room_id = 1
> ORDER BY created_at ASC;
> ```
> この検索パターンに合わせて `(chat_room_id, created_at)` の複合インデックスを作る。

> ⚠️ **`ChatMessage` に `UpdatedAt` がない理由**
> メッセージは一度送信したら変更しない（編集機能なし）。
> 不要なカラムは作らない。これもDB設計のポイント。

#### 3. main.go に AutoMigrate を追加

```go
if err := db.AutoMigrate(
    &model.Company{},
    &model.User{},
    &model.Vehicle{},
    &model.DispatchPlan{},
    &model.TripLeg{},
    &model.Match{},
    &model.Payment{},
    &model.ChatRoom{},     // ← 追加
    &model.ChatMessage{},  // ← 追加
); err != nil {
    log.Fatal("❌ マイグレーション失敗:", err)
}
```

### ✅ 動作確認

```bash
docker compose up --build
docker compose exec db mysql -u root -p delivery_db

DESCRIBE chat_rooms;
DESCRIBE chat_messages;
SHOW INDEX FROM chat_messages;
```

### 🟢 期待される出力（chat_rooms）

```
+------------+-----------------+------+-----+---------+----------------+
| Field      | Type            | Null | Key | Default | Extra          |
+------------+-----------------+------+-----+---------+----------------+
| id         | bigint unsigned | NO   | PRI | NULL    | auto_increment |
| match_id   | bigint unsigned | YES  | UNI | NULL    |                |
| status     | varchar(20)     | YES  |     | active  |                |
| created_at | datetime(3)     | YES  |     | NULL    |                |
| updated_at | datetime(3)     | YES  |     | NULL    |                |
+------------+-----------------+------+-----+---------+----------------+
```

### 🧪 追加チャレンジ（余裕があれば）
- [ ] `handler/chat.go` を作成し、`SendMessage` と `GetMessages` のAPIハンドラーを書いてみよう
- [ ] マッチング承認時（`ApproveMatch`）にチャットルームを自動作成するロジックを追加してみよう
- [ ] WebSocket を調べて、リアルタイムチャットの実装方法を考えてみよう

### 📖 学習ポイント

| 概念 | 説明 |
|------|------|
| 1:1 リレーション | `uniqueIndex` で1つのマッチングに1つのルームを保証 |
| HasMany | 1つのルームに複数のメッセージ |
| 複合インデックス | 検索パターン（WHERE + ORDER BY）に合わせて設計 |
| 不要なカラムは作らない | `UpdatedAt` が不要なら省略する |

### 🏷️ ラベル
`feature` `backend` `database` `model` `初心者向け`

---

## Issue #012: `blocked_patterns` / `user_violations` テーブル作成（チャットフィルタリング）

### 📋 概要
チャットメッセージに含まれる電話番号・メールアドレス・URLを検知しブロックする仕組みの土台となる2つのテーブルを作成する。

### 🎯 ゴール
- `model/blocked_pattern.go` に BlockedPattern 構造体を定義する
- `model/user_violation.go` に UserViolation 構造体を定義する
- 初期パターンデータ（電話番号・メール・URL）を投入するシーダーを作成する

### 📝 前提条件
- ✅ Issue #011（chat_rooms / chat_messages テーブル作成）が完了していること

### 📝 タスク

#### 1. BlockedPattern モデルの作成

`internal/model/blocked_pattern.go` を新規作成する。

```go
package model

import "time"

// BlockedPattern はチャットでブロックする正規表現パターンを表す
type BlockedPattern struct {
    ID          uint      `json:"id" gorm:"primaryKey"`
    Pattern     string    `json:"pattern" gorm:"type:varchar(500)" binding:"required"`
    Description string    `json:"description" gorm:"type:varchar(255)"`
    IsActive    bool      `json:"is_active" gorm:"default:true"`
    CreatedAt   time.Time `json:"created_at"`
}

// TableName specifies the table name for GORM
func (BlockedPattern) TableName() string {
    return "blocked_patterns"
}
```

> 💡 **初期パターンデータ**:
> | Pattern | Description | 検知例 |
> |---------|-------------|-------|
> | `\d{2,4}-\d{2,4}-\d{4}` | 電話番号（ハイフンあり） | `090-1234-5678` |
> | `0\d{9,10}` | 電話番号（ハイフンなし） | `09012345678` |
> | `[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+` | メールアドレス | `test@example.com` |
> | `https?://\S+` | URL | `https://example.com` |

#### 2. UserViolation モデルの作成

`internal/model/user_violation.go` を新規作成する。

```go
package model

import "time"

// UserViolation はユーザーの違反記録を表す
type UserViolation struct {
    ID            uint        `json:"id" gorm:"primaryKey"`
    UserID        uint        `json:"user_id" gorm:"index:idx_user_violations_user"`
    User          User        `json:"user" gorm:"foreignKey:UserID" binding:"-"`
    ChatMessageID uint        `json:"chat_message_id"`
    ChatMessage   ChatMessage `json:"chat_message" gorm:"foreignKey:ChatMessageID" binding:"-"`
    ViolationType string      `json:"violation_type" gorm:"type:varchar(30)"`
    Detail        string      `json:"detail"`
    CreatedAt     time.Time   `json:"created_at"`
}

// TableName specifies the table name for GORM
func (UserViolation) TableName() string {
    return "user_violations"
}
```

> 💡 **ViolationType の種類**:
> | 値 | 意味 |
> |----|------|
> | `contact_info_leak` | 連絡先（電話番号・メール等）の送信を試みた |
> | `spam` | スパムメッセージ |
> | `abuse` | 暴言・嫌がらせ |

> 💡 **なぜ違反を記録するの？**
> 繰り返し違反するユーザーを特定し、アカウント停止（`is_active = false`）などの措置を取るため。
> 違反回数が閾値を超えたら自動で警告メールを送る、といった拡張も可能。

#### 3. 初期データ投入（シーダー）

`internal/seed/blocked_patterns.go` を新規作成する。

```go
package seed

import (
    "log"

    "github.com/delivery-app/delivery-api/internal/model"
    "gorm.io/gorm"
)

// SeedBlockedPatterns は初期のフィルタリングパターンを投入する
func SeedBlockedPatterns(db *gorm.DB) {
    patterns := []model.BlockedPattern{
        {Pattern: `\d{2,4}-\d{2,4}-\d{4}`, Description: "電話番号（ハイフンあり）", IsActive: true},
        {Pattern: `0\d{9,10}`, Description: "電話番号（ハイフンなし）", IsActive: true},
        {Pattern: `[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+`, Description: "メールアドレス", IsActive: true},
        {Pattern: `https?://\S+`, Description: "URL", IsActive: true},
    }

    for _, p := range patterns {
        // 既に同じパターンが存在する場合はスキップ
        var count int64
        db.Model(&model.BlockedPattern{}).Where("pattern = ?", p.Pattern).Count(&count)
        if count == 0 {
            db.Create(&p)
            log.Printf("✅ パターン追加: %s", p.Description)
        }
    }
}
```

> 💡 **シーダー（Seeder）とは？**
> テーブル作成後に初期データを投入するスクリプトのこと。
> `AutoMigrate` はテーブルの**構造**を作るだけで、**データ**は入れない。
> blocked_patterns のような「最初からデータが必要なテーブル」にはシーダーが必要。

> ⚠️ **`Count` で存在チェックしてからの `Create`**
> シーダーは何度実行しても安全（**冪等性**）であるべき。
> 既にデータがある場合にまた INSERT すると重複する。
> だから「既に存在するかチェック → なければ追加」という流れにする。

#### 4. main.go でシーダーを呼ぶ

```go
import "github.com/delivery-app/delivery-api/internal/seed"

// AutoMigrate の後に追加
seed.SeedBlockedPatterns(db)
```

#### 5. main.go の AutoMigrate を更新

```go
if err := db.AutoMigrate(
    // ... 既存のモデル ...
    &model.BlockedPattern{},   // ← 追加
    &model.UserViolation{},    // ← 追加
); err != nil {
    log.Fatal("❌ マイグレーション失敗:", err)
}
```

### ✅ 動作確認

```bash
docker compose up --build
docker compose exec db mysql -u root -p delivery_db

DESCRIBE blocked_patterns;
DESCRIBE user_violations;

-- 初期データの確認
SELECT * FROM blocked_patterns;
```

### 🟢 期待される出力（SELECT）

```
+----+-------------------------------------+---------------------------------+-----------+---------------------+
| id | pattern                             | description                     | is_active | created_at          |
+----+-------------------------------------+---------------------------------+-----------+---------------------+
|  1 | \d{2,4}-\d{2,4}-\d{4}              | 電話番号（ハイフンあり）          | 1         | 2026-03-12 ...      |
|  2 | 0\d{9,10}                           | 電話番号（ハイフンなし）          | 1         | 2026-03-12 ...      |
|  3 | [a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+   | メールアドレス                   | 1         | 2026-03-12 ...      |
|  4 | https?://\S+                        | URL                             | 1         | 2026-03-12 ...      |
+----+-------------------------------------+---------------------------------+-----------+---------------------+
```

### 🧪 追加チャレンジ（余裕があれば）
- [ ] Go の `regexp` パッケージでメッセージ文字列にパターンマッチを試してみよう
- [ ] `handler/chat.go` の `SendMessage` にフィルタリングロジックを組み込んでみよう
- [ ] LINE風の「ちょっと似た文字列」（例: `零九零`）もフィルタリングする方法を考えてみよう

### 📖 学習ポイント

| 概念 | 説明 |
|------|------|
| シーダー (Seeder) | 初期データを投入するスクリプト |
| 冪等性 | 何度実行しても同じ結果になる性質 |
| 正規表現 | パターンマッチングの仕組み。Go では `regexp` パッケージを使う |
| `is_active` フラグ | レコードを削除せず無効化するパターン（ソフト無効化） |

### 🏷️ ラベル
`feature` `backend` `database` `model` `初心者向け`

---

## Issue #013: `reviews` テーブル作成（評価・レビュー）

### 📋 概要
マッチング完了後の相互評価（星1〜5 + コメント）を管理する `reviews` テーブルを作成する。

### 🎯 ゴール
- `model/review.go` に Review 構造体を定義する
- `main.go` の `AutoMigrate` に追加し、`reviews` テーブルが自動作成される
- 1マッチングにつき各者1回のみ評価可能な制約を確認する

### 📝 前提条件
- ✅ Issue #009（matches テーブル作成）が完了していること
- ✅ Issue #005（companies テーブル作成）が完了していること

### 📝 タスク

#### 1. Review モデルの作成

`internal/model/review.go` を新規作成する。

```go
package model

import "time"

// Review はマッチング完了後の相互評価を表す
type Review struct {
    ID                uint    `json:"id" gorm:"primaryKey"`
    MatchID           uint    `json:"match_id" gorm:"uniqueIndex:idx_review_unique,priority:1"`
    Match             Match   `json:"match" gorm:"foreignKey:MatchID" binding:"-"`
    ReviewerID        uint    `json:"reviewer_id" gorm:"uniqueIndex:idx_review_unique,priority:2"`
    Reviewer          User    `json:"reviewer" gorm:"foreignKey:ReviewerID" binding:"-"`
    RevieweeCompanyID uint    `json:"reviewee_company_id" gorm:"index:idx_reviews_company"`
    RevieweeCompany   Company `json:"reviewee_company" gorm:"foreignKey:RevieweeCompanyID" binding:"-"`
    Rating            int     `json:"rating" binding:"required,min=1,max=5"`
    Comment           string  `json:"comment"`
    CreatedAt         time.Time `json:"created_at"`
}

// TableName specifies the table name for GORM
func (Review) TableName() string {
    return "reviews"
}
```

> 💡 **複合ユニーク制約 `idx_review_unique`**:
> ```go
> MatchID    uint `gorm:"uniqueIndex:idx_review_unique,priority:1"`
> ReviewerID uint `gorm:"uniqueIndex:idx_review_unique,priority:2"`
> ```
> **同じインデックス名 `idx_review_unique` を2つのフィールドに付ける**と、
> GORMは `(match_id, reviewer_id)` の複合ユニーク制約を作成する。
> これにより「同じ人が同じマッチングに2回評価する」ことをDB側で防止する。
>
> `priority` はインデックス内のカラム順序を指定する。
> `priority:1` が先、`priority:2` が後。

> 💡 **なぜ個人ではなく「会社に対する評価」なの？**
> ドライバーは日替わりで変わるかもしれないが、会社の信頼性は蓄積される。
> `RevieweeCompanyID` で「どの会社への評価か」を記録し、
> `companies.rating_avg` に集計結果を反映する。

> 💡 **`binding:"required,min=1,max=5"` の意味**
> Ginのバリデーション。APIリクエストで `rating` が必須かつ1〜5の範囲であることを強制する。
> これにより不正な値（0や100等）がDBに入るのを防ぐ。

#### 2. main.go の AutoMigrate を更新

```go
if err := db.AutoMigrate(
    // ... 既存のモデル ...
    &model.Review{},  // ← 追加
); err != nil {
    log.Fatal("❌ マイグレーション失敗:", err)
}
```

### ✅ 動作確認

```bash
docker compose up --build
docker compose exec db mysql -u root -p delivery_db

DESCRIBE reviews;
SHOW INDEX FROM reviews;
```

### 🟢 期待される出力

```
+---------------------+-----------------+------+-----+---------+----------------+
| Field               | Type            | Null | Key | Default | Extra          |
+---------------------+-----------------+------+-----+---------+----------------+
| id                  | bigint unsigned | NO   | PRI | NULL    | auto_increment |
| match_id            | bigint unsigned | YES  | MUL | NULL    |                |
| reviewer_id         | bigint unsigned | YES  |     | NULL    |                |
| reviewee_company_id | bigint unsigned | YES  | MUL | NULL    |                |
| rating              | bigint          | YES  |     | NULL    |                |
| comment             | longtext        | YES  |     | NULL    |                |
| created_at          | datetime(3)     | YES  |     | NULL    |                |
+---------------------+-----------------+------+-----+---------+----------------+
```

### 🧪 追加チャレンジ（余裕があれば）
- [ ] `companies.rating_avg` を更新するバッチ処理（`AVG(rating)` でSELECTして UPDATE）を書いてみよう
- [ ] マッチング完了後にしか評価できないバリデーション（`match.Status == "completed"` のチェック）を追加してみよう
- [ ] `repository/review.go` を作成し、`GetByCompanyID` で会社の全レビューを取得する機能を実装してみよう

### 📖 学習ポイント

| 概念 | 説明 |
|------|------|
| 複合ユニーク制約 | 2つのカラムの組み合わせが一意であることを保証 |
| `priority` | 複合インデックス内のカラム順序を指定 |
| Ginのバリデーション | `binding` タグでリクエストの値を検証 |
| 集計と非正規化 | `rating_avg` は毎回 AVG() で計算せず、companies に持たせる（パフォーマンスのため） |

### 🏷️ ラベル
`feature` `backend` `database` `model` `初心者向け`

---

## Issue #014: `subscriptions` テーブル作成（サブスクリプション / Stripe Billing）

### 📋 概要
運送会社向けの月額サブスクリプション管理テーブルを作成する。Stripe Billing API と連携する前提。

### 🎯 ゴール
- `model/subscription.go` に Subscription 構造体を定義する
- `main.go` の `AutoMigrate` に追加し、`subscriptions` テーブルが自動作成される

### 📝 前提条件
- ✅ Issue #005（companies テーブル作成）が完了していること

### 📝 タスク

#### 1. Subscription モデルの作成

`internal/model/subscription.go` を新規作成する。

```go
package model

import "time"

// Subscription plan constants
const (
    PlanFree     = "free"      // 無料: ソロモードのみ、月3件マッチング
    PlanStandard = "standard"  // ¥9,800/月: マッチング無制限、チャット
    PlanPremium  = "premium"   // ¥29,800/月: 全機能、API連携、優先表示
)

// Subscription status constants
const (
    SubscriptionStatusActive    = "active"
    SubscriptionStatusCancelled = "cancelled"
    SubscriptionStatusPastDue   = "past_due"
)

// Subscription は運送会社の月額サブスクリプションを表す
type Subscription struct {
    ID                   uint      `json:"id" gorm:"primaryKey"`
    CompanyID            uint      `json:"company_id" gorm:"index:idx_subscriptions_company"`
    Company              Company   `json:"company" gorm:"foreignKey:CompanyID" binding:"-"`
    Plan                 string    `json:"plan" gorm:"type:varchar(20);default:free"`
    StripeSubscriptionID string    `json:"stripe_subscription_id" gorm:"type:varchar(255)"`
    Status               string    `json:"status" gorm:"type:varchar(20);default:active"`
    CurrentPeriodStart   time.Time `json:"current_period_start"`
    CurrentPeriodEnd     time.Time `json:"current_period_end"`
    CreatedAt            time.Time `json:"created_at"`
    UpdatedAt            time.Time `json:"updated_at"`
}

// TableName specifies the table name for GORM
func (Subscription) TableName() string {
    return "subscriptions"
}
```

> 💡 **プラン一覧**:
> | プラン | 料金 | 機能 |
> |--------|------|------|
> | `free` | 無料 | ソロモード（配車管理）のみ。公開マッチングは月3件まで |
> | `standard` | ¥9,800/月 | 公開マッチング無制限。チャット機能。基本統計 |
> | `premium` | ¥29,800/月 | 全機能。API連携。優先表示。分析レポート |

> 💡 **Stripe Billing との連携イメージ**:
> ```
> 1. 会社オーナーが「Standard」プランを選択
> 2. バックエンド: Stripe Checkout Session を作成
> 3. フロント:     Stripeの決済画面にリダイレクト
> 4. ユーザー:     カード情報を入力して決済
> 5. Stripe:       Webhook で invoice.paid を通知
> 6. バックエンド: subscriptions.status = "active" に UPDATE
> 7. バックエンド: companies.plan = "standard" に UPDATE
>
> 毎月自動で:
> 8. Stripe:       自動課金 → invoice.paid
> 9. バックエンド: current_period_start/end を更新
>
> 支払い失敗時:
> 10. Stripe:      customer.subscription.updated (status: past_due)
> 11. バックエンド: subscriptions.status = "past_due" に UPDATE
> ```

> ⚠️ **`CurrentPeriodStart/End` の用途**
> マッチング件数の月間カウントに使う。
> free プランは月3件までなので、「今月の開始日〜終了日」の間の
> マッチング数をカウントして制限する。

#### 2. main.go の AutoMigrate を更新

```go
if err := db.AutoMigrate(
    // ... 既存のモデル ...
    &model.Subscription{},  // ← 追加
); err != nil {
    log.Fatal("❌ マイグレーション失敗:", err)
}
```

### ✅ 動作確認

```bash
docker compose up --build
docker compose exec db mysql -u root -p delivery_db

DESCRIBE subscriptions;
```

### 🟢 期待される出力

```
+------------------------+-----------------+------+-----+---------+----------------+
| Field                  | Type            | Null | Key | Default | Extra          |
+------------------------+-----------------+------+-----+---------+----------------+
| id                     | bigint unsigned | NO   | PRI | NULL    | auto_increment |
| company_id             | bigint unsigned | YES  | MUL | NULL    |                |
| plan                   | varchar(20)     | YES  |     | free    |                |
| stripe_subscription_id | varchar(255)    | YES  |     | NULL    |                |
| status                 | varchar(20)     | YES  |     | active  |                |
| current_period_start   | datetime(3)     | YES  |     | NULL    |                |
| current_period_end     | datetime(3)     | YES  |     | NULL    |                |
| created_at             | datetime(3)     | YES  |     | NULL    |                |
| updated_at             | datetime(3)     | YES  |     | NULL    |                |
+------------------------+-----------------+------+-----+---------+----------------+
```

### 🧪 追加チャレンジ（余裕があれば）
- [ ] free プランの月間マッチング数チェック（3件制限）のロジックを書いてみよう
- [ ] Stripe Webhook のハンドラー（`invoice.paid` イベント処理）を考えてみよう
- [ ] `Company` モデルに `Subscription` リレーション（HasOne）を追加してみよう

### 📖 学習ポイント

| 概念 | 説明 |
|------|------|
| Stripe Billing | Stripeの定期課金サービス。毎月自動で課金してくれる |
| Checkout Session | Stripeが用意する決済画面。自分でフォームを作る必要がない |
| Webhook | Stripeが決済完了時にこちらのAPIを呼んで通知する仕組み |
| `past_due` | 支払い失敗。カード期限切れ等。猶予期間内に解決しないと `cancelled` になる |

### 🏷️ ラベル
`feature` `backend` `database` `model` `初心者向け`

---

## Issue #015: `trackings` テーブル作成（位置記録・予測位置の基盤）

### 📋 概要
トラックの位置情報を記録する `trackings` テーブルを作成する。配車計画・区間に紐づける形で設計する。

### 🎯 ゴール
- `model/tracking.go` に Tracking 構造体を定義する
- `main.go` の `AutoMigrate` に追加し、`trackings` テーブルが自動作成される

### 📝 前提条件
- ✅ Issue #007（dispatch_plans テーブル作成）が完了していること
- ✅ Issue #008（trip_legs テーブル作成）が完了していること

### 📝 タスク

#### 1. Tracking モデルの作成

`internal/model/tracking.go` を新規作成する。

```go
package model

import "time"

// Tracking は配車計画の位置情報記録を表す
type Tracking struct {
    ID             uint         `json:"id" gorm:"primaryKey"`
    DispatchPlanID uint         `json:"dispatch_plan_id" gorm:"index:idx_trackings_plan_recorded"`
    DispatchPlan   DispatchPlan `json:"dispatch_plan" gorm:"foreignKey:DispatchPlanID" binding:"-"`
    TripLegID      *uint        `json:"trip_leg_id"`
    TripLeg        TripLeg      `json:"trip_leg" gorm:"foreignKey:TripLegID" binding:"-"`
    Lat            float64      `json:"lat"`
    Lng            float64      `json:"lng"`
    RecordedAt     time.Time    `json:"recorded_at" gorm:"index:idx_trackings_plan_recorded"`
}

// TableName specifies the table name for GORM
func (Tracking) TableName() string {
    return "trackings"
}
```

> 💡 **`TripLegID` が `*uint`（nullable）の理由**
> 位置記録は配車計画全体に紐づくが、「今どの区間にいるか」は必ずしも特定できない。
> 例: 区間1の到着地と区間2の出発地の間（休憩中など）はどちらの区間でもない。
> なので `TripLegID` はNULL許容にしている。

> 💡 **`(dispatch_plan_id, recorded_at)` のインデックス**
> 位置情報の取得パターン:
> ```sql
> SELECT * FROM trackings
> WHERE dispatch_plan_id = 1
> ORDER BY recorded_at DESC
> LIMIT 1;
> ```
> 「ある配車計画の最新位置」を取得するクエリに最適化。

> ⚠️ **このプロジェクトではGPSは使わない**
> ドライバーがGPSをONにすることを嫌がるため、位置情報は
> 「出発時刻・到着時刻・ルート情報」から**予測**する。
> `trackings` テーブルは「ワンタップ報告」（出発ボタン・到着ボタン）の
> 記録や、将来的なGPS対応の基盤として用意している。

#### 2. main.go の AutoMigrate を更新

```go
if err := db.AutoMigrate(
    // ... 既存のモデル ...
    &model.Tracking{},  // ← 追加
); err != nil {
    log.Fatal("❌ マイグレーション失敗:", err)
}
```

### ✅ 動作確認

```bash
docker compose up --build
docker compose exec db mysql -u root -p delivery_db

DESCRIBE trackings;
SHOW INDEX FROM trackings;
```

### 🟢 期待される出力

```
+------------------+-----------------+------+-----+---------+----------------+
| Field            | Type            | Null | Key | Default | Extra          |
+------------------+-----------------+------+-----+---------+----------------+
| id               | bigint unsigned | NO   | PRI | NULL    | auto_increment |
| dispatch_plan_id | bigint unsigned | YES  | MUL | NULL    |                |
| trip_leg_id      | bigint unsigned | YES  | MUL | NULL    |                |
| lat              | double          | YES  |     | NULL    |                |
| lng              | double          | YES  |     | NULL    |                |
| recorded_at      | datetime(3)     | YES  |     | NULL    |                |
+------------------+-----------------+------+-----+---------+----------------+
```

### 🧪 追加チャレンジ（余裕があれば）
- [ ] `util/predict_location.go` を読んで、予測位置の計算ロジックを理解してみよう
- [ ] ワンタップ報告のAPIエンドポイント（`POST /api/v1/trackings`）を設計してみよう
- [ ] `EXPLAIN` を使って `(dispatch_plan_id, recorded_at)` インデックスが使われるか確認してみよう

### 📖 学習ポイント

| 概念 | 説明 |
|------|------|
| nullable FK | NULL許容の外部キー。紐づけ先が不明な場合に使う |
| 時系列データ | `recorded_at` でソートして取得する位置情報は典型的な時系列データ |
| インデックスの効果 | 大量のレコードでも特定の配車計画の最新位置を高速に取得できる |

### 🏷️ ラベル
`feature` `backend` `database` `model` `初心者向け`

---

## Issue #016: 段階的情報開示の API ロジック実装

### 📋 概要
マッチングの段階に応じて会社情報の公開範囲を制御するAPIレスポンスフィルタリングを実装する。

### 🎯 ゴール
- 検索時: エリア・車種・積載量・料金のみ表示
- マッチング申請後: 評価・実績件数・匿名表示名を追加
- 承認後: 正式会社名・チャット開放
- 決済完了後: 電話番号・住所・運転手名を開示

### 📝 前提条件
- ✅ Issue #005〜#010（全モデル作成）が完了していること

### 📝 タスク

#### 1. レスポンス用のDTO構造体を定義

`internal/dto/company_response.go` を新規作成する。

```go
package dto

// CompanySearchResponse は検索時の会社情報（最小限）
type CompanySearchResponse struct {
    DisplayName string  `json:"display_name"`
    // 電話番号・正式名称・住所は含めない
}

// CompanyMatchedResponse はマッチング申請後の会社情報
type CompanyMatchedResponse struct {
    DisplayName string  `json:"display_name"`
    RatingAvg   float64 `json:"rating_avg"`
    TotalDeals  int     `json:"total_deals"`
}

// CompanyApprovedResponse は承認後の会社情報
type CompanyApprovedResponse struct {
    Name        string  `json:"name"`
    DisplayName string  `json:"display_name"`
    RatingAvg   float64 `json:"rating_avg"`
    TotalDeals  int     `json:"total_deals"`
}

// CompanyFullResponse は決済完了後の会社情報（フル開示）
type CompanyFullResponse struct {
    Name        string  `json:"name"`
    DisplayName string  `json:"display_name"`
    Address     string  `json:"address"`
    Phone       string  `json:"phone"`
    Email       string  `json:"email"`
    RatingAvg   float64 `json:"rating_avg"`
    TotalDeals  int     `json:"total_deals"`
}
```

> 💡 **DTO（Data Transfer Object）とは？**
> DBのモデルをそのままAPIレスポンスにすると、不要なフィールド（パスワードハッシュ等）や
> 今回のように「まだ見せたくない情報」まで含まれてしまう。
> DTOはAPIレスポンス専用の構造体で、「どの段階でどの情報を返すか」を明確に制御する。

> 💡 **段階的情報開示の一覧**:
> | 段階 | 表示される情報 |
> |------|---------------|
> | 検索時 | `display_name` のみ |
> | マッチング申請後 | + `rating_avg`, `total_deals` |
> | 承認後 | + `name`（正式会社名） |
> | 決済完了後 | + `phone`, `address`, `email` |

#### 2. ハンドラーでステータスに応じたレスポンスを返す

```go
// handler/trip_leg.go の検索API例
func (h *TripLegHandler) SearchPublicLegs(c *gin.Context) {
    // ... 検索ロジック ...

    // 検索段階なので最小限の情報だけ返す
    response := dto.CompanySearchResponse{
        DisplayName: leg.DispatchPlan.Company.DisplayName,
    }
    // ⚠️ leg.DispatchPlan.Company.Phone は含めない！
}
```

#### 3. ディレクトリ構成の確認

```
delivery-api/
├── internal/
│   ├── dto/
│   │   └── company_response.go   ← ⭐ 今回作成
│   ├── handler/
│   │   └── trip_leg.go           ← 検索APIでDTO使用
│   └── model/
│       └── ...
```

### ✅ 動作確認

各APIエンドポイントを叩いて、段階に応じた情報が返ることを確認:

```bash
# 検索API（ログインなし or 未マッチング）
curl http://localhost:8080/api/v1/trip-legs/search
# → display_name のみ

# マッチング詳細（申請後）
curl -H "Authorization: Bearer <token>" http://localhost:8080/api/v1/matches/1
# → display_name + rating_avg + total_deals

# 決済完了後
# → phone, address 含む
```

### 🧪 追加チャレンジ（余裕があれば）
- [ ] ミドルウェアでマッチング段階を判定し、自動的にレスポンスをフィルタリングする仕組みを考えてみよう
- [ ] テストを書いて「検索時に phone が含まれていないこと」を確認してみよう

### 📖 学習ポイント

| 概念 | 説明 |
|------|------|
| DTO | APIレスポンス専用の構造体。モデルとは分離して管理 |
| 情報開示制御 | ビジネスロジックに基づいてレスポンスを制御するパターン |
| マネタイズ設計 | 「情報を段階的に開示する」ことで決済を促す仕組み |

### 🏷️ ラベル
`feature` `backend` `api` `初心者向け`

---

## Issue #017: 配車ボード画面（複数ルート同時表示 / フロントエンド）

### 📋 概要
1つの地図に自社の全トラックのルートを色分けして重ねて表示し、各トラックの予測位置をリアルタイム風に表示する「配車ボード」画面を新規作成する。

### 🎯 ゴール
- `DispatchBoardPage.jsx` を新規作成し、`/dispatch-board` でアクセスできるようにする
- 1つの地図に複数ルート（色分け）が表示される
- 各トラックの予測位置マーカーが60秒ごとに自動更新される

### 📝 前提条件
- ✅ Issue #007（dispatch_plans テーブル作成）が完了していること
- ✅ Issue #008（trip_legs テーブル作成）が完了していること
- ✅ バックエンド: `GET /api/v1/dispatch-plans?company_id=X&date=YYYY-MM-DD` が実装済み
- ✅ バックエンド: `GET /api/v1/dispatch-plans/:id/predict` が実装済み

### 📝 タスク

#### 1. 配車ボードの全体像

```
┌──────────────────────────────────────────────────────────┐
│  配車ボード（2026-03-12）                  [日付選択]     │
│                                                          │
│  ┌────────────────────────────────────────────────────┐  │
│  │              Google Map                            │  │
│  │                                                    │  │
│  │   🔴 トラック① (赤ルート + 予測位置マーカー)       │  │
│  │   🟢 トラック② (緑ルート + 予測位置マーカー)       │  │
│  │   🔵 トラック③ (青ルート + 予測位置マーカー)       │  │
│  │                                                    │  │
│  └────────────────────────────────────────────────────┘  │
│                                                          │
│  トラック一覧:                                            │
│  ┌───────┬─────────────┬────────┬───────┬──────────┐     │
│  │ 車番   │ ルート       │ 進捗   │ 状態   │ 帰り荷   │    │
│  │ ① 2t  │ 横浜→川崎   │ 70%    │ 配送中 │ あり     │    │
│  │ ② 4t  │ 横浜→厚木   │ 45%    │ 配送中 │ あり     │    │
│  │ ③ 10t │ 横浜→小田原  │ 30%   │ 配送中 │ 募集中!  │    │
│  └───────┴─────────────┴────────┴───────┴──────────┘     │
└──────────────────────────────────────────────────────────┘
```

#### 2. 主要な実装ポイント

```jsx
// 複数ルートの色分け表示（DirectionsRendererを複数使う）
const ROUTE_COLORS = ['#FF0000', '#00AA00', '#0000FF', '#FF6600', '#9900CC'];

// 各配車計画ごとにDirectionsRendererを作成
plans.forEach((plan, index) => {
    const directionsRenderer = new google.maps.DirectionsRenderer({
        map: map,
        polylineOptions: {
            strokeColor: ROUTE_COLORS[index % ROUTE_COLORS.length],
            strokeWeight: 4,
        },
        suppressMarkers: false,
    });
    // DirectionsServiceでルートを取得して描画
});
```

> 💡 **複数の `DirectionsRenderer`**
> Google Maps の `DirectionsRenderer` は1つにつき1ルートしか描画できない。
> 複数ルートを同時に表示するには、トラックの数だけ `DirectionsRenderer` を作成する。

> 💡 **予測位置の自動更新**
> ```jsx
> useEffect(() => {
>     const interval = setInterval(() => {
>         // 全トラックの予測位置を再取得
>         fetchAllPredictions();
>     }, 60000); // 60秒ごと
>     return () => clearInterval(interval);
> }, []);
> ```

#### 3. バックエンドAPIの設計

```
GET /api/v1/dispatch-plans?company_id=1&date=2026-03-12

レスポンス:
[
  {
    "id": 1,
    "vehicle": { "plate_number": "横浜 100 あ 1234", "vehicle_type": "10t" },
    "driver": { "name": "田中太郎" },
    "trip_legs": [
      { "leg_order": 1, "origin_address": "横浜", "dest_address": "小田原", ... },
      { "leg_order": 2, "origin_address": "小田原", "dest_address": "横浜", ... }
    ]
  },
  ...
]

GET /api/v1/dispatch-plans/1/predict

レスポンス:
{
  "lat": 35.3192,
  "lng": 139.2434,
  "progress_percent": 50.0,
  "current_leg_order": 1,
  "elapsed_seconds": 10800,
  "remaining_seconds": 10800
}
```

#### 4. ディレクトリ構成（フロントエンド）

```
delivery-frontend/
├── src/
│   ├── pages/
│   │   └── DispatchBoardPage.jsx    ← ⭐ 今回作成
│   └── components/
│       └── MultiRouteMap.jsx        ← ⭐ 今回作成
```

### ✅ 動作確認

1. テストデータとして3台分の配車計画を作成
2. `/dispatch-board` にアクセス
3. 地図に3本のルートが色分けで表示されることを確認
4. 各トラックの予測位置マーカーが表示されることを確認
5. 60秒後にマーカーが更新されることを確認

### 🧪 追加チャレンジ（余裕があれば）
- [ ] トラック一覧テーブルの行クリックで地図上の該当ルートをハイライトする機能を追加してみよう
- [ ] マーカークリックで InfoWindow（車番・進捗・残り時間）を表示してみよう
- [ ] 帰り荷が「募集中」の区間だけを目立つ色で表示してみよう

### 📖 学習ポイント

| 概念 | 説明 |
|------|------|
| 複数 DirectionsRenderer | Google Mapsで複数ルートを同時表示する方法 |
| `setInterval` | 一定間隔で処理を繰り返すJavaScriptの関数 |
| `useEffect` + `clearInterval` | Reactでタイマーを安全に管理するパターン |
| Preload (GORM) | 配車計画 → 車両 → 区間 を1回のクエリで取得するテクニック |

### 🏷️ ラベル
`feature` `frontend` `google-maps` `初心者向け`

---

## 📊 全体の依存関係マップ

```
Issue #004 (users) ← 完了済み
      │
      ▼
Issue #005 (companies + users修正)
      │
      ├──────────────────────┐
      ▼                      ▼
Issue #006 (vehicles)   Issue #014 (subscriptions)
      │
      ▼
Issue #007 (dispatch_plans)
      │
      ▼
Issue #008 (trip_legs)
      │
      ├──────────────┐──────────────────┐
      ▼              ▼                  ▼
Issue #009      Issue #015         Issue #017
(matches)       (trackings)        (配車ボード)
      │
      ├──────────────┐──────────────────┐
      ▼              ▼                  ▼
Issue #010      Issue #011         Issue #013
(payments)      (chat_rooms/       (reviews)
      │          messages)
      │              │
      ▼              ▼
Issue #016      Issue #012
(段階的情報     (blocked_patterns/
 開示API)        user_violations)
```

> 💡 **進め方のおすすめ**:
> 1. まず #005 → #006 → #007 → #008 を順にやる（基盤テーブル）
> 2. 次に #009 → #010（マッチング・決済）
> 3. 並行して #011 → #012（チャット・フィルタリング）
> 4. 最後に #013, #014, #015, #016, #017（付加機能）
