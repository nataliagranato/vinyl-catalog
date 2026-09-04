import { NextRequest, NextResponse } from "next/server";

export async function POST(req: NextRequest) {
  const { text, targetLang } = await req.json();

  if (!text || !targetLang) {
    return NextResponse.json({ error: "Missing text or targetLang" }, { status: 400 });
  }

  const url = new URL("https://translate.googleapis.com/translate_a/single");
  url.searchParams.set("client", "gtx");
  url.searchParams.set("sl", "auto");
  url.searchParams.set("tl", targetLang);
  url.searchParams.set("dt", "t");
  url.searchParams.set("q", text);

  const res = await fetch(url.toString(), { headers: { "User-Agent": "Mozilla/5.0" } });

  if (!res.ok) {
    return NextResponse.json({ error: "Translation service unavailable" }, { status: 502 });
  }

  const data: [[string, string][]] = await res.json();
  const translated = data[0].map((seg) => seg[0]).join("");
  return NextResponse.json({ translated });
}
