import { NextRequest, NextResponse } from "next/server";
import { cookies } from "next/headers";
import { log } from "@/lib/logger";

export async function POST(
  _req: NextRequest,
  { params }: { params: Promise<{ id: string }> }
) {
  const start = Date.now();
  const { id } = await params;
  const cookieStore = await cookies();
  const token = cookieStore.get("token")?.value ?? "";

  const res = await fetch(`${process.env.API_URL}/api/v1/vinyls/${id}/favorite`, {
    method: "POST",
    headers: { Authorization: `Bearer ${token}` },
  });

  const data = await res.json();
  log(res.ok ? "info" : "error", "/api/vinyls/[id]/favorite", {
    vinyl_id: id,
    status: res.status,
    favorited: data?.favorited,
    duration_ms: Date.now() - start,
  });
  return NextResponse.json(data, { status: res.status });
}
