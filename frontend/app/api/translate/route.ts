import { NextRequest, NextResponse } from "next/server";
import { log } from "@/lib/logger";

export async function GET(req: NextRequest) {
  const start = Date.now();
  const { searchParams } = req.nextUrl;
  const q = searchParams.get("q");
  const langpair = searchParams.get("langpair");

  if (!q || !langpair) {
    log("warn", "/api/translate", { error: "missing params", langpair });
    return NextResponse.json({ error: "Missing q or langpair" }, { status: 400 });
  }

  const [sl, tl] = langpair.split("|");
  const url = new URL("https://translate.googleapis.com/translate_a/single");
  url.searchParams.set("client", "gtx");
  url.searchParams.set("sl", sl);
  url.searchParams.set("tl", tl);
  url.searchParams.set("dt", "t");
  url.searchParams.set("q", q);

  const res = await fetch(url.toString(), { headers: { "User-Agent": "Mozilla/5.0" } });

  if (res.status === 429) {
    log("warn", "/api/translate", { langpair, status: 429, error: "quota_exceeded", duration_ms: Date.now() - start });
    return NextResponse.json({ error: "Quota de tradução esgotada. Tente mais tarde." }, { status: 429 });
  }
  if (!res.ok) {
    log("error", "/api/translate", { langpair, status: res.status, error: "upstream_error", duration_ms: Date.now() - start });
    return NextResponse.json({ error: "Serviço de tradução indisponível." }, { status: 502 });
  }

  const data: [[string, string][]] = await res.json();
  const translated = data[0].map((seg) => seg[0]).join("");
  log("info", "/api/translate", { langpair, chars: q.length, status: 200, duration_ms: Date.now() - start });
  return NextResponse.json({ translatedText: translated });
}
