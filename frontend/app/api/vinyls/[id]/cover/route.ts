import { NextRequest, NextResponse } from "next/server";
import { cookies } from "next/headers";
import { log } from "@/lib/logger";

export async function POST(
  req: NextRequest,
  { params }: { params: Promise<{ id: string }> }
) {
  const start = Date.now();
  const { id } = await params;
  const cookieStore = await cookies();
  const token = cookieStore.get("token")?.value ?? "";

  const formData = await req.formData();

  const res = await fetch(`${process.env.API_URL}/api/v1/vinyls/${id}/cover`, {
    method: "POST",
    headers: { Authorization: `Bearer ${token}` },
    body: formData,
  });

  const data = await res.json();
  log(res.ok ? "info" : "error", "/api/vinyls/[id]/cover", {
    vinyl_id: id,
    status: res.status,
    duration_ms: Date.now() - start,
  });
  return NextResponse.json(data, { status: res.status });
}
