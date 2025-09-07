-- extensions（UUID生成に pgcrypto を利用）
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- ============= staging / analysis =============
CREATE TABLE IF NOT EXISTS imports (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  status          TEXT NOT NULL CHECK (status IN ('uploaded','mapping','validating','ready_to_commit','committed','failed')),
  original_filename TEXT NOT NULL,
  row_count       INTEGER NOT NULL DEFAULT 0,
  sample          JSONB,              -- 先頭数行のサンプル
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS import_rows_raw (
  import_id       UUID NOT NULL REFERENCES imports(id) ON DELETE CASCADE,
  row_index       INTEGER NOT NULL,   -- 0-based
  raw_json        JSONB NOT NULL,     -- 1行分の生データ（ヘッダ名: 値）
  detected_errors JSONB,              -- ["invalid_email", "missing_name", ...]
  PRIMARY KEY (import_id, row_index)
);

CREATE TABLE IF NOT EXISTS import_mappings (
  id              BIGSERIAL PRIMARY KEY,
  import_id       UUID NOT NULL REFERENCES imports(id) ON DELETE CASCADE,
  source_col      TEXT NOT NULL,      -- CSVの列名
  target_field    TEXT NOT NULL,      -- 例: "email" / "name" / ...
  confidence      REAL NOT NULL DEFAULT 0,
  is_override     BOOLEAN NOT NULL DEFAULT FALSE, -- UIで手動上書き
  UNIQUE(import_id, source_col)
);

-- ============= core domain (MVP: contacts) =============
CREATE TABLE IF NOT EXISTS contacts (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name            TEXT,
  email           TEXT,
  phone           TEXT,
  address_line1   TEXT,
  city            TEXT,
  postal_code     TEXT,
  country         TEXT,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 代表的インデックス
CREATE INDEX IF NOT EXISTS idx_contacts_email_lower ON contacts ((lower(email)));
CREATE INDEX IF NOT EXISTS idx_import_rows_raw_import ON import_rows_raw(import_id);
CREATE INDEX IF NOT EXISTS idx_import_mappings_import ON import_mappings(import_id);

-- ============= audit =============
CREATE TABLE IF NOT EXISTS import_audit_logs (
  id              BIGSERIAL PRIMARY KEY,
  import_id       UUID REFERENCES imports(id) ON DELETE SET NULL,
  action          TEXT NOT NULL,      -- "commit" / "mapping.update" など
  actor           TEXT,               -- MVPでは任意（APIキー名など）
  metadata        JSONB,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
