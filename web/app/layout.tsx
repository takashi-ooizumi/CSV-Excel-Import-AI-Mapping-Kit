import type { ReactNode } from "react";
import "./globals.css";

export const metadata = {
  title: "CSV Import Kit",
  description: "CSV/Excel Import AI Mapping Kit (MVP)",
};

export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    <html lang="ja">
      <body>{children}</body>
    </html>
  );
}
