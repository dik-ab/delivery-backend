# 🔗 Issue #005: フロントエンドからHealth Checkエンドポイントを叩く

## 📋 概要

React フロントエンドから Go バックエンドの `/health` エンドポイントに API リクエストを送り、レスポンスを画面に表示する。フロントとバックエンドの疎通確認を行う最初の一歩。

## 🎯 ゴール

- フロントエンドから `GET /health` を叩いてレスポンスを取得できる
- 画面に DB の接続状態が表示される
- CORS の設定を理解している

## 📝 前提条件

- ✅ Issue #002（Go + MySQL の Docker Compose 環境構築）が完了していること
- ✅ Issue #003（Health Check エンドポイントの実装）が完了していること
- ✅ Issue #004（React の初期環境構築）が完了していること

## 📝 タスク

### 1. バックエンドに CORS 設定を追加

フロントエンド（localhost:3000）からバックエンド（localhost:8080）にリクエストを送るには、CORS（Cross-Origin Resource Sharing）の設定が必要。

`internal/middleware/cors.go` を作成する。

```go
package middleware

import (
    "github.com/gin-gonic/gin"
)

// CORSMiddleware はクロスオリジンリクエストを許可するミドルウェア
func CORSMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Writer.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
        c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

        // プリフライトリクエストの処理
        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(204)
            return
        }

        c.Next()
    }
}
```

> 💡 **ポイント**: ブラウザはセキュリティのため、異なるオリジン（ドメイン:ポート）へのリクエストをデフォルトでブロックする。CORS ヘッダーを設定することで「このオリジンからのアクセスは許可する」と明示する。

### 2. main.go に CORS ミドルウェアを適用

`cmd/server/main.go` を修正する。

```go
import (
    // ... 既存のインポート
    "github.com/delivery-app/delivery-api/internal/middleware"
)

func main() {
    // ... DB接続部分は変更なし

    r := gin.Default()

    // CORS ミドルウェアを適用（ルート定義の前に！）
    r.Use(middleware.CORSMiddleware())

    // ヘルスチェック
    healthHandler := handler.NewHealthHandler(db)
    r.GET("/health", healthHandler.Check)

    // ... 残りのコードは変更なし
}
```

> ⚠️ **注意**: `r.Use()` は必ずルート定義（`r.GET()` 等）の **前** に書くこと。後に書くとミドルウェアが適用されない。

### 3. フロントエンドに API クライアントを作成

`src/api/client.js` を作成する。外部ライブラリは使わず、ブラウザ標準の `fetch` API でラップする。

