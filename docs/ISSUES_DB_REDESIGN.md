# GitHub Issues: データベース再設計

要件定義書・テーブル設計書に基づくissue一覧。
上から順に着手することを想定（依存関係順）。

---

## Issue #1: `companies` テーブル新設と users の所属関係構築

**Labels:** `database`, `priority: critical`

### 概要
運送会社を管理する `companies` テーブルを新設し、`users` テーブルに `company_id` FK を追加する。

### 背景
現状 `users.company` はただの文字列フィールド。運送会社が主体のサービスにおいて、会社単位での配車管理・段階的情報開示・評価の蓄積に対応できない。

### やること
- [ ] `companies` テーブル作成（name, display_name, address, headquarters_lat/lng, phone, email, logo_url, rating_avg, total_deals, plan）
- [ ] `users` テーブルに `company_id` カラム追加（nullable FK → companies.id）
- [ ] `users.role` の値を拡張: `owner` / `driver` / `dispatcher` / `shipper` / `admin`
- [ ] 既存 `users.company`（文字列）から `companies` レコードを生成するマイグレーションスクリプト作成
- [ ] Go model 定義: `model/company.go`
- [ ] Go repository 定義: `repository/company.go`
- [ ] `model/user.go` に `CompanyID` フィールドと `Company` リレーション追加

### 受け入れ条件
- companies テーブルが作成され、既存ユーザーの会社情報が移行されている
- users.company_id で companies と JOIN できる
- 荷主（shipper）は company_id が NULL でも登録可能

---

## Issue #2: `vehicles` テーブルを会社所有に変更

**Labels:** `database`, `priority: critical`

### 概要
既存の `vehicles` テーブルを `user_id` ベースから `company_id` ベースに変更する。

### 背景
車両は個人ではなく会社が所有する。同じ車両を日によって別のドライバーが運転するケースに対応する必要がある。

### やること
- [ ] `vehicles.user_id` → `vehicles.company_id` に変更
- [ ] `plate_number` に UNIQUE 制約を追加
- [ ] `status` カラム追加（active / maintenance / retired）
- [ ] 既存データのマイグレーション: `user_id` → そのユーザーの `company_id` に変換
- [ ] `model/vehicle.go` と `repository/vehicle.go` を更新

### 受け入れ条件
- 車両が会社単位で管理され、plate_number で一意に特定できる
- 既存データが正しく移行されている

---

## Issue #3: `dispatch_plans` テーブル新設（配車計画）

**Labels:** `database`, `priority: critical`

### 概要
「1台のトラック × 1日」を表す配車計画テーブルを新設する。ソロモードの核心。

### 背景
現状の `trips` テーブルは1便=1レコードで独立しており、「トラック③の今日の配車」のようなグルーピングができない。

### やること
- [ ] `dispatch_plans` テーブル作成（company_id FK, vehicle_id FK, driver_id FK, plan_date, status, note）
- [ ] ステータス: `planned` / `in_progress` / `completed` / `cancelled`
- [ ] インデックス: `(company_id, plan_date)`, `(driver_id)`, `(status)`
- [ ] Go model 定義: `model/dispatch_plan.go`
- [ ] Go repository 定義: `repository/dispatch_plan.go`（GetByCompanyAndDate, GetByDriverなど）

### 受け入れ条件
- 配車計画が会社・車両・ドライバー・日付で作成できる
- 会社IDと日付で当日の配車一覧を取得できる

---

## Issue #4: `trip_legs` テーブル新設（区間）/ 旧 `trips` からの移行

**Labels:** `database`, `priority: critical`

### 概要
配車計画を構成する各区間のテーブルを新設し、旧 `trips` テーブルのデータを移行する。

### 背景
1つの配車計画は複数の区間（往路・帰路）で構成される。区間ごとに荷物状態・公開設定を持たせることで、帰り便の空車区間だけを公開マッチング対象にできる。

### やること
- [ ] `trip_legs` テーブル作成（dispatch_plan_id FK, leg_order, origin/destination, departure_at, arrival_at, leg_type, cargo_status, visibility, route_polyline, route_steps_json, route_duration_sec, delay_minutes, price, status）
- [ ] `cargo_status`: `loaded` / `empty_seeking` / `empty_private`
- [ ] `visibility`: `private` / `company_only` / `public`
- [ ] `status`: `scheduled` / `in_transit` / `completed` / `cancelled`
- [ ] インデックス: `(dispatch_plan_id)`, `(visibility, cargo_status)`, `(departure_at)`, `(status)`
- [ ] Go model: `model/trip_leg.go`
- [ ] Go repository: `repository/trip_leg.go`（SearchPublicLegs, GetByDispatchPlanなど）
- [ ] 既存 `trips` → `dispatch_plans` + `trip_legs` へのマイグレーションスクリプト

### 受け入れ条件
- 区間ごとに荷物ステータスと公開範囲を設定できる
- `cargo_status=empty_seeking` かつ `visibility=public|company_only` の区間のみが検索対象
- 既存 trips データが正しく移行されている

---

