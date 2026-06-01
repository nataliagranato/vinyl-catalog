"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import Link from "next/link";
import { motion, AnimatePresence } from "framer-motion";
import { ArrowLeft, Trash2, Plus, ChevronDown, ChevronUp, Music } from "lucide-react";
import { useTranslation, LyricsTranslateButton, LyricsSideBySide } from "@/components/LyricsTranslator";
import { VinylResponse, TrackResponse, resolveUploadUrl } from "@/lib/api";
import { VinylForm, VinylFormData } from "@/components/VinylForm";
import { artistToHsl } from "@/lib/vinylColor";
import { useToast } from "@/components/ToastProvider";
import { Button } from "@/components/ui/Button";
import { getTracer } from "@/lib/telemetry";
import { SpanStatusCode } from "@opentelemetry/api";

// ─── Track item ─────────────────────────────────────────────────────────────

function TrackItem({
  track,
  vinylId,
  onUpdated,
  onDeleted,
}: {
  track: TrackResponse;
  vinylId: string;
  onUpdated: (t: TrackResponse) => void;
  onDeleted: (id: string) => void;
}) {
  const [showLyrics, setShowLyrics] = useState(false);
  const [editing, setEditing] = useState(false);
  const [title, setTitle] = useState(track.title);
  const [lyrics, setLyrics] = useState(track.lyrics);
  const { state: translationState, translate, reset: resetTranslation, setLang } = useTranslation(track.lyrics ?? "");
  const { toast } = useToast();

  const save = async () => {
    const res = await fetch(`/api/vinyls/${vinylId}/tracks/${track.id}`, {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ title, position: track.position, lyrics }),
    });
    if (res.ok) {
      const updated = await res.json();
      onUpdated(updated);
      setEditing(false);
      toast("Track updated", "success");
    } else {
      toast("Failed to update track", "error");
    }
  };

  const del = async () => {
    if (!confirm(`Delete "${track.title}"?`)) return;
    const res = await fetch(`/api/vinyls/${vinylId}/tracks/${track.id}`, { method: "DELETE" });
    if (res.ok) {
      onDeleted(track.id);
      toast("Track deleted", "success");
    } else {
      toast("Failed to delete track", "error");
    }
  };

  const canShowLyrics = !!track.lyrics && !editing;

  return (
    <div className="border border-border rounded-lg overflow-hidden">
      <div
        className={`flex items-center gap-3 px-4 py-3 bg-surface ${canShowLyrics ? "cursor-pointer hover:bg-surface/80 transition-colors" : ""}`}
        onClick={() => { if (canShowLyrics) setShowLyrics((v) => !v); }}
      >
        <span className="text-muted text-xs w-5 text-right shrink-0">{track.position}</span>
        {editing ? (
          <input
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            onClick={(e) => e.stopPropagation()}
            className="flex-1 bg-background border border-border rounded px-2 py-1 text-sm text-foreground focus:outline-none focus:border-accent"
          />
        ) : (
          <span className="flex-1 text-sm font-medium text-foreground">{track.title}</span>
        )}

        <div className="flex items-center gap-1 shrink-0" onClick={(e) => e.stopPropagation()}>
          {editing ? (
            <>
              <button onClick={save} className="text-xs text-accent hover:text-accent/80 px-2 py-1">Save</button>
              <button onClick={() => setEditing(false)} className="text-xs text-muted hover:text-foreground px-2 py-1">Cancel</button>
            </>
          ) : (
            <button onClick={() => setEditing(true)} className="text-xs text-muted hover:text-accent px-2 py-1">Edit</button>
          )}
          <button onClick={del} className="text-muted hover:text-red-400 p-1"><Trash2 size={13} /></button>
        </div>

        {canShowLyrics && (
          <>
            <LyricsTranslateButton
              state={translationState}
              onTranslate={translate}
              onReset={resetTranslation}
              onLangChange={setLang}
            />
            <span className="text-muted shrink-0">
              {showLyrics ? <ChevronUp size={14} /> : <ChevronDown size={14} />}
            </span>
          </>
        )}
      </div>

      <AnimatePresence>
        {(showLyrics || editing) && (
          <motion.div
            initial={{ height: 0, opacity: 0 }}
            animate={{ height: "auto", opacity: 1 }}
            exit={{ height: 0, opacity: 0 }}
            transition={{ duration: 0.2 }}
            className="overflow-hidden"
          >
            <div className="px-6 py-4 border-t border-border bg-background/50">
              {editing ? (
                <textarea
                  value={lyrics}
                  onChange={(e) => setLyrics(e.target.value)}
                  rows={6}
                  placeholder="Paste lyrics here…"
                  className="w-full bg-surface border border-border rounded px-3 py-2 text-sm text-foreground placeholder:text-muted focus:outline-none focus:border-accent resize-none"
                />
              ) : (
                <LyricsSideBySide original={track.lyrics ?? ""} state={translationState} />
              )}
            </div>
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
}

// ─── Add track form ──────────────────────────────────────────────────────────

function AddTrackForm({ vinylId, onAdded }: { vinylId: string; onAdded: (t: TrackResponse) => void }) {
  const [open, setOpen] = useState(false);
  const [title, setTitle] = useState("");
  const [position, setPosition] = useState(1);
  const [lyrics, setLyrics] = useState("");
  const { toast } = useToast();

  const submit = async () => {
    if (!title.trim()) return;
    const res = await fetch(`/api/vinyls/${vinylId}/tracks`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ title, position, lyrics }),
    });
    if (res.ok) {
      const t = await res.json();
      onAdded(t);
      setTitle(""); setLyrics(""); setOpen(false);
      toast("Track added", "success");
    } else {
      toast("Failed to add track", "error");
    }
  };

  if (!open) {
    return (
      <button
        onClick={() => setOpen(true)}
        className="flex items-center gap-2 text-sm text-muted hover:text-accent transition-colors border border-dashed border-border rounded-lg px-4 py-3 w-full"
      >
        <Plus size={14} /> Add track
      </button>
    );
  }

  return (
    <div className="border border-accent/30 rounded-lg p-4 flex flex-col gap-3">
      <div className="flex gap-3">
        <input
          type="number"
          value={position}
          onChange={(e) => setPosition(Number(e.target.value))}
          className="w-16 bg-surface border border-border rounded px-2 py-1.5 text-sm text-foreground focus:outline-none focus:border-accent"
          placeholder="#"
          min={1}
        />
        <input
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          className="flex-1 bg-surface border border-border rounded px-3 py-1.5 text-sm text-foreground placeholder:text-muted focus:outline-none focus:border-accent"
          placeholder="Track title"
        />
      </div>
      <textarea
        value={lyrics}
        onChange={(e) => setLyrics(e.target.value)}
        rows={4}
        placeholder="Lyrics (optional)"
        className="bg-surface border border-border rounded px-3 py-2 text-sm text-foreground placeholder:text-muted focus:outline-none focus:border-accent resize-none"
      />
      <div className="flex gap-2">
        <button onClick={submit} className="bg-accent text-background px-4 py-1.5 rounded text-sm font-medium hover:bg-accent/90">Add</button>
        <button onClick={() => setOpen(false)} className="text-sm text-muted hover:text-foreground px-3">Cancel</button>
      </div>
    </div>
  );
}

