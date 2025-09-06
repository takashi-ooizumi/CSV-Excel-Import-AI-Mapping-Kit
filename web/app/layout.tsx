export const metadata = {
  title: "CSV Import Kit",
  description: "CSV/Excel Import AI Mapping Kit (MVP)",
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="ja">
      <body>{children}</body>
    </html>
  );
}
