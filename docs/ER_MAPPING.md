# ER Mapping（エンティティ関連図）

## ER図（Mermaid）

```mermaid
erDiagram
    users {
        uint id PK
        string email UK "ユニーク"
        string password_hash
        string name
        string role "driver / shipper / admin"
        string company
        string phone
        datetime created_at
        datetime updated_at
    }

    trips {
        uint id PK
        uint driver_id FK "users.id"
        string origin_address
        float origin_lat
        float origin_lng
        string destination_address
        float destination_lat
        float destination_lng
        datetime departure_at
        datetime estimated_arrival "nullable"
        string vehicle_type "軽トラ / 2t / 4t / 10t / 大型"
        float available_weight "kg"
        int price "円"
        string status "open / matched / in_transit / completed / cancelled"
        string trip_type "outbound / return"
        bool is_public "外部公開フラグ"
        bool is_solo_mode "ソロモード"
        string note
        int delay_minutes "渋滞遅延（分）"
        text route_polyline "エンコード済みポリライン"
        int route_duration_sec "所要時間（秒）"
        mediumtext route_steps_json "ルートステップJSON"
        datetime created_at
        datetime updated_at
    }

    matches {
        uint id PK
        uint trip_id FK "trips.id"
        uint shipper_id FK "users.id"
        float cargo_weight "kg"
        string cargo_description
        string status "pending / approved / rejected / completed"
        string message "リクエスト時メッセージ"
        string reject_reason "拒否理由"
        string request_type "shipper_to_company / company_to_company"
        datetime created_at
        datetime updated_at
    }

    vehicles {
        uint id PK
        uint user_id FK "users.id"
        string type "軽トラ / 2t / 4t / 10t / 大型"
        float max_weight "kg"
        string plate_number
        datetime created_at
        datetime updated_at
    }

    payments {
        uint id PK
        uint match_id FK "matches.id"
        uint payer_id FK "users.id"
        int amount "円"
        string currency "jpy"
        string stripe_payment_id
        string status "pending / succeeded / failed / refunded"
        datetime created_at
        datetime updated_at
    }

    trackings {
        uint id PK
        uint trip_id FK "trips.id"
        float lat
        float lng
        datetime recorded_at
    }

    deliveries {
        uint id PK
        string name
        string address
        float lat
        float lng
        string status "pending / in_progress / completed"
        string note
        datetime created_at
        datetime updated_at
    }

    users ||--o{ trips : "driver_id（ドライバーが便を登録）"
    users ||--o{ matches : "shipper_id（荷主がリクエスト）"
    users ||--o{ vehicles : "user_id（車両を所有）"
    users ||--o{ payments : "payer_id（支払者）"
    trips ||--o{ matches : "trip_id（便に対するマッチング）"
    trips ||--o{ trackings : "trip_id（位置情報記録）"
    matches ||--o{ payments : "match_id（マッチングへの決済）"
```

## テーブル一覧

| テーブル名 | 説明 | 主な関連 |
|-----------|------|---------|
| users | ユーザー（ドライバー/荷主/管理者） | trips, matches, vehicles, payments |
| trips | 配送便（往路・帰り便） | users(driver), matches, trackings |
| matches | マッチングリクエスト | trips, users(shipper), payments |
| vehicles | 車両情報 | users |
| payments | 決済情報（Stripe） | matches, users(payer) |
| trackings | 位置情報トラッキング | trips |
| deliveries | 配送先（レガシー） | なし |

## リレーション詳細

### users → trips（1対多）
ドライバー（`role = driver / transport_company`）が配送便を登録する。`trips.driver_id` が `users.id` を参照。

### users → matches（1対多）
荷主（`role = shipper`）がマッチングリクエストを送信する。`matches.shipper_id` が `users.id` を参照。

### trips → matches（1対多）
1つの便に対して複数の荷主からマッチングリクエストが届く。`matches.trip_id` が `trips.id` を参照。

### matches → payments（1対多）
承認済みマッチングに対して決済が行われる。`payments.match_id` が `matches.id` を参照。

### users → payments（1対多）
荷主が決済を行う。`payments.payer_id` が `users.id` を参照。

### users → vehicles（1対多）
ドライバーが複数の車両を登録できる。`vehicles.user_id` が `users.id` を参照。

### trips → trackings（1対多）
便の位置情報を時系列で記録。`trackings.trip_id` が `trips.id` を参照。

## ステータス遷移

### Trip Status
```
open → matched → in_transit → completed
  ↓
cancelled
```

### Match Status
```
pending → approved → completed
  ↓
rejected
```
※ 現状は `approved → completed` の間に決済が強制されていない（改善対象）

### Payment Status
```
pending → succeeded
  ↓
failed
  ↓
refunded
```
