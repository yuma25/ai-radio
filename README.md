# AI Radio

Google Gemini API による対話推論と，ローカルの VOICEVOX Engine による音声合成を組み合わせた，自律型の双方向AIラジオ生放送 Web アプリケーションである．

---

## 1. ドキュメント案内

本システムの詳細なアーキテクチャ仕様，インフラ設計，技術選定の背景，および通信制御の仕組みについては，以下のドキュメントを参照すること．

* **[システムアーキテクチャおよびインフラ設計書](docs/architecture.md):**
  * システム構成図とリクエスト・ストリーミング通信フロー
  * 各コンポーネント（ブラウザ，Nginx，Go，VOICEVOX，Gemini API）の技術的責務
  * Nginx をリバースプロキシ兼静的配信サーバーとして採用している理由と CORS 対策
  * ローカル環境と本番環境における Node.js の動作実態（SPA と SSR の違い）
  * バックエンドおよびフロントエンドのクリーンアーキテクチャ全8層構造の定義
  * Gemini 2.0 Flash の選定背景とステートレス API 仕様

---

## 2. ディレクトリ構成

フロントエンド・バックエンド・ドキュメントが独立して自己完結する構成となっている．

```text
ai-radio/
├── docs/                             # システム仕様・アーキテクチャドキュメント
│   └── architecture.md               # 総合アーキテクチャ解説書
│
├── backend/                          # 【バックエンド】Go 1.24 / クリーンアーキテクチャ全8層
│   ├── cmd/08_server/main.go         # 【第8層: Main】エントリーポイント
│   ├── internal/
│   │   ├── 01_domain_model/          # 【第1層: Domain Model】中核ルール・エンティティ
│   │   ├── 02_domain_repository/     # 【第2層: Domain Repository】インターフェース契約
│   │   ├── 03_usecase/               # 【第3層: UseCase】業務シナリオ
│   │   ├── 04_adapter_gateway/       # 【第4層: Adapter Output】外部通信・マッピング
│   │   ├── 05_adapter_handler/       # 【第5層: Adapter Input】リクエスト受付・DTO変換
│   │   ├── 06_infrastructure/        # 【第6層: Infrastructure】HTTPサーバー・ルーティング
│   │   └── 07_di/                    # 【第7層: DI Container】依存性注入・配線
│   ├── Dockerfile                    # Go マルチステージビルド
│   └── go.mod                        # Go モジュール定義
│
├── frontend/                         # 【フロントエンド】React 19 / TypeScript / Vite / 全8層
│   ├── src/
│   │   ├── 01_domain_model/          # 【第1層: Domain Model】純粋TS型・不変条件
│   │   ├── 02_domain_repository/     # 【第2層: Domain Repository】外部通信契約
│   │   ├── 03_usecase/               # 【第3層: UseCase】業務手順シナリオ
│   │   ├── 04_adapter_gateway/       # 【第4層: Adapter Output】API通信・マッピング
│   │   ├── 05_infrastructure/        # 【第5層: Infrastructure】HTTP通信クライアント基盤
│   │   ├── 06_adapter_presenter/     # 【第6層: Adapter Input】カスタムフック・状態管理
│   │   ├── 07_presentation_view/     # 【第7層: Presentation】受動的UI描画
│   │   └── 08_di/                    # 【第8層: Application Setup】DIコンテナ・配線
│   ├── Dockerfile                    # Node ビルド ➔ Nginx 配信 (マルチステージビルド)
│   ├── nginx.conf                    # Nginx設定（SPAルーティング・APIプロキシ・SSE対応）
│   ├── package.json                  # フロント依存パッケージ定義
│   └── vite.config.ts                # Vite 設定
│
├── docker-compose.yml                # 全サービス一括起動設定 (VOICEVOX / Backend / Frontend)
├── .env                              # プロジェクト環境変数設定
├── .env.example                      # 環境変数テンプレート
└── README.md                         # 本ドキュメント
```

---

## 3. 環境変数設定

ルートディレクトリの `.env.example` をコピーして `.env` を作成し，Google Gemini API キーを設定する．

```bash
cp .env.example .env
```


| 変数名           | 説明                                     | 設定例                   |
| :----------------- | :----------------------------------------- | :------------------------- |
| `PORT`           | バックエンドの待受ポート                 | `8080`                   |
| `ENV`            | 動作環境                                 | `development`            |
| `VOICEVOX_URL`   | VOICEVOX Engine の接続先 URL             | `http://localhost:50021` |
| `GEMINI_API_KEY` | Google Gemini API キー                   | `AIzaSy...`              |
| `GEMINI_MODEL`   | 使用する Gemini モデル                   | `gemini-2.0-flash`       |
| `VITE_API_URL`   | フロントエンドが接続するバックエンド URL | `http://localhost:8080`  |

---

## 4. 起動手順

### 方式A: Docker Compose による一括起動

すべてのサービス（フロントエンド，バックエンド，VOICEVOX）をコンテナ上で一括起動する．

```bash
docker compose up -d
```

起動後，ブラウザで `http://localhost:5173` にアクセスする．

### 方式B: 開発環境での個別起動

開発時にホットリロード等を有効にして個別起動する場合の手順である．

1. **VOICEVOX Engine の起動 (Docker):**
   ```bash
   docker compose up -d voicevox
   ```
2. **バックエンドの起動 (Go):**
   ```bash
   cd backend
   go run cmd/08_server/main.go
   ```
3. **フロントエンドの起動 (Vite / Node.js):**
   ```bash
   cd frontend
   pnpm install
   pnpm run dev
   ```
