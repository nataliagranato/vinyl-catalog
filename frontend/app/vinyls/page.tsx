import { buildVinylsApi } from "@/lib/api";
import { getToken } from "@/lib/getToken";
import { VinylListClient } from "./VinylListClient";

export default async function VinylsPage() {
  const token = await getToken();
  const api = buildVinylsApi(process.env.API_URL!, token);
  const vinyls = await api.list().catch(() => []);

  return <VinylListClient initialVinyls={vinyls} />;
}
