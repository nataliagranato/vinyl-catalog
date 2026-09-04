import { VinylResponse } from "./api";

export interface Filters {
  search: string;
  genre: string;
  year: string;
}

export function filterVinyls(vinyls: VinylResponse[], filters: Filters): VinylResponse[] {
  return vinyls.filter((vinyl) => {
    const searchMatch =
      !filters.search ||
      vinyl.title.toLowerCase().includes(filters.search.toLowerCase()) ||
      vinyl.artist.toLowerCase().includes(filters.search.toLowerCase());

    const genreMatch = !filters.genre || vinyl.genre === filters.genre;
    const yearMatch = !filters.year || vinyl.year.toString() === filters.year;

    return searchMatch && genreMatch && yearMatch;
  });
}
