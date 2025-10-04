-- 依存拡張（Supabase では既に有効なことが多いが保険として）
create extension if not exists pgcrypto;

-- マッピングテンプレート
create table if not exists public.mapping_templates (
  id          uuid primary key default gen_random_uuid(),
  name        text        not null,
  schema_key  text        not null,          -- 例: 'orders_v1'
  rules       jsonb       not null,          -- 例: {"order_id":"Order ID", ...}
  description text        null,              -- 任意メモ
  created_at  timestamptz not null default now(),
  updated_at  timestamptz not null default now()
);

-- ルールは JSON オブジェクトであることを保証
alter table public.mapping_templates
  add constraint mapping_templates_rules_is_object
    check (jsonb_typeof(rules) = 'object');

-- よく使うキーで索引
create index if not exists idx_mapping_templates_schema_key
  on public.mapping_templates (schema_key);

create index if not exists idx_mapping_templates_created_at
  on public.mapping_templates (created_at desc);

-- 更新時刻の自動更新
create or replace function public.set_updated_at()
returns trigger language plpgsql as $$
begin
  new.updated_at := now();
  return new;
end $$;

drop trigger if exists trg_mapping_templates_updated_at on public.mapping_templates;
create trigger trg_mapping_templates_updated_at
before update on public.mapping_templates
for each row execute function public.set_updated_at();

-- RLS（Supabaseの警告を消しつつ、まずは全許可にしておく）
alter table public.mapping_templates enable row level security;

-- まずは誰でも操作可（デモ/MVP用）。運用時はポリシーを絞ってください。
drop policy if exists "mt_select_all" on public.mapping_templates;
create policy "mt_select_all" on public.mapping_templates
  for select using (true);

drop policy if exists "mt_insert_all" on public.mapping_templates;
create policy "mt_insert_all" on public.mapping_templates
  for insert with check (true);

drop policy if exists "mt_update_all" on public.mapping_templates;
create policy "mt_update_all" on public.mapping_templates
  for update using (true) with check (true);

drop policy if exists "mt_delete_all" on public.mapping_templates;
create policy "mt_delete_all" on public.mapping_templates
  for delete using (true);
