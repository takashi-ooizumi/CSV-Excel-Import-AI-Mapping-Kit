"use client";

import { useEffect, useMemo, useState } from "react";
import { listTemplates, createTemplate, TemplateItem } from "./templatesApi";

type Props = {
  sourceHeaders: string[];
  rows: string[][]; // プレビュー用（先頭N行）
  schema: readonly string[];
  apiBase: string; // ★ 追加：APIベースURL（親から渡す）
};

type Rules = Record<string, string | null>;

function normalizeKey(s: string) {
  return s
    .trim()
    .toLowerCase()
    .replace(/[ \-]+/g, "_");
}

function guessRules(schema: readonly string[], headers: string[]): Rules {
  const normalized = headers.map((h) => ({ raw: h, key: normalizeKey(h) }));
  const rules: Partial<Rules> = {};
  for (const dest of schema) {
    const dk = normalizeKey(dest);
    const hit =
      normalized.find((h) => h.key === dk)?.raw ??
      normalized.find((h) => h.key.includes(dk))?.raw ??
      null;
    rules[dest] = hit;
  }
  return rules as Rules;
}

function toCSV(headers: string[], rows: string[][]) {
  const esc = (s: string) => (/[",\n]/.test(s) ? `"${s.replace(/"/g, '""')}"` : s);
  const lines = [headers.map(esc).join(","), ...rows.map((r) => r.map(esc).join(","))];
  return lines.join("\n");
}

export default function MappingUI({ sourceHeaders, rows, schema, apiBase }: Props) {
  const [rules, setRules] = useState<Rules>(() => guessRules(schema, sourceHeaders));
  const [preview, setPreview] = useState<{ headers: string[]; rows: string[][] } | null>(null);
  const [loading, setLoading] = useState(false);
  const [tpls, setTpls] = useState<TemplateItem[]>([]);
  const [tplLoading, setTplLoading] = useState(false);
  const [tplName, setTplName] = useState("");
  const [tplDesc, setTplDesc] = useState("");

  useEffect(() => {
    setRules(guessRules(schema, sourceHeaders));
    setPreview(null);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [sourceHeaders.join("|")]);

  const options = useMemo(() => ["", ...sourceHeaders], [sourceHeaders]); // "" は未割当

  const onApply = async () => {
    setLoading(true);
    try {
      // サーバ適用：/api/mappings/apply
      const res = await fetch(`${apiBase}/api/mappings/apply`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          headers: sourceHeaders,
          rows,
          rules, // Record<dest, source|null>
        }),
      });
      if (!res.ok) throw new Error(await res.text());
      const data: { normalizedHeaders: string[]; normalizedRows: string[][] } = await res.json();
      setPreview({ headers: data.normalizedHeaders, rows: data.normalizedRows });
    } catch (e) {
      const msg = e instanceof Error ? e.message : "Unknown error";
      alert(`Apply failed: ${msg}`);
    } finally {
      setLoading(false);
    }
  };

  const onDownloadCSV = () => {
    if (!preview) return;
    const csv = toCSV(preview.headers, preview.rows);
    const blob = new Blob([csv], { type: "text/csv;charset=utf-8" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = "normalized.csv";
    a.click();
    URL.revokeObjectURL(url);
  };

  const onDownloadJSON = () => {
    if (!preview) return;
    const objs = preview.rows.map((r) =>
      Object.fromEntries(preview.headers.map((h, i) => [h, r[i] ?? ""]))
    );
    const blob = new Blob([JSON.stringify(objs, null, 2)], {
      type: "application/json",
    });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = "normalized.json";
    a.click();
    URL.revokeObjectURL(url);
  };

  // テンプレ一覧取得
  const fetchTemplates = async () => {
    setTplLoading(true);
    try {
      const data = await listTemplates(apiBase);
      setTpls(data);
    } catch (e) {
      console.error(e);
      alert("Failed to load templates");
    } finally {
      setTplLoading(false);
    }
  };

  // 保存
  const onSaveTemplate = async () => {
    if (!tplName.trim()) {
      alert("Template name is required");
      return;
    }
    // rules から null を除去して送る（好みでOK）
    const cleaned: Record<string, string> = {};
    Object.entries(rules).forEach(([k, v]) => {
      if (v) cleaned[k] = v;
    });

    try {
      const { id } = await createTemplate(apiBase, {
        name: tplName.trim(),
        schema_key: "orders_v1",
        rules: cleaned,
        description: tplDesc || undefined,
      });
      alert(`Saved as template (${id})`);
      setTplName("");
      setTplDesc("");
      fetchTemplates();
    } catch (e) {
      const msg = e instanceof Error ? e.message : "Unknown error";
      alert(`Save failed: ${msg}`);
    }
  };

  // 適用（読込）
  const onLoadTemplate = (id: string) => {
    const t = tpls.find((x) => x.id === id);
    if (!t) return;
    const r = (t.rules ?? {}) as Record<string, string>;
    // schema のキーだけを反映（安全）
    const next: Rules = {} as Rules;
    schema.forEach((k) => {
      next[k] = (r[k] as string) ?? null;
    });
    setRules(next);
  };

  return (
    <div className="space-y-6">
      <div className="rounded-xl border p-4">
        <h3 className="font-semibold mb-3">1) マッピング（標準スキーマ → 入力ヘッダ）</h3>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
          {schema.map((dest) => (
            <label key={dest} className="flex items-center gap-3">
              <div className="w-44 shrink-0 text-sm font-medium">{dest}</div>
              <select
                className="border rounded-md px-2 py-1 w-full"
                value={rules[dest] ?? ""}
                onChange={(e) => setRules((prev) => ({ ...prev, [dest]: e.target.value || null }))}
              >
                {options.map((h) => (
                  <option key={h} value={h}>
                    {h === "" ? "（未割当）" : h}
                  </option>
                ))}
              </select>
            </label>
          ))}
        </div>

        <div className="mt-4 flex gap-2">
          <button
            className="rounded-lg bg-black text-white px-4 py-2 disabled:opacity-50"
            onClick={onApply}
            disabled={loading}
          >
            {loading ? "Applying..." : "Apply mapping (server)"}
          </button>
        </div>
      </div>

      {/* テンプレ保存/読込 */}
      <div className="rounded-xl border p-4">
        <h3 className="font-semibold mb-3">テンプレート</h3>

        <div className="grid md:grid-cols-2 gap-4">
          {/* 保存 */}
          <div className="space-y-2">
            <div className="text-sm font-medium">Save as template</div>
            <input
              className="border rounded-md px-2 py-1 w-full"
              placeholder="Template name"
              value={tplName}
              onChange={(e) => setTplName(e.target.value)}
            />
            <input
              className="border rounded-md px-2 py-1 w-full"
              placeholder="Description (optional)"
              value={tplDesc}
              onChange={(e) => setTplDesc(e.target.value)}
            />
            <button
              className="rounded-lg border px-3 py-2"
              onClick={onSaveTemplate}
            >
              Save template
            </button>
          </div>

          {/* 読込 */}
          <div className="space-y-2">
            <div className="flex items-center gap-2">
              <div className="text-sm font-medium">Load template</div>
              <button
                className="rounded-lg border px-2 py-1 text-sm"
                onClick={fetchTemplates}
                disabled={tplLoading}
                title="Refresh"
              >
                {tplLoading ? "Loading..." : "Refresh"}
              </button>
            </div>
            <select
              className="border rounded-md px-2 py-1 w-full"
              onChange={(e) => e.target.value && onLoadTemplate(e.target.value)}
              defaultValue=""
            >
              <option value="" disabled>
                Select a template
              </option>
              {tpls.map((t) => (
                <option key={t.id} value={t.id}>
                  {t.name} ({new Date(t.created_at).toLocaleString()})
                </option>
              ))}
            </select>
          </div>
        </div>
      </div>

      {preview && (
        <div className="rounded-xl border p-4">
          <h3 className="font-semibold mb-3">2) 正規化プレビュー（server result）</h3>
          <div className="overflow-auto border rounded-md">
            <table className="min-w-full text-sm">
              <thead>
                <tr className="bg-gray-50">
                  {preview.headers.map((h) => (
                    <th key={h} className="px-3 py-2 text-left font-medium">
                      {h}
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {preview.rows.map((row, i) => (
                  <tr key={i} className="border-t">
                    {row.map((cell, j) => (
                      <td key={j} className="px-3 py-2 whitespace-pre">
                        {cell}
                      </td>
                    ))}
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          <div className="mt-4 flex gap-2">
            <button className="rounded-lg border px-4 py-2" onClick={onDownloadCSV}>
              Download CSV
            </button>
            <button className="rounded-lg border px-4 py-2" onClick={onDownloadJSON}>
              Download JSON
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