## Issue #5: `matches` テーブル改修（trip_id → trip_leg_id / ステータス追加）

**Labels:** `database`, `priority: high`

### 概要
マッチングの対象を「便」から「区間」に変更し、決済必須のステータス遷移を実装する。

### 背景
- マッチング対象が便全体ではなく区間（trip_leg）単位になる
- `approved` → 直接 `completed` にできてしまう問題を修正（`payment_pending` を挟む）
- 荷主だけでなく他の運送会社もリクエスト可能に（`requester_id` + `requester_company_id`）

### やること
- [ ] `matches.trip_id` → `matches.trip_leg_id` に変更
- [ ] `matches.shipper_id` → `matches.requester_id` にリネーム
- [ ] `matches.requester_company_id` カラム追加（nullable FK → companies.id）
- [ ] ステータスに `payment_pending` 追加: `pending → approved → payment_pending → completed`
- [ ] `CompleteMatch` ハンドラーで決済済みチェックを追加
- [ ] 既存データのマイグレーション: trip_id → 対応する trip_leg_id
- [ ] model/match.go, repository/match.go, handler/match.go を更新

### 受け入れ条件
- マッチングが区間単位で作成される
- 決済なしで completed にできない
- 運送会社間マッチング（company_to_company）が正しく動作する

---

## Issue #6: `chat_rooms` / `chat_messages` テーブル新設

**Labels:** `database`, `feature`, `priority: high`

### 概要
マッチング承認後に自動でチャットルームを作成し、メッセージのやり取りを可能にする。

### 背景
マッチング成立後のコミュニケーション手段がない。決済完了前は電話番号を非公開にするため、チャットが唯一の連絡手段になる。

### やること
- [ ] `chat_rooms` テーブル作成（match_id FK UNIQUE, status）
- [ ] `chat_messages` テーブル作成（chat_room_id FK, sender_id FK, content, status, filtered_reason）
- [ ] メッセージステータス: `sent` / `filtered` / `warned`
- [ ] インデックス: `chat_messages(chat_room_id, created_at)`
- [ ] Go model: `model/chat_room.go`, `model/chat_message.go`
- [ ] Go repository: `repository/chat.go`
- [ ] Go handler: `handler/chat.go`（SendMessage, GetMessages, GetRoomByMatch）
- [ ] ApproveMatch 時にチャットルームを自動作成するロジック追加

### 受け入れ条件
- マッチング承認時にチャットルームが自動作成される
- 双方がメッセージを送受信できる
- チャットルームはマッチングに1:1で紐づく

---

## Issue #7: チャットのコンテンツフィルタリング実装

**Labels:** `database`, `feature`, `priority: high`

### 概要
チャットメッセージに含まれる電話番号・メールアドレス・URLを検知しブロックする仕組みを実装する。

### 背景
決済前にチャットで連絡先を交換されると、プラットフォームを介さず直接取引される（システムただ使い問題）。

### やること
- [ ] `blocked_patterns` テーブル作成（pattern, description, is_active）
- [ ] `user_violations` テーブル作成（user_id FK, chat_message_id FK, violation_type, detail）
- [ ] 初期パターンデータ投入:
  - 電話番号: `\d{2,4}-\d{2,4}-\d{4}`, `0\d{9,10}`
  - メールアドレス: `[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+`
  - URL: `https?://\S+`
- [ ] メッセージ送信時のフィルタリングロジック（handler/chat.go）
- [ ] 検知時: メッセージを `filtered` にし、`user_violations` に記録
- [ ] API レスポンスで送信者に警告を返す

### 受け入れ条件
- 電話番号・メール・URLを含むメッセージがブロックされる
- 送信者に警告が表示される
- 違反が user_violations に記録される

---

## Issue #8: `reviews` テーブル新設（評価・レビュー）

**Labels:** `database`, `feature`, `priority: medium`

### 概要
マッチング完了後の相互評価システムを実装する。

### 背景
「荷物がのるか、信用できるかは完全自己責任」問題への対策。会社の信頼度を蓄積する仕組みが必要。

### やること
- [ ] `reviews` テーブル作成（match_id FK, reviewer_id FK, reviewee_company_id FK, rating 1-5, comment）
- [ ] インデックス: `(reviewee_company_id)`
- [ ] Go model / repository / handler 作成
- [ ] マッチング完了時に評価を促す通知ロジック
- [ ] `companies.rating_avg` と `companies.total_deals` を更新するトリガー or バッチ処理
- [ ] 1マッチングにつき各者1回のみ評価可能の制約

### 受け入れ条件
- マッチング完了後に双方が評価を投稿できる
- companies.rating_avg が評価に応じて更新される
- 重複評価が防止される

---

## Issue #9: 段階的情報開示のAPI実装

**Labels:** `backend`, `feature`, `priority: critical`

### 概要
マッチングの段階に応じて会社情報の公開範囲を制御するAPIロジックを実装する。

### 背景
「直電問題」対策の核心。検索時には匿名、決済完了後に初めて連絡先を開示する。