```js
const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080'

/**
 * fetch をラップした API クライアント
 * - ベースURLの自動付与
 * - JSON の自動パース
 * - エラーハンドリング
 */
const apiClient = {
  async get(path) {
    const response = await fetch(`${API_URL}${path}`)

    if (!response.ok) {
      throw new Error(`HTTP Error: ${response.status} ${response.statusText}`)
    }

    return response.json()
  },

  async post(path, body) {
    const response = await fetch(`${API_URL}${path}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
    })

    if (!response.ok) {
      throw new Error(`HTTP Error: ${response.status} ${response.statusText}`)
    }

    return response.json()
  },

  async put(path, body) {
    const response = await fetch(`${API_URL}${path}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
    })

    if (!response.ok) {
      throw new Error(`HTTP Error: ${response.status} ${response.statusText}`)
    }

    return response.json()
  },

  async delete(path) {
    const response = await fetch(`${API_URL}${path}`, {
      method: 'DELETE',
    })

    if (!response.ok) {
      throw new Error(`HTTP Error: ${response.status} ${response.statusText}`)
    }

    return response.json()
  },
}

export default apiClient
```

> 💡 **ポイント**:
> - `fetch` はブラウザ標準の API なので追加パッケージ不要
> - axios と違って、`fetch` は HTTP エラー（400, 500 等）でも reject **しない**。`response.ok` を自分でチェックする必要がある
> - `response.json()` で JSON に自動パースできる
> - `import.meta.env.VITE_API_URL` で `.env.local` の値を読み込む

### 4. App.jsx から Health Check を呼び出す

`src/App.jsx` を修正する。

```jsx
import { useState, useEffect } from 'react'
import apiClient from './api/client'
import './App.css'

function App() {
  const [health, setHealth] = useState(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)

  useEffect(() => {
    const checkHealth = async () => {
      try {
        // fetch ベースの apiClient で GET リクエスト
        const data = await apiClient.get('/health')
        setHealth(data)
      } catch (err) {
        setError(err.message)
      } finally {
        setLoading(false)
      }
    }

    checkHealth()
  }, [])

  return (
    <div className="App">
      <h1>🚚 Delivery App</h1>

      <div style={{
        padding: '20px',
        margin: '20px',
        borderRadius: '8px',
        backgroundColor: health?.status === 'healthy' ? '#d4edda' : '#f8d7da',
        border: `1px solid ${health?.status === 'healthy' ? '#c3e6cb' : '#f5c6cb'}`,
      }}>
        <h2>サーバーステータス</h2>

        {loading && <p>⏳ 確認中...</p>}

        {error && (
          <div>
            <p>❌ エラー: {error}</p>
            <p>バックエンドが起動しているか確認してください</p>
          </div>
        )}

        {health && (
          <div>
            <p>ステータス: {health.status === 'healthy' ? '🟢 正常' : '🔴 異常'}</p>
            <p>データベース: {health.database === 'connected' ? '✅ 接続済み' : '❌ 未接続'}</p>
          </div>
        )}
      </div>
    </div>
  )
}

export default App
```

> 💡 **axios との違い**: axios は `response.data` でデータにアクセスするけど、fetch ベースの apiClient は直接データを返す。`const data = await apiClient.get('/health')` でそのまま使える。

### 5. ディレクトリ構成の確認

```
delivery-frontend/
├── src/
│   ├── api/
│   │   └── client.js      ← ⭐ 今回作成
│   ├── App.jsx             ← ⭐ 今回修正
│   ├── App.css
│   ├── index.jsx
│   └── index.css
├── .env.local
├── index.html
├── package.json
└── vite.config.js

delivery-api/
├── cmd/server/main.go      ← ⭐ CORS追加
├── internal/
│   ├── handler/
│   │   └── health.go
│   └── middleware/
│       └── cors.go          ← ⭐ 今回作成
├── docker-compose.yml
└── Dockerfile
```

## ✅ 動作確認

```bash
# ターミナル1: バックエンド起動
cd delivery-api
docker compose up --build

# ターミナル2: フロントエンド起動
cd delivery-frontend
pnpm dev
```

ブラウザで `http://localhost:3000` にアクセスして、以下が表示されればOK：

```
🚚 Delivery App

サーバーステータス
ステータス: 🟢 正常
データベース: ✅ 接続済み
```

### CORS エラーが出た場合

ブラウザの開発者ツール（F12）→ Console に以下のようなエラーが出る場合：

```
Access to fetch at 'http://localhost:8080/health' from origin 'http://localhost:3000' has been blocked by CORS policy
```

→ バックエンドの CORS ミドルウェアが正しく設定されているか確認！

## 🧪 追加チャレンジ（余裕があれば）

- [ ] 「再チェック」ボタンを追加して、クリックで再度 `/health` を叩くようにしてみよう
- [ ] レスポンスのタイムスタンプを表示してみよう
- [ ] CSS をカスタマイズしてカッコよくしてみよう
- [ ] バックエンドの Health Check レスポンスに `version` を追加し、フロントでも表示してみよう

## 📖 学習ポイント

| 概念 | 説明 |
|------|------|
| CORS | Cross-Origin Resource Sharing。異なるオリジン間の通信を安全に行う仕組み |
| Preflight Request | ブラウザが本リクエストの前に `OPTIONS` メソッドで送る確認リクエスト |
| `fetch` | ブラウザ標準の HTTP クライアント API。外部ライブラリ不要で使える |
| `response.ok` | fetch はエラーでも reject しないので、自分でステータスチェックが必要 |
| `response.json()` | レスポンスボディを JSON としてパースする。Promise を返す |
| `async/await` | 非同期処理を同期的に書けるJSの構文。`try/catch` でエラーハンドリング |
| `useEffect` | 第2引数 `[]` で「マウント時に1回だけ」実行される |
| 環境変数 | `import.meta.env.VITE_*` で `.env.local` の値にアクセスできる |

## 🎉 ここまでできたら

おめでとう！フロントエンドとバックエンドが繋がった状態になったよ。
次のステップでは、実際のデータ（配送先情報）の CRUD 機能を実装していく。

## 🏷️ ラベル

`feature` `frontend` `backend` `api` `初心者向け`
