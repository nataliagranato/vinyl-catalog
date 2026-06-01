"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { motion } from "framer-motion";
import { ArrowLeft, Plus, X } from "lucide-react";
import { ProfileResponse, resolveUploadUrl } from "@/lib/api";
import { Button } from "@/components/ui/Button";
import { Input } from "@/components/ui/Input";
import { useToast } from "@/components/ToastProvider";

export default function ProfileEditPage() {
  const router = useRouter();
  const { toast } = useToast();
  const [profile, setProfile] = useState<ProfileResponse | null>(null);
  const [displayName, setDisplayName] = useState("");
  const [bio, setBio] = useState("");
  const [links, setLinks] = useState<string[]>([]);
  const [genres, setGenres] = useState<string[]>([]);
  const [newLink, setNewLink] = useState("");
  const [newGenre, setNewGenre] = useState("");
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    fetch("/api/profile").then((r) => r.json()).then((p: ProfileResponse) => {
      setProfile(p);
      setDisplayName(p.display_name || "");
      setBio(p.bio || "");
      setLinks(p.links || []);
      setGenres(p.preferred_genres || []);
    });
  }, []);

  const save = async () => {
    setSaving(true);
    try {
      const res = await fetch("/api/profile", {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ display_name: displayName, bio, links, preferred_genres: genres }),
      });
      if (!res.ok) throw new Error("Save failed");
      toast("Profile saved", "success");
      router.push("/profile");
    } catch {
      toast("Failed to save profile", "error");
    } finally {
      setSaving(false);
    }
  };

  const uploadPhoto = async (file: File) => {
    const fd = new FormData();
    fd.append("file", file);
    const res = await fetch("/api/profile/photo", { method: "POST", body: fd });
    if (res.ok) {
      const data = await res.json();
      setProfile((prev) => prev ? { ...prev, photo_url: data.photo_url } : prev);
      toast("Photo updated", "success");
    } else {
      toast("Failed to upload photo", "error");
    }
  };

  if (!profile) return (
    <div className="min-h-screen bg-background flex items-center justify-center">
      <div className="w-8 h-8 border-2 border-accent border-t-transparent rounded-full animate-spin" />
    </div>
  );

  return (
    <div className="min-h-screen bg-background px-4 py-8">
      <div className="max-w-lg mx-auto">
        <Link href="/vinyls" className="inline-flex items-center gap-2 text-muted hover:text-accent text-sm mb-8 transition-colors">
          <ArrowLeft size={14} /> Back
        </Link>

        <motion.div initial={{ opacity: 0, y: 16 }} animate={{ opacity: 1, y: 0 }}>
          <h1 className="font-serif text-3xl font-bold mb-8">Edit Profile</h1>

          <div className="bg-surface border border-border rounded-xl p-8 flex flex-col gap-6">

            {/* Photo */}
            <div className="flex items-center gap-4">
              <div className="w-16 h-16 rounded-full overflow-hidden bg-background border border-border flex items-center justify-center flex-shrink-0">
                {profile.photo_url ? (
                  <img src={resolveUploadUrl(profile.photo_url)} alt="" className="w-full h-full object-cover" />
                ) : (
                  <span className="text-2xl text-muted">◉</span>
                )}
              </div>
              <label className="text-sm text-muted hover:text-accent cursor-pointer transition-colors">
                <input type="file" accept="image/jpeg,image/png,image/webp" className="hidden"
                  onChange={(e) => { const f = e.target.files?.[0]; if (f) uploadPhoto(f); }} />
                Change photo
              </label>
            </div>

            {/* Display name */}
            <Input
              id="display_name"
              label="Display Name"
              value={displayName}
              onChange={(e) => setDisplayName(e.target.value)}
              placeholder="Your name"
            />

            {/* Bio */}
            <div className="flex flex-col gap-1.5">
              <label className="text-sm font-medium text-muted">Bio</label>
              <textarea
                value={bio}
                onChange={(e) => setBio(e.target.value)}
                rows={3}
                placeholder="Tell the world about your music taste…"
                className="bg-background border border-border rounded px-4 py-2.5 text-foreground placeholder:text-muted focus:outline-none focus:border-accent transition-colors resize-none text-sm"
              />
            </div>

            {/* Preferred genres */}
            <div className="flex flex-col gap-2">
              <label className="text-sm font-medium text-muted">Preferred Genres</label>
              <div className="flex flex-wrap gap-2 mb-2">
                {genres.map((g) => (
                  <span key={g} className="flex items-center gap-1 px-2.5 py-1 rounded-full bg-background border border-border text-sm text-foreground">
                    {g}
                    <button onClick={() => setGenres(genres.filter((x) => x !== g))} className="text-muted hover:text-red-400 ml-0.5"><X size={11} /></button>
                  </span>
                ))}
              </div>
              <div className="flex gap-2">
                <input
                  value={newGenre}
                  onChange={(e) => setNewGenre(e.target.value)}
                  onKeyDown={(e) => { if (e.key === "Enter" && newGenre.trim()) { setGenres([...genres, newGenre.trim()]); setNewGenre(""); } }}
                  placeholder="Add genre (Enter to add)"
                  className="flex-1 bg-background border border-border rounded px-3 py-1.5 text-sm text-foreground placeholder:text-muted focus:outline-none focus:border-accent transition-colors"
                />
                <button onClick={() => { if (newGenre.trim()) { setGenres([...genres, newGenre.trim()]); setNewGenre(""); } }}
                  className="p-1.5 border border-border rounded hover:border-accent text-muted hover:text-accent transition-colors">
                  <Plus size={16} />
                </button>
              </div>
            </div>

            {/* Links */}
            <div className="flex flex-col gap-2">
              <label className="text-sm font-medium text-muted">Links</label>
              <div className="flex flex-col gap-1.5 mb-2">
                {links.map((l) => (
                  <div key={l} className="flex items-center gap-2">
                    <span className="flex-1 text-sm text-accent truncate">{l}</span>
                    <button onClick={() => setLinks(links.filter((x) => x !== l))} className="text-muted hover:text-red-400"><X size={12} /></button>
                  </div>
                ))}
              </div>
              <div className="flex gap-2">
                <input
                  value={newLink}
                  onChange={(e) => setNewLink(e.target.value)}
                  onKeyDown={(e) => { if (e.key === "Enter" && newLink.trim()) { setLinks([...links, newLink.trim()]); setNewLink(""); } }}
                  placeholder="https://… (Enter to add)"
                  className="flex-1 bg-background border border-border rounded px-3 py-1.5 text-sm text-foreground placeholder:text-muted focus:outline-none focus:border-accent transition-colors"
                />
                <button onClick={() => { if (newLink.trim()) { setLinks([...links, newLink.trim()]); setNewLink(""); } }}
                  className="p-1.5 border border-border rounded hover:border-accent text-muted hover:text-accent transition-colors">
                  <Plus size={16} />
                </button>
              </div>
            </div>

            <Button onClick={save} loading={saving} className="w-full mt-2">Save Profile</Button>
          </div>

          <div className="mt-4 text-center">
            <Link href="/profile" className="text-sm text-muted hover:text-accent transition-colors">
              View public profile →
            </Link>
          </div>
        </motion.div>
      </div>
    </div>
  );
}
