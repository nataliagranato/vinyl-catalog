import { NextRequest, NextResponse } from "next/server";
import { cookies } from "next/headers";
import { log } from "@/lib/logger";

export async function POST(req: NextRequest) {
  const start = Date.now();
  const cookieStore = await cookies();
  const token = cookieStore.get("token")?.value ?? "";
  const formData = await req.formData();
  const res = await fetch(`${process.env.API_URL}/api/v1/profile/photo`, {
    method: "POST",
    headers: { Authorization: `Bearer ${token}` },
    body: formData,
  });
  const data = await res.json();
  log(res.ok ? "info" : "error", "/api/profile/photo", {
    status: res.status,
    duration_ms: Date.now() - start,
  });
  return NextResponse.json(data, { status: res.status });
}
