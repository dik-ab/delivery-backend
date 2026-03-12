# 全体設計書（アーキテクチャ）

## システム概要

帰り便マッチングプラットフォーム。運送会社/ドライバーが登録した配送便（特に帰り便＝空車）に対して、荷主が荷物の配送をリクエストし、マッチング・決済・追跡までを一気通貫で提供するWebアプリケーション。

## 技術スタック

### バックエンド（delivery-api）

| 項目 | 技術 |
|------|------|
| 言語 | Go |
| Webフレームワーク | Gin |
| ORM | GORM |
| データベース | MySQL |
| 認証 | JWT + bcrypt |
| 決済 | Stripe Payment Intent API (v76) |
| 地理計算 | Haversine距離計算 |
| 外部API | Google Geocoding API, Google Directions API |

### フロントエンド（delivery-frontend）

| 項目 | 技術 |
|------|------|
| 言語 | JavaScript (React) |
| ビルドツール | Vite |
| 地図 | @react-google-maps/api |
| 決済UI | Stripe.js (CDN) |
| 状態管理 | React Context + Custom Hooks |
| ルーティング | React Router |

## ディレクトリ構成

```
google-maps-projects/
├── delivery-api/                    # バックエンド
│   ├── cmd/server/main.go          # エントリポイント
│   ├── internal/
│   │   ├── model/                  # データモデル定義
│   │   │   ├── user.go
│   │   │   ├── trip.go
│   │   │   ├── match.go
│   │   │   ├── vehicle.go
│   │   │   ├── payment.go
│   │   │   ├── tracking.go
│   │   │   └── delivery.go        # レガシー
│   │   ├── repository/             # データアクセス層
│   │   │   ├── user.go
│   │   │   ├── trip.go
│   │   │   ├── match.go
│   │   │   ├── payment.go
│   │   │   ├── tracking.go
│   │   │   └── delivery.go
│   │   ├── handler/                # APIハンドラー（コントローラー）
│   │   │   ├── auth.go
│   │   │   ├── trip.go
│   │   │   ├── match.go
│   │   │   ├── payment.go
│   │   │   ├── tracking.go
│   │   │   ├── predict.go
│   │   │   ├── admin.go
│   │   │   └── delivery.go
│   │   ├── middleware/             # ミドルウェア
│   │   │   ├── auth.go            # JWT認証
│   │   │   ├── cors.go            # CORS
│   │   │   └── admin.go           # 管理者権限
│   │   ├── router/                # ルーティング
│   │   │   └── router.go
│   │   └── util/                  # ユーティリティ
│   │       ├── haversine.go       # 距離計算
│   │       └── predict_location.go # 位置予測
│   └── go.mod
│
├── delivery-frontend/               # フロントエンド
│   ├── src/
│   │   ├── api/                    # APIクライアント
│   │   │   ├── client.js          # fetchラッパー
│   │   │   ├── trips.js
│   │   │   ├── matches.js
│   │   │   ├── payments.js
│   │   │   ├── tracking.js
│   │   │   └── admin.js
│   │   ├── hooks/                  # カスタムフック
│   │   │   ├── useAuth.js
│   │   │   ├── useTrips.js
│   │   │   ├── useMatches.js
│   │   │   └── useDeliveries.js
│   │   ├── context/               # React Context
│   │   │   └── AuthContext.jsx
│   │   ├── pages/                 # ページコンポーネント
│   │   │   ├── LoginPage.jsx
│   │   │   ├── RegisterPage.jsx
│   │   │   ├── DashboardPage.jsx
│   │   │   ├── TripCreatePage.jsx
│   │   │   ├── TripSearchPage.jsx
│   │   │   ├── TripDetailPage.jsx
│   │   │   ├── MyTripsPage.jsx
│   │   │   ├── MyMatchesPage.jsx
│   │   │   ├── TrackingPage.jsx
│   │   │   ├── PaymentPage.jsx
│   │   │   ├── AdminDashboardPage.jsx
│   │   │   ├── AdminTripsPage.jsx
│   │   │   └── AdminUsersPage.jsx
│   │   ├── components/            # 共通コンポーネント
│   │   │   ├── TripCard.jsx
│   │   │   ├── MatchCard.jsx
│   │   │   ├── StatsCard.jsx
│   │   │   ├── Map.jsx           # レガシー
│   │   │   ├── TripRouteMap.jsx
│   │   │   └── PredictedLocationMap.jsx
│   │   └── App.jsx               # ルーティング定義
│   ├── index.html
│   └── vite.config.js
│
└── docs/                           # ドキュメント
```

## レイヤードアーキテクチャ

```
┌─────────────────────────────────────────────┐
│                 Frontend (React)             │
│  Pages → Hooks → API Client → fetch()       │
└──────────────────────┬──────────────────────┘
                       │ HTTP (REST JSON)
┌──────────────────────▼──────────────────────┐
│                 Backend (Go/Gin)             │
│                                             │
│  Router → Middleware → Handler → Repository  │
│           (Auth/CORS)  (Logic)   (GORM/DB)  │
└──────────────────────┬──────────────────────┘
                       │
┌──────────────────────▼──────────────────────┐
│              MySQL Database                  │
└─────────────────────────────────────────────┘
          │                    │
    ┌─────▼─────┐      ┌──────▼──────┐
    │  Stripe   │      │ Google Maps │
    │  API      │      │ API         │
    └───────────┘      └─────────────┘
```

