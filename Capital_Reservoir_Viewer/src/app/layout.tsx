import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "Capital Reservoir Viewer",
  description: "Read-only live and replay inspection of DSE_JEH Capital Reservoir evidence",
};

export default function RootLayout({ children }: LayoutProps<"/">) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}
