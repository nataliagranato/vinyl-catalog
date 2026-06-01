# Frontend (Next.js) Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build a "Dark Groove" Next.js 15 frontend inside `frontend/` that consumes the vinyl catalog API with full CRUD, client-side filters, JWT auth via httpOnly cookie, and high-impact visual design.

**Architecture:** Next.js 15 App Router with Server and Client Components. JWT stored in httpOnly cookie via Route Handler (never localStorage). Framer Motion for animations, Tailwind for styling, react-hook-form + zod for forms.

**Tech Stack:** Next.js 15, React 19, Tailwind CSS 4, Framer Motion 11, react-hook-form 7, zod 3, @fontsource/playfair-display, @fontsource/inter

---

### Task 1: Scaffold Next.js project

**Files:**
- Create: `frontend/` (entire project scaffold)

**Step 1: Scaffold the project**

Run from repo root:
```bash
cd /Users/natalia.granato/Downloads/vinyl-catalog
npx create-next-app@latest frontend \
  --typescript \
  --tailwind \
  --eslint \
  --app \
  --no-src-dir \
  --import-alias "@/*"
```

**Step 2: Install additional dependencies**

```bash
cd frontend
npm install framer-motion react-hook-form zod @hookform/resolvers
npm install @fontsource/playfair-display @fontsource/inter
```

**Step 3: Verify it starts**

```bash
npm run dev
```
Expected: Server running on http://localhost:3000

**Step 4: Commit**

```bash
cd ..
git add frontend/
git commit -m "feat(frontend): scaffold Next.js 15 project"
```

---

### Task 2: Configure Tailwind, fonts and global styles

**Files:**
- Modify: `frontend/tailwind.config.ts`
- Modify: `frontend/app/globals.css`
- Modify: `frontend/app/layout.tsx`

**Step 1: Update tailwind.config.ts**

Replace the entire file content:
```ts
import type { Config } from "tailwindcss";

const config: Config = {
  content: [
    "./pages/**/*.{js,ts,jsx,tsx,mdx}",
    "./components/**/*.{js,ts,jsx,tsx,mdx}",
    "./app/**/*.{js,ts,jsx,tsx,mdx}",
  ],
  theme: {
    extend: {
      colors: {
        background: "#0A0A0A",
        surface: "#141414",
        border: "#2A2A2A",
        accent: "#D4A017",
        foreground: "#F5F5F0",
        muted: "#6B6B6B",
      },
      fontFamily: {
        serif: ["Playfair Display", "Georgia", "serif"],
        sans: ["Inter", "system-ui", "sans-serif"],
      },
    },
  },
  plugins: [],
};

export default config;
```

**Step 2: Update globals.css**

Replace the entire file:
```css
@import "@fontsource/playfair-display/400.css";
@import "@fontsource/playfair-display/700.css";
@import "@fontsource/inter/400.css";
@import "@fontsource/inter/500.css";
@import "@fontsource/inter/600.css";
@tailwind base;
@tailwind components;
@tailwind utilities;

@layer base {
  body {
    @apply bg-background text-foreground font-sans antialiased;
  }

  * {
    @apply border-border;
  }
}

@layer utilities {
  .scrollbar-hide {
    -ms-overflow-style: none;
    scrollbar-width: none;
  }
  .scrollbar-hide::-webkit-scrollbar {
    display: none;
  }
}
```

**Step 3: Update layout.tsx**

```tsx
import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "Vinyl Catalog",
  description: "Your personal vinyl record collection",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}
```

**Step 4: Verify build has no errors**

```bash
cd frontend && npm run build
```
Expected: Build completes with no type errors.

**Step 5: Commit**

```bash
cd ..
git add frontend/
git commit -m "feat(frontend): configure Tailwind dark theme and fonts"
```

---

### Task 3: API client library

**Files:**
- Create: `frontend/lib/api.ts`
- Create: `frontend/lib/api.test.ts`

**Step 1: Write failing tests**

Create `frontend/lib/api.test.ts`:
```ts
import { buildVinylsApi } from "./api";

describe("buildVinylsApi", () => {
  it("calls list with GET and no body", async () => {
    const mockFetch = jest.fn().mockResolvedValue({
      ok: true,
      json: async () => [{ id: "1", title: "Kind of Blue" }],
    });
    const api = buildVinylsApi("http://test", "token123", mockFetch as any);
    await api.list();
    expect(mockFetch).toHaveBeenCalledWith(
      "http://test/api/v1/vinyls",
      expect.objectContaining({ method: "GET" })
    );
  });

  it("calls create with POST and body", async () => {
    const mockFetch = jest.fn().mockResolvedValue({
      ok: true,
      json: async () => ({ id: "1" }),
    });
    const api = buildVinylsApi("http://test", "token123", mockFetch as any);
    const data = { title: "A", artist: "B", year: 1980, genre: "Rock", label: "X" };
    await api.create(data);
    expect(mockFetch).toHaveBeenCalledWith(
      "http://test/api/v1/vinyls",
      expect.objectContaining({ method: "POST" })
    );
  });

  it("throws on non-ok response", async () => {
    const mockFetch = jest.fn().mockResolvedValue({
      ok: false,
      status: 404,
      json: async () => ({ error: "not found" }),
    });
    const api = buildVinylsApi("http://test", "token123", mockFetch as any);
    await expect(api.get("abc")).rejects.toThrow("not found");
  });
});
```

**Step 2: Install test runner**

```bash
cd frontend
npm install -D jest @types/jest ts-jest jest-environment-jsdom
```

Add to `frontend/package.json` scripts:
```json
"test": "jest"
```

Create `frontend/jest.config.ts`:
```ts
import type { Config } from "jest";

const config: Config = {
  preset: "ts-jest",
  testEnvironment: "node",
  moduleNameMapper: {
    "^@/(.*)$": "<rootDir>/$1",
  },
};

export default config;
```

**Step 3: Run to verify it fails**

```bash
cd frontend && npm test
```
Expected: FAIL — `buildVinylsApi` not found

**Step 4: Implement lib/api.ts**

