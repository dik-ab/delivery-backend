# 機能仕様書

## 1. ユーザー管理

### 1.1 ユーザー登録
荷主またはドライバー（運送会社）としてアカウントを作成する。

| 項目 | 詳細 |
|------|------|
| 必須入力 | 名前、メールアドレス、パスワード、ロール（driver/shipper） |
| 任意入力 | 会社名、電話番号 |
| パスワード | bcryptでハッシュ化して保存 |
| 認証トークン | 登録完了と同時にJWTを発行、レスポンスで返却 |
| バリデーション | メールアドレスの重複チェック |

### 1.2 ログイン
メールアドレスとパスワードで認証し、JWTトークンを取得する。

| 項目 | 詳細 |
|------|------|
| 認証方式 | メールアドレス + パスワード（bcrypt照合） |
| トークン | JWT（Claims: user_id, email, role） |
| フロント保存先 | localStorage（auth_token, auth_user） |
| 自動リダイレクト | 401レスポンス時にlocalStorageクリア → /login |

### 1.3 ロール

| ロール | 説明 | 主な操作 |
|--------|------|---------|
| driver / transport_company | ドライバー・運送会社 | 便登録、マッチング承認/拒否/完了、位置情報記録 |
| shipper | 荷主 | 便検索、マッチングリクエスト送信、決済 |
| admin | 管理者 | 統計閲覧、ユーザー管理、全データ閲覧 |

## 2. 便（Trip）管理

### 2.1 便の登録
ドライバーが配送便を登録する。

| フィールド | 型 | 説明 |
|-----------|------|------|
| origin_address | string | 出発地住所（Google Geocodingで緯度経度を自動取得） |
| destination_address | string | 到着地住所（同上） |
| departure_at | datetime | 出発日時 |
| vehicle_type | string | 車両タイプ（軽トラ/2t/4t/10t/大型） |
| available_weight | float | 積載可能重量（kg） |
| price | int | 料金（円） |
| trip_type | string | `outbound`（往路）/ `return`（帰り便） |
| is_public | bool | 外部公開するか（荷主から検索可能にする） |
| is_solo_mode | bool | ソロモード（自社便管理のみ、マッチング対象外） |
| delay_minutes | int | 想定渋滞遅延（分） |
| note | string | 備考 |

ルート情報（Google Directions APIから自動取得）:
route_polyline, route_duration_sec, route_steps_json

### 2.2 便の検索
荷主が条件に合う便を検索する。

| 検索パラメータ | 説明 |
|--------------|------|
| origin (lat/lng) | 出発地（住所からGeocodingで変換） |
| destination (lat/lng) | 到着地（同上） |
| radius_km | 検索半径（10/25/50/100/500km） |
| date | 出発日（任意） |
| trip_type | all / outbound / return |

検索ロジック:
Haversine距離計算を用いて、出発地/到着地がそれぞれ指定半径内にある便を抽出。帰り便検索時は出発地⇔到着地を反転して検索する「逆方向マッチング」にも対応。

レスポンス: `{ normal_matches, return_matches, total_matches }`

### 2.3 便のステータス

| ステータス | 説明 |
|-----------|------|
| open | 公開中（マッチング受付中） |
| matched | マッチング成立（承認済み） |
| in_transit | 配送中 |
| completed | 配送完了 |
| cancelled | キャンセル |

## 3. マッチング

### 3.1 マッチングリクエスト送信
荷主が便に対してマッチングリクエストを送信する。

| フィールド | 型 | 説明 |
|-----------|------|------|
| trip_id | uint | 対象便ID |
| cargo_weight | float | 荷物重量（kg） |
| cargo_description | string | 荷物内容 |
| message | string | ドライバーへのメッセージ |

バリデーション: `cargo_weight ≤ trip.available_weight`（積載量超過チェック）

### 3.2 マッチング承認/拒否
ドライバーがリクエストを承認または拒否する。

承認: `PUT /matches/:id/approve` → status を `approved` に更新
拒否: `PUT /matches/:id/reject` → status を `rejected` に更新、reject_reason を保存

