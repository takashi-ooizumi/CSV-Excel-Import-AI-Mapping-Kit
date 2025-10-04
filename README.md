# CSV / Excel Import AI Mapping Kit

**CSV/Excel の取り込み → ヘッダー推定 → スキーマへのマッピング → 正規化データのダウンロード**  
フロントは **Next.js**、API は **Go(chi)**。**Supabase(Postgres)** を使った永続化、**Render/Vercel** でのデプロイに対応します。

> 目的：スタートアップ/中小企業で頻出の「バラバラな CSV/Excel を毎回手で揃える」作業を、**スピード導入 × 安定性**で解決するミニ基盤。

---

## ✨ 特徴（MVP）

- **自動マッピング（推測）**：列名の同義語/類似や値サンプルから候補を提示（手動修正可）
- **正規化プレビュー**：サーバ API でマッピングを適用して結果を表示（将来の保存/大規模化に対応）
- **ダウンロード**：正規化データを **CSV / JSON** で保存
- **CORS/ヘルス**：`/readyz` / `/livez` 実装、Render/Vercel でそのまま動作
- **軽量スタック**：Go API（chi + pgx）/ Next.js（App Router）/ Postgres / Docker

---

## 🧱 Tech Stack

- **Web**: Next.js (App Router, TypeScript), Tailwind CSS  
- **API**: Go 1.23+, Chi Router, Distroless で軽量ランタイム  
- **DB**: Supabase(Postgres 15)  
- **Infra**: Docker / docker-compose（ローカル用）、Render(API) / Vercel(Web)  
- **DX**: `make fmt`（Go fmt + Prettier）、`make check`（go vet/test + tsc + prettier:check）、Husky pre-commit（format only）

---

## 🏗️ アーキテクチャ

```mermaid
flowchart LR
  subgraph Web["Next.js (Vercel/Local)"]
    U[Upload UI] --> MUI[Mapping UI]
    MUI --> PV[Preview / Download]
  end

  subgraph API["Go API (Render/Local)"]
    IMP[/POST /api/imports/]:::ep
    APPLY[/POST /api/mappings/apply/]:::ep
    READY[/GET /readyz]:::ep
    LIVE[/GET /livez]:::ep
  end

  subgraph DB["Supabase (Postgres)"]
    MIG[(migrations)]
    TPL[(mapping_templates - 予定)]
  end

  U -->|multipart/form-data| IMP
  MUI -->|JSON rules| APPLY
  IMP --> DB
  APPLY --> DB
  MIG --> DB
  classDef ep fill:#eef,stroke:#99f,color:#000;
```

---

## 🔗 本番デプロイ（参考）

- **Web (Vercel)**: `https://csv-excel-import-ai-mapping-kit.vercel.app/`  
- **API (Render)**: `https://csv-import-kit-api-prod.onrender.com`  
- **DB (Supabase)**: 管理は Supabase ダッシュボードから

> Vercel のプロジェクトに **`NEXT_PUBLIC_API_BASE_URL`**（例：Render の API URL）を設定してください。

---

## 🚀 クイックスタート（ローカル）

### 0) 依存
- Node.js 20+
- Go 1.23+
- Docker（任意：ローカルで DB/コンテナを使う場合）

### 1) 環境変数ファイル（リポジトリ直下）

**`.env`（API ランタイム向け）**
```
API_PORT=8080
ALLOWED_ORIGIN=http://localhost:3000
# ローカルDBの例（compose の db サービスを使う場合）
DATABASE_URL=postgresql://app:app@localhost:5432/app
```

**`.env.migrations`（migrate コンテナ向け）**
```
# 直結(5432) の URL（sslmode は環境に応じて）
DATABASE_URL=postgresql://app:app@db:5432/app?sslmode=disable
```

> Supabase を使う場合：  
> - **アプリ実行時**はプーリング(例: `aws-...:6543`) の接続文字列を `.env` に  
> - **マイグレーション**は直結(5432) の接続文字列を `.env.migrations` に設定してください（sslmode=require 推奨）。

### 2) Docker（compose はリポジトリ直下）

