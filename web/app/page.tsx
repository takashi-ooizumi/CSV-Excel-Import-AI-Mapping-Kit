export default function Page() {
  return (
    <main style={{ maxWidth: 720, margin: "40px auto", padding: 16, fontFamily: "system-ui, sans-serif" }}>
      <h1 style={{ fontSize: 28, marginBottom: 8 }}>CSV/Excel Import AI Mapping Kit</h1>
      <p style={{ color: "#666", marginBottom: 16 }}>
        モノレポ初期化済み（Next.js + Go + Postgres）。今後、アップロード→マッピング→プレビュー→コミットの
        ウィザードを実装していきます。
      </p>
      <ul style={{ lineHeight: 1.8 }}>
        <li>
          API health: <a href="http://localhost:8080/healthz" target="_blank">http://localhost:8080/healthz</a>
        </li>
        <li>
          README: <code>README.md</code> を参照（アーキテクチャ・ロードマップ）
        </li>
      </ul>
    </main>
  );
}
