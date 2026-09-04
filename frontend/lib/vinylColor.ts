// Generate consistent HSL color from artist name
export function artistToHsl(artist: string): string {
  let hash = 0;
  for (let i = 0; i < artist.length; i++) {
    hash = artist.charCodeAt(i) + ((hash << 5) - hash);
  }
  
  const h = Math.abs(hash % 360);
  const s = 60 + (Math.abs(hash) % 20); // 60-80% saturation
  const l = 45 + (Math.abs(hash) % 15); // 45-60% lightness
  
  return `hsl(${h}, ${s}%, ${l}%)`;
}
