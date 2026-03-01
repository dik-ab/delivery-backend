# 配送ルート管理API

Go + GinフレームワークとGoogle Mapsを統合した、配送ルート管理アプリケーションのバックエンドAPI。

## 概要

このプロジェクトは、配送先の管理とGoogle Mapsの統合を行うRESTful APIです。配送先の追加、編集、削除、および一覧表示機能を提供します。

## 必要要件

- Go 1.21以上
- MySQL 8.0以上
- Docker & Docker Compose（オプション）
- swag (Swagger生成用): `go install github.com/swaggo/swag/cmd/swag@latest`

## セットアップ

### 1. 環境変数の設定

`.env.example`を`.env`にコピーして設定を編集します：

```bash
cp .env.example .env
```

`.env`ファイルの内容：

```
DB_HOST=localhost
DB_PORT=3306
DB_USER=delivery_user
DB_PASSWORD=delivery_pass
DB_NAME=delivery_db
PORT=8080
GOOGLE_MAPS_API_KEY=your_api_key_here
```

### 2. 依存関係のインストール

```bash
go mod download
```

### 3a. ローカルでの実行（MySQLが別途必要）

```bash
# Swagger生成
make swagger

# アプリケーション実行
make run
```

APIはデフォルトで `http://localhost:8080` で起動します。

### 3b. Dockerでの実行

```bash
# コンテナの起動
make docker-up

# コンテナの停止
make docker-down
```

Dockerで実行する場合、MySQLは自動的にセットアップされます。

## APIエンドポイント

### ヘルスチェック
- `GET /api/v1/health` - APIのステータス確認

### 配送先の操作

#### 一覧取得
```bash
GET /api/v1/deliveries
```

#### 単一取得
```bash
GET /api/v1/deliveries/:id
```

#### 作成
```bash
POST /api/v1/deliveries

Body:
{
  "name": "配送先名",
  "address": "住所",
  "lat": 35.6762,
  "lng": 139.7674,
  "note": "メモ"
}
```

#### 更新
```bash
PUT /api/v1/deliveries/:id

Body:
{
  "name": "配送先名",
  "address": "住所",
  "lat": 35.6762,
  "lng": 139.7674,
  "status": "pending|in_progress|completed",
  "note": "メモ"
}
```

#### 削除
```bash
DELETE /api/v1/deliveries/:id
```

## Swagger API ドキュメント

APIドキュメントにアクセス：

```
http://localhost:8080/swagger/index.html
```

## プロジェクト構造

```
delivery-api/
├── cmd/server/           # エントリーポイント
├── internal/
│   ├── handler/          # HTTPハンドラー
│   ├── model/            # データモデル
│   ├── repository/       # データベース操作
│   ├── router/           # ルーティング設定
│   └── middleware/       # ミドルウェア
├── docs/                 # Swagger生成ファイル
├── docker-compose.yml    # Docker設定
├── Dockerfile            # コンテナイメージ定義
├── .env.example          # 環境変数テンプレート
├── go.mod               # Go依存管理
├── Makefile             # ビルドスクリプト
└── README.md            # このファイル
```

## コマンド

```bash
# アプリケーション実行
make run

# ビルド
make build

# Docker起動
make docker-up

# Docker停止
make docker-down

# Swagger生成
make swagger

# コードフォーマット
make fmt
```

## データベーススキーマ

### deliveries テーブル

| カラム    | 型       | 説明                              |
|-----------|----------|-----------------------------------|
| id        | INT      | プライマリーキー（自動増分）      |
| name      | VARCHAR  | 配送先名                          |
| address   | VARCHAR  | 住所                              |
| lat       | DOUBLE   | 緯度                              |
| lng       | DOUBLE   | 経度                              |
| status    | VARCHAR  | ステータス（pending/in_progress/completed） |
| note      | TEXT     | メモ                              |
| created_at| DATETIME | 作成日時                          |
| updated_at| DATETIME | 更新日時                          |

## テクノロジースタック

- **フレームワーク**: Gin (Go)
- **ORM**: GORM
- **データベース**: MySQL
- **API ドキュメント**: Swagger (swaggo)
- **コンテナ化**: Docker & Docker Compose

## CORS設定

APIはCORSが有効です。すべてのオリジンからのリクエストを受け付けています。

## エラーハンドリング

APIは以下の形式でエラーレスポンスを返します：

```json
{
  "error": "エラーメッセージ"
}
```

## ライセンス

このプロジェクトはMITライセンスの下で公開されています。

## サポート

問題が発生した場合は、GitHubのissueを作成してください。
