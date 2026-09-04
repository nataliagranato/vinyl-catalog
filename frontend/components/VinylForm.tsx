"use client";

import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Input } from "@/components/ui/Input";
import { Button } from "@/components/ui/Button";

const schema = z.object({
  title: z.string().min(1, "Title is required"),
  artist: z.string().min(1, "Artist is required"),
  year: z.number().min(1900).max(new Date().getFullYear() + 1),
  genre: z.string().min(1, "Genre is required"),
  condition: z.string().min(1, "Condition is required"),
  label: z.string().optional(),
  description: z.string().optional(),
});

export type VinylFormData = z.infer<typeof schema>;

interface Props {
  onSubmit: (data: VinylFormData) => Promise<void>;
  submitLabel: string;
  defaultValues?: Partial<VinylFormData>;
}

const GENRES = [
  "Rock", "Jazz", "Blues", "Classical", "Electronic", "Hip Hop",
  "Pop", "R&B", "Soul", "Funk", "Country", "Folk", "Reggae",
  "Punk", "Metal", "Indie", "Alternative", "Other"
];

const CONDITIONS = ["Mint", "Near Mint", "Very Good Plus", "Very Good", "Good", "Fair", "Poor"];

export function VinylForm({ onSubmit, submitLabel, defaultValues }: Props) {
  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<VinylFormData>({
    resolver: zodResolver(schema),
    defaultValues: defaultValues || {
      year: new Date().getFullYear(),
      genre: "",
      condition: "",
    },
  });

  return (
    <form onSubmit={handleSubmit((data) => onSubmit(data))} className="flex flex-col gap-5">
      <Input
        id="title"
        label="Album Title"
        placeholder="Enter album title"
        {...register("title")}
        error={errors.title?.message}
      />
      <Input
        id="artist"
        label="Artist"
        placeholder="Enter artist name"
        {...register("artist")}
        error={errors.artist?.message}
      />
      <div className="grid grid-cols-2 gap-4">
        <Input
          id="year"
          label="Year"
          type="number"
          placeholder="2024"
          {...register("year", { valueAsNumber: true })}
          error={errors.year?.message}
        />
        <div className="flex flex-col gap-1">
          <label htmlFor="genre" className="text-sm font-medium text-foreground">
            Genre
          </label>
          <select
            id="genre"
            {...register("genre")}
            className="px-3 py-2 bg-surface border border-border rounded text-sm focus:outline-none focus:ring-2 focus:ring-accent"
          >
            <option value="">Select genre</option>
            {GENRES.map((genre) => (
              <option key={genre} value={genre}>
                {genre}
              </option>
            ))}
          </select>
          {errors.genre?.message && (
            <p className="text-xs text-red-400">{errors.genre.message}</p>
          )}
        </div>
      </div>
      <div className="flex flex-col gap-1">
        <label htmlFor="condition" className="text-sm font-medium text-foreground">
          Condition
        </label>
        <select
          id="condition"
          {...register("condition")}
          className="px-3 py-2 bg-surface border border-border rounded text-sm focus:outline-none focus:ring-2 focus:ring-accent"
        >
          <option value="">Select condition</option>
          {CONDITIONS.map((condition) => (
            <option key={condition} value={condition}>
              {condition}
            </option>
          ))}
        </select>
        {errors.condition?.message && (
          <p className="text-xs text-red-400">{errors.condition.message}</p>
        )}
      </div>
      <Input
        id="label"
        label="Record Label"
        placeholder="e.g., Columbia, Atlantic, Warner Bros."
        {...register("label")}
        error={errors.label?.message}
      />
      <div className="flex flex-col gap-1">
        <label htmlFor="description" className="text-sm font-medium text-foreground">
          Description
        </label>
        <textarea
          id="description"
          {...register("description")}
          rows={3}
          placeholder="Additional notes about this record..."
          className="px-3 py-2 bg-surface border border-border rounded text-sm focus:outline-none focus:ring-2 focus:ring-accent resize-none"
        />
        {errors.description?.message && (
          <p className="text-xs text-red-400">{errors.description.message}</p>
        )}
      </div>
      <Button type="submit" loading={isSubmitting} className="w-full mt-2">
        {submitLabel}
      </Button>
    </form>
  );
}
