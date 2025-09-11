BEGIN;

-- RLS を有効化（FORCEは付けない）
ALTER TABLE public.imports            ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.import_rows_raw    ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.import_mappings    ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.contacts           ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.import_audit_logs  ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.schema_migrations  ENABLE ROW LEVEL SECURITY;

-- 一旦は全許可（後で厳格化する前提の暫定ポリシー）
DO $$
DECLARE t text;
BEGIN
  FOR t IN SELECT unnest(ARRAY[
    'imports',
    'import_rows_raw',
    'import_mappings',
    'contacts',
    'import_audit_logs',
    'schema_migrations'
  ]) LOOP
    EXECUTE format('DROP POLICY IF EXISTS allow_all ON public.%I;', t);
    EXECUTE format('CREATE POLICY allow_all ON public.%I FOR ALL USING (true) WITH CHECK (true);', t);
  END LOOP;
END $$;

COMMIT;
