drop trigger if exists trg_mapping_templates_updated_at on public.mapping_templates;
drop function if exists public.set_updated_at();
drop table if exists public.mapping_templates;
