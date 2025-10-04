"use client";

import { useEffect, useMemo, useState } from "react";

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
