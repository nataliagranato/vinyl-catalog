"use client";

import { Filters } from "@/lib/filterVinyls";
import { VinylResponse } from "@/lib/api";

interface Props {
  filters: Filters;
  onChange: (filters: Filters) => void;
  vinyls: VinylResponse[];
}

export function FilterBar({ filters, onChange, vinyls }: Props) {
  const genres = Array.from(new Set(vinyls.map((v) => v.genre))).sort();
  const years = Array.from(new Set(vinyls.map((v) => v.year.toString()))).sort().reverse();

  return (
    <div className="flex flex-wrap gap-3">
      <input
        type="text"
        placeholder="Search records..."
        value={filters.search}
        onChange={(e) => onChange({ ...filters, search: e.target.value })}
        className="px-4 py-2 bg-surface border border-border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-accent"
      />
      <select
        value={filters.genre}
        onChange={(e) => onChange({ ...filters, genre: e.target.value })}
        className="px-4 py-2 bg-surface border border-border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-accent"
      >
        <option value="">All genres</option>
        {genres.map((genre) => (
          <option key={genre} value={genre}>
            {genre}
          </option>
        ))}
      </select>
      <select
        value={filters.year}
        onChange={(e) => onChange({ ...filters, year: e.target.value })}
        className="px-4 py-2 bg-surface border border-border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-accent"
      >
        <option value="">All years</option>
        {years.map((year) => (
          <option key={year} value={year}>
            {year}
          </option>
        ))}
      </select>
    </div>
  );
}