Create `frontend/lib/api.ts`:
```ts
export type VinylResponse = {
  id: string;
  title: string;
  artist: string;
  year: number;
  genre: string;
  label: string;
  created_at: string;
  updated_at: string;
};

export type CreateVinylInput = {
  title: string;
  artist: string;
  year: number;
  genre: string;
  label: string;
};

export type ApiError = { error: string };

export class HttpError extends Error {
  constructor(public status: number, message: string) {
    super(message);
  }
}

export function buildVinylsApi(
  baseUrl: string,
  token: string,
  fetchFn: typeof fetch = fetch
) {
  const headers = {
    "Content-Type": "application/json",
    Authorization: `Bearer ${token}`,
  };

  async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
    const res = await fetchFn(`${baseUrl}${path}`, { ...options, headers });
    const json = await res.json();
    if (!res.ok) {
      throw new HttpError(res.status, (json as ApiError).error ?? "Request failed");
    }
    return json as T;
  }

  return {
    list: () => request<VinylResponse[]>("/api/v1/vinyls", { method: "GET" }),
    get: (id: string) => request<VinylResponse>(`/api/v1/vinyls/${id}`, { method: "GET" }),
    create: (data: CreateVinylInput) =>
      request<VinylResponse>("/api/v1/vinyls", {
        method: "POST",
        body: JSON.stringify(data),
      }),
    update: (id: string, data: CreateVinylInput) =>
      request<VinylResponse>(`/api/v1/vinyls/${id}`, {
        method: "PUT",
        body: JSON.stringify(data),
      }),
    remove: (id: string) =>
      request<void>(`/api/v1/vinyls/${id}`, { method: "DELETE" }),
  };
}
```

**Step 5: Run tests to verify they pass**

```bash
cd frontend && npm test
```
Expected: 3 tests PASS

**Step 6: Commit**

```bash
cd ..
git add frontend/lib/ frontend/jest.config.ts frontend/package.json
git commit -m "feat(frontend): add API client with tests"
```

---

### Task 4: Auth — Route Handler, middleware and context

**Files:**
- Create: `frontend/app/api/auth/login/route.ts`
- Create: `frontend/app/api/auth/logout/route.ts`
- Create: `frontend/middleware.ts`
- Create: `frontend/lib/auth.ts`
- Create: `frontend/components/AuthProvider.tsx`

**Step 1: Create .env.local**

Create `frontend/.env.local`:
```
NEXT_PUBLIC_API_URL=http://localhost:8080
API_URL=http://localhost:8080
```

**Step 2: Create login Route Handler**

Create `frontend/app/api/auth/login/route.ts`:
```ts
import { NextRequest, NextResponse } from "next/server";

export async function POST(req: NextRequest) {
  const { username, password } = await req.json();

  const res = await fetch(`${process.env.API_URL}/api/v1/auth/login`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ username, password }),
  });

  const data = await res.json();

  if (!res.ok) {
    return NextResponse.json({ error: data.error ?? "Login failed" }, { status: res.status });
  }

  const response = NextResponse.json({ ok: true });
  response.cookies.set("token", data.token, {
    httpOnly: true,
    secure: process.env.NODE_ENV === "production",
    sameSite: "lax",
    path: "/",
    maxAge: 60 * 60 * 24, // 24 hours
  });

  return response;
}
```

**Step 3: Create logout Route Handler**

Create `frontend/app/api/auth/logout/route.ts`:
```ts
import { NextResponse } from "next/server";

export async function POST() {
  const response = NextResponse.json({ ok: true });
  response.cookies.delete("token");
  return response;
}
```

**Step 4: Create middleware**

Create `frontend/middleware.ts`:
```ts
import { NextRequest, NextResponse } from "next/server";

const PUBLIC_PATHS = ["/login", "/api/auth/login"];

export function middleware(req: NextRequest) {
  const { pathname } = req.nextUrl;
  const isPublic = PUBLIC_PATHS.some((p) => pathname.startsWith(p));
  const token = req.cookies.get("token")?.value;

  if (!isPublic && !token) {
    return NextResponse.redirect(new URL("/login", req.url));
  }

  if (pathname === "/login" && token) {
    return NextResponse.redirect(new URL("/vinyls", req.url));
  }

  return NextResponse.next();
}

export const config = {
  matcher: ["/((?!_next/static|_next/image|favicon.ico).*)"],
};
```

**Step 5: Create auth helper**

Create `frontend/lib/auth.ts`:
```ts
export async function loginRequest(username: string, password: string) {
  const res = await fetch("/api/auth/login", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ username, password }),
  });
  const data = await res.json();
  if (!res.ok) throw new Error(data.error ?? "Login failed");
  return data;
}

export async function logoutRequest() {
  await fetch("/api/auth/logout", { method: "POST" });
}
```

**Step 6: Create a server helper to get token from cookies**

Create `frontend/lib/getToken.ts`:
```ts
import { cookies } from "next/headers";

export async function getToken(): Promise<string> {
  const cookieStore = await cookies();
  return cookieStore.get("token")?.value ?? "";
}
```

**Step 7: Verify TypeScript compiles**

```bash
cd frontend && npx tsc --noEmit
```
Expected: No errors.

**Step 8: Commit**

```bash
cd ..
git add frontend/app/api/ frontend/middleware.ts frontend/lib/auth.ts frontend/lib/getToken.ts frontend/.env.local
git commit -m "feat(frontend): add JWT auth via httpOnly cookie and middleware"
```

---

### Task 5: Root redirect and login page

**Files:**
- Modify: `frontend/app/page.tsx`
- Create: `frontend/app/login/page.tsx`
- Create: `frontend/components/ui/Button.tsx`
- Create: `frontend/components/ui/Input.tsx`

**Step 1: Root redirect**

Replace `frontend/app/page.tsx`:
```tsx
import { redirect } from "next/navigation";

export default function Home() {
  redirect("/vinyls");
}
```

**Step 2: Create Button component**

