import type { Metadata } from "next";
import "./globals.css";
import { ToastProvider } from "@/components/ToastProvider";
import FaroInit from "@/components/FaroInit";

export const metadata: Metadata = {
  title: "Vinyl Catalog",
  description: "Your personal vinyl record collection",
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body>
        <FaroInit />
        <ToastProvider>{children}</ToastProvider>
      </body>
    </html>
  );
}
