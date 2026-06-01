import { ProfileResponse } from "@/lib/api";
import { ProfilePublicClient } from "./ProfilePublicClient";

async function fetchProfile(): Promise<ProfileResponse> {
  const res = await fetch(`${process.env.API_URL}/api/v1/profile`, { cache: "no-store" });
  if (!res.ok) return { username: "admin", display_name: "", bio: "", photo_url: "", links: [], preferred_genres: [], favorite_vinyl_ids: [], favorite_vinyls: [] as import("@/lib/api").FavoriteVinylWithTracks[] };
  return res.json();
}

export default async function ProfilePage() {
  const profile = await fetchProfile();
  return <ProfilePublicClient profile={profile} />;
}