Create `frontend/components/ui/Button.tsx`:
```tsx
import { ButtonHTMLAttributes, forwardRef } from "react";

type Props = ButtonHTMLAttributes<HTMLButtonElement> & {
  variant?: "primary" | "ghost" | "danger";
  loading?: boolean;
};

export const Button = forwardRef<HTMLButtonElement, Props>(
  ({ variant = "primary", loading, children, className = "", disabled, ...props }, ref) => {
    const base = "inline-flex items-center justify-center font-sans font-medium rounded transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-accent/50 disabled:opacity-50 disabled:cursor-not-allowed";
    const variants = {
      primary: "bg-accent text-background px-6 py-2.5 hover:bg-accent/90 active:scale-95",
      ghost: "border border-border text-foreground px-6 py-2.5 hover:border-accent hover:text-accent",
      danger: "border border-red-800 text-red-400 px-6 py-2.5 hover:bg-red-900/20",
    };
    return (
      <button
        ref={ref}
        disabled={disabled || loading}
        className={`${base} ${variants[variant]} ${className}`}
        {...props}
      >
        {loading ? (
          <span className="w-4 h-4 border-2 border-current border-t-transparent rounded-full animate-spin mr-2" />
        ) : null}
        {children}
      </button>
    );
  }
);

Button.displayName = "Button";
```

**Step 3: Create Input component**

Create `frontend/components/ui/Input.tsx`:
```tsx
import { InputHTMLAttributes, forwardRef } from "react";

type Props = InputHTMLAttributes<HTMLInputElement> & {
  label?: string;
  error?: string;
};

export const Input = forwardRef<HTMLInputElement, Props>(
  ({ label, error, className = "", id, ...props }, ref) => (
    <div className="flex flex-col gap-1.5">
      {label && (
        <label htmlFor={id} className="text-sm font-medium text-muted">
          {label}
        </label>
      )}
      <input
        ref={ref}
        id={id}
        className={`bg-surface border ${error ? "border-red-700" : "border-border"} rounded px-4 py-2.5 text-foreground placeholder:text-muted focus:outline-none focus:border-accent transition-colors ${className}`}
        {...props}
      />
      {error && <p className="text-xs text-red-400">{error}</p>}
    </div>
  )
);

Input.displayName = "Input";
```

**Step 4: Create login page**

Create `frontend/app/login/page.tsx`:
```tsx
"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { motion } from "framer-motion";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { loginRequest } from "@/lib/auth";
import { Button } from "@/components/ui/Button";
import { Input } from "@/components/ui/Input";

const schema = z.object({
  username: z.string().min(1, "Username is required"),
  password: z.string().min(1, "Password is required"),
});

type FormData = z.infer<typeof schema>;

export default function LoginPage() {
  const router = useRouter();
  const [serverError, setServerError] = useState("");
  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<FormData>({ resolver: zodResolver(schema) });

  const onSubmit = async (data: FormData) => {
    setServerError("");
    try {
      await loginRequest(data.username, data.password);
      router.push("/vinyls");
    } catch (err) {
      setServerError(err instanceof Error ? err.message : "Login failed");
    }
  };

  return (
    <main className="min-h-screen flex items-center justify-center bg-background px-4">
      {/* Vinyl groove background */}
      <div
        className="absolute inset-0 opacity-[0.03] pointer-events-none"
        style={{
          backgroundImage: `repeating-radial-gradient(circle at 50% 50%, transparent 0, transparent 20px, #D4A017 20px, #D4A017 21px)`,
        }}
      />

      <motion.div
        initial={{ opacity: 0, y: 24 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.5, ease: "easeOut" }}
        className="relative w-full max-w-sm"
      >
        {/* Logo */}
        <div className="text-center mb-10">
          <div className="w-16 h-16 rounded-full bg-accent/10 border-2 border-accent/30 flex items-center justify-center mx-auto mb-4">
            <span className="text-2xl">◉</span>
          </div>
          <h1 className="font-serif text-3xl font-bold text-foreground">Vinyl Catalog</h1>
          <p className="text-muted text-sm mt-1">Sign in to your collection</p>
        </div>

        {/* Form */}
        <form
          onSubmit={handleSubmit(onSubmit)}
          className="bg-surface border border-border rounded-xl p-8 flex flex-col gap-5"
        >
          <Input
            id="username"
            label="Username"
            placeholder="admin"
            autoComplete="username"
            {...register("username")}
            error={errors.username?.message}
          />
          <Input
            id="password"
            label="Password"
            type="password"
            placeholder="••••••••"
            autoComplete="current-password"
            {...register("password")}
            error={errors.password?.message}
          />

          {serverError && (
            <p className="text-sm text-red-400 text-center">{serverError}</p>
          )}

          <Button type="submit" loading={isSubmitting} className="w-full mt-2">
            Sign in
          </Button>
        </form>
      </motion.div>
    </main>
  );
}
```

**Step 5: Verify TypeScript**

```bash
cd frontend && npx tsc --noEmit
```
Expected: No errors.

**Step 6: Manual test**

Start dev server (`npm run dev`), navigate to http://localhost:3000 — should redirect to `/login`. Login form should render with amber accents on dark background.

**Step 7: Commit**

```bash
cd ..
git add frontend/app/ frontend/components/ui/
git commit -m "feat(frontend): add login page with Framer Motion and form validation"
```

---

### Task 6: VinylCard component

**Files:**
- Create: `frontend/lib/vinylColor.ts`
- Create: `frontend/lib/vinylColor.test.ts`
- Create: `frontend/components/VinylCard.tsx`

**Step 1: Write failing test for color utility**

Create `frontend/lib/vinylColor.test.ts`:
```ts
import { artistToHsl } from "./vinylColor";

describe("artistToHsl", () => {
  it("returns a valid HSL string", () => {
    const color = artistToHsl("Miles Davis");
    expect(color).toMatch(/^hsl\(\d+, \d+%, \d+%\)$/);
  });

  it("returns the same color for the same artist", () => {
    expect(artistToHsl("Coltrane")).toBe(artistToHsl("Coltrane"));
  });

  it("returns different colors for different artists", () => {
    expect(artistToHsl("Miles Davis")).not.toBe(artistToHsl("John Coltrane"));
  });
});
```

**Step 2: Run to verify it fails**

```bash
cd frontend && npm test -- --testPathPattern=vinylColor
```
Expected: FAIL

**Step 3: Implement vinylColor.ts**

Create `frontend/lib/vinylColor.ts`:
```ts
export function artistToHsl(artist: string): string {
  let hash = 0;
  for (let i = 0; i < artist.length; i++) {
    hash = artist.charCodeAt(i) + ((hash << 5) - hash);
  }
  const hue = Math.abs(hash) % 360;
  return `hsl(${hue}, 35%, 22%)`;
}
```

**Step 4: Run tests**

```bash
cd frontend && npm test -- --testPathPattern=vinylColor
```
Expected: 3 tests PASS

**Step 5: Create VinylCard component**

Create `frontend/components/VinylCard.tsx`:
```tsx
"use client";

