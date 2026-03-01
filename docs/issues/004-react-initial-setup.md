# ⚛️ Issue #004: React（Vite）の初期環境構築

## 📋 概要

フロントエンドの開発環境を React + Vite で構築する。
別リポジトリ（`delivery-frontend`）として作成し、開発サーバーが起動するところまで確認する。

## 🎯 ゴール

- Vite + React のプロジェクトが作成できている
- `pnpm dev` で開発サーバーが起動する
- ブラウザで `http://localhost:3000` にアクセスして画面が表示される

## 📝 タスク

### 1. プロジェクトの作成

```bash
# プロジェクト作成
pnpm create vite delivery-frontend --template react

# ディレクトリに移動
cd delivery-frontend

# 依存関係インストール
pnpm install
```

> 💡 **ポイント**: `pnpm` は `npm` と比べてディスク効率が良く、インストールも高速。パッケージマネージャーとしておすすめ。

### 2. 追加パッケージのインストール

```bash
# 後のIssueで使うけど、先に入れておく
pnpm add react-router-dom
```

> 💡 **ポイント**: API リクエストにはブラウザ標準の `fetch` API を使うので、axios のような外部ライブラリは不要。依存を減らせるし、ブラウザの標準仕様を学べる。

### 3. Vite の設定変更

`vite.config.js` を編集して、ポートを 3000 に変更する。

```js
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  server: {
    port: 3000,
  },
})
```

> 💡 **ポイント**: デフォルトは 5173 番ポートだけど、3000 に変えておくとバックエンドの CORS 設定と合わせやすい。

### 4. 環境変数ファイルの作成

`.env.local` を作成する（Gitにはコミットしない）。

```
VITE_API_URL=http://localhost:8080
```

`.env.example` も作成する（こっちはコミットする）。

```
VITE_API_URL=http://localhost:8080
```

> 💡 **ポイント**: Vite では環境変数の名前を `VITE_` で始める必要がある。`VITE_` プレフィックスがないと、クライアントサイドのコードからアクセスできない。

### 5. 初期の App.jsx を整理

`src/App.jsx` をシンプルにする。

```jsx
import { useState, useEffect } from 'react'
import './App.css'

function App() {
  const [message, setMessage] = useState('Loading...')

  useEffect(() => {
    // 後のIssueでAPIを叩くようにする
    setMessage('🚚 Delivery App - Frontend Ready!')
  }, [])

  return (
    <div className="App">
      <h1>{message}</h1>
      <p>フロントエンドの環境構築が完了しました！</p>
    </div>
  )
}

export default App
```

### 6. .gitignore の確認

Vite がデフォルトで作成する `.gitignore` に以下が含まれていることを確認：

```
node_modules
dist
.env.local
```

### 7. ディレクトリ構成の確認

```
delivery-frontend/
├── public/
│   └── index.html
├── src/
│   ├── App.jsx          ← メインコンポーネント
│   ├── App.css
│   ├── index.jsx        ← エントリーポイント
│   └── index.css
├── .env.example
├── .env.local           ← ※Git管理外
├── .gitignore
├── index.html           ← Viteのエントリーポイント
├── package.json
├── pnpm-lock.yaml
└── vite.config.js
```

> 💡 **ポイント**: Vite では `index.html` がプロジェクトルートに置かれる。これが SPA のエントリーポイントになる。`public/` の中ではなくルートにあることに注意。

## ✅ 動作確認

```bash
# 開発サーバー起動
pnpm dev

# ブラウザで確認
# http://localhost:3000 にアクセス
# 「🚚 Delivery App - Frontend Ready!」が表示されればOK
```

## 🧪 追加チャレンジ（余裕があれば）

- [ ] `App.css` を自分好みにカスタマイズしてみよう
- [ ] コンポーネントを1つ追加して、`App.jsx` から読み込んでみよう
- [ ] React DevTools（Chrome拡張）をインストールして使ってみよう

## 📖 学習ポイント

| 概念 | 説明 |
|------|------|
| Vite | 次世代のフロントエンドビルドツール。HMR（Hot Module Replacement）が超高速 |
| JSX | JavaScript の中に HTML ライクな構文を書ける。React のコア機能 |
| `useState` | React の状態管理フック。値の変更で自動的に画面が再レンダリングされる |
| `useEffect` | 副作用フック。コンポーネントのマウント時や依存値の変更時に処理を実行 |
| SPA | Single Page Application。ページ遷移なしで画面を切り替えるアーキテクチャ |

## 🏷️ ラベル

`setup` `frontend` `react` `初心者向け`
