import { NextRequest, NextResponse } from "next/server";
import { cookies } from "next/headers";

async function getAuthHeader() {
  const cookieStore = await cookies();
  const token = cookieStore.get("token")?.value ?? "";
  return { Authorization: `Bearer ${token}` };
}

export async function GET(
  _req: NextRequest,
  { params }: { params: Promise<{ id: string }> }
) {
  const { id } = await params;
  const headers = await getAuthHeader();
  const res = await fetch(`${process.env.API_URL}/api/v1/vinyls/${id}`, { headers });
  const data = await res.json();
  return NextResponse.json(data, { status: res.status });
}

export async function PUT(
  req: NextRequest,
  { params }: { params: Promise<{ id: string }> }
) {
  const { id } = await params;
  const headers = { ...(await getAuthHeader()), "Content-Type": "application/json" };
  const body = await req.json();
  const res = await fetch(`${process.env.API_URL}/api/v1/vinyls/${id}`, {
    method: "PUT",
    headers,
    body: JSON.stringify(body),
  });
  const data = await res.json();
  return NextResponse.json(data, { status: res.status });
}

export async function DELETE(
  _req: NextRequest,
  { params }: { params: Promise<{ id: string }> }
) {
  const { id } = await params;
  const headers = await getAuthHeader();
  const res = await fetch(`${process.env.API_URL}/api/v1/vinyls/${id}`, {
    method: "DELETE",
    headers,
  });
  if (!res.ok) return NextResponse.json({ error: "Delete failed" }, { status: res.status });
  return NextResponse.json({ ok: true });
}