import Link from "next/link";
import { motion } from "framer-motion";
import { Pencil, Trash2 } from "lucide-react";
import { VinylResponse } from "@/lib/api";
import { artistToHsl } from "@/lib/vinylColor";

type Props = {
  vinyl: VinylResponse;
  onDelete?: (id: string) => void;
};

export function VinylCard({ vinyl, onDelete }: Props) {
  const bgColor = artistToHsl(vinyl.artist);

  return (
    <motion.div
      initial={{ opacity: 0, scale: 0.95 }}
      animate={{ opacity: 1, scale: 1 }}
      transition={{ duration: 0.3 }}
      whileHover={{ y: -4 }}
      className="group relative"
    >
      <Link href={`/vinyls/${vinyl.id}`} className="block">
        {/* Cover */}
        <div
          className="aspect-square rounded-lg flex flex-col items-center justify-center p-4 mb-3 relative overflow-hidden"
          style={{ backgroundColor: bgColor }}
        >
          {/* Vinyl groove rings */}
          <div className="absolute inset-0 opacity-20">
            {[20, 35, 50, 65].map((size) => (
              <div
                key={size}
                className="absolute rounded-full border border-black/30"
                style={{
                  width: `${size}%`,
                  height: `${size}%`,
                  top: `${(100 - size) / 2}%`,
                  left: `${(100 - size) / 2}%`,
                }}
              />
            ))}
          </div>
          {/* Center hole */}
          <div className="w-4 h-4 rounded-full bg-background/60 z-10" />

          {/* Hover shadow glow */}
          <div className="absolute inset-0 rounded-lg ring-0 group-hover:ring-2 group-hover:ring-accent/50 transition-all duration-200" />
        </div>

        {/* Info */}
        <div className="px-1">
          <p className="font-serif text-foreground font-semibold truncate leading-tight">
            {vinyl.title}
          </p>
          <p className="text-muted text-sm truncate">{vinyl.artist}</p>
          <p className="text-muted/60 text-xs mt-0.5">
            {vinyl.year}{vinyl.label ? ` · ${vinyl.label}` : ""}
          </p>
        </div>
      </Link>

      {/* Action buttons (hover) */}
      {onDelete && (
        <div className="absolute top-2 right-2 flex gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
          <Link
            href={`/vinyls/${vinyl.id}`}
            onClick={(e) => e.stopPropagation()}
            className="p-1.5 bg-surface/90 rounded hover:text-accent transition-colors"
          >
            <Pencil size={13} />
          </Link>
          <button
            onClick={(e) => { e.preventDefault(); onDelete(vinyl.id); }}
            className="p-1.5 bg-surface/90 rounded hover:text-red-400 transition-colors"
          >
            <Trash2 size={13} />
          </button>
        </div>
      )}
    </motion.div>
  );
}
```

**Step 6: Install lucide-react**

```bash
cd frontend && npm install lucide-react
```

**Step 7: Commit**

```bash
cd ..
git add frontend/components/VinylCard.tsx frontend/lib/vinylColor.ts frontend/lib/vinylColor.test.ts
git commit -m "feat(frontend): add VinylCard with deterministic artist colors"
```

---

### Task 7: FilterBar component

**Files:**
- Create: `frontend/lib/filterVinyls.ts`
- Create: `frontend/lib/filterVinyls.test.ts`
- Create: `frontend/components/FilterBar.tsx`

**Step 1: Write failing tests**

Create `frontend/lib/filterVinyls.test.ts`:
```ts
import { filterVinyls } from "./filterVinyls";

const vinyls = [
  { id: "1", title: "Kind of Blue", artist: "Miles Davis", year: 1959, genre: "Jazz", label: "Columbia", created_at: "", updated_at: "" },
  { id: "2", title: "A Love Supreme", artist: "John Coltrane", year: 1965, genre: "Jazz", label: "Impulse!", created_at: "", updated_at: "" },
  { id: "3", title: "Purple Rain", artist: "Prince", year: 1984, genre: "Pop", label: "Warner", created_at: "", updated_at: "" },
];

describe("filterVinyls", () => {
  it("returns all when no filters", () => {
    expect(filterVinyls(vinyls, { search: "", genre: "", year: "" })).toHaveLength(3);
  });

  it("filters by search (title)", () => {
    const result = filterVinyls(vinyls, { search: "blue", genre: "", year: "" });
    expect(result).toHaveLength(1);
    expect(result[0].title).toBe("Kind of Blue");
  });

  it("filters by search (artist)", () => {
    const result = filterVinyls(vinyls, { search: "coltrane", genre: "", year: "" });
    expect(result).toHaveLength(1);
    expect(result[0].artist).toBe("John Coltrane");
  });

  it("filters by genre", () => {
    const result = filterVinyls(vinyls, { search: "", genre: "Jazz", year: "" });
    expect(result).toHaveLength(2);
  });

  it("filters by year", () => {
    const result = filterVinyls(vinyls, { search: "", genre: "", year: "1984" });
    expect(result).toHaveLength(1);
    expect(result[0].title).toBe("Purple Rain");
  });

  it("combines filters", () => {
    const result = filterVinyls(vinyls, { search: "love", genre: "Jazz", year: "" });
    expect(result).toHaveLength(1);
  });
});
```

**Step 2: Run to verify it fails**

```bash
cd frontend && npm test -- --testPathPattern=filterVinyls
```
Expected: FAIL

**Step 3: Implement filterVinyls.ts**

Create `frontend/lib/filterVinyls.ts`:
```ts
import { VinylResponse } from "./api";

