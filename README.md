title="README.md"
# CSV/Excel Import AI Mapping Kit

中小〜スタートアップ向けに**「顧客/取引/商品などのCSV・Excelを“10分で取り込める”」**を目指した、**自動マッピング＋検証＋正規化**のミニ基盤です。
**実務の入口になりやすい課題**（他サービスからのデータ移行・他社SaaS連携）を、「スピード導入 × 信頼性 × 可観測性」で解きます。

> **MVPのゴール**：**1万行のCSVをP95 < 30秒でプレビュー**し、列の自動マッピング提案 → バリデーション → **アトミックに本テーブルへコミット**。失敗行は理由を可視化。

---

## 特徴（MVP）
- **自動マッピング**：列名の同義語・類似度・サンプル値から**候補**を提案（手動修正可、confidence付き）
- **バリデーション & 正規化**：email/phone/date/currency/address 等の型検証・整形
- **アトミックコミット**：`/commit` で1トランザクション投入。**監査ログ**と**冪等性キー**で二重適用防止
- **可観測性**：OpenTelemetry対応（トレース/メトリクス/ログ）。**ボトルネックが見える**
- **UI ウィザード**：アップロード → マッピング調整 → 検証結果レビュー → コミット
- **軽量スタック**：Go API（chi + pgx）/ Next.js（App Router）/ Postgres / Docker

> 将来（任意）：LLM連携でマッピング精度補助（OpenAI/ローカル）。まずは**規則ベース**で安定化。

---

## なぜモノレポ？
- `docker compose up` で **web/api/db** を一括起動（デモ・検収・E2Eが楽）
- **Issue/PR/CI** を一本化：共通の品質ゲート（lint/test/build/k6）を適用
- スキーマ変更・API変更が **UIと同じPRで同期** できる（齟齬が減る）
- 将来は `pkg/` で共通ライブラリ（型/バリデーション）も共有可能

---

## アーキテクチャ

```mermaid
flowchart LR
  %% ===== Web (Next.js) =====
  subgraph Web["Next.js (Edge + SSR)"]
    U[Upload UI] --> MUI[Mapping UI]
    MUI --> PV[Preview / Errors]
  end

  %% ===== API (Go) =====
  subgraph Api["Go API"]
    UP[Upload Handler] --> INF[Mapping Inference]
    INF --> VAL[Validation / Normalization]
    VAL --> STG[(staging tables)]
    COMMIT[Commit Handler] --> TX[(atomic tx)]
    TX --> CORE[(core tables)]
    OTel[(traces / metrics / logs)]
  end

  %% ===== DB =====
  DB[(Postgres)]
  DB --- STG
  DB --- CORE

  %% ===== Edges between groups =====
  U -->|multipart/form-data| UP
  MUI -->|POST /imports/{id}/mappings| INF
  PV -->|GET /imports/{id}/preview| VAL
  COMMIT -->|POST /imports/{id}/commit| TX
```
## 起動（Docker）

# リポジトリ直下で
docker compose -f deploy/docker-compose.yml build --no-cache
docker compose -f deploy/docker-compose.yml up -d

# コンテナ削除
docker compose -f deploy/docker-compose.yml down

# コンテナ削除（ボリュームも削除）
docker compose -f deploy/docker-compose.yml down -v

# 停止
docker compose -f deploy/docker-compose.yml stop

# 再開
docker compose -f deploy/docker-compose.yml up -d

# web: http://localhost:3000
# api: http://localhost:8080/healthz

# db migration

# 適用（up）
docker compose --env-file .env.migrations -f deploy/docker-compose.yml run --rm migrator up

# 現在バージョン表示
docker compose --env-file .env.migrations -f deploy/docker-compose.yml run --rm migrator version

# 1つ戻す（ロールバック）
docker compose --env-file .env.migrations -f deploy/docker-compose.yml run --rm migrator down 1

# 2つ戻す
docker compose --env-file .env.migrations -f deploy/docker-compose.yml run --rm migrator steps -2

.env.migrations には Direct(5432) の postgres:// 形式の URL を DATABASE_URL= で記載してください（sslmode=require 推奨）。

アプリ実行時は .env の Pooling(6543) を使用します。

# deploy
web: vercel
api: Render
db: supabase https://supabase.com/dashboard/org/egudvdmoiftfuyzkhdfp
