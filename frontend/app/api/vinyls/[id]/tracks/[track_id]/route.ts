import { NextRequest, NextResponse } from "next/server";
import { cookies } from "next/headers";

async function getToken() {
  const cookieStore = await cookies();
  return cookieStore.get("token")?.value ?? "";
}

export async function PUT(
  req: NextRequest,
  { params }: { params: Promise<{ id: string; track_id: string }> }
) {
  const { id, track_id } = await params;
  const token = await getToken();
  const body = await req.json();
  const res = await fetch(`${process.env.API_URL}/api/v1/vinyls/${id}/tracks/${track_id}`, {
    method: "PUT",
    headers: { "Content-Type": "application/json", Authorization: `Bearer ${token}` },
    body: JSON.stringify(body),
  });
  const data = await res.json();
  return NextResponse.json(data, { status: res.status });
}

export async function DELETE(
  _req: NextRequest,
  { params }: { params: Promise<{ id: string; track_id: string }> }
) {
  const { id, track_id } = await params;
  const token = await getToken();
  const res = await fetch(`${process.env.API_URL}/api/v1/vinyls/${id}/tracks/${track_id}`, {
    method: "DELETE",
    headers: { Authorization: `Bearer ${token}` },
  });
  if (!res.ok) return NextResponse.json({ error: "Delete failed" }, { status: res.status });
  return NextResponse.json({ ok: true });
}
