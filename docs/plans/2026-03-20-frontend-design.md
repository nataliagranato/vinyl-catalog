# Frontend Design — Vinyl Catalog

**Date:** 2026-03-20
**Status:** Approved
**Approach:** "Dark Groove" — Next.js 15 + Tailwind + Framer Motion

## Overview

A Next.js frontend for the vinyl catalog API. Showcase-quality visual with a dark vinyl-themed aesthetic: near-black background, amber/gold accents, Playfair Display serif titles. Target: portfolio/demonstration use.

## Architecture

```
vinyl-catalog/
├── frontend/
│   ├── app/
│   │   ├── layout.tsx
│   │   ├── page.tsx              # Redirects to /login or /vinyls
│   │   ├── login/page.tsx
│   │   └── vinyls/
│   │       ├── page.tsx          # List + filters
│   │       ├── [id]/page.tsx     # Detail
│   │       └── new/page.tsx      # Create
│   ├── components/
│   │   ├── VinylCard.tsx
│   │   ├── VinylForm.tsx
│   │   ├── FilterBar.tsx
│   │   └── ui/                   # Button, Input, Modal, Toast
│   ├── lib/
│   │   ├── api.ts                # Centralized fetch wrapper
│   │   └── auth.ts               # JWT cookie management
│   ├── .env.local
│   ├── next.config.ts
│   ├── tailwind.config.ts
│   └── package.json
```

**State management:** useState + useContext for JWT. No Redux. Server Components where possible.
**Auth:** JWT stored in httpOnly cookie via Next.js Route Handler. Middleware protects all `/vinyls/*` routes.

## Visual Design

### Color Palette
| Token | Value | Use |
|---|---|---|
| Background | `#0A0A0A` | Page background |
| Surface | `#141414` | Cards, modals |
| Border | `#2A2A2A` | Subtle borders |
| Accent | `#D4A017` | Hover, highlights, CTAs |
| Text | `#F5F5F0` | Primary text |
| Muted | `#6B6B6B` | Secondary text |

### Typography
- **Playfair Display** (serif) — titles and headings
- **Inter** (sans-serif) — body, labels, inputs

## Components

### Login (`/login`)
Centered layout with project logo, username/password fields, submit button. SVG vinyl groove texture as background. Framer Motion fade+slide-up entrance animation.

### Vinyl List (`/vinyls`)
Responsive grid of `VinylCard`. Top: `FilterBar` with text search (title/artist) and selects (genre, year) — all client-side filtering, no extra requests. Floating "Add vinyl" button (bottom-right).

### VinylCard
Square card (1:1 aspect ratio, LP cover proportions). Background color generated deterministically from artist name (unique hue per artist). Hover: amber box-shadow + scale via Framer Motion. Shows title, artist, year, label. Edit/delete icons appear on hover.

### Vinyl Detail (`/vinyls/[id]`)
Two-panel layout: large animated "cover" left, all fields right. Edit and delete buttons with confirmation modal.

### VinylForm (`/vinyls/new` and edit)
Client-side validation with react-hook-form + zod. Fields: title, artist, year, genre, label.

## Data Flow

```
Browser
  → Next.js Middleware (validates httpOnly cookie)
  → Client Component → lib/api.ts → fetch()
  → Go API (app:8080)
  → Response → local state
```

`lib/api.ts` provides a base `apiFetch()` that automatically injects the token, plus a `vinyls` object with `list`, `get`, `create`, `update`, `remove` methods.

## Error Handling

| Scenario | Behavior |
|---|---|
| 401 Unauthorized | Redirect to `/login`, clear cookie |
| 404 Not Found | Inline error, no redirect |
| 500 / Network error | Framer Motion toast (bottom corner) |
| Form validation | Inline zod messages, blocks submit |

## Docker

New `frontend` service added to existing `docker-compose.yml`:

```yaml
frontend:
  build:
    context: ./frontend
    dockerfile: Dockerfile
  ports:
    - "3000:3000"
  environment:
    - NEXT_PUBLIC_API_URL=http://app:8080
  depends_on:
    - app
```

Multi-stage Dockerfile (`node:20-alpine`), final image ~150MB.

## Dependencies

```json
{
  "next": "15.x",
  "react": "19.x",
  "tailwindcss": "4.x",
  "framer-motion": "^11",
  "react-hook-form": "^7",
  "zod": "^3",
  "@fontsource/playfair-display": "*",
  "@fontsource/inter": "*"
}
```
