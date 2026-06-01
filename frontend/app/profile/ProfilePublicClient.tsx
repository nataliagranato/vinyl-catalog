"use client";

import { useState } from "react";
import { motion, AnimatePresence } from "framer-motion";
import Link from "next/link";
import { ExternalLink, Music2, Heart, ChevronDown, ChevronUp, Music } from "lucide-react";
import { ProfileResponse, FavoriteVinylWithTracks, TrackResponse, resolveUploadUrl } from "@/lib/api";
import { useTranslation, LyricsTranslateButton, LyricsSideBySide } from "@/components/LyricsTranslator";

const MAX_FAVORITES = 6;

function TrackRow({ track }: { track: TrackResponse }) {
  const [open, setOpen] = useState(false);
  const hasLyrics = !!track.lyrics;
  const { state, translate, reset, setLang } = useTranslation(track.lyrics ?? "");

  return (
    <div className="border border-border rounded-lg overflow-hidden">
      <div
        className={`flex items-center gap-3 px-4 py-2.5 bg-surface ${hasLyrics ? "cursor-pointer hover:bg-surface/70 transition-colors" : ""}`}
        onClick={() => hasLyrics && setOpen((v) => !v)}
      >
        <span className="text-muted text-xs w-4 text-right shrink-0">{track.position}</span>
        <span className="flex-1 text-sm text-foreground">{track.title}</span>
        {hasLyrics && (
          <>
            <LyricsTranslateButton
              state={state}
              onTranslate={translate}
              onReset={reset}
              onLangChange={setLang}
            />
            <span className="text-muted shrink-0">
              {open ? <ChevronUp size={13} /> : <ChevronDown size={13} />}
            </span>
          </>
        )}
      </div>
      <AnimatePresence>
        {open && (
          <motion.div
            initial={{ height: 0, opacity: 0 }}
            animate={{ height: "auto", opacity: 1 }}
            exit={{ height: 0, opacity: 0 }}
            transition={{ duration: 0.2 }}
            className="overflow-hidden"
          >
            <div className="px-6 py-4 border-t border-border bg-background/50">
              <LyricsSideBySide original={track.lyrics ?? ""} state={state} />
            </div>
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
}

function FavoriteAlbum({ vinyl }: { vinyl: FavoriteVinylWithTracks }) {
  const [expanded, setExpanded] = useState(false);
  const hasTracks = vinyl.tracks?.length > 0;

  return (
    <motion.div
      layout
      className="bg-surface border border-border rounded-xl overflow-hidden"
    >
      {/* Album header */}
      <div
        className={`flex gap-4 p-4 ${hasTracks ? "cursor-pointer hover:bg-surface/70 transition-colors" : ""}`}
        onClick={() => hasTracks && setExpanded((v) => !v)}
      >
        <div className="w-16 h-16 rounded-lg overflow-hidden bg-background border border-border shrink-0">
          {vinyl.cover_url ? (
            <img
              src={resolveUploadUrl(vinyl.cover_url)}
              alt={vinyl.title}
              className="w-full h-full object-cover"
            />
          ) : (
            <div className="w-full h-full flex items-center justify-center text-muted/40 text-xl">◉</div>
          )}
        </div>
        <div className="flex-1 min-w-0">
          <p className="font-serif font-semibold text-foreground truncate">{vinyl.title}</p>
          <p className="text-sm text-accent truncate">{vinyl.artist}</p>
          <p className="text-xs text-muted mt-0.5">{vinyl.year}{vinyl.label ? ` · ${vinyl.label}` : ""}</p>
          {hasTracks && (
            <p className="text-xs text-muted/60 mt-1 flex items-center gap-1">
              <Music size={10} />
              {vinyl.tracks.length} {vinyl.tracks.length === 1 ? "track" : "tracks"}
            </p>
          )}
        </div>
        {hasTracks && (
          <span className="text-muted self-center shrink-0">
            {expanded ? <ChevronUp size={15} /> : <ChevronDown size={15} />}
          </span>
        )}
      </div>

      {/* Tracklist */}
      <AnimatePresence>
        {expanded && (
          <motion.div
            initial={{ height: 0, opacity: 0 }}
            animate={{ height: "auto", opacity: 1 }}
            exit={{ height: 0, opacity: 0 }}
            transition={{ duration: 0.25 }}
            className="overflow-hidden"
          >
            <div className="px-4 pb-4 flex flex-col gap-1.5 border-t border-border pt-3">
              {vinyl.tracks.map((t) => (
                <TrackRow key={t.id} track={t} />
              ))}
            </div>
          </motion.div>
        )}
      </AnimatePresence>
    </motion.div>
  );
}

export function ProfilePublicClient({ profile }: { profile: ProfileResponse }) {
  const displayName = profile.display_name || profile.username;
  const favorites = profile.favorite_vinyls ?? [];
  const visible = favorites.slice(0, MAX_FAVORITES);
  const overflow = favorites.length - MAX_FAVORITES;

  return (
    <div className="min-h-screen bg-background px-4 py-12">
      <div className="max-w-2xl mx-auto">
        <motion.div initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }} transition={{ duration: 0.5 }}>

          {/* Avatar + name */}
          <div className="flex items-center gap-6 mb-8">
            <div className="w-20 h-20 rounded-full overflow-hidden bg-surface border border-border flex-shrink-0 flex items-center justify-center">
              {profile.photo_url ? (
                <img src={resolveUploadUrl(profile.photo_url)} alt={displayName} className="w-full h-full object-cover" />
              ) : (
                <span className="text-3xl text-muted">◉</span>
              )}
            </div>
            <div>
              <h1 className="font-serif text-3xl font-bold text-foreground">{displayName}</h1>
              {profile.bio && <p className="text-muted text-sm mt-1 leading-relaxed">{profile.bio}</p>}
            </div>
          </div>

          {/* Preferred genres */}
          {profile.preferred_genres.length > 0 && (
            <div className="mb-6">
              <div className="flex items-center gap-2 mb-3">
                <Music2 size={14} className="text-accent" />
                <h2 className="text-xs text-muted uppercase tracking-wider">Preferred Genres</h2>
              </div>
              <div className="flex flex-wrap gap-2">
                {profile.preferred_genres.map((g) => (
                  <span key={g} className="px-3 py-1 rounded-full bg-surface border border-border text-sm text-foreground">
                    {g}
                  </span>
                ))}
              </div>
            </div>
          )}

          {/* Links */}
          {profile.links.length > 0 && (
            <div className="mb-8">
              <h2 className="text-xs text-muted uppercase tracking-wider mb-3">Links</h2>
              <div className="flex flex-col gap-2">
                {profile.links.map((link) => (
                  <a
                    key={link}
                    href={link.startsWith("http") ? link : `https://${link}`}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="flex items-center gap-2 text-sm text-accent hover:text-accent/80 transition-colors"
                  >
                    <ExternalLink size={13} />
                    {link}
                  </a>
                ))}
              </div>
            </div>
          )}

          {/* Favorite albums */}
          {visible.length > 0 && (
            <div className="mb-8">
              <div className="flex items-center gap-2 mb-4">
                <Heart size={14} className="text-red-400" fill="currentColor" />
                <h2 className="text-xs text-muted uppercase tracking-wider">Favorites</h2>
              </div>
              <div className="flex flex-col gap-3">
                {visible.map((vinyl) => (
                  <FavoriteAlbum key={vinyl.id} vinyl={vinyl} />
                ))}
              </div>
              {overflow > 0 && (
                <div className="mt-4 text-center">
                  <Link href="/vinyls" className="text-xs text-muted hover:text-accent transition-colors">
                    +{overflow} more · View collection →
                  </Link>
                </div>
              )}
            </div>
          )}

          <div className="pt-6 border-t border-border text-center">
            <Link href="/vinyls" className="text-sm text-muted hover:text-accent transition-colors">
              View collection →
            </Link>
          </div>
        </motion.div>
      </div>
    </div>
  );
}
