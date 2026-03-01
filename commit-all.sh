#!/bin/bash
# 1ファイルずつコミットするスクリプト（Backend）
# 使い方: bash commit-all.sh

set -e

echo "🚀 Backend: 1ファイルずつコミット開始"

# .envは除外（APIキーが入ってるため）
git add .gitignore
git commit -m "🙈 .gitignore: 環境変数ファイル・ビルド成果物を除外"

# Models
git add internal/model/user.go
git commit -m "✨ model: Userモデル追加（ドライバー/荷主/管理者ロール）"

git add internal/model/vehicle.go
git commit -m "✨ model: Vehicleモデル追加（車種・積載量管理）"

git add internal/model/trip.go
git commit -m "✨ model: Tripモデル追加（便の登録・出発地/到着地/出発日時）"

git add internal/model/match.go
git commit -m "✨ model: Matchモデル追加（荷主↔ドライバーのマッチング）"

git add internal/model/tracking.go
git commit -m "✨ model: Trackingモデル追加（10分毎の位置情報記録）"

git add internal/model/delivery.go
git commit -m "📦 model: 既存Deliveryモデル（配送先CRUD用）"

# Utilities
git add internal/util/haversine.go
git commit -m "🔧 util: Haversine距離計算（帰り便の逆方向検索用）"

git add internal/util/jwt.go
git commit -m "🔐 util: JWT トークン生成・検証ユーティリティ"

# Middleware
git add internal/middleware/cors.go
git commit -m "🌐 middleware: CORS設定（フロントエンドからのアクセス許可）"

git add internal/middleware/auth.go
git commit -m "🔐 middleware: JWT認証 + 管理者権限チェック"

# Repositories
git add internal/repository/user.go
git commit -m "💾 repository: Userリポジトリ（CRUD + メール検索）"

git add internal/repository/trip.go
git commit -m "💾 repository: Tripリポジトリ（CRUD + 位置検索 + 帰り便検索）"

git add internal/repository/match.go
git commit -m "💾 repository: Matchリポジトリ（マッチング管理）"

git add internal/repository/tracking.go
git commit -m "💾 repository: Trackingリポジトリ（位置情報履歴）"

git add internal/repository/delivery.go
git commit -m "💾 repository: 既存Deliveryリポジトリ"

# Handlers
git add internal/handler/auth.go
git commit -m "🔑 handler: 認証API（ユーザー登録・ログイン・bcrypt + JWT）"

git add internal/handler/trip.go
git commit -m "🚚 handler: 便管理API（CRUD + 帰り便検索エンドポイント）"

git add internal/handler/match.go
git commit -m "🤝 handler: マッチングAPI（リクエスト・承認・拒否・完了）"

git add internal/handler/tracking.go
git commit -m "📍 handler: 位置追跡API（記録・履歴・最新位置取得）"

git add internal/handler/admin.go
git commit -m "👑 handler: 管理者API（統計・ユーザー管理・全便/マッチング一覧）"

git add internal/handler/delivery.go
git commit -m "📦 handler: 既存配送先CRUD API"

# Router
git add internal/router/router.go
git commit -m "🛣️ router: 全ルーティング設定（認証・便・マッチング・管理者）"

# Config files
git add go.mod go.sum
git commit -m "📦 deps: Go依存関係（Gin, GORM, JWT, bcrypt）"

git add Dockerfile
git commit -m "🐳 docker: マルチステージビルド Dockerfile"

git add docker-compose.yml
git commit -m "🐳 docker-compose: MySQL + API サービス構成"

git add .env.example
git commit -m "⚙️ config: 環境変数テンプレート"

git add Makefile
git commit -m "🔨 makefile: ビルド・実行・Docker操作コマンド"

git add .golangci.yml
git commit -m "🔧 lint: golangci-lint 設定"

git add cmd/server/main.go
git commit -m "🚀 server: エントリーポイント（全モデルAutoMigrate）"

# CI/CD
git add .github/workflows/ci.yml
git commit -m "🔄 ci: GitHub Actions（lint → build → test → docker）"

# Documentation
git add README.md
git commit -m "📝 docs: README（セットアップ手順・API概要）"

git add API_DOCUMENTATION.md
git commit -m "📝 docs: API完全ドキュメント（全エンドポイント解説）"

git add IMPLEMENTATION_NOTES.md
git commit -m "📝 docs: 実装ノート（アーキテクチャ解説）"

git add QUICK_START.md
git commit -m "📝 docs: クイックスタートガイド"

git add docs/curriculum/backend-curriculum.md
git commit -m "📚 curriculum: バックエンドカリキュラム（Go + Gin 学習用）"

git add docs/guides/go-basics.md
git commit -m "📚 guide: Go言語基礎解説"

git add docs/issues/001-docker-compose-setup.md
git commit -m "📋 issue: #001 Docker Compose環境構築"

# Cleanup files
git add EXPANSION_SUMMARY.txt FILES_SUMMARY.txt 2>/dev/null && git commit -m "📝 docs: 拡張サマリー" || true

echo "✅ Backend: 全ファイルのコミット完了！"
echo ""
echo "次のコマンドでpush:"
echo "  git push -u origin main"