### やること
- [ ] APIレスポンスのフィールドフィルタリングロジック作成
- [ ] 段階定義:
  - 検索時: エリア・車種・積載量・料金のみ
  - マッチング申請後: 評価・実績件数・display_name
  - 承認後: 会社名（name）
  - 決済完了後: 電話番号・住所・運転手名
- [ ] trip_legs 検索APIで company 情報をフィルタリングして返す
- [ ] matches APIで段階に応じた company 情報を返す

### 受け入れ条件
- 各段階で開示される情報が正しく制御されている
- 決済前に電話番号・住所が取得できないことをテストで確認

---

## Issue #10: `subscriptions` テーブル新設（Stripe Billing）

**Labels:** `database`, `feature`, `priority: low`

### 概要
運送会社向けの月額サブスクリプション管理テーブルを新設し、Stripe Billing と連携する。

### 背景
無料プラン（ソロモードのみ、月3件マッチング）と有料プラン（無制限マッチング）を分けてマネタイズする。

### やること
- [ ] `subscriptions` テーブル作成（company_id FK, plan, stripe_subscription_id, status, current_period_start/end）
- [ ] プラン定義: `free`（月3件）/ `standard` ¥9,800/月 / `premium` ¥29,800/月
- [ ] Stripe Billing APIとの連携:
  - Checkout Session 作成
  - Webhook: `invoice.paid`, `customer.subscription.updated`, `customer.subscription.deleted`
- [ ] マッチング作成時にプランの上限チェック（freeなら月3件）
- [ ] Go model / repository / handler 作成

### 受け入れ条件
- 会社がプランを選択しStripeで決済できる
- freeプランで月4件目のマッチング作成時にエラーが返る
- Webhookで subscription status が正しく更新される

---

## Issue #11: `trackings` テーブル改修（配車計画対応）

**Labels:** `database`, `priority: medium`

### 概要
trackings テーブルを配車計画・区間に紐づける形に改修する。

### 背景
現状 `trackings.trip_id` で便に紐づいているが、新設計では `dispatch_plan_id` と `trip_leg_id` に紐づける。

### やること
- [ ] `trackings.trip_id` → `trackings.dispatch_plan_id` に変更
- [ ] `trackings.trip_leg_id` カラム追加（nullable FK）
- [ ] インデックス: `(dispatch_plan_id, recorded_at)`
- [ ] 位置予測ロジック（`util/predict_location.go`）を trip_legs 対応に更新
- [ ] 既存データのマイグレーション

### 受け入れ条件
- 配車計画単位で位置情報を記録・取得できる
- 位置予測が trip_legs の departure_at / arrival_at / route_steps_json を使って動作する

---

## Issue #12: 配車ボード画面（複数ルート同時表示）

**Labels:** `frontend`, `feature`, `priority: high`

### 概要
1つの地図に自社の全トラックのルートを色分けして重ねて表示し、各トラックの予測位置をリアルタイム風に表示する「配車ボード」画面を新規作成する。

### 背景
ソロモードの核心機能。現状は1便ずつしか地図表示できず、「トラック①②③が今どこにいるか」を一覧管理できない。運送会社が最初に価値を感じる機能。

### やること
- [ ] `DispatchBoardPage.jsx` 新規作成（`/dispatch-board`）
- [ ] 日付選択 → 当日の dispatch_plans 一覧を取得する API 呼び出し
- [ ] `MultiRouteMap` コンポーネント新規作成:
  - 複数の `DirectionsRenderer` をトラックごとに色分けして描画
  - 各トラックの予測位置マーカーを表示（predict API を複数呼び出し）
  - 60秒ごとに予測位置を自動更新
  - マーカークリックで InfoWindow（車番・進捗・残り時間）
- [ ] トラック一覧テーブル（車番・ルート・進捗%・配送状態・帰り荷状態）
- [ ] テーブル行クリック → 地図上の該当ルートをハイライト
- [ ] バックエンド: `GET /api/v1/dispatch-plans?company_id=X&date=YYYY-MM-DD` エンドポイント追加
- [ ] バックエンド: 複数区間の予測位置を一括取得する `GET /api/v1/dispatch-plans/:id/predict` エンドポイント追加

### 受け入れ条件
- 1つの地図に3台以上のトラックのルートが色分けで同時表示される
- 各トラックの予測位置が地図上にマーカーで表示される
- 60秒ごとに予測位置が更新される
- トラック一覧テーブルで進捗が確認できる

---

## Issue #13: 旧 `trips` / `deliveries` テーブルの廃止

**Labels:** `database`, `cleanup`, `priority: low`

### 概要
新テーブル構造への移行完了後、旧テーブルを廃止する。

### やること
- [ ] 全APIが新テーブル（dispatch_plans + trip_legs）を参照していることを確認
- [ ] 旧 `trips` テーブルへの参照をコードから完全に削除
- [ ] `deliveries` テーブル（レガシー）を削除
- [ ] `delivery.go`（model / repository / handler）を削除
- [ ] 旧テーブルの DROP マイグレーション作成

### 受け入れ条件
- 旧テーブルへの参照がコードに残っていない
- 全機能が新テーブル構造で正しく動作する
