"use client";

import { motion } from "framer-motion";
import { Trash2, Heart } from "lucide-react";
import { VinylResponse } from "@/lib/api";

interface Props {
  vinyl: VinylResponse;
  onDelete: (id: string) => void;
  isFavorited: boolean;
  onToggleFavorite: (id: string) => void;
}

export function VinylCard({ vinyl, onDelete, isFavorited, onToggleFavorite }: Props) {
  return (
    <motion.div
      whileHover={{ y: -4 }}
      className="bg-surface border border-border rounded-lg overflow-hidden group"
    >
      <div className="aspect-square bg-muted relative flex items-center justify-center">
        {vinyl.cover_url ? (
          <img
            src={vinyl.cover_url}
            alt={vinyl.title}
            className="w-full h-full object-cover"
          />
        ) : (
          <div className="text-6xl opacity-20">◉</div>
        )}
        <button
          onClick={() => onToggleFavorite(vinyl.id)}
          className="absolute top-2 right-2 p-2 bg-black/50 rounded-full opacity-0 group-hover:opacity-100 transition-opacity"
        >
          <Heart
            size={16}
            className={isFavorited ? "fill-red-500 text-red-500" : "text-white"}
          />
        </button>
      </div>
      <div className="p-3">
        <h3 className="font-serif font-semibold text-foreground truncate">{vinyl.title}</h3>
        <p className="text-sm text-muted truncate">{vinyl.artist}</p>
        <div className="flex items-center justify-between mt-2">
          <span className="text-xs text-muted">{vinyl.year} • {vinyl.genre}</span>
          <button
            onClick={() => onDelete(vinyl.id)}
            className="p-1 text-muted hover:text-red-400 transition-colors"
          >
            <Trash2 size={14} />
          </button>
        </div>
      </div>
    </motion.div>
  );
}