export type Filters = {
  search: string;
  genre: string;
  year: string;
};

export function filterVinyls(vinyls: VinylResponse[], filters: Filters): VinylResponse[] {
  const search = filters.search.toLowerCase();
  return vinyls.filter((v) => {
    if (search && !v.title.toLowerCase().includes(search) && !v.artist.toLowerCase().includes(search)) {
      return false;
    }
    if (filters.genre && v.genre !== filters.genre) return false;
    if (filters.year && String(v.year) !== filters.year) return false;
    return true;
  });
}
```

**Step 4: Run tests**

```bash
cd frontend && npm test -- --testPathPattern=filterVinyls
```
Expected: 6 tests PASS

**Step 5: Create FilterBar component**

Create `frontend/components/FilterBar.tsx`:
```tsx
"use client";

import { VinylResponse } from "@/lib/api";
import { Filters } from "@/lib/filterVinyls";
import { Search } from "lucide-react";

type Props = {
  filters: Filters;
  onChange: (f: Filters) => void;
  vinyls: VinylResponse[];
};

export function FilterBar({ filters, onChange, vinyls }: Props) {
  const genres = Array.from(new Set(vinyls.map((v) => v.genre).filter(Boolean))).sort();
  const years = Array.from(new Set(vinyls.map((v) => String(v.year)))).sort((a, b) => Number(b) - Number(a));

  const select = "bg-surface border border-border rounded px-3 py-2 text-sm text-foreground focus:outline-none focus:border-accent transition-colors";

  return (
    <div className="flex flex-wrap gap-3 items-center">
      <div className="relative flex-1 min-w-48">
        <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-muted" />
        <input
          type="text"
          placeholder="Search title or artist…"
          value={filters.search}
          onChange={(e) => onChange({ ...filters, search: e.target.value })}
          className="w-full bg-surface border border-border rounded px-3 py-2 pl-8 text-sm text-foreground placeholder:text-muted focus:outline-none focus:border-accent transition-colors"
        />
      </div>

      <select
        value={filters.genre}
        onChange={(e) => onChange({ ...filters, genre: e.target.value })}
        className={select}
      >
        <option value="">All genres</option>
        {genres.map((g) => <option key={g} value={g}>{g}</option>)}
      </select>

      <select
        value={filters.year}
        onChange={(e) => onChange({ ...filters, year: e.target.value })}
        className={select}
      >
        <option value="">All years</option>
        {years.map((y) => <option key={y} value={y}>{y}</option>)}
      </select>

      {(filters.search || filters.genre || filters.year) && (
        <button
          onClick={() => onChange({ search: "", genre: "", year: "" })}
          className="text-xs text-muted hover:text-accent transition-colors"
        >
          Clear
        </button>
      )}
    </div>
  );
}
```

**Step 6: Commit**

```bash
cd ..
git add frontend/components/FilterBar.tsx frontend/lib/filterVinyls.ts frontend/lib/filterVinyls.test.ts
git commit -m "feat(frontend): add FilterBar with client-side filtering (tested)"
```

---

### Task 8: Toast notification component

**Files:**
- Create: `frontend/components/Toast.tsx`
- Create: `frontend/components/ToastProvider.tsx`

**Step 1: Create Toast component**

Create `frontend/components/Toast.tsx`:
```tsx
"use client";

import { motion, AnimatePresence } from "framer-motion";
import { X, CheckCircle, AlertCircle } from "lucide-react";

export type ToastData = {
  id: string;
  message: string;
  type: "success" | "error";
};

type Props = {
  toasts: ToastData[];
  onDismiss: (id: string) => void;
};

export function ToastContainer({ toasts, onDismiss }: Props) {
  return (
    <div className="fixed bottom-6 right-6 flex flex-col gap-2 z-50">
      <AnimatePresence>
        {toasts.map((t) => (
          <motion.div
            key={t.id}
            initial={{ opacity: 0, x: 40, scale: 0.95 }}
            animate={{ opacity: 1, x: 0, scale: 1 }}
            exit={{ opacity: 0, x: 40, scale: 0.95 }}
            transition={{ duration: 0.25 }}
            className={`flex items-center gap-3 px-4 py-3 rounded-lg border text-sm max-w-sm shadow-lg ${
              t.type === "success"
                ? "bg-surface border-accent/30 text-foreground"
                : "bg-surface border-red-800/50 text-foreground"
            }`}
          >
            {t.type === "success" ? (
              <CheckCircle size={16} className="text-accent flex-shrink-0" />
            ) : (
              <AlertCircle size={16} className="text-red-400 flex-shrink-0" />
            )}
            <span className="flex-1">{t.message}</span>
            <button onClick={() => onDismiss(t.id)} className="text-muted hover:text-foreground ml-1">
              <X size={14} />
            </button>
          </motion.div>
        ))}
      </AnimatePresence>
    </div>
  );
}
```

**Step 2: Create ToastProvider**

Create `frontend/components/ToastProvider.tsx`:
```tsx
"use client";

import { createContext, useCallback, useContext, useState } from "react";
import { ToastContainer, ToastData } from "./Toast";

type ToastContextType = {
  toast: (message: string, type?: "success" | "error") => void;
};

const ToastContext = createContext<ToastContextType>({ toast: () => {} });

export function ToastProvider({ children }: { children: React.ReactNode }) {
  const [toasts, setToasts] = useState<ToastData[]>([]);

  const toast = useCallback((message: string, type: "success" | "error" = "success") => {
    const id = Math.random().toString(36).slice(2);
    setToasts((prev) => [...prev, { id, message, type }]);
    setTimeout(() => setToasts((prev) => prev.filter((t) => t.id !== id)), 4000);
  }, []);

  const dismiss = useCallback((id: string) => {
    setToasts((prev) => prev.filter((t) => t.id !== id));
  }, []);

  return (
    <ToastContext.Provider value={{ toast }}>
      {children}
      <ToastContainer toasts={toasts} onDismiss={dismiss} />
    </ToastContext.Provider>
  );
}

