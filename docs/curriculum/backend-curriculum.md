# デリバリー管理API - Go + Gin バックエンド 完全カリキュラム

**対象：** Go基礎を習得済みで、既にGo/Ginでポートフォリオを構築した学生
**目標：** エンタープライズ級のバックエンド開発の実践的な知識を習得する
**作成日：** 2026年3月

---

## 目次

1. [アーキテクチャ概要](#アーキテクチャ概要)
2. [ファイル構成と詳細解説](#ファイル構成と詳細解説)
3. [APIエンドポイント](#apiエンドポイント)
4. [データベース](#データベース)
5. [Docker環境構築](#docker環境構築)
6. [実装パターンと学習ポイント](#実装パターンと学習ポイント)

---

## アーキテクチャ概要

このアプリケーションは**クリーンアーキテクチャ**の原則に従っており、以下の4つのレイヤーで構成されています：

```
HTTP リクエスト
     ↓
┌─────────────────────────────────────┐
│  Handler Layer （ハンドラレイヤー）     │
│  HTTP入出力処理、リクエスト解析       │
└─────────────────────────────────────┘
     ↓
┌─────────────────────────────────────┐
│  Repository Layer （リポジトリレイヤー） │
│  データベース操作（CRUD）            │
└─────────────────────────────────────┘
     ↓
┌─────────────────────────────────────┐
│  Model Layer （モデルレイヤー）       │
│  データ構造定義、ビジネスロジック    │
└─────────────────────────────────────┘
     ↓
   MySQL データベース
```

### レイヤーの役割

**Handler Layer（ハンドラレイヤー）**
- GinフレームワークのHTTPハンドラー
- クライアントのリクエストを受け取る
- JSON形式のデータをバリデーションしながらパース
- レスポンスをJSON形式で返す
- HTTPステータスコードを適切に設定する

**Repository Layer（リポジトリレイヤー）**
- データベースとのやり取りを担当
- GORM ORM（Object-Relational Mapping）を使用
- SQL文は直接書かず、Go構造体を操作
- ビジネスロジックはHandler層に任せ、ここはデータ操作のみ

**Model Layer（モデルレイヤー）**
- Delivery構造体の定義
- データベーステーブルとの対応付け
- struct tagで、JSON形式やGORMの挙動を定義

**Router Layer（ルーターレイヤー）**
- すべてのエンドポイントを定義
- Middleware（ミドルウェア）を適用
- Dependency Injection（依存性注入）により、Handler層に必要なデータを渡す

---

## ファイル構成と詳細解説

### ディレクトリ構造

```
delivery-api/
├── cmd/
│   └── server/
│       └── main.go                 # エントリーポイント
├── internal/
│   ├── model/
│   │   └── delivery.go            # データモデル定義
│   ├── repository/
│   │   └── delivery.go            # データベース操作
│   ├── handler/
│   │   └── delivery.go            # HTTPハンドラー
│   ├── middleware/
│   │   └── cors.go                # CORS設定
│   └── router/
│       └── router.go              # ルート定義
├── docker-compose.yml             # Docker構成
├── Dockerfile                     # コンテナイメージ定義
├── Makefile                       # コマンド短縮
├── .env                           # 本番環境変数（Git除外）
├── .env.example                   # 環境変数テンプレート
└── go.mod                         # Go依存性管理
```

---

### 1. `cmd/server/main.go` - アプリケーションエントリーポイント

**ファイルの役割：** アプリケーション全体の起動処理を担当

#### コード全体

```go
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/delivery-app/delivery-api/internal/model"
	"github.com/delivery-app/delivery-api/internal/router"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// @title Delivery API
// @version 1.0
// @description A delivery route management API with Google Maps integration
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @host localhost:8080
// @BasePath /
// @schemes http https
func main() {
	// Load environment variables
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "3306")
	dbUser := getEnv("DB_USER", "delivery_user")
	dbPassword := getEnv("DB_PASSWORD", "delivery_pass")
	dbName := getEnv("DB_NAME", "delivery_db")
	port := getEnv("PORT", "8080")

	// Setup database connection
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto migrate the models
	if err := db.AutoMigrate(&model.Delivery{}); err != nil {
		log.Fatalf("Failed to migrate models: %v", err)
	}

	log.Println("Database connected and migrated successfully")

	// Setup router
	r := router.SetupRouter(db)

	// Start server
	log.Printf("Starting server on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}
```

#### 詳細解説

**1. Swagger コメント（3～24行目）**
```go
// @title Delivery API
// @version 1.0
```
これらのコメントは `swag init` コマンドで自動的にSwagger/OpenAPI定義に変換されます。API仕様書を自動生成するための宣言です。

**2. 環境変数の読み込み（28～33行目）**
```go
dbHost := getEnv("DB_HOST", "localhost")
```
- `getEnv()` 関数で環境変数を安全に読み込む
- 環境変数が存在しない場合はデフォルト値を使用
- Docker環境では `docker-compose.yml` で `DB_HOST=db` に設定される

**3. DSN（Data Source Name）の構築（36～37行目）**
```go
dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
	dbUser, dbPassword, dbHost, dbPort, dbName)
```

これはMySQL接続文字列です。パラメータの意味：
- `%s:%s@tcp(...)` : ユーザー名:パスワード@TCP接続先
- `charset=utf8mb4` : 日本語を含む4バイト文字セット
- `parseTime=True` : MySQLのTIMESTAMP型をGoの`time.Time`に自動変換
- `loc=Local` : 時刻をローカルタイムゾーンで処理

**4. データベース接続（39～42行目）**
```go
db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
if err != nil {
	log.Fatalf("Failed to connect to database: %v", err)
}
```
- `gorm.Open()` でMySQL接続を確立
- 接続失敗時は `log.Fatalf()` でエラーログを出力してアプリを終了

**5. AutoMigrate（44～47行目）**
```go
if err := db.AutoMigrate(&model.Delivery{}); err != nil {
	log.Fatalf("Failed to migrate models: %v", err)
}
```
最も重要な部分です。これが何をするのかを詳しく説明します：

- `AutoMigrate()` はGORMの機能で、Go構造体をMySQL テーブルに自動的に同期
- `&model.Delivery{}` の型情報を読み取り、必要なテーブルを作成/更新
- すでにテーブルが存在する場合、構造体に新しいフィールドが追加されていれば列を追加

**6. ルーターセットアップ（51～52行目）**
```go
r := router.SetupRouter(db)
```
`router.go` で定義された `SetupRouter()` 関数にDB接続を渡す。この関数がすべてのHTTPエンドポイントを設定します。

**7. サーバー起動（55～58行目）**
```go
if err := r.Run(":" + port); err != nil {
	log.Fatalf("Failed to start server: %v", err)
}
```
- Ginフレームワークを起動
- `:8080` でリッスン開始
- 起動に失敗したら`log.Fatalf()`でエラー出力

**8. getEnv() ヘルパー関数（61～67行目）**
```go
func getEnv(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}
```
- `os.LookupEnv()` は環境変数の値と存在フラグを返す
- 存在する場合は環境変数の値、しない場合はデフォルト値を返す

---

### 2. `internal/model/delivery.go` - データモデル定義

**ファイルの役割：** Delivery（配送）を表す構造体を定義し、MySQLテーブルへのマッピングを指定

#### コード全体

```go
package model

import "time"

// Delivery represents a delivery destination
type Delivery struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name" binding:"required"`
	Address   string    `json:"address" binding:"required"`
	Lat       float64   `json:"lat"`
	Lng       float64   `json:"lng"`
	Status    string    `json:"status" gorm:"default:pending"`
	Note      string    `json:"note"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName specifies the table name for GORM
func (Delivery) TableName() string {
	return "deliveries"
}
```

#### 詳細解説

**構造体の定義**

```go
type Delivery struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name" binding:"required"`
	Address   string    `json:"address" binding:"required"`
	Lat       float64   `json:"lat"`
	Lng       float64   `json:"lng"`
	Status    string    `json:"status" gorm:"default:pending"`
	Note      string    `json:"note"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
```

**各フィールドの説明**

| フィールド | 型 | 説明 |
|----------|-----|------|
| `ID` | `uint` | 主キー。配送先を一意に識別 |
| `Name` | `string` | 配送先の名前（例：「渋谷支店」） |
| `Address` | `string` | 住所 |
| `Lat` | `float64` | 緯度（Google Maps用） |
| `Lng` | `float64` | 経度（Google Maps用） |
| `Status` | `string` | ステータス（pending/in_progress/completed） |
| `Note` | `string` | メモ・備考 |
| `CreatedAt` | `time.Time` | レコード作成日時（自動設定） |
| `UpdatedAt` | `time.Time` | レコード更新日時（自動設定） |

**Struct Tag の詳細解説**

Go構造体の各フィールドには「struct tag」というメタデータを付けることができます。

**JSON Tag：`json:"fieldName"`**
```go
ID        uint      `json:"id"`
```
- APIのリクエスト/レスポンスをJSON形式に変換する際、フィールドをどう表現するかを指定
- クライアントに返すJSONでは`"id"`という名前になる
- Go側では大文字`ID`（エクスポート可能），JSON側は小文字`"id"`という命名規則

**Binding Tag：`binding:"required"`**
```go
Name      string    `json:"name" binding:"required"`
```
- Ginフレームワークの入力検証機能
- `binding:"required"` : このフィールドは必須（リクエストに含まれなければエラー）
- バリデーションは自動的にHttp ハンドラーで行われる

**GORM Tag：`gorm:"primaryKey"`**
```go
ID        uint      `json:"id" gorm:"primaryKey"`
```
- GORMがこのフィールドをどう扱うかを指定
- `primaryKey` : このフィールドをデータベースの主キーとしている

**GORM Default：`gorm:"default:pending"`**
```go
Status    string    `json:"status" gorm:"default:pending"`
```
- テーブルの列定義時に、デフォルト値を`pending`に設定
- 新規Deliveryを作成する際、Statusを指定しなければ自動的に`"pending"`になる

**TableName() メソッド**

```go
func (Delivery) TableName() string {
	return "deliveries"
}
```
- GORMに対して「この構造体はMySQLの `deliveries` テーブルにマップされる」と明示
- 指定しない場合、GORMは複数形の「deliveries」を自動推測するが、明示的に書く方がベストプラクティス

---

### 3. `internal/repository/delivery.go` - データベース操作層

**ファイルの役割：** MySQLへのすべてのデータベース操作（CRUD）を担当。GORMを使用してSQL文を書かずにデータを操作

#### コード全体

```go
package repository

import (
	"errors"

	"github.com/delivery-app/delivery-api/internal/model"
	"gorm.io/gorm"
)

// DeliveryRepository handles delivery database operations
type DeliveryRepository struct {
	db *gorm.DB
}

// NewDeliveryRepository creates a new delivery repository
func NewDeliveryRepository(db *gorm.DB) *DeliveryRepository {
	return &DeliveryRepository{db: db}
}

// GetAll retrieves all deliveries
func (r *DeliveryRepository) GetAll() ([]model.Delivery, error) {
	var deliveries []model.Delivery
	if err := r.db.Find(&deliveries).Error; err != nil {
		return nil, err
	}
	return deliveries, nil
}

// GetByID retrieves a delivery by ID
func (r *DeliveryRepository) GetByID(id uint) (*model.Delivery, error) {
	var delivery model.Delivery
	if err := r.db.First(&delivery, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("delivery not found")
		}
		return nil, err
	}
	return &delivery, nil
}

// Create creates a new delivery
func (r *DeliveryRepository) Create(delivery *model.Delivery) error {
	return r.db.Create(delivery).Error
}

// Update updates an existing delivery
func (r *DeliveryRepository) Update(id uint, delivery *model.Delivery) error {
	if err := r.db.Model(&model.Delivery{}).Where("id = ?", id).Updates(delivery).Error; err != nil {
		return err
	}
	return nil
}

// Delete deletes a delivery by ID
func (r *DeliveryRepository) Delete(id uint) error {
	if err := r.db.Delete(&model.Delivery{}, id).Error; err != nil {
		return err
	}
	return nil
}
```

#### 詳細解説

**DeliveryRepository 構造体**

```go
type DeliveryRepository struct {
	db *gorm.DB
}
```
- `db *gorm.DB` : MySQL接続を保持
- Dependency Injection パターン：外部からDB接続を受け取る

**NewDeliveryRepository 構造体ファクトリー**

```go
func NewDeliveryRepository(db *gorm.DB) *DeliveryRepository {
	return &DeliveryRepository{db: db}
}
```
- 新しい`DeliveryRepository`インスタンスを生成
- 設計パターン：「ファクトリー関数」。Goには`new`キーワードがあるが、初期化をカプセル化するため関数にする

**GetAll() メソッド - すべてのレコードを取得**

```go
func (r *DeliveryRepository) GetAll() ([]model.Delivery, error) {
	var deliveries []model.Delivery
	if err := r.db.Find(&deliveries).Error; err != nil {
		return nil, err
	}
	return deliveries, nil
}
```

詳細：
- `var deliveries []model.Delivery` : Delivery構造体のスライス（配列）を宣言
- `r.db.Find(&deliveries)` : GORMのFind()メソッド
  - SQL的には `SELECT * FROM deliveries`
  - `&deliveries`（ポインタ）を渡すことで、内部で構造体に値が詰められる
- `.Error` : GORMはメソッドチェーン。最後に`.Error`でエラーを取得
- エラーがなければ`(deliveries, nil)`でスライスを返す

**GetByID() メソッド - IDで1件取得**

```go
func (r *DeliveryRepository) GetByID(id uint) (*model.Delivery, error) {
	var delivery model.Delivery
	if err := r.db.First(&delivery, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("delivery not found")
		}
		return nil, err
	}
	return &delivery, nil
}
```

詳細：
- `r.db.First(&delivery, id)` : SQL的には `SELECT * FROM deliveries WHERE id = ? LIMIT 1`
- `errors.Is(err, gorm.ErrRecordNotFound)` : GORMが「レコードが見つからない」というエラーを返したかチェック
- 見つからない場合は、わかりやすいエラーメッセージ「delivery not found」を返す
- 戻り値は`*model.Delivery`（ポインタ）：呼び出し側で値を変更できるようにするため

**Create() メソッド - 新規作成**

```go
func (r *DeliveryRepository) Create(delivery *model.Delivery) error {
	return r.db.Create(delivery).Error
}
```

詳細：
- `r.db.Create(delivery)` : SQL的には `INSERT INTO deliveries ...`
- GORMは`delivery`の`ID`フィールドに自動採番IDを設定
- `delivery`はポインタなので、呼び出し側の元の値も更新される
- 非常に短いメソッド：GORMはシンプルな操作を簡潔に書ける

**Update() メソッド - 更新**

```go
func (r *DeliveryRepository) Update(id uint, delivery *model.Delivery) error {
	if err := r.db.Model(&model.Delivery{}).Where("id = ?", id).Updates(delivery).Error; err != nil {
		return err
	}
	return nil
}
```

詳細：
- `r.db.Model(&model.Delivery{})` : 「Delivery テーブルを対象にする」という宣言
- `.Where("id = ?", id)` : SQL的な`WHERE`句。`?`はプレースホルダー、`id`の値が代入される
- `.Updates(delivery)` : `delivery`構造体の全フィールドを使って更新
  - 注意：ゼロ値（`""`や`0`）のフィールドも更新対象に含まれる
- 戻り値はエラーのみ（成功なら`nil`）

**Delete() メソッド - 削除**

```go
func (r *DeliveryRepository) Delete(id uint) error {
	if err := r.db.Delete(&model.Delivery{}, id).Error; err != nil {
		return err
	}
	return nil
}
```

詳細：
- `r.db.Delete(&model.Delivery{}, id)` : SQL的には `DELETE FROM deliveries WHERE id = ?`
- GORMは「第2引数をWHERE条件」と自動判定
- エラーチェックして返す

---

### 4. `internal/handler/delivery.go` - HTTPハンドラー層

**ファイルの役割：** HTTPリクエストを受け取り、リポジトリを呼び出してビジネスロジックを実行し、レスポンスをJSON形式で返す

#### コード全体（見出しのみ、詳細は下記）

```go
package handler

import (
	"net/http"
	"strconv"

	"github.com/delivery-app/delivery-api/internal/model"
	"github.com/delivery-app/delivery-api/internal/repository"
	"github.com/gin-gonic/gin"
)

// DeliveryHandler handles delivery-related requests
type DeliveryHandler struct {
	repo *repository.DeliveryRepository
}

// NewDeliveryHandler creates a new delivery handler
func NewDeliveryHandler(repo *repository.DeliveryRepository) *DeliveryHandler {
	return &DeliveryHandler{repo: repo}
}

// GetDeliveries, GetDelivery, CreateDelivery, UpdateDelivery, DeleteDelivery, HealthCheck メソッド
```

#### Gin フレームワーク基礎

Ginは軽量で高速なGoのWebフレームワークです。ハンドラー関数は以下のシグネチャを持ちます：

```go
func (h *DeliveryHandler) GetDeliveries(c *gin.Context) {
	// c は HTTP リクエスト/レスポンスを操作するコンテキスト
}
```

`gin.Context` の主なメソッド：
- `c.Param("id")` : URLパラメータを取得（例：`/deliveries/:id` の`:id`）
- `c.Query("name")` : クエリパラメータを取得（例：`?name=value`）
- `c.ShouldBindJSON(&data)` : リクエストボディをJSONとしてパース
- `c.JSON(statusCode, data)` : JSONレスポンスを返す
- `c.AbortWithStatus(code)` : リクエスト処理を中止し、ステータスコードを返す

#### GetDeliveries() - すべての配送先を取得

```go
// GetDeliveries godoc
// @Summary Get all deliveries
// @Description Get a list of all delivery destinations
// @Tags deliveries
// @Accept json
// @Produce json
// @Success 200 {array} model.Delivery
// @Failure 500 {object} map[string]string
// @Router /api/v1/deliveries [get]
func (h *DeliveryHandler) GetDeliveries(c *gin.Context) {
	deliveries, err := h.repo.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch deliveries"})
		return
	}
	c.JSON(http.StatusOK, deliveries)
}
```

**詳細説明：**
1. Swagger コメント（3～10行目）：API仕様を自動ドキュメント化
2. `h.repo.GetAll()` : リポジトリ層の処理を実行
3. `err != nil` : エラーチェック。データベースアクセス失敗時
4. `c.JSON(http.StatusInternalServerError, ...)` : ステータスコード500でエラーを返す
5. `gin.H` : `map[string]interface{}` の短記法。JSON変換時に`{"error":"..."}`形式になる
6. エラーがなければ、200とDeliveryスライスを返す

#### GetDelivery() - IDで1件取得

```go
func (h *DeliveryHandler) GetDelivery(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid delivery ID"})
		return
	}

	delivery, err := h.repo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, delivery)
}
```

**詳細説明：**
1. `c.Param("id")` : URL パスパラメータ `{id}` の値を文字列で取得
2. `strconv.ParseUint(c.Param("id"), 10, 32)` : 文字列をuint32に変換
   - 第2引数の `10` : 10進数
   - 第3引数の `32` : 32ビット整数
3. `uint(id)` : ParseUintで返されたuint64をuintに型変換
4. ステータスコードの使い分け：
   - `400 Bad Request` : クライアントエラー（無効な入力）
   - `404 Not Found` : リソースが見つからない
   - `200 OK` : 成功

#### CreateDelivery() - 新規作成

```go
func (h *DeliveryHandler) CreateDelivery(c *gin.Context) {
	var delivery model.Delivery
	if err := c.ShouldBindJSON(&delivery); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.repo.Create(&delivery); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create delivery"})
		return
	}

	c.JSON(http.StatusCreated, delivery)
}
```

**詳細説明：**
1. `var delivery model.Delivery` : Delivery構造体の変数を宣言
2. `c.ShouldBindJSON(&delivery)` : Ginの強力な機能
   - リクエストボディをJSON として解析
   - struct tag `binding:"required"` などを使用してバリデーション
   - Name, Address が含まれていなければエラーを返す
3. `h.repo.Create(&delivery)` : リポジトリでDB操作
   - 成功すると`delivery.ID`にDB生成のIDが詰められる
4. `http.StatusCreated` (201) : リソース作成成功を表す標準ステータスコード

#### UpdateDelivery() - 更新

```go
func (h *DeliveryHandler) UpdateDelivery(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid delivery ID"})
		return
	}

	var delivery model.Delivery
	if err := c.ShouldBindJSON(&delivery); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if delivery exists
	_, err = h.repo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	delivery.ID = uint(id)
	if err := h.repo.Update(uint(id), &delivery); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update delivery"})
		return
	}

	c.JSON(http.StatusOK, delivery)
}
```

**詳細説明：**
1. URLパラメータからIDを抽出し、型変換（GetDelivery同様）
2. リクエストボディをJSONからパース
3. `h.repo.GetByID(uint(id))` : 更新対象が本当に存在するか確認（存在確認パターン）
4. `delivery.ID = uint(id)` : 重要な行。IDをセット（リポジトリのWHERE句で使用される）
5. リポジトリで更新実行
6. 200で成功を返す

#### DeleteDelivery() - 削除

```go
func (h *DeliveryHandler) DeleteDelivery(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid delivery ID"})
		return
	}

	// Check if delivery exists
	_, err = h.repo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	if err := h.repo.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete delivery"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Delivery deleted successfully"})
}
```

**詳細説明：**
1. パターンとしてはUpdateと似ている
2. リクエストボディはなく、URLパラメータから削除対象IDを取得
3. 削除前に存在確認
4. リポジトリで削除実行
5. 成功時は`gin.H`を使って`{"message":"..."}`を返す

#### HealthCheck() - ヘルスチェック

```go
func (h *DeliveryHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "healthy"})
}
```

**詳細説明：**
- Docker、Kubernetesなどのオーケストレーションツールがサーバーの生存確認に使う
- 常に`200 OK`を返す
- ビジネスロジック不要、単なるステータス確認

---

### 5. `internal/router/router.go` - ルーター設定

**ファイルの役割：** すべてのHTTPエンドポイントを定義し、ハンドラーに接続

#### コード全体

```go
package router

import (
	"github.com/delivery-app/delivery-api/internal/handler"
	"github.com/delivery-app/delivery-api/internal/middleware"
	"github.com/delivery-app/delivery-api/internal/repository"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SetupRouter creates and configures the Gin router
func SetupRouter(db *gorm.DB) *gin.Engine {
	router := gin.Default()

	// Apply middleware
	router.Use(middleware.CORSMiddleware())

	// Initialize repository and handler
	deliveryRepo := repository.NewDeliveryRepository(db)
	deliveryHandler := handler.NewDeliveryHandler(deliveryRepo)

	// Health check endpoint
	router.GET("/api/v1/health", deliveryHandler.HealthCheck)

	// TODO: Swagger documentation（swag init 実行後に有効化）
	// swaggerfiles "github.com/swaggo/files"
	// ginswagger "github.com/swaggo/gin-swagger"
	// router.GET("/swagger/*any", ginswagger.WrapHandler(swaggerfiles.Handler))

	// Delivery routes
	api := router.Group("/api/v1")
	{
		api.GET("/deliveries", deliveryHandler.GetDeliveries)
		api.GET("/deliveries/:id", deliveryHandler.GetDelivery)
		api.POST("/deliveries", deliveryHandler.CreateDelivery)
		api.PUT("/deliveries/:id", deliveryHandler.UpdateDelivery)
		api.DELETE("/deliveries/:id", deliveryHandler.DeleteDelivery)
	}

	return router
}
```

#### 詳細解説

**Ginエンジンの初期化**

```go
router := gin.Default()
```
- `gin.Default()` : Ginフレームワークを初期化
- デフォルトミドルウェア（Logger、Recovery）を含む

**ミドルウェア適用**

```go
router.Use(middleware.CORSMiddleware())
```
- `router.Use()` : すべてのリクエストに対してミドルウェアを適用
- CORS設定（詳細は次のセクション）

**依存性注入（Dependency Injection）**

```go
deliveryRepo := repository.NewDeliveryRepository(db)
deliveryHandler := handler.NewDeliveryHandler(deliveryRepo)
```
- `db`をRepositoryに注入
- Repositoryを Handlerに注入
- このパターンにより、テスト時にモック（偽物）のリポジトリを注入することが容易になる

**ルートグループ**

```go
api := router.Group("/api/v1")
{
	api.GET("/deliveries", deliveryHandler.GetDeliveries)
	api.GET("/deliveries/:id", deliveryHandler.GetDelivery)
	// ...
}
```
- `router.Group("/api/v1")` : プレフィックス `/api/v1` を持つルートグループを作成
- これにより、各エンドポイントの完全なパスは `/api/v1/deliveries` などになる
- 複数のバージョンを同時サポートする場合、`/api/v2` などを追加できる

**ルート定義**

```go
api.GET("/deliveries", deliveryHandler.GetDeliveries)
```
- HTTPメソッド（GET、POST等）とパスを指定
- ハンドラー関数をマッピング
- URLパラメータは `:id` のようにコロンで表記（Gin独自の構文）

---

### 6. `internal/middleware/cors.go` - CORS設定

**ファイルの役割：** Cross-Origin Resource Sharing (CORS) の設定により、ブラウザからのクロスオリジンリクエストを許可

#### コード全体

```go
package middleware

import (
	"github.com/gin-gonic/gin"
)

// CORSMiddleware adds CORS headers to responses
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
```

#### CORSとは何か

**CORS（Cross-Origin Resource Sharing）の必要性：**

ブラウザはセキュリティ上、異なるオリジン（ドメイン、ポート、プロトコル）からのリクエストをブロックします。

例：
- フロントエンド：`http://localhost:3000` （React、Vue等）
- バックエンド：`http://localhost:8080` （このAPI）

フロントエンドからのJavaScript fetch/axios は、デフォルトではバックエンドへのリクエストをブロックします。

**CORSの仕組み：**

ブラウザは自動的にプリフライトリクエスト（OPTIONS メソッド）を事前に送信し、バックエンドが「このクロスオリジンリクエストを許可する」というHTTPヘッダーを返していることを確認します。

#### コード詳細解説

**ミドルウェア関数の返却**

```go
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// ...
	}
}
```
- `gin.HandlerFunc` の定義：`func(c *gin.Context)`
- クロージャーパターンを使用

**CORSレスポンスヘッダーの設定**

```go
c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
```
- `c.Writer` : HTTPレスポンスへアクセス
- `Header().Set(key, value)` : ヘッダーに値を追加
- `"Access-Control-Allow-Origin", "*"` : **すべてのオリジンからのリクエストを許可**
  - 本番環境では `"*"` ではなく特定ドメインに限定すべき：`"https://example.com"`

```go
c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
```
- クライアントがクッキーや認証情報を送信できるようにする

```go
c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, ...")
```
- リクエストで送信可能なHTTPヘッダーを指定
- `Content-Type` : `application/json` などを含む
- `Authorization` : JWTトークンなど認証情報

```go
c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")
```
- 許可するHTTPメソッド

**OPTIONSメソッド（プリフライト）の処理**

```go
if c.Request.Method == "OPTIONS" {
	c.AbortWithStatus(204)
	return
}
```
- ブラウザのプリフライトリクエスト（OPTIONS）に対して、204 No Content を返す
- これ以上の処理（c.Next()）を行わない

**リクエスト処理の継続**

```go
c.Next()
```
- CORSヘッダーを設定した後、次のミドルウェア/ハンドラーへ処理を継続

---

### 7. `docker-compose.yml` - Docker構成定義

**ファイルの役割：** APIとMySQL を含む複数のDockerコンテナを定義し、ネットワーク・ボリューム・環境変数を設定

#### コード全体

```yaml
services:
  api:
    build: .
    ports:
      - "8080:8080"
    depends_on:
      db:
        condition: service_healthy
    env_file:
      - .env
    environment:
      - DB_HOST=db

  db:
    image: mysql:8.0
    ports:
      - "3306:3306"
    environment:
      MYSQL_ROOT_PASSWORD: root
      MYSQL_DATABASE: delivery_db
      MYSQL_USER: delivery_user
      MYSQL_PASSWORD: delivery_pass
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost"]
      interval: 10s
      timeout: 5s
      retries: 5
    volumes:
      - mysql_data:/var/lib/mysql

volumes:
  mysql_data:
```

#### 詳細解説

**services セクション**

Docker Composeは複数のサービス（コンテナ）を定義します。

**api サービス**

```yaml
api:
  build: .
```
- `build: .` : カレントディレクトリの`Dockerfile` を使用してイメージをビルド
- Go/Ginアプリケーション本体

```yaml
ports:
  - "8080:8080"
```
- `ホストマシンポート:コンテナ内ポート`
- ホスト側で `http://localhost:8080` にアクセスすると、コンテナの8080ポートに到達

```yaml
depends_on:
  db:
    condition: service_healthy
```
- apiサービスはdbサービスに依存
- `condition: service_healthy` : db のヘルスチェック（healthcheck）が成功するまで api を起動しない
- これにより、MySQLが完全に起動してから Go アプリを起動する

```yaml
env_file:
  - .env
```
- `.env` ファイルから環境変数を読み込む

```yaml
environment:
  - DB_HOST=db
```
- `DB_HOST` 環境変数を `db` に設定
- **重要なポイント：** Dockerネットワーク内では、サービス名がホスト名として機能
- `localhost` ではなく `db` を使用（dbという名前のサービスと通信）

**db サービス（MySQL）**

```yaml
db:
  image: mysql:8.0
```
- MySQLのオフィシャルイメージ version 8.0 を使用
- `build` ではなく `image` を指定（既製イメージをそのまま使用）

```yaml
ports:
  - "3306:3306"
```
- ホストマシンから直接MySQLにアクセス可能
- ホスト側から `mysql -h localhost -u delivery_user -p` でアクセス可能

```yaml
environment:
  MYSQL_ROOT_PASSWORD: root
  MYSQL_DATABASE: delivery_db
  MYSQL_USER: delivery_user
  MYSQL_PASSWORD: delivery_pass
```
- MySQLコンテナの初期設定
- `MYSQL_DATABASE` : 起動時に自動作成されるデータベース
- `MYSQL_USER` と `MYSQL_PASSWORD` : 初期ユーザーとパスワード

```yaml
healthcheck:
  test: ["CMD", "mysqladmin", "ping", "-h", "localhost"]
  interval: 10s
  timeout: 5s
  retries: 5
```
- MySQLコンテナの健康状態をチェック
- `test` : `mysqladmin ping` で MySQLの応答を確認
- `interval: 10s` : 10秒ごとにチェック
- `timeout: 5s` : 5秒以内に応答がなければ失敗と判定
- `retries: 5` : 5回連続失敗でコンテナを不健康と判定

```yaml
volumes:
  - mysql_data:/var/lib/mysql
```
- データボリュームの指定
- `mysql_data` という名前のボリュームを、コンテナの `/var/lib/mysql` にマウント

**volumes セクション**

```yaml
volumes:
  mysql_data:
```
- `mysql_data` という名前のボリュームを定義
- `docker-compose down` を実行しても、このボリュームは削除されない（`-v` オプション付きでのみ削除）
- MySQLのデータが永続化される

#### Docker Compose の実行フロー

```
$ docker-compose up -d

1. Dockerfile をビルド → apiイメージを作成
2. db (MySQL 8.0) コンテナを起動
3. MySQL の healthcheck が成功するまで待機
4. api コンテナを起動
5. api は環境変数 DB_HOST=db でMySQL に接続
```

---

### 8. `Dockerfile` - マルチステージビルド

**ファイルの役割：** Go アプリケーションをコンテナイメージにビルドする

#### コード全体

```dockerfile
FROM golang:1.21-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod ./

RUN go mod download

COPY . .

RUN go mod tidy

RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

COPY --from=builder /app/server .

EXPOSE 8080

CMD ["./server"]
```

#### マルチステージビルドの概念

通常、Goアプリケーションのビルドには大きなイメージ（golang:1.21）が必要ですが、実行には不要です。マルチステージビルドにより、ビルド用と実行用で異なるイメージを使用し、最終的なイメージサイズを削減します。

#### ステージ1：ビルドステージ

```dockerfile
FROM golang:1.21-alpine AS builder
```
- `golang:1.21-alpine` : Goのコンパイラとツールチェーンをすべて含む
- `AS builder` : このステージを `builder` という名前で参照

```dockerfile
RUN apk add --no-cache git
```
- Alpine Linuxのパッケージマネージャー `apk` で `git` をインストール
- `--no-cache` : パッケージキャッシュを保存しない（イメージサイズを削減）

```dockerfile
WORKDIR /app
```
- コンテナ内の作業ディレクトリを `/app` に設定
- 以下の `COPY`、`RUN` コマンドはすべてこのディレクトリで実行される

```dockerfile
COPY go.mod ./
```
- ホストマシンの `go.mod` をコンテナの `/app/go.mod` にコピー

```dockerfile
RUN go mod download
```
- `go.mod` 内の依存ライブラリをすべてダウンロード
- このステップを早めに実行することで、Docker のレイヤーキャッシュが有効になる
- 以降のステップで `go.mod` 変更がなければ、このステップはキャッシュから復用される

```dockerfile
COPY . .
```
- プロジェクト全体をコンテナにコピー

```dockerfile
RUN go mod tidy
```
- 使用されていない依存ライブラリを削除し、`go.mod` をクリーンアップ

```dockerfile
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server
```
- **ビルドコマンドの詳細：**
  - `CGO_ENABLED=0` : C言語ライブラリとのリンクを無効化（クロスプラットフォーム対応）
  - `GOOS=linux` : ターゲットOSを Linux に指定
  - `go build -o server ./cmd/server` : `cmd/server/main.go` をコンパイルし、実行ファイル `server` を生成
  - 結果：`/app/server` という実行可能ファイルが生成される

#### ステージ2：実行ステージ

```dockerfile
FROM alpine:latest
```
- **新しいフェーズを開始**
- `alpine:latest` : 軽量なLinuxディストリビューション（約5MB）
- ビルドステージの Goコンパイラなどは含まない

```dockerfile
RUN apk --no-cache add ca-certificates tzdata
```
- `ca-certificates` : HTTPS通信に必要な認証局の証明書
- `tzdata` : タイムゾーン情報（アプリが時刻処理する場合）

```dockerfile
WORKDIR /root/
```
- 実行時の作業ディレクトリ

```dockerfile
COPY --from=builder /app/server .
```
- **マルチステージビルド特有の機能**
- `--from=builder` : builderステージから `/app/server` をコピー
- 結果：`/root/server` に実行ファイルが配置される
- Goのコンパイラ、ライブラリなどは一切含まない

```dockerfile
EXPOSE 8080
```
- コンテナがポート 8080 をリッスンすることを宣言
- `docker-compose.yml` の `ports` と連携

```dockerfile
CMD ["./server"]
```
- コンテナ起動時に実行するコマンド
- 配列形式（exec形式）で指定（シェル経由ではなく直接実行）

#### イメージサイズの比較

| ステージ | イメージサイズ |
|---------|--------------|
| golang:1.21-alpine | 約400MB |
| 実行ファイル（Go 単体） | 約15MB |
| alpine:latest | 約5MB |
| **最終イメージ** | **～20MB** |

---

### 9. `Makefile` - コマンド自動化

**ファイルの役割：** よく使うコマンドを簡潔に実行するための自動化スクリプト

#### コード全体

```makefile
.PHONY: run build docker-up docker-down swagger lint fmt help

help:
	@echo "Available targets:"
	@echo "  make run         - Run the application"
	@echo "  make build       - Build the application"
	@echo "  make docker-up   - Start Docker containers"
	@echo "  make docker-down - Stop Docker containers"
	@echo "  make swagger     - Generate Swagger documentation"
	@echo "  make lint        - Run linter"
	@echo "  make fmt         - Format code"

run:
	go run ./cmd/server

build:
	go build -o server ./cmd/server

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

swagger:
	swag init -g cmd/server/main.go

lint:
	golangci-lint run

fmt:
	gofmt -s -w .
```

#### 詳細解説

**`.PHONY`**

```makefile
.PHONY: run build docker-up docker-down swagger lint fmt help
```
- `.PHONY` で指定されたターゲットは、実際のファイルではなく仮想的なターゲット
- 同名のファイル/ディレクトリが存在してもコマンドを実行

**help ターゲット**

```makefile
help:
	@echo "Available targets:"
```
- `@echo` : `@` により、コマンドそのものをエコーしない（結果のみ表示）
- ユーザーが `make help` で利用可能なコマンドを確認できる

**run ターゲット**

```makefile
run:
	go run ./cmd/server
```
- `go run` : ビルドなしで直接実行（開発時に便利）
- `main.go` の構文エラーがあれば即座に検出

**build ターゲット**

```makefile
build:
	go build -o server ./cmd/server
```
- `go build` : 実行ファイルをコンパイル
- `-o server` : 出力ファイル名を `server` に指定

**docker-up ターゲット**

```makefile
docker-up:
	docker-compose up -d
```
- `docker-compose up` : all services を起動
- `-d` : デタッチモード（バックグラウンド実行）
- コンテナを起動後、すぐにターミナルに制御が戻る

**docker-down ターゲット**

```makefile
docker-down:
	docker-compose down
```
- `docker-compose down` : すべてのコンテナを停止・削除
- ボリュームは削除されない（ `docker-compose down -v` で削除可能）

**swagger ターゲット**

```makefile
swagger:
	swag init -g cmd/server/main.go
```
- `swag init` : Swagger/OpenAPI定義を自動生成
- `-g cmd/server/main.go` : swagger コメント を読み込むエントリーポイント
- `docs/` ディレクトリにSwagger UIの定義が生成される

**lint ターゲット**

```makefile
lint:
	golangci-lint run
```
- `golangci-lint` : 複数の linter（コード品質チェッカー）を統合したツール
- スタイル、可能なバグ、パフォーマンス問題などを検出

**fmt ターゲット**

```makefile
fmt:
	gofmt -s -w .
```
- `gofmt` : Go コードのフォーマッター
- `-s` : シンプルな形式に簡潔化
- `-w` : ファイルをインプレースで変更

---

### 10. `.env` / `.env.example` - 環境変数

**ファイルの役割：** アプリケーション設定を環境変数で管理。機密情報をコード外に置く

#### `.env.example` （テンプレート）

```
DB_HOST=localhost
DB_PORT=3306
DB_USER=delivery_user
DB_PASSWORD=delivery_pass
DB_NAME=delivery_db
PORT=8080
GOOGLE_MAPS_API_KEY=your_api_key_here
```

#### `.env` （実際の値、Git除外）

```
DB_HOST=localhost
DB_PORT=3306
DB_USER=delivery_user
DB_PASSWORD=delivery_pass
DB_NAME=delivery_db
PORT=8080
GOOGLE_MAPS_API_KEY=AIzaSyAblUdDUJGq64edIGbZfZmB86B9LZjhUs4
```

#### 詳細解説

**環境変数の利点**

1. **セキュリティ** : APIキーやパスワードをコードに含めない
2. **環境による切り替え** : 開発/本番で異なる設定を簡単に切り替え
3. **CI/CD連携** : GitHubActionsなどで環境変数を自動注入

**使用方法**

```go
// main.go
dbHost := getEnv("DB_HOST", "localhost")
```

**`.env.example` と `.gitignore`**

```
# .gitignore
.env
```
- `.env` ファイルをGitで追跡しない
- チームメンバーは `.env.example` をコピーして `.env` を作成
- 各自の環境に応じて値を設定

**Docker Compose での自動読み込み**

```yaml
# docker-compose.yml
env_file:
  - .env
```
- `docker-compose up` 時に `.env` から環境変数を自動読み込み

---

## APIエンドポイント

このセクションでは、すべてのAPIエンドポイントをHTTPメソッド・パス・リクエスト/レスポンス形式とともに説明します。

### 1. GET /api/v1/deliveries - すべて取得

**説明：** 登録されているすべての配送先情報を取得します

**リクエスト**
```bash
curl -X GET http://localhost:8080/api/v1/deliveries
```

**レスポンス (200 OK)**
```json
[
  {
    "id": 1,
    "name": "渋谷支店",
    "address": "東京都渋谷区道玄坂1-2-3",
    "lat": 35.6595,
    "lng": 139.7004,
    "status": "completed",
    "note": "営業時間 9:00-18:00",
    "created_at": "2026-03-01T10:30:00Z",
    "updated_at": "2026-03-01T15:45:00Z"
  },
  {
    "id": 2,
    "name": "新宿支店",
    "address": "東京都新宿区西新宿1-1-1",
    "lat": 35.6894,
    "lng": 139.6917,
    "status": "in_progress",
    "note": "",
    "created_at": "2026-03-01T11:00:00Z",
    "updated_at": "2026-03-01T14:00:00Z"
  }
]
```

**ステータスコード**
- `200 OK` : 成功（データが0件でも200）
- `500 Internal Server Error` : データベースエラー

---

### 2. GET /api/v1/deliveries/{id} - IDで取得

**説明：** 指定されたIDの配送先情報を取得します

**リクエスト**
```bash
curl -X GET http://localhost:8080/api/v1/deliveries/1
```

**レスポンス (200 OK)**
```json
{
  "id": 1,
  "name": "渋谷支店",
  "address": "東京都渋谷区道玄坂1-2-3",
  "lat": 35.6595,
  "lng": 139.7004,
  "status": "completed",
  "note": "営業時間 9:00-18:00",
  "created_at": "2026-03-01T10:30:00Z",
  "updated_at": "2026-03-01T15:45:00Z"
}
```

**エラーレスポンス - 無効なID (400 Bad Request)**
```bash
curl -X GET http://localhost:8080/api/v1/deliveries/abc
```
```json
{
  "error": "Invalid delivery ID"
}
```

**エラーレスポンス - 存在しないID (404 Not Found)**
```bash
curl -X GET http://localhost:8080/api/v1/deliveries/999
```
```json
{
  "error": "delivery not found"
}
```

**ステータスコード**
- `200 OK` : 成功
- `400 Bad Request` : IDが数字でない
- `404 Not Found` : IDが見つからない
- `500 Internal Server Error` : データベースエラー

---

### 3. POST /api/v1/deliveries - 新規作成

**説明：** 新しい配送先を作成します

**リクエスト**
```bash
curl -X POST http://localhost:8080/api/v1/deliveries \
  -H "Content-Type: application/json" \
  -d '{
    "name": "品川支店",
    "address": "東京都港区港南1-2-3",
    "lat": 35.6283,
    "lng": 139.7380,
    "status": "pending",
    "note": "新規オープン"
  }'
```

**リクエストボディの詳細**

| フィールド | 型 | 必須 | 説明 |
|----------|-----|------|------|
| `name` | string | ✓ | 配送先名 |
| `address` | string | ✓ | 住所 |
| `lat` | float64 | ✗ | 緯度 |
| `lng` | float64 | ✗ | 経度 |
| `status` | string | ✗ | ステータス（pending/in_progress/completed） |
| `note` | string | ✗ | メモ |

**レスポンス (201 Created)**
```json
{
  "id": 3,
  "name": "品川支店",
  "address": "東京都港区港南1-2-3",
  "lat": 35.6283,
  "lng": 139.7380,
  "status": "pending",
  "note": "新規オープン",
  "created_at": "2026-03-01T16:00:00Z",
  "updated_at": "2026-03-01T16:00:00Z"
}
```

**エラーレスポンス - 必須フィールド不足 (400 Bad Request)**
```bash
curl -X POST http://localhost:8080/api/v1/deliveries \
  -H "Content-Type: application/json" \
  -d '{"address": "..."}'
```
```json
{
  "error": "Key: 'Delivery.Name' Error:Field validation for 'Name' failed on the 'required' tag"
}
```

**ステータスコード**
- `201 Created` : 成功（新しいリソースが作成された）
- `400 Bad Request` : 必須フィールド不足またはJSONパース失敗
- `500 Internal Server Error` : データベースエラー

---

### 4. PUT /api/v1/deliveries/{id} - 更新

**説明：** 指定されたIDの配送先情報を更新します

**リクエスト**
```bash
curl -X PUT http://localhost:8080/api/v1/deliveries/1 \
  -H "Content-Type: application/json" \
  -d '{
    "name": "渋谷支店",
    "address": "東京都渋谷区道玄坂1-2-3",
    "lat": 35.6595,
    "lng": 139.7004,
    "status": "completed",
    "note": "営業時間 9:00-18:00 ※金曜は20:00まで"
  }'
```

**レスポンス (200 OK)**
```json
{
  "id": 1,
  "name": "渋谷支店",
  "address": "東京都渋谷区道玄坂1-2-3",
  "lat": 35.6595,
  "lng": 139.7004,
  "status": "completed",
  "note": "営業時間 9:00-18:00 ※金曜は20:00まで",
  "created_at": "2026-03-01T10:30:00Z",
  "updated_at": "2026-03-01T16:15:00Z"
}
```

**エラーレスポンス - 存在しないID (404 Not Found)**
```json
{
  "error": "delivery not found"
}
```

**ステータスコード**
- `200 OK` : 成功
- `400 Bad Request` : IDが数字でない、またはJSONパース失敗
- `404 Not Found` : IDが見つからない
- `500 Internal Server Error` : データベースエラー

---

### 5. DELETE /api/v1/deliveries/{id} - 削除

**説明：** 指定されたIDの配送先を削除します

**リクエスト**
```bash
curl -X DELETE http://localhost:8080/api/v1/deliveries/1
```

**レスポンス (200 OK)**
```json
{
  "message": "Delivery deleted successfully"
}
```

**エラーレスポンス - 存在しないID (404 Not Found)**
```json
{
  "error": "delivery not found"
}
```

**ステータスコード**
- `200 OK` : 成功
- `400 Bad Request` : IDが数字でない
- `404 Not Found` : IDが見つからない
- `500 Internal Server Error` : データベースエラー

---

### 6. GET /api/v1/health - ヘルスチェック

**説明：** APIが正常に動作しているか確認します。Docker/Kubernetesのヘルスチェックで使用

**リクエスト**
```bash
curl -X GET http://localhost:8080/api/v1/health
```

**レスポンス (200 OK)**
```json
{
  "status": "healthy"
}
```

**ステータスコード**
- `200 OK` : APIが健全な状態

---

## データベース

### MySQL テーブル構造

`AutoMigrate(&model.Delivery{})` により、以下のテーブルが自動作成されます：

```sql
CREATE TABLE deliveries (
  id bigint unsigned PRIMARY KEY AUTO_INCREMENT,
  name varchar(255) NOT NULL,
  address varchar(255) NOT NULL,
  lat double,
  lng double,
  status varchar(255) DEFAULT 'pending',
  note text,
  created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
```

### 列の詳細

| 列名 | 型 | 制約 | 説明 |
|------|-----|------|------|
| `id` | bigint unsigned | PRIMARY KEY, AUTO_INCREMENT | 自動採番される主キー |
| `name` | varchar(255) | NOT NULL | 配送先名（必須） |
| `address` | varchar(255) | NOT NULL | 住所（必須） |
| `lat` | double | - | 緯度（オプション） |
| `lng` | double | - | 経度（オプション） |
| `status` | varchar(255) | DEFAULT 'pending' | ステータス（デフォルト値：pending） |
| `note` | text | - | メモ（任意の長さのテキスト） |
| `created_at` | timestamp | NOT NULL | レコード作成日時（自動設定） |
| `updated_at` | timestamp | NOT NULL | レコード更新日時（自動更新） |

### GORM AutoMigrate の詳細

#### 1. テーブルが存在しない場合

初回起動時、`db.AutoMigrate(&model.Delivery{})` は以下を実行します：

```sql
CREATE TABLE deliveries (
  id bigint unsigned PRIMARY KEY AUTO_INCREMENT,
  name varchar(255) NOT NULL,
  address varchar(255) NOT NULL,
  -- ...
);
```

#### 2. テーブルは存在するが新しい列が追加された場合

例えば、`Delivery` 構造体に新しいフィールドを追加した場合：

```go
type Delivery struct {
  // ... 既存フィールド
  PhoneNumber string `json:"phone_number"`  // 新規追加
}
```

次回起動時、GORMは自動的に列を追加します：

```sql
ALTER TABLE deliveries ADD COLUMN phone_number varchar(255);
```

#### 3. 列を削除した場合

**注意：** GORMの`AutoMigrate`は自動的に列を削除しません（データ損失を防ぐため）。

削除が必要な場合は、手動でマイグレーションスクリプトを実行するか、以下のように明示的にドロップします：

```go
// 開発環境のみ
db.Migrator().DropColumn(&model.Delivery{}, "some_field")
```

### GORM が Go 構造体を SQL に変換する仕組み

#### struct tag から型を推測

```go
ID        uint      `json:"id" gorm:"primaryKey"`
```

- `uint` (unsigned int) → MySQL の `bigint unsigned`
- `string` → `varchar(255)` (255はGORMのデフォルト)
- `float64` → `double`
- `time.Time` → `timestamp`
- `text` (別途指定) → `text`

#### フィールド名から列名を推測

```go
CreatedAt time.Time `json:"created_at"`
```

- Goのフィールドスネークケース（`CreatedAt`）をスネークケース（`created_at`）に自動変換
- JSON tag とは別に、データベース列名を指定する `db` tag も使用可能：

```go
CreatedAt time.Time `json:"created_at" db:"create_time"`  // 列名は create_time
```

### 実際のデータの流れ

```
クライアント (JSON)
    ↓
{"name": "渋谷支店", "address": "..."}
    ↓
Gin の c.ShouldBindJSON(&delivery)
    ↓
Go 構造体 (Delivery)
    {ID: 0, Name: "渋谷支店", Address: "..."}
    ↓
GORM の db.Create(&delivery)
    ↓
MySQL INSERT 文の生成・実行
INSERT INTO deliveries (name, address, ...) VALUES (?, ?, ...)
    ↓
MySQL がデータを保存・ID を自動採番
    ↓
GORMが delivery.ID に採番IDを設定
delivery.ID = 3
    ↓
レスポンス (201 Created)
{"id": 3, "name": "渋谷支店", ...}
```

---

## Docker環境構築

### docker-compose の起動手順

#### 1. イメージのビルドとコンテナの起動

```bash
make docker-up
# または
docker-compose up -d
```

**実行内容：**
1. `Dockerfile` から apiイメージをビルド
2. MySQLコンテナを起動
3. MySQLヘルスチェック（healthcheck）を開始
4. ヘルスチェック成功後、apiコンテナを起動

**出力例：**
```
Creating delivery-db_1   ... done
Creating delivery-api_1  ... done
```

#### 2. ログの確認

```bash
docker-compose logs -f api
```

期待される出力：
```
api_1  | 2026/03/01 16:00:00 Database connected and migrated successfully
api_1  | 2026/03/01 16:00:01 Starting server on port 8080
```

#### 3. APIのテスト

```bash
curl http://localhost:8080/api/v1/health
```

期待される出力：
```json
{"status":"healthy"}
```

#### 4. MySQL への直接接続

```bash
docker exec -it delivery-api-db-1 mysql -u delivery_user -p delivery_db
# パスワード: delivery_pass
```

```sql
> SELECT * FROM deliveries;
> DESCRIBE deliveries;
```

#### 5. コンテナの停止・削除

```bash
make docker-down
# または
docker-compose down
```

- **コンテナは削除される**
- **ボリューム（mysql_data）は保持される**（次回起動時にデータが復元される）

**ボリュームも含めて削除：**
```bash
docker-compose down -v
```

---

### Dockerfile のマルチステージビルドの詳細

#### ビルドプロセスの可視化

```
ステージ1：builder
┌─────────────────────────────────────┐
│ FROM golang:1.21-alpine             │ ← 約400MB
│                                     │
│ - COPY go.mod                       │
│ - go mod download                   │
│ - COPY . .                          │
│ - go build → /app/server            │ ← 15MB (バイナリ)
└─────────────────────────────────────┘
             ↓
          一時レイヤーとして保持

ステージ2：最終イメージ
┌─────────────────────────────────────┐
│ FROM alpine:latest                  │ ← 約5MB
│                                     │
│ - COPY --from=builder /app/server . │ ← ビルドステージからコピー
│                                     │
│ EXPOSE 8080                         │
│ CMD ["./server"]                    │
└─────────────────────────────────────┘
         ↓
   最終イメージ：～20MB
```

#### イメージレイヤーの確認

```bash
docker image history delivery-api:latest
```

出力例：
```
IMAGE          CREATED       CREATED BY
abc123def      10 seconds    /bin/sh -c #(nop) CMD ["./server"]
def456ghi      10 seconds    /bin/sh -c #(nop) EXPOSE 8080
ghi789jkl      10 seconds    COPY --from=builder /app/server .
...
```

#### クロスプラットフォーム対応

```dockerfile
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server
```

- `CGO_ENABLED=0` : C言語ライブラリとのリンクを無効化
  - 理由：WindowsでビルドしたGo実行ファイルをLinuxで実行するため
  - Cの依存がなければ、単一の静的リンク実行ファイルが生成される
- `GOOS=linux` : ターゲットOSを指定
  - `GOARCH=amd64` も指定可能（CPU アーキテクチャ）

**例：ARM64（Apple Silicon）向けビルド**
```dockerfile
RUN GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o server ./cmd/server
```

---

### サービス間通信

#### Docker Compose ネットワーク

```yaml
services:
  api:
    environment:
      - DB_HOST=db  # ← サービス名
```

**仕組み：**
- Docker Compose は自動的にネットワークを作成
- サービス名（api、db）が DNS 解決される
- コンテナ内から `db` と接続すると、MySQL コンテナ の`3306` ポートに到達

**ホストネームの自動設定：**
```
api コンテナから見ると：
- localhost → api コンテナ自身
- db → db コンテナの MySQL

db コンテナから見ると：
- localhost → db コンテナ自身
- api → api コンテナのアプリケーション
```

#### ホストマシンからのアクセス

```bash
# ホストマシンから api にアクセス
curl http://localhost:8080/api/v1/health

# ホストマシンから MySQL に直接アクセス
mysql -h localhost -P 3306 -u delivery_user -p delivery_db
```

#### 環境変数のオーバーライド

`docker-compose.yml` では：
```yaml
environment:
  - DB_HOST=db  # ← Docker Compose ネットワーク内
```

本番環境では、この値を環境変数でオーバーライド可能：
```bash
DB_HOST=prod-db.example.com docker-compose up -d
```

---

## 実装パターンと学習ポイント

### 1. クリーンアーキテクチャとレイヤー分離

このプロジェクトで実装されているパターン：

```
HTTP Request
    ↓
[Handler] ← JSON validation, status codes
    ↓
[Repository] ← Database operations only
    ↓
[Model] ← Data structures
    ↓
Database
```

**利点：**
- **テストが容易** : Repository をモック化し、Handler のテストを単独で実行可能
- **保守性向上** : ビジネスロジックの変更が DB 操作に影響しない
- **再利用性** : Repository は複数の Handler から利用可能

**実装例 - テスト**

```go
// テスト用の Mock Repository
type MockRepository struct {}

func (m *MockRepository) GetAll() ([]model.Delivery, error) {
	return []model.Delivery{{ID: 1, Name: "Test"}}, nil
}

// Handler のテスト
handler := handler.NewDeliveryHandler(&MockRepository{})
// ... handler の動作を検証
```

### 2. 依存性注入（Dependency Injection）

```go
// router.go
deliveryRepo := repository.NewDeliveryRepository(db)
deliveryHandler := handler.NewDeliveryHandler(deliveryRepo)
```

**パターン：**
- コンストラクタ関数（`NewXxxx`）で依存関係を注入
- グローバル変数を使わない
- テスト時に異なるインスタンスを注入可能

### 3. エラーハンドリング

複数のレベルでエラー処理が行われます：

```go
// main.go - 致命的エラー
if err != nil {
    log.Fatalf("Failed to connect to database: %v", err)  // アプリ終了
}

// handler.go - ユーザーへのレスポンス
if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "..."})
    return  // レスポンス後に処理を中止
}

// repository.go - エラー情報の詳細化
if errors.Is(err, gorm.ErrRecordNotFound) {
    return nil, errors.New("delivery not found")  // わかりやすいエラーメッセージ
}
```

**ベストプラクティス：**
1. **エラー時は即座に return** : 処理を続行しない
2. **適切なHTTPステータスコード** : 4xx（クライアント問題）と 5xx（サーバー問題）を区別
3. **エラーログ** : デバッグのため、詳細なログを記録

### 4. HTTP ステータスコードの使い分け

| コード | 用途 | 例 |
|-------|------|-----|
| 200 | リクエスト成功 | GET 成功、PUT 成功 |
| 201 | リソース作成成功 | POST 成功 |
| 204 | コンテンツなし | OPTIONS プリフライト応答 |
| 400 | クライアント入力エラー | 無効なID、必須フィールド不足 |
| 404 | リソース未検出 | ID が見つからない |
| 500 | サーバーエラー | DB 接続失敗、予期しない例外 |

### 5. 構造体 tag の活用

```go
// JSON : API入出力の形式
// binding : Ginの入力検証
// gorm : GORM のデータベース操作
type Delivery struct {
    ID      uint      `json:"id" gorm:"primaryKey"`
    Name    string    `json:"name" binding:"required"`
    Status  string    `json:"status" gorm:"default:pending"`
}
```

複数の tool が同じフィールドに異なる目的で tag を付与。Go の柔軟な設計パターン。

### 6. ポインタの使い分け

```go
// GetByID は単一レコードを返す可能性があるため、ポインタ
func (r *DeliveryRepository) GetByID(id uint) (*model.Delivery, error)

// GetAll はスライスを返すため、スライスそのもの（ポインタ不要）
func (r *DeliveryRepository) GetAll() ([]model.Delivery, error)

// Create/Update は元の値を変更するため、ポインタ
func (r *DeliveryRepository) Create(delivery *model.Delivery) error
```

**ルール：**
- スライス/マップ を返す場合は、通常ポインタ不要（既に参照型）
- 構造体を返す場合で、呼び出し側で変更する可能性があれば、ポインタを返す

### 7. メソッドレシーバーの `*` の意味

```go
// DeliveryRepository のメソッドとしてメソッドを定義
// (r *DeliveryRepository) : DeliveryRepository へのポインタレシーバー
func (r *DeliveryRepository) GetAll() ([]model.Delivery, error) {
    // r.db にアクセス可能
}
```

- ポインタレシーバー `(r *DeliveryRepository)` : メソッド内でレシーバーを変更可能
- 値レシーバー `(r DeliveryRepository)` : メソッド内の変更が呼び出し側に反映されない

このプロジェクトではすべてポインタレシーバーを使用（DB操作のため）。

### 8. インターフェース（Interface）の活用

```go
// Repositoryインターフェース（実装されていないが、設計パターンの例）
type DeliveryRepositoryInterface interface {
    GetAll() ([]model.Delivery, error)
    GetByID(id uint) (*model.Delivery, error)
    Create(delivery *model.Delivery) error
    Update(id uint, delivery *model.Delivery) error
    Delete(id uint) error
}

// これを使用することで、テスト時に Mock を実装できる
type MockRepository struct {}
func (m *MockRepository) GetAll() ([]model.Delivery, error) { ... }
```

Go の interface は暗黙的に実装（duck typing）される強力なパターン。

---

## デプロイメントとベストプラクティス

### ローカル開発フロー

```bash
# 1. コード編集
# 2. 動作確認（ローカル実行）
make run

# 3. Docker で同じ環境で実行
make docker-up
curl http://localhost:8080/api/v1/health

# 4. コード品質チェック
make lint
make fmt

# 5. 停止
make docker-down
```

### 本番環境への展開のポイント

1. **環境変数の管理**
   - `.env` は Git に含めない
   - 本番環境では別途環境変数を設定

2. **ログ出力**
   - 本番環境では JSON 形式のログが推奨
   - スタックトレースの記録

3. **パフォーマンス**
   - コネクションプーリング
   - キャッシング（Redis 等）
   - インデックスの最適化

4. **セキュリティ**
   - HTTPS の強制
   - CORS の制限（`"*"` ではなく特定ドメイン）
   - 入力検証の強化
   - SQL インジェクション対策（GORMが自動で実装）

---

## まとめ

このカリキュラムで学んだ内容：

1. **Clean Architecture** - レイヤー分離によるスケーラブルな設計
2. **Go基本** - struct、interface、ポインタ、エラーハンドリング
3. **Gin Framework** - HTTPハンドラー、ルーティング、ミドルウェア
4. **GORM ORM** - DB 操作の抽象化、AutoMigrate
5. **MySQL** - テーブル設計、データベース操作
6. **Docker** - マルチステージビルド、環境構築、サービス連携
7. **REST API 設計** - HTTPメソッド、ステータスコード、JSON フォーマット
8. **運用** - ログ出力、エラー処理、デプロイメント

これらのパターンと知識は、Go を使った本番レベルのバックエンド開発に必須です。

---

**最後に：** このコードを完全に理解した後、以下のような拡張を試してみることをお勧めします：

- [ ] JWTによる認証機能の追加
- [ ] バリデーションルールの拡張
- [ ] ページネーション機能（GET /deliveries?page=1&limit=10）
- [ ] フィルター機能（GET /deliveries?status=completed）
- [ ] ユニットテストの追加
- [ ] PostgreSQL への移行
- [ ] Kubernetes デプロイメント
- [ ] CI/CD パイプライン（GitHub Actions）

日本語学習資料として、このドキュメントを何度も読み返し、コードと照らし合わせることで、深い理解が得られるでしょう。
