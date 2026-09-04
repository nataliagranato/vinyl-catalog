import { NextRequest, NextResponse } from "next/server";
import { cookies } from "next/headers";

export async function GET() {
  try {
    const base = process.env.API_URL ?? '';
    const res = await fetch(`${base}/api/v1/profile`, {
      headers: { 'User-Agent': 'Mozilla/5.0', Accept: 'application/json' },
    });
    const text = await res.text();
    try {
      return NextResponse.json(JSON.parse(text), { status: res.status });
    } catch {
      return NextResponse.json(
        { error: 'proxy_bad_json', status: res.status, preview: text.slice(0, 500) },
        { status: 500 }
      );
    }
  } catch (e) {
    return NextResponse.json(
      { error: 'proxy_failed', hasApiUrl: !!process.env.API_URL, detail: String(e) },
      { status: 500 }
    );
  }
}

export async function PUT(req: NextRequest) {
  const cookieStore = await cookies();
  const token = cookieStore.get("token")?.value ?? "";
  const body = await req.json();
  const res = await fetch(`${process.env.API_URL}/api/v1/profile`, {
    method: "PUT",
    headers: { "Content-Type": "application/json", Authorization: `Bearer ${token}` },
    body: JSON.stringify(body),
  });
  const data = await res.json();
  return NextResponse.json(data, { status: res.status });
}
