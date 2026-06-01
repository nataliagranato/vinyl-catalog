"use client";

import { useState, useEffect } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { motion } from "framer-motion";
import { Plus, LogOut, User } from "lucide-react";
import { VinylResponse } from "@/lib/api";
import { VinylCard } from "@/components/VinylCard";
import { FilterBar } from "@/components/FilterBar";
import { filterVinyls, Filters } from "@/lib/filterVinyls";
import { useToast } from "@/components/ToastProvider";
import { logoutRequest } from "@/lib/auth";
import { getTracer } from "@/lib/telemetry";
import { SpanStatusCode } from "@opentelemetry/api";

type Props = { initialVinyls: VinylResponse[]; initialFavoriteIds?: string[] };

export function VinylListClient({ initialVinyls, initialFavoriteIds = [] }: Props) {
  const [vinyls, setVinyls] = useState(initialVinyls);
  const [filters, setFilters] = useState<Filters>({ search: "", genre: "", year: "" });
  const [deleting, setDeleting] = useState<string | null>(null);
  const [favoriteIds, setFavoriteIds] = useState<Set<string>>(new Set(initialFavoriteIds));
  const { toast } = useToast();
  const router = useRouter();

  useEffect(() => {
    fetch("/api/profile")
      .then((r) => r.json())
      .then((p) => {
        if (Array.isArray(p.favorite_vinyl_ids)) {
          setFavoriteIds(new Set(p.favorite_vinyl_ids));
        }
      })
      .catch(() => {});
  }, []);

  const filtered = filterVinyls(vinyls, filters);

  const handleDelete = async (id: string) => {
    if (!confirm("Delete this vinyl?")) return;
    setDeleting(id);
    try {
      const res = await fetch(`/api/vinyls/${id}`, { method: "DELETE" });
      if (!res.ok) throw new Error("Delete failed");
      setVinyls((prev) => prev.filter((v) => v.id !== id));
      toast("Vinyl deleted", "success");
    } catch {
      toast("Failed to delete vinyl", "error");
    } finally {
      setDeleting(null);
    }
  };

  const handleToggleFavorite = async (id: string) => {
    const tracer = getTracer();
    await tracer.startActiveSpan("favorite.toggle", async (span) => {
      span.setAttributes({ "vinyl.id": id, "favorite.action": favoriteIds.has(id) ? "remove" : "add" });
      try {
        const res = await fetch(`/api/vinyls/${id}/favorite`, { method: "POST" });
        if (res.ok) {
          const data = await res.json();
          setFavoriteIds(new Set(data.favorite_vinyl_ids));
          toast(data.favorited ? "Added to favorites" : "Removed from favorites", "success");
        } else {
          toast("Failed to update favorites", "error");
        }
        span.setStatus({ code: SpanStatusCode.OK });
      } catch (e) {
        span.setStatus({ code: SpanStatusCode.ERROR });
        throw e;
      } finally {
        span.end();
      }
    });
  };

  const handleLogout = async () => {
    await logoutRequest();
    router.push("/login");
  };

  return (
    <div className="min-h-screen bg-background">
      <header className="border-b border-border px-6 py-4 flex items-center justify-between">
        <h1 className="font-serif text-2xl font-bold text-foreground">
          Vinyl Catalog
        </h1>
        <div className="flex items-center gap-3">
          <Link href="/vinyls/new">
            <motion.button
              whileHover={{ scale: 1.03 }}
              whileTap={{ scale: 0.97 }}
              className="flex items-center gap-2 bg-accent text-background px-4 py-2 rounded font-medium text-sm"
            >
              <Plus size={15} />
              Add vinyl
            </motion.button>
          </Link>
          <Link href="/profile/edit" className="p-2 text-muted hover:text-foreground transition-colors" title="Edit profile">
            <User size={16} />
          </Link>
          <button
            onClick={handleLogout}
            className="p-2 text-muted hover:text-foreground transition-colors"
            title="Sign out"
          >
            <LogOut size={16} />
          </button>
        </div>
      </header>

      <main className="max-w-7xl mx-auto px-6 py-8">
        <div className="mb-8">
          <FilterBar filters={filters} onChange={setFilters} vinyls={vinyls} />
        </div>

        <p className="text-muted text-sm mb-6">
          {filtered.length} {filtered.length === 1 ? "record" : "records"}
          {vinyls.length !== filtered.length ? ` of ${vinyls.length}` : ""}
        </p>

        {filtered.length === 0 ? (
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            className="text-center py-24 text-muted"
          >
            <p className="text-4xl mb-4">◉</p>
            <p className="font-serif text-xl">No records found</p>
            <p className="text-sm mt-2">Try adjusting your filters</p>
          </motion.div>
        ) : (
          <motion.div
            layout
            className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 gap-6"
          >
            {filtered.map((vinyl) => (
              <motion.div key={vinyl.id} layout style={{ opacity: deleting === vinyl.id ? 0.4 : 1 }}>
                <VinylCard
                  vinyl={vinyl}
                  onDelete={handleDelete}
                  isFavorited={favoriteIds.has(vinyl.id)}
                  onToggleFavorite={handleToggleFavorite}
                />
              </motion.div>
            ))}
          </motion.div>
        )}
      </main>
    </div>
  );
}