## APIエンドポイント一覧

### 公開エンドポイント（認証不要）

| メソッド | パス | 説明 |
|---------|------|------|
| GET | `/api/v1/health` | ヘルスチェック |
| POST | `/api/v1/auth/register` | ユーザー登録 |
| POST | `/api/v1/auth/login` | ログイン |
| POST | `/api/v1/webhook/stripe` | Stripe Webhook |

### 認証必須エンドポイント

| メソッド | パス | 説明 | 権限 |
|---------|------|------|------|
| GET | `/api/v1/trips` | 全便取得 | 全ロール |
| GET | `/api/v1/trips/:id` | 便詳細 | 全ロール |
| POST | `/api/v1/trips` | 便登録 | driver |
| PUT | `/api/v1/trips/:id` | 便更新 | driver |
| DELETE | `/api/v1/trips/:id` | 便削除 | driver |
| POST | `/api/v1/trips/search` | 便検索 | 全ロール |
| GET | `/api/v1/trips/:trip_id/predict` | 位置予測 | 全ロール |
| GET | `/api/v1/matches` | 全マッチング取得 | 全ロール |
| GET | `/api/v1/matches/:id` | マッチング詳細 | 全ロール |
| POST | `/api/v1/matches` | マッチングリクエスト送信 | shipper |
| PUT | `/api/v1/matches/:id/approve` | マッチング承認 | driver |
| PUT | `/api/v1/matches/:id/reject` | マッチング拒否 | driver |
| PUT | `/api/v1/matches/:id/complete` | マッチング完了 | driver |
| POST | `/api/v1/payments/create-intent` | 決済Intent作成 | shipper |
| GET | `/api/v1/payments/match/:match_id` | マッチングの決済情報 | 全ロール |
| PUT | `/api/v1/payments/:id/confirm` | 手動決済確認 | 全ロール |
| POST | `/api/v1/tracking` | 位置情報記録 | driver |
| GET | `/api/v1/tracking/:trip_id` | 追跡履歴 | 全ロール |
| GET | `/api/v1/tracking/:trip_id/latest` | 最新位置 | 全ロール |

### 管理者エンドポイント

| メソッド | パス | 説明 |
|---------|------|------|
| GET | `/api/v1/admin/stats` | 統計情報 |
| GET | `/api/v1/admin/users` | 全ユーザー |
| GET | `/api/v1/admin/trips` | 全便 |
| GET | `/api/v1/admin/matches` | 全マッチング |
| PUT | `/api/v1/admin/users/:id/role` | ロール変更 |

## 認証フロー

```
1. POST /auth/register → { token, user }
2. POST /auth/login    → { token, user }
3. token を localStorage に保存
4. 以降のリクエストに Authorization: Bearer <token> を付与
5. 401レスポンス → localStorage クリア → /login にリダイレクト
```

JWT Claims に含まれる情報: `user_id`, `email`, `role`

## 決済フロー（Stripe）

```
[荷主]                     [Frontend]              [Backend]               [Stripe]
  │                           │                       │                      │
  │── 金額入力 ──────────────→│                       │                      │
  │                           │── create-intent ─────→│                      │
  │                           │                       │── PaymentIntent ────→│
  │                           │                       │←── client_secret ────│
  │                           │←── client_secret ─────│                      │
  │── カード情報入力 ────────→│                       │                      │
  │                           │── confirmCardPayment ─────────────────────→│
  │                           │                       │                      │
  │                           │←── 結果 ──────────────────────────────────│
  │                           │                       │←── Webhook ─────────│
  │                           │                       │  (payment_intent.    │
  │                           │                       │   succeeded)         │
  │                           │                       │── match.status =     │
  │                           │                       │   "completed"        │
```

## 位置予測アルゴリズム

リアルタイムGPSを使わず、ルートデータと出発時刻から現在位置を推定する仕組み。

```
入力: route_steps_json, departure_at, delay_minutes, target_time
  ↓
経過時間を計算: elapsed = target_time - departure_at
  ↓
遅延を考慮: effective_elapsed = elapsed - delay_minutes
  ↓
ルートステップを順に辿り、elapsed に対応する位置を補間
  ↓
出力: { lat, lng, progress_percent, remaining_seconds }
```

## 環境変数

### バックエンド
| 変数 | 説明 |
|------|------|
| DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME | MySQL接続情報 |
| JWT_SECRET | JWT署名キー |
| STRIPE_SECRET_KEY | Stripe秘密鍵 |
| STRIPE_WEBHOOK_SECRET | Stripe Webhook署名検証キー |
| GOOGLE_MAPS_API_KEY | Google Maps API キー |

### フロントエンド
| 変数 | 説明 |
|------|------|
| VITE_API_URL | APIベースURL（デフォルト: http://localhost:8080/api/v1） |
| VITE_GOOGLE_MAPS_API_KEY | Google Maps API キー |
| VITE_STRIPE_PUBLISHABLE_KEY | Stripe公開鍵 |
