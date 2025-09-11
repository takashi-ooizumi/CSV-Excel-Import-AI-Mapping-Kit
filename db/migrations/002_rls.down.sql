BEGIN;

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
    EXECUTE format('ALTER TABLE public.%I DISABLE ROW LEVEL SECURITY;', t);
  END LOOP;
END $$;

COMMIT;