### 3.3 マッチング完了
`PUT /matches/:id/complete` → status を `completed` に更新

現状の仕様: ドライバーが手動で完了ボタンを押す。また、Stripe Webhookで決済成功時にも自動で completed に更新される。

### 3.4 マッチングステータス

| ステータス | 説明 | UIラベル |
|-----------|------|---------|
| pending | リクエスト送信済み、ドライバー承認待ち | 待機中 |
| approved | ドライバーが承認、決済待ち | 承認済み |
| rejected | ドライバーが拒否 | 拒否 |
| completed | 配送完了 | 完了 |

### 3.5 リクエストタイプ

| タイプ | 説明 |
|--------|------|
| shipper_to_company | 荷主 → 運送会社（標準） |
| company_to_company | 運送会社間マッチング |

## 4. 決済（Stripe）

### 4.1 決済フロー

1. 荷主がマッチング詳細画面で「決済へ進む」をクリック
2. 金額を入力（最低50円）
3. バックエンドでStripe PaymentIntentを作成、client_secretを返却
4. フロントエンドでStripe Card Elementをマウント
5. confirmCardPayment()で決済実行（3D Secure自動対応）
6. 決済成功 → Stripe Webhookがバックエンドに通知
7. バックエンドでpaymentステータスを`succeeded`に更新、matchステータスを`completed`に更新

### 4.2 決済ステータス

| ステータス | 説明 |
|-----------|------|
| pending | 決済処理中 |
| succeeded | 決済成功 |
| failed | 決済失敗 |
| refunded | 返金済み |

### 4.3 Webhook処理
Stripe署名検証 → イベントタイプに応じた処理:
`payment_intent.succeeded` → payment更新 + match完了
`payment_intent.payment_failed` → payment失敗更新

## 5. 位置情報トラッキング

### 5.1 位置記録
ドライバーが現在位置を記録する。

| フィールド | 型 | 説明 |
|-----------|------|------|
| trip_id | uint | 対象便ID |
| lat | float | 緯度 |
| lng | float | 経度 |

### 5.2 位置予測
リアルタイムGPSではなく、ルートデータと出発時刻から現在位置を推定する。

入力: trip_id, at（予測対象時刻、RFC3339形式）

計算ロジック:
1. 出発時刻からの経過時間を計算
2. delay_minutesを考慮した実効経過時間を算出
3. route_steps_jsonの各ステップを順に辿り、経過時間に対応する位置を線形補間
4. 進捗率、残り時間も同時に算出

レスポンス: `{ lat, lng, progress_percent, elapsed_seconds, remaining_seconds }`

### 5.3 追跡画面
地図上に出発地・到着地・予測位置を表示。60秒ごとに自動リフレッシュ。進捗バーで配送進捗を可視化。

## 6. 管理者機能

### 6.1 統計ダッシュボード
全ユーザー数、全便数、全マッチング数、アクティブ便数、完了便数、待機中/承認済みマッチング数を表示。

### 6.2 ユーザー管理
全ユーザー一覧表示。ロール変更が可能。

### 6.3 データ閲覧
全便・全マッチングのデータを管理画面から閲覧可能。

## 7. フロントエンド画面一覧

| パス | ページ | 説明 |
|------|--------|------|
| /login | LoginPage | ログイン |
| /register | RegisterPage | ユーザー登録 |
| /dashboard | DashboardPage | ダッシュボード（ロール別表示） |
| /trips/new | TripCreatePage | 便登録 |
| /trips/search | TripSearchPage | 便検索 |
| /trips/:id | TripDetailPage | 便詳細・マッチングリクエスト送信 |
| /my-trips | MyTripsPage | 自分の便一覧 |
| /my-matches | MyMatchesPage | マッチング管理 |
| /tracking/:tripId | TrackingPage | 位置追跡 |
| /payment/:matchId | PaymentPage | 決済 |
| /admin | AdminDashboardPage | 管理者ダッシュボード |
| /admin/trips | AdminTripsPage | 管理者：便一覧 |
| /admin/users | AdminUsersPage | 管理者：ユーザー一覧 |