// ─── Main page ───────────────────────────────────────────────────────────────

export default function VinylDetailPage() {
  const { id } = useParams<{ id: string }>();
  const router = useRouter();
  const { toast } = useToast();
  const [vinyl, setVinyl] = useState<VinylResponse | null>(null);
  const [tracks, setTracks] = useState<TrackResponse[]>([]);
  const [loading, setLoading] = useState(true);
  const [editing, setEditing] = useState(false);
  const [deleting, setDeleting] = useState(false);

  useEffect(() => {
    Promise.all([
      fetch(`/api/vinyls/${id}`).then((r) => r.json()),
      fetch(`/api/vinyls/${id}/tracks`).then((r) => r.json()),
    ])
      .then(([v, t]) => { setVinyl(v); setTracks(Array.isArray(t) ? t : []); })
      .catch(() => toast("Failed to load vinyl", "error"))
      .finally(() => setLoading(false));
  }, [id]);

  const handleUpdate = async (data: VinylFormData) => {
    try {
      const res = await fetch(`/api/vinyls/${id}`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(data),
      });
      if (!res.ok) throw new Error("Update failed");
      const updated = await res.json();
      setVinyl(updated);
      setEditing(false);
      toast("Vinyl updated", "success");
    } catch {
      toast("Failed to update vinyl", "error");
    }
  };

  const handleDelete = async () => {
    if (!confirm("Delete this vinyl permanently?")) return;
    setDeleting(true);
    const res = await fetch(`/api/vinyls/${id}`, { method: "DELETE" });
    if (res.ok) {
      toast("Vinyl deleted", "success");
      router.push("/vinyls");
    } else {
      toast("Failed to delete", "error");
      setDeleting(false);
    }
  };

  if (loading) return (
    <div className="min-h-screen bg-background flex items-center justify-center">
      <div className="w-8 h-8 border-2 border-accent border-t-transparent rounded-full animate-spin" />
    </div>
  );

  if (!vinyl) return (
    <div className="min-h-screen bg-background flex flex-col items-center justify-center gap-4">
      <p className="font-serif text-2xl">Record not found</p>
      <Link href="/vinyls" className="text-accent text-sm">← Back to collection</Link>
    </div>
  );

  return (
    <div className="min-h-screen bg-background px-4 py-8">
      <div className="max-w-5xl mx-auto">
        <Link href="/vinyls" className="inline-flex items-center gap-2 text-muted hover:text-accent transition-colors text-sm mb-8">
          <ArrowLeft size={14} /> Back to collection
        </Link>

        {/* Top: cover + info */}
        <div className="grid md:grid-cols-2 gap-12 items-start mb-12">
          {/* Cover */}
          <motion.div initial={{ opacity: 0, x: -20 }} animate={{ opacity: 1, x: 0 }}>
            <div
              className="aspect-square rounded-2xl relative overflow-hidden flex items-center justify-center"
              style={{ backgroundColor: artistToHsl(vinyl.artist) }}
            >
              {vinyl.cover_url ? (
                <img src={resolveUploadUrl(vinyl.cover_url)} alt={`${vinyl.title} cover`} className="absolute inset-0 w-full h-full object-cover" />
              ) : (
                <>
                  {[20, 35, 50, 65, 78].map((size) => (
                    <div key={size} className="absolute rounded-full border border-black/20"
                      style={{ width: `${size}%`, height: `${size}%`, top: `${(100-size)/2}%`, left: `${(100-size)/2}%` }} />
                  ))}
                  <div className="w-6 h-6 rounded-full bg-background/50 z-10" />
                </>
              )}
            </div>
            {/* Upload cover button */}
            {!editing && (
              <label className="mt-3 flex items-center justify-center gap-2 text-xs text-muted hover:text-accent cursor-pointer transition-colors">
                <input type="file" accept="image/jpeg,image/png,image/webp" className="hidden"
                  onChange={async (e) => {
                    const file = e.target.files?.[0];
                    if (!file) return;
                    const tracer = getTracer();
                    await tracer.startActiveSpan("cover.upload", async (span) => {
                      span.setAttributes({
                        "upload.file_size_bytes": file.size,
                        "upload.file_type": file.type,
                        "vinyl.id": id,
                      });
                      try {
                        const fd = new FormData();
                        fd.append("file", file);
                        const res = await fetch(`/api/vinyls/${id}/cover`, { method: "POST", body: fd });
                        if (res.ok) {
                          const data = await res.json();
                          setVinyl((prev) => prev ? { ...prev, cover_url: data.cover_url } : prev);
                          toast("Cover updated", "success");
                          span.setAttributes({ "upload.status": "success" });
                          span.setStatus({ code: SpanStatusCode.OK });
                        } else {
                          toast("Failed to upload cover", "error");
                          span.setAttributes({ "upload.status": "error", "upload.http_status": res.status });
                          span.setStatus({ code: SpanStatusCode.ERROR, message: "upload failed" });
                        }
                      } catch (err) {
                        span.setStatus({ code: SpanStatusCode.ERROR, message: String(err) });
                        toast("Failed to upload cover", "error");
                      } finally {
                        span.end();
                      }
                    });
                  }}
                />
                Change cover
              </label>
            )}
          </motion.div>

          {/* Info / edit */}
          <motion.div initial={{ opacity: 0, x: 20 }} animate={{ opacity: 1, x: 0 }}>
            {editing ? (
              <>
                <h2 className="font-serif text-2xl font-bold mb-6">Edit Record</h2>
                <VinylForm
                  defaultValues={{ title: vinyl.title, artist: vinyl.artist, year: vinyl.year, genre: vinyl.genre, label: vinyl.label, description: vinyl.description }}
                  onSubmit={handleUpdate}
                  submitLabel="Save changes"
                />
                <button onClick={() => setEditing(false)} className="mt-4 text-sm text-muted hover:text-foreground transition-colors w-full text-center">Cancel</button>
              </>
            ) : (
              <>
                <h1 className="font-serif text-4xl font-bold leading-tight">{vinyl.title}</h1>
                <p className="text-accent text-xl mt-1">{vinyl.artist}</p>

                <dl className="mt-6 grid grid-cols-2 gap-4">
                  {[
                    ["Year", vinyl.year],
                    ["Genre", vinyl.genre || "—"],
                    ["Label", vinyl.label || "—"],
                    ["Added", new Date(vinyl.created_at).toLocaleDateString()],
                  ].map(([k, v]) => (
                    <div key={String(k)}>
                      <dt className="text-xs text-muted uppercase tracking-wider">{k}</dt>
                      <dd className="text-foreground mt-0.5">{v}</dd>
                    </div>
                  ))}
                </dl>

                {vinyl.description && (
                  <p className="mt-6 text-sm text-muted leading-relaxed">{vinyl.description}</p>
                )}

                <div className="flex gap-3 mt-8">
                  <Button onClick={() => setEditing(true)} variant="ghost">Edit</Button>
                  <Button onClick={handleDelete} variant="danger" loading={deleting}>
                    <Trash2 size={14} className="mr-1.5" /> Delete
                  </Button>
                </div>
              </>
            )}
          </motion.div>
        </div>

        {/* Track list */}
        <section>
          <div className="flex items-center gap-2 mb-4">
            <Music size={16} className="text-accent" />
            <h2 className="font-serif text-xl font-bold">Tracklist</h2>
            <span className="text-muted text-sm ml-1">({tracks.length})</span>
          </div>

          <div className="flex flex-col gap-2">
            {tracks.map((t) => (
              <TrackItem
                key={t.id}
                track={t}
                vinylId={id}
                onUpdated={(updated) => setTracks((prev) => prev.map((x) => x.id === updated.id ? updated : x))}
                onDeleted={(tid) => setTracks((prev) => prev.filter((x) => x.id !== tid))}
              />
            ))}
            <AddTrackForm vinylId={id} onAdded={(t) => setTracks((prev) => [...prev, t].sort((a,b) => a.position - b.position))} />
          </div>
        </section>
      </div>
    </div>
  );
}