export const useToast = () => useContext(ToastContext);
```

**Step 3: Add ToastProvider to root layout**

Modify `frontend/app/layout.tsx`:
```tsx
import type { Metadata } from "next";
import "./globals.css";
import { ToastProvider } from "@/components/ToastProvider";

export const metadata: Metadata = {
  title: "Vinyl Catalog",
  description: "Your personal vinyl record collection",
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body>
        <ToastProvider>{children}</ToastProvider>
      </body>
    </html>
  );
}
```

**Step 4: Commit**

```bash
cd ..
git add frontend/components/Toast.tsx frontend/components/ToastProvider.tsx frontend/app/layout.tsx
git commit -m "feat(frontend): add Framer Motion toast notification system"
```

---

### Task 9: Vinyl list page

**Files:**
- Create: `frontend/app/vinyls/page.tsx`

**Step 1: Create the page**

Create `frontend/app/vinyls/page.tsx`:
```tsx
import { buildVinylsApi } from "@/lib/api";
import { getToken } from "@/lib/getToken";
import { VinylListClient } from "./VinylListClient";

export default async function VinylsPage() {
  const token = await getToken();
  const api = buildVinylsApi(process.env.API_URL!, token);
  const vinyls = await api.list().catch(() => []);

  return <VinylListClient initialVinyls={vinyls} />;
}
```

**Step 2: Create VinylListClient**

Create `frontend/app/vinyls/VinylListClient.tsx`:
```tsx
"use client";

import { useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { motion } from "framer-motion";
import { Plus, LogOut } from "lucide-react";
import { VinylResponse, buildVinylsApi } from "@/lib/api";
import { VinylCard } from "@/components/VinylCard";
import { FilterBar } from "@/components/FilterBar";
import { filterVinyls, Filters } from "@/lib/filterVinyls";
import { useToast } from "@/components/ToastProvider";
import { logoutRequest } from "@/lib/auth";

type Props = { initialVinyls: VinylResponse[] };

export function VinylListClient({ initialVinyls }: Props) {
  const [vinyls, setVinyls] = useState(initialVinyls);
  const [filters, setFilters] = useState<Filters>({ search: "", genre: "", year: "" });
  const [deleting, setDeleting] = useState<string | null>(null);
  const { toast } = useToast();
  const router = useRouter();

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

  const handleLogout = async () => {
    await logoutRequest();
    router.push("/login");
  };

  return (
    <div className="min-h-screen bg-background">
      {/* Header */}
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
        {/* Filters */}
        <div className="mb-8">
          <FilterBar filters={filters} onChange={setFilters} vinyls={vinyls} />
        </div>

        {/* Results count */}
        <p className="text-muted text-sm mb-6">
          {filtered.length} {filtered.length === 1 ? "record" : "records"}
          {vinyls.length !== filtered.length ? ` of ${vinyls.length}` : ""}
        </p>

        {/* Grid */}
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
              <motion.div key={vinyl.id} layout opacity={deleting === vinyl.id ? 0.4 : 1}>
                <VinylCard vinyl={vinyl} onDelete={handleDelete} />
              </motion.div>
            ))}
          </motion.div>
        )}
      </main>
    </div>
  );
}
```

**Step 3: Create proxy Route Handler for vinyls DELETE (passes token from cookie)**

Create `frontend/app/api/vinyls/[id]/route.ts`:
```ts
import { NextRequest, NextResponse } from "next/server";
import { cookies } from "next/headers";

export async function DELETE(
  _req: NextRequest,
  { params }: { params: Promise<{ id: string }> }
) {
  const { id } = await params;
  const cookieStore = await cookies();
  const token = cookieStore.get("token")?.value ?? "";

  const res = await fetch(`${process.env.API_URL}/api/v1/vinyls/${id}`, {
    method: "DELETE",
    headers: { Authorization: `Bearer ${token}` },
  });

  if (!res.ok) {
    return NextResponse.json({ error: "Delete failed" }, { status: res.status });
  }
  return NextResponse.json({ ok: true });
}
```

**Step 4: Commit**

```bash
cd ..
git add frontend/app/vinyls/
git commit -m "feat(frontend): add vinyl list page with grid and filters"
```

---

### Task 10: VinylForm component and Create page

**Files:**
- Create: `frontend/components/VinylForm.tsx`
- Create: `frontend/app/vinyls/new/page.tsx`
- Create: `frontend/app/api/vinyls/route.ts`

**Step 1: Create VinylForm**

Create `frontend/components/VinylForm.tsx`:
```tsx
"use client";

import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Input } from "@/components/ui/Input";
import { Button } from "@/components/ui/Button";

export const vinylSchema = z.object({
  title: z.string().min(1, "Title is required"),
  artist: z.string().min(1, "Artist is required"),
  year: z.coerce.number().min(1860, "Year must be 1860 or later").max(new Date().getFullYear(), "Year cannot be in the future"),
  genre: z.string().optional().default(""),
  label: z.string().optional().default(""),
});

export type VinylFormData = z.infer<typeof vinylSchema>;

type Props = {
  defaultValues?: Partial<VinylFormData>;
  onSubmit: (data: VinylFormData) => Promise<void>;
  submitLabel?: string;
};

export function VinylForm({ defaultValues, onSubmit, submitLabel = "Save" }: Props) {
  const { register, handleSubmit, formState: { errors, isSubmitting } } = useForm<VinylFormData>({
    resolver: zodResolver(vinylSchema),
    defaultValues,
  });

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="flex flex-col gap-5">
      <Input id="title" label="Title *" placeholder="Kind of Blue" {...register("title")} error={errors.title?.message} />
      <Input id="artist" label="Artist *" placeholder="Miles Davis" {...register("artist")} error={errors.artist?.message} />
      <Input id="year" label="Year *" type="number" placeholder="1959" {...register("year")} error={errors.year?.message} />
      <Input id="genre" label="Genre" placeholder="Jazz" {...register("genre")} error={errors.genre?.message} />
      <Input id="label" label="Label" placeholder="Columbia" {...register("label")} error={errors.label?.message} />
      <Button type="submit" loading={isSubmitting} className="w-full mt-2">{submitLabel}</Button>
    </form>
  );
}
```

**Step 2: Create proxy Route Handler for POST /api/vinyls**

Create `frontend/app/api/vinyls/route.ts`:
```ts
import { NextRequest, NextResponse } from "next/server";
import { cookies } from "next/headers";

