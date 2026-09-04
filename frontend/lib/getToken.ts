import { cookies } from "next/headers";

export async function getToken(): Promise<string> {
  const cookieStore = await cookies();
  return cookieStore.get("token")?.value ?? "";
}
