"use client";

interface Props {
  id: string;
  label: string;
  type?: string;
  placeholder?: string;
  autoComplete?: string;
  error?: string;
  [key: string]: any;
}

export function Input({ id, label, type = "text", placeholder, autoComplete, error, ...rest }: Props) {
  return (
    <div className="flex flex-col gap-1">
      <label htmlFor={id} className="text-sm font-medium text-foreground">
        {label}
      </label>
      <input
        id={id}
        type={type}
        placeholder={placeholder}
        autoComplete={autoComplete}
        className={`px-3 py-2 bg-surface border border-border rounded text-sm focus:outline-none focus:ring-2 focus:ring-accent ${
          error ? "border-red-400" : ""
        }`}
        {...rest}
      />
      {error && <p className="text-xs text-red-400">{error}</p>}
    </div>
  );
}