export async function POST(req: NextRequest) {
  const cookieStore = await cookies();
  const token = cookieStore.get("token")?.value ?? "";
  const body = await req.json();

  const res = await fetch(`${process.env.API_URL}/api/v1/vinyls`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${token}`,
    },
    body: JSON.stringify(body),
  });

  const data = await res.json();
  return NextResponse.json(data, { status: res.status });
}
```

**Step 3: Create new vinyl page**

Create `frontend/app/vinyls/new/page.tsx`:
```tsx
"use client";

import { useRouter } from "next/navigation";
import Link from "next/link";
import { motion } from "framer-motion";
import { ArrowLeft } from "lucide-react";
import { VinylForm, VinylFormData } from "@/components/VinylForm";
import { useToast } from "@/components/ToastProvider";

export default function NewVinylPage() {
  const router = useRouter();
  const { toast } = useToast();

  const handleSubmit = async (data: VinylFormData) => {
    const res = await fetch("/api/vinyls", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(data),
    });
    if (!res.ok) {
      const err = await res.json();
      throw new Error(err.error ?? "Failed to create vinyl");
    }
    toast("Vinyl added to collection", "success");
    router.push("/vinyls");
  };

  return (
    <div className="min-h-screen bg-background px-4 py-8">
      <div className="max-w-md mx-auto">
        <Link href="/vinyls" className="inline-flex items-center gap-2 text-muted hover:text-accent transition-colors text-sm mb-8">
          <ArrowLeft size={14} /> Back to collection
        </Link>

        <motion.div initial={{ opacity: 0, y: 16 }} animate={{ opacity: 1, y: 0 }}>
          <h1 className="font-serif text-3xl font-bold mb-2">Add a Record</h1>
          <p className="text-muted text-sm mb-8">Add a new vinyl to your collection</p>

          <div className="bg-surface border border-border rounded-xl p-8">
            <VinylForm onSubmit={handleSubmit} submitLabel="Add to collection" />
          </div>
        </motion.div>
      </div>
    </div>
  );
}
```

**Step 4: Commit**

```bash
cd ..
git add frontend/components/VinylForm.tsx frontend/app/vinyls/new/ frontend/app/api/vinyls/
git commit -m "feat(frontend): add VinylForm and create vinyl page"
```

---

### Task 11: Vinyl detail and edit page

**Files:**
- Create: `frontend/app/vinyls/[id]/page.tsx`
- Create: `frontend/app/api/vinyls/[id]/route.ts` (add GET and PUT handlers)

**Step 1: Add GET and PUT to proxy Route Handler**

Replace `frontend/app/api/vinyls/[id]/route.ts` with full content:
```ts
import { NextRequest, NextResponse } from "next/server";
import { cookies } from "next/headers";

async function getAuthHeader() {
  const cookieStore = await cookies();
  const token = cookieStore.get("token")?.value ?? "";
  return { Authorization: `Bearer ${token}` };
}

export async function GET(
  _req: NextRequest,
  { params }: { params: Promise<{ id: string }> }
) {
  const { id } = await params;
  const headers = await getAuthHeader();
  const res = await fetch(`${process.env.API_URL}/api/v1/vinyls/${id}`, { headers });
  const data = await res.json();
  return NextResponse.json(data, { status: res.status });
}

export async function PUT(
  req: NextRequest,
  { params }: { params: Promise<{ id: string }> }
) {
  const { id } = await params;
  const headers = { ...(await getAuthHeader()), "Content-Type": "application/json" };
  const body = await req.json();
  const res = await fetch(`${process.env.API_URL}/api/v1/vinyls/${id}`, {
    method: "PUT",
    headers,
    body: JSON.stringify(body),
  });
  const data = await res.json();
  return NextResponse.json(data, { status: res.status });
}

export async function DELETE(
  _req: NextRequest,
  { params }: { params: Promise<{ id: string }> }
) {
  const { id } = await params;
  const headers = await getAuthHeader();
  const res = await fetch(`${process.env.API_URL}/api/v1/vinyls/${id}`, {
    method: "DELETE",
    headers,
  });
  if (!res.ok) return NextResponse.json({ error: "Delete failed" }, { status: res.status });
  return NextResponse.json({ ok: true });
}
```

**Step 2: Create detail page**

Create `frontend/app/vinyls/[id]/page.tsx`:
```tsx
"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import Link from "next/link";
import { motion } from "framer-motion";
import { ArrowLeft, Trash2 } from "lucide-react";
import { VinylResponse } from "@/lib/api";
import { VinylForm, VinylFormData } from "@/components/VinylForm";
import { artistToHsl } from "@/lib/vinylColor";
import { useToast } from "@/components/ToastProvider";
import { Button } from "@/components/ui/Button";

export default function VinylDetailPage() {
  const { id } = useParams<{ id: string }>();
  const router = useRouter();
  const { toast } = useToast();
  const [vinyl, setVinyl] = useState<VinylResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [editing, setEditing] = useState(false);
  const [deleting, setDeleting] = useState(false);

  useEffect(() => {
    fetch(`/api/vinyls/${id}`)
      .then((r) => r.json())
      .then(setVinyl)
      .catch(() => toast("Failed to load vinyl", "error"))
      .finally(() => setLoading(false));
  }, [id]);

  const handleUpdate = async (data: VinylFormData) => {
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

  if (loading) {
    return (
      <div className="min-h-screen bg-background flex items-center justify-center">
        <div className="w-8 h-8 border-2 border-accent border-t-transparent rounded-full animate-spin" />
      </div>
    );
  }

  if (!vinyl) {
    return (
      <div className="min-h-screen bg-background flex flex-col items-center justify-center gap-4">
        <p className="font-serif text-2xl">Record not found</p>
        <Link href="/vinyls" className="text-accent text-sm">← Back to collection</Link>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-background px-4 py-8">
      <div className="max-w-4xl mx-auto">
        <Link href="/vinyls" className="inline-flex items-center gap-2 text-muted hover:text-accent transition-colors text-sm mb-8">
          <ArrowLeft size={14} /> Back to collection
        </Link>

        <div className="grid md:grid-cols-2 gap-12 items-start">
          {/* Cover panel */}
          <motion.div initial={{ opacity: 0, x: -20 }} animate={{ opacity: 1, x: 0 }}>
            <div
              className="aspect-square rounded-2xl flex items-center justify-center relative overflow-hidden"
              style={{ backgroundColor: artistToHsl(vinyl.artist) }}
            >
              {[20, 35, 50, 65, 78].map((size) => (
                <div
                  key={size}
                  className="absolute rounded-full border border-black/20"
                  style={{
                    width: `${size}%`,
                    height: `${size}%`,
                    top: `${(100 - size) / 2}%`,
                    left: `${(100 - size) / 2}%`,
                  }}
                />
              ))}
              <div className="w-6 h-6 rounded-full bg-background/50 z-10" />
            </div>
          </motion.div>

          {/* Info / edit panel */}
          <motion.div initial={{ opacity: 0, x: 20 }} animate={{ opacity: 1, x: 0 }}>
            {editing ? (
              <>
                <h2 className="font-serif text-2xl font-bold mb-6">Edit Record</h2>
                <VinylForm
                  defaultValues={{ title: vinyl.title, artist: vinyl.artist, year: vinyl.year, genre: vinyl.genre, label: vinyl.label }}
                  onSubmit={handleUpdate}
                  submitLabel="Save changes"
                />
                <button onClick={() => setEditing(false)} className="mt-4 text-sm text-muted hover:text-foreground transition-colors w-full text-center">
                  Cancel
                </button>
              </>
            ) : (
              <>
                <h1 className="font-serif text-4xl font-bold leading-tight">{vinyl.title}</h1>
                <p className="text-accent text-xl mt-1">{vinyl.artist}</p>

                <dl className="mt-8 grid grid-cols-2 gap-4">
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

                <div className="flex gap-3 mt-10">
                  <Button onClick={() => setEditing(true)} variant="ghost">Edit</Button>
                  <Button onClick={handleDelete} variant="danger" loading={deleting}>
                    <Trash2 size={14} className="mr-1.5" /> Delete
                  </Button>
                </div>
              </>
            )}
          </motion.div>
        </div>
      </div>
    </div>
  );
}
```

**Step 3: Verify TypeScript**

```bash
cd frontend && npx tsc --noEmit
```
Expected: No errors.

**Step 4: Commit**

```bash
cd ..
git add frontend/app/vinyls/[id]/ frontend/app/api/vinyls/[id]/
git commit -m "feat(frontend): add vinyl detail page with edit and delete"
```

---

### Task 12: Dockerfile and docker-compose integration

**Files:**
- Create: `frontend/Dockerfile`
- Create: `frontend/.dockerignore`
- Modify: `docker-compose.yml`

**Step 1: Create Dockerfile**

Create `frontend/Dockerfile`:
```dockerfile
# Stage 1: deps
FROM node:20-alpine AS deps
WORKDIR /app
COPY package.json package-lock.json* ./
RUN npm ci

# Stage 2: build
FROM node:20-alpine AS builder
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY . .
ENV NEXT_TELEMETRY_DISABLED=1
RUN npm run build

# Stage 3: runner
FROM node:20-alpine AS runner
WORKDIR /app
ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1
RUN addgroup --system --gid 1001 nodejs && adduser --system --uid 1001 nextjs
COPY --from=builder /app/public ./public
COPY --from=builder --chown=nextjs:nodejs /app/.next/standalone ./
COPY --from=builder --chown=nextjs:nodejs /app/.next/static ./.next/static
USER nextjs
EXPOSE 3000
ENV PORT=3000
ENV HOSTNAME="0.0.0.0"
CMD ["node", "server.js"]
```

**Step 2: Enable standalone output in next.config.ts**

Modify `frontend/next.config.ts` to add `output: "standalone"`:
```ts
import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  output: "standalone",
};

export default nextConfig;
```

**Step 3: Create .dockerignore**

Create `frontend/.dockerignore`:
```
node_modules
.next
.env*.local
*.md
```

**Step 4: Add frontend service to docker-compose.yml**

Open `docker-compose.yml` from repo root and add the `frontend` service after the existing `app` service:
```yaml
  frontend:
    build:
      context: ./frontend
      dockerfile: Dockerfile
    ports:
      - "3000:3000"
    environment:
      - API_URL=http://app:8080
      - NEXT_PUBLIC_API_URL=http://localhost:8080
    depends_on:
      - app
    networks:
      - vinyl-network
```

Note: Check the existing `networks` key name in `docker-compose.yml` and use the same one.

**Step 5: Test Docker build**

```bash
cd frontend && docker build -t vinyl-frontend .
```
Expected: Build succeeds, image created.

**Step 6: Commit**

```bash
cd ..
git add frontend/Dockerfile frontend/.dockerignore frontend/next.config.ts docker-compose.yml
git commit -m "feat(frontend): add Dockerfile and docker-compose integration"
```

---

### Task 13: Final verification

**Step 1: Run all frontend tests**

```bash
cd frontend && npm test
```
Expected: All tests pass (vinylColor, filterVinyls, api).

**Step 2: Run full TypeScript check**

```bash
cd frontend && npx tsc --noEmit
```
Expected: No errors.

**Step 3: Run lint**

```bash
cd frontend && npm run lint
```
Expected: No errors.

**Step 4: Build production**

```bash
cd frontend && npm run build
```
Expected: Build succeeds.

**Step 5: Manual smoke test**

Start with docker-compose:
```bash
cd .. && docker-compose up --build
```

Verify:
- http://localhost:3000 redirects to `/login`
- Login with admin credentials works
- Vinyl list shows grid with colored cards
- Search/genre/year filters work client-side
- Create new vinyl → appears in list
- Click card → detail page with cover art
- Edit vinyl → updates in place
- Delete vinyl → removed from list
- Logout → redirects to `/login`

**Step 6: Final commit**

```bash
git add -A
git commit -m "feat(frontend): complete Next.js vinyl catalog frontend"
```