```bash
# 1. ビルド（No cache は任意）
docker compose build --no-cache
# 2. 起動
docker compose up -d
# 3. DB マイグレーション（.env.migrations を使用）
docker compose --env-file .env.migrations run --rm migrator up
```

確認：
```bash
# API ヘルス
curl -i http://localhost:8080/readyz
# Web
open http://localhost:3000
```

停止 / 再開：
```bash
docker compose down           # コンテナ停止+削除
docker compose down -v        # ボリュームも削除（DB初期化）
docker compose stop           # 停止のみ
docker compose up -d          # 再開
```

---

## 🧪 試してみる（テスト CSV）

`/imports` 画面から以下をアップロードして確認できます。

- `orders_comma.csv`（カンマ区切り / ヘッダあり）
- `orders_semicolon.csv`（セミコロン区切り / ヘッダあり）
- `logs_pipe.csv`（パイプ区切り / ヘッダなし）

フロー：**Upload → ヘッダ/区切りの推定確認 → マッピング調整 → Apply(server) → プレビュー → CSV/JSON ダウンロード**

> 推定が誤った場合は、UI の **「ヘッダあり/なしトグル」**で上書きできます。

---

## 🔌 API Endpoints（現状）

- `POST /api/imports`  
  CSV を受け取り、`{ delimiter, hasHeader, headers, sampleRows, countGuessed }` を返します。

- `POST /api/mappings/apply`  
  リクエスト：  
  ```json
  {
    "headers": ["Order ID", "Customer", "..."],
    "rows": [["1001","1","..."], ["1002","1","..."]],
    "rules": { "order_id": "Order ID", "customer_id": "Customer", "...": null }
  }
  ```
  レスポンス：  
  ```json
  {
    "normalizedHeaders": ["order_id","customer_id","product","quantity","unit_price","order_date"],
    "normalizedRows": [["1001","1","Notebook","2","980","2024-06-01"], ...]
  }
  ```

- `GET /readyz` / `GET /livez`  
  ヘルスチェック用。

> CORS は `ALLOWED_ORIGIN` で制御。`OPTIONS` は 204 を返します。

---

## 🛠️ 開発コマンド（最小運用）

ルートの `Makefile` で統一：

```bash
# 整形（Go fmt + goimports + Prettier write）
make fmt

# チェック（Go: vet/test, Web: prettier:check + tsc）
make check

# まとめ（整形 → チェック）
make all
```

> Husky（pre-commit）は **format のみ**（`make fmt`）を実行する軽量運用です。

---

## 🗃️ DB Migrations（docker-compose から）

```bash
# up
docker compose --env-file .env.migrations run --rm migrator up

# 現在バージョン
docker compose --env-file .env.migrations run --rm migrator version

# 1つ戻す
docker compose --env-file .env.migrations run --rm migrator down 1

# 2つ戻す
docker compose --env-file .env.migrations run --rm migrator steps -2
```

> マイグレーションファイルは `db/migrations`（**3桁連番** `000_xxx.up.sql` / `.down.sql`）で管理。  
> 適用後のファイル名変更は履歴の齟齬になるため避けてください。

---

## 🧭 今後のロードマップ

- **テンプレ保存**：`mapping_templates` テーブル（`name`, `schema_key`, `rules`）＋ `/api/templates`（list/create）
- **結果保存**：正規化済みデータ（行ごとのエラー含む）の保存、または Supabase Storage への CSV 保存
- **バリデーション**：必須項目未マッピング・型エラーの検出/表示
- **観測**：リクエストID・処理時間ログ・簡易トレース（OpenTelemetry）

---

## 🐞 Troubleshooting

- **CORS で 403/ERR_FAILED**  
  `ALLOWED_ORIGIN` が Web の URL と一致しているか確認（ローカルは `http://localhost:3000`）。

- **`database driver: unknown driver`**  
  migrator の `DATABASE_URL` のスキームが `postgres://` / `postgresql://` になっているか確認。

- **ESLint が騒がしい**  
  現状は **Prettier + tsc** の最小構成（ESLint は任意）で運用しています。必要になったら段階的に追加してください。

---

## 📄 License

MIT
