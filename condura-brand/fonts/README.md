# fonts/

Condura does **not** vendor font binaries in-repo.

## What we use

| Role | Family | Source |
|---|---|---|
| UI sans | System stack (`-apple-system`, `Segoe UI`, `Inter` if installed) | OS / optional user install |
| Marketing display | Loaded by `condura-ui` via its own CSS / next/font | Next.js app, not this folder |
| Mono | `ui-monospace`, `SF Mono`, `Menlo`, `Consolas` | OS |

## Why empty of `.ttf` / `.woff2`

- Avoids shipping multi-MB binaries and license redistribution risk.
- Desktop (Wails) and marketing (Next.js) already resolve type independently.
- Brand identity is carried by **color + motion + mark**, not a custom typeface.

## If we add a licensed family later

1. Drop binaries under `fonts/<family>/` with a `LICENSE` file.
2. Add `@font-face` rules to `tokens/` and wire `make brand` (when added).
3. Do not claim a "full brand kit" until both the face and the license ship.
