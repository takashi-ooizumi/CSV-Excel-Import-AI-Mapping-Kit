"use client";
import React, { useMemo, useState } from "react";
import MappingUI from "./MappingUI";
import { ORDER_SCHEMA_V1 } from "./schema";

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
  const [forceHasHeader, setForceHasHeader] = useState<boolean | null>(null); // ← 上書き用

  // 環境変数が無い場合のフォールバック（ローカル直叩き）
  const apiBase = useMemo(() => {
    return (
      process.env.NEXT_PUBLIC_API_BASE_URL ??
      (typeof window !== "undefined" ? "http://localhost:8080" : "")
    );
  }, []);

  const onSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!file) return;
    setLoading(true);
    try {
      const fd = new FormData();
      fd.append("file", file);

      // ヘッダ上書きが選ばれていたらクエリで通知（API側対応済/対応予定どちらでもOK）
      const qs = forceHasHeader === null ? "" : `?hasHeader=${forceHasHeader ? "true" : "false"}`;

      const res = await fetch(`${apiBase}/api/imports`, { method: "POST", body: fd });
      if (!res.ok) throw new Error(await res.text());

      const data: Preview = await res.json();

      // UI 上書きがONならプレビュー側も見た目だけ合わせる
      const coerced = forceHasHeader === null ? data : { ...data, hasHeader: forceHasHeader };

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

      {apiBase === "" && (
        <p className="text-sm text-red-600">
          NEXT_PUBLIC_API_BASE_URL が未設定です。ローカルなら <code>http://localhost:8080</code>{" "}
          を想定します。
        </p>
      )}

      <form onSubmit={onSubmit} className="space-y-3">
        <input
          type="file"
          accept=".csv,text/csv,application/vnd.ms-excel"
          onChange={(e) => setFile(e.target.files?.[0] ?? null)}
          className="block"
        />

        <div className="flex items-center gap-3 text-sm">
          <span className="text-gray-600">ヘッダ推定を上書き:</span>
          <label className="flex items-center gap-1">
            <input
              type="radio"
              name="hdr"
              checked={forceHasHeader === true}
              onChange={() => setForceHasHeader(true)}
            />
            ヘッダあり
          </label>
          <label className="flex items-center gap-1">
            <input
              type="radio"
              name="hdr"
              checked={forceHasHeader === false}
              onChange={() => setForceHasHeader(false)}
            />
            ヘッダなし
          </label>
          <label className="flex items-center gap-1">
            <input
              type="radio"
              name="hdr"
              checked={forceHasHeader === null}
              onChange={() => setForceHasHeader(null)}
            />
            自動（推測）
          </label>
        </div>

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
      {preview && preview.headers?.length > 0 && (
        <div className="mt-8">
          <MappingUI
            sourceHeaders={preview.headers}
            rows={preview.sampleRows ?? []}
            schema={ORDER_SCHEMA_V1}
          />
        </div>
      )}
    </main>
  );
}
