export interface VinylResponse {
  id: string;
  title: string;
  artist: string;
  year: number;
  genre: string;
  condition: string;
  cover_url?: string;
  label?: string;
  description?: string;
  created_at: string;
  updated_at: string;
}

export interface CreateVinylRequest {
  title: string;
  artist: string;
  year: number;
  genre: string;
  condition: string;
  label?: string;
  description?: string;
}

export interface UpdateVinylRequest {
  title?: string;
  artist?: string;
  year?: number;
  genre?: string;
  condition?: string;
  label?: string;
  description?: string;
}

export interface Track {
  id: string;
  title: string;
  duration?: string;
  position: number;
}

export interface TrackResponse extends Track {
  lyrics?: string;
}

export interface FavoriteVinylWithTracks extends VinylResponse {
  tracks: TrackResponse[];
  label?: string;
}

export function resolveUploadUrl(url: string): string {
  if (url.startsWith("http://") || url.startsWith("https://")) {
    return url;
  }
  return `${process.env.NEXT_PUBLIC_API_URL ?? 'http://localhost:8787'}/uploads/${url}`;
}

export interface ProfileResponse {
  username: string;
  display_name: string;
  bio: string;
  photo_url: string;
  links: Array<{ title: string; url: string }>;
  preferred_genres: string[];
  favorite_vinyl_ids: string[];
  favorite_vinyls: FavoriteVinylWithTracks[];
}

class VinylsApi {
  constructor(private baseUrl: string, private token: string) {}

  private headers() {
    return {
      "Content-Type": "application/json",
      Authorization: `Bearer ${this.token}`,
    };
  }

  async list(): Promise<VinylResponse[]> {
    const res = await fetch(`${this.baseUrl}/api/v1/vinyls`, {
      headers: this.headers(),
    });
    if (!res.ok) throw new Error("Failed to fetch vinyls");
    return res.json();
  }

  async create(data: CreateVinylRequest): Promise<VinylResponse> {
    const res = await fetch(`${this.baseUrl}/api/v1/vinyls`, {
      method: "POST",
      headers: this.headers(),
      body: JSON.stringify(data),
    });
    if (!res.ok) throw new Error("Failed to create vinyl");
    return res.json();
  }

  async update(id: string, data: UpdateVinylRequest): Promise<VinylResponse> {
    const res = await fetch(`${this.baseUrl}/api/v1/vinyls/${id}`, {
      method: "PUT",
      headers: this.headers(),
      body: JSON.stringify(data),
    });
    if (!res.ok) throw new Error("Failed to update vinyl");
    return res.json();
  }

  async delete(id: string): Promise<void> {
    const res = await fetch(`${this.baseUrl}/api/v1/vinyls/${id}`, {
      method: "DELETE",
      headers: this.headers(),
    });
    if (!res.ok) throw new Error("Failed to delete vinyl");
  }

  async uploadCover(id: string, file: File): Promise<{ cover_url: string }> {
    const formData = new FormData();
    formData.append("cover", file);

    const res = await fetch(`${this.baseUrl}/api/v1/vinyls/${id}/cover`, {
      method: "POST",
      headers: {
        Authorization: `Bearer ${this.token}`,
      },
      body: formData,
    });
    if (!res.ok) throw new Error("Failed to upload cover");
    return res.json();
  }
}

export function buildVinylsApi(baseUrl: string, token: string): VinylsApi {
  return new VinylsApi(baseUrl, token);
}
