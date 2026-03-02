# 🚀 Issue #001: GitHub リポジトリ作成 & Git 初期設定

## 📋 概要

GitHub にバックエンド・フロントエンドの2つのリポジトリを作成し、ローカル環境と紐づける。
すべての開発作業の土台となるので、最初にこれをやる。

## 🎯 ゴール

- GitHub に `delivery-api`（バックエンド）と `delivery-frontend`（フロントエンド）の2つのリポジトリが作成されている
- ローカルに clone して、最初のコミット & push ができている
- Git の基本設定（ユーザー名・メールアドレス）が完了している

## 📝 前提条件

- GitHub アカウントを持っていること
- Git がインストールされていること

### Git がインストールされているか確認

```bash
git --version
```

`git version 2.x.x` のように表示されればOK。表示されない場合は以下からインストール：

- **Mac**: `brew install git`（Homebrewが必要）
- **Windows**: [Git for Windows](https://gitforwindows.org/) からダウンロード

## 📝 タスク

### 1. Git のグローバル設定

まだ設定していない場合、Git に自分の名前とメールアドレスを登録する。
これはコミット時に「誰がこの変更をしたか」を記録するための設定。

```bash
git config --global user.name "あなたのGitHubユーザー名"
git config --global user.email "あなたのメールアドレス"
```

設定を確認：

```bash
git config --global --list
```

以下のように表示されればOK：

```
user.name=あなたのGitHubユーザー名
user.email=あなたのメールアドレス
```

> 💡 **ポイント**: `--global` をつけると、すべてのリポジトリで共通の設定になる。リポジトリごとに変えたい場合は `--global` を外す。

### 2. GitHub でリポジトリを作成

GitHub にログインして、リポジトリを **2つ** 作成する。

#### バックエンド用リポジトリ

1. GitHub 右上の「+」→「New repository」をクリック
2. 以下を入力：
   - **Repository name**: `delivery-api`
   - **Description**: `配送管理APIサーバー（Go + Gin + MySQL）`（任意）
   - **Public / Private**: どちらでもOK（ポートフォリオにするなら Public がおすすめ）
   - **Initialize this repository with**: 何もチェックしない（READMEもなし）
3. 「Create repository」をクリック

#### フロントエンド用リポジトリ

同じ手順で以下の内容で作成：

- **Repository name**: `delivery-frontend`
- **Description**: `配送管理フロントエンド（React + Vite）`（任意）

> ⚠️ **注意**: 「Add a README file」にチェックを入れないこと！チェックを入れると、ローカルから push するときに競合が起きる。

### 3. ローカルにバックエンドのプロジェクトを作成

```bash
# 作業用ディレクトリに移動（場所はお好みで）
cd ~

# バックエンドのディレクトリ作成 & 初期化
mkdir delivery-api
cd delivery-api

# Git リポジトリを初期化
git init

# ブランチ名を main に変更（Git のデフォルトは master だけど main が主流）
git branch -m main

# GitHub リモートを追加（URLはあなたのGitHubユーザー名に置き換えてね）
git remote add origin https://github.com/あなたのユーザー名/delivery-api.git
```

### 4. 最初のファイルを作成してコミット

まずは `.gitignore` を作成する。これは Git で管理しないファイルを指定するための設定ファイル。

`.gitignore` を作成：

```
# 環境変数（パスワード等が書かれているのでGitに上げない）
.env

# Goのバイナリ
delivery-api
*.exe

# IDE設定
.idea/
.vscode/
*.swp

# OS生成ファイル
.DS_Store
Thumbs.db
```

> 💡 **ポイント**: `.env` ファイルにはデータベースのパスワードなどの秘密情報が入る。これを GitHub に push すると世界中に公開されてしまうので、必ず `.gitignore` に含める。

コミット & push：

```bash
# .gitignore をステージング
git add .gitignore

# コミット
git commit -m "🙈 .gitignore: 環境変数・ビルド成果物を除外"

# GitHub に push
git push -u origin main
```

> 💡 **コマンド解説**:
> - `git add`: 変更をステージングエリアに追加（コミットする準備）
> - `git commit -m "メッセージ"`: 変更を記録する。`-m` でコミットメッセージを指定
> - `git push -u origin main`: GitHub にアップロード。`-u` は次回から `git push` だけで済むようにする設定

### 5. フロントエンドも同様にセットアップ

```bash
# ホームに戻る
cd ~

# フロントエンドのディレクトリ作成 & 初期化
mkdir delivery-frontend
cd delivery-frontend

git init
git branch -m main
git remote add origin https://github.com/あなたのユーザー名/delivery-frontend.git
```

`.gitignore` を作成：

```
# 依存パッケージ（容量が大きいのでGitに上げない）
node_modules

# ビルド成果物
dist

# 環境変数
.env.local
.env

# OS生成ファイル
.DS_Store
Thumbs.db
```

コミット & push：

```bash
git add .gitignore
git commit -m "🙈 .gitignore: node_modules・ビルド成果物・環境変数を除外"
git push -u origin main
```

### 6. push できたか確認

GitHub の各リポジトリページにアクセスして、`.gitignore` ファイルが表示されていればOK！

## 🔍 トラブルシューティング

| 症状 | 原因 | 対処法 |
|------|------|--------|
| `git push` で認証エラー | GitHub の認証設定がない | 下記「GitHub 認証の設定」を参照 |
| `error: remote origin already exists` | すでに remote が設定されてる | `git remote remove origin` してからやり直す |
| `error: failed to push some refs` | リモートに既にコミットがある | README付きでリポジトリを作った場合。リポジトリを削除して作り直すのが一番簡単 |

### GitHub 認証の設定

`git push` 時に認証を求められた場合、以下のどちらかで設定する：

#### 方法A: HTTPS + Personal Access Token（おすすめ）

1. GitHub → Settings → Developer settings → Personal access tokens → Tokens (classic) → Generate new token
2. スコープは `repo` にチェック
3. 生成されたトークンをコピー
4. `git push` 時のパスワード入力でこのトークンを貼り付ける

#### 方法B: SSH キー

```bash
# SSHキーを生成
ssh-keygen -t ed25519 -C "あなたのメールアドレス"

# 公開鍵を表示してコピー
cat ~/.ssh/id_ed25519.pub
```

GitHub → Settings → SSH and GPG keys → New SSH key に貼り付ける。

SSH を使う場合は、remote URL を変更：

```bash
git remote set-url origin git@github.com:あなたのユーザー名/delivery-api.git
```

## 📖 学習ポイント

| 概念 | 説明 |
|------|------|
| `git init` | ディレクトリを Git リポジトリとして初期化する |
| `git remote` | ローカルリポジトリとリモート（GitHub）を紐づける |
| `git add` | 変更をステージングエリアに追加する |
| `git commit` | ステージングした変更を記録する |
| `git push` | ローカルの記録をリモートにアップロードする |
| `.gitignore` | Git で追跡しないファイル・ディレクトリを指定する |
| ブランチ | コードの変更を分岐させる仕組み。`main` がメインブランチ |

## 🎉 ここまでできたら

GitHub に2つのリポジトリができて、最初のコミットが push された状態！
次の Issue #002 では、バックエンドの Docker Compose 環境構築に進む。

## 🏷️ ラベル

`setup` `git` `github` `初心者向け`
