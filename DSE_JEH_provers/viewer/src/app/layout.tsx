import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "DSE_JEH Proving Viewer",
  description: "Observational viewer for deterministic DSE_JEH evidence",
};

export default function RootLayout({ children }: LayoutProps<"/">) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}
