"use client";
import React, { useState } from "react";

type Preview = {
  delimiter: string;
  hasHeader: boolean;
  headers: string[];
  sampleRows: string[][];
  countGuessed: number;
};

export default function ImportPage() {
  const [file, setFile] = useState<File | null>(null);
  const [preview, setPreview] = useState<Preview | null>(null);
  const [loading, setLoading] = useState(false);
  const apiBase = process.env.NEXT_PUBLIC_API_BASE_URL!; // Vercel/ローカルで設定済み前提

  const onSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!file) return;
    setLoading(true);
    try {
      const fd = new FormData();
      fd.append("file", file);
      const res = await fetch(`${apiBase}/api/imports`, { method: "POST", body: fd });
      if (!res.ok) throw new Error(await res.text());
      const data: Preview = await res.json();
      setPreview(data);
    } catch (e: any) {
      alert(`Upload failed: ${e.message}`);
    } finally {
      setLoading(false);
    }
  };

  return (
    <main className="p-6 space-y-6">
      <h1 className="text-2xl font-bold">CSV Import Preview</h1>

      <form onSubmit={onSubmit} className="space-y-3">
        <input
          type="file"
          accept=".csv,text/csv,application/vnd.ms-excel"
          onChange={(e) => setFile(e.target.files?.[0] ?? null)}
          className="block"
        />
        <button
          disabled={!file || loading}
          className="px-4 py-2 rounded bg-black text-white disabled:opacity-50"
        >
          {loading ? "Uploading..." : "Upload & Preview"}
        </button>
      </form>

      {preview && (
        <section className="overflow-auto">
          <div className="text-sm text-gray-600 mb-2">
            delimiter: <b>{preview.delimiter}</b> / hasHeader: <b>{String(preview.hasHeader)}</b>
          </div>
          <table className="min-w-[800px] border-collapse">
            {" "}
            {/* ← Tailwind の bracket 構文 */}
            <thead>
              <tr>
                {preview.headers.map((h, i) => (
                  <th key={i} className="border px-2 py-1 text-left bg-gray-50">
                    {h}
                  </th>
                ))}
              </tr>
            </thead>
            <tbody>
              {" "}
              {/* ← tbody を thead の外へ */}
              {preview.sampleRows.map((row, r) => (
                <tr key={r}>
                  {row.map((cell, c) => (
                    <td key={c} className="border px-2 py-1">
                      {cell}
                    </td>
                  ))}
                </tr>
              ))}
            </tbody>
          </table>
        </section>
      )}
    </main>
  );
}
