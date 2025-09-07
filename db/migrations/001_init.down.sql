DROP TABLE IF EXISTS import_audit_logs;
DROP INDEX IF EXISTS idx_import_mappings_import;
DROP INDEX IF EXISTS idx_import_rows_raw_import;
DROP INDEX IF EXISTS idx_contacts_email_lower;
DROP TABLE IF EXISTS contacts;
DROP TABLE IF EXISTS import_mappings;
DROP TABLE IF EXISTS import_rows_raw;
DROP TABLE IF EXISTS imports;
-- 拡張は落とさない（他用途で使われる可能性があるため）
