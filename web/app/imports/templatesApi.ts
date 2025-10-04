// web/app/imports/templatesApi.ts
export type TemplateItem = {
  id: string;
  name: string;
  schema_key: string;
  rules: Record<string, unknown>;
  description?: string | null;
  created_at: string;
  updated_at: string;
};

export async function listTemplates(apiBase: string): Promise<TemplateItem[]> {
  const res = await fetch(`${apiBase}/api/templates`, { cache: "no-store" });
  if (!res.ok) throw new Error(await res.text());
  return res.json();
}

export async function createTemplate(
  apiBase: string,
  input: { name: string; schema_key: string; rules: Record<string, unknown>; description?: string }
): Promise<{ id: string }> {
  const res = await fetch(`${apiBase}/api/templates`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  });
  if (!res.ok) throw new Error(await res.text());
  return res.json();
}
