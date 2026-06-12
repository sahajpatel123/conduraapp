/*
  Self-host the type system so rendering needs no network. Fetches the Google
  Fonts CSS for each family (latin), downloads the woff2 files into
  public/fonts/, and writes public/fonts.css with local @font-face rules.
*/
import { mkdir, writeFile } from "node:fs/promises";
import { createWriteStream } from "node:fs";
import { Readable } from "node:stream";
import { pipeline } from "node:stream/promises";
import path from "node:path";

const UA =
  "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0 Safari/537.36";

const FAMILIES = [
  { css: "Archivo:wght@600;700;800" },
  { css: "Instrument+Serif:ital@1" },
  { css: "Geist:wght@400;500;600" },
  { css: "Geist+Mono:wght@400;500" },
];

const outDir = path.resolve("public/fonts");
await mkdir(outDir, { recursive: true });

let combined = "";
let idx = 0;

for (const fam of FAMILIES) {
  const url = `https://fonts.googleapis.com/css2?family=${fam.css}&display=swap`;
  const css = await (await fetch(url, { headers: { "User-Agent": UA } })).text();

  // Keep only latin @font-face blocks (drop the others to stay small).
  const blocks = css.split("@font-face").slice(1);
  for (const raw of blocks) {
    const block = "@font-face" + raw.split("}")[0] + "}";
    // Keep only the latin subset (its unicode-range always covers U+0000-00FF).
    if (!block.includes("U+0000")) continue;
    const m = block.match(/url\((https:[^)]+\.woff2)\)/);
    if (!m) continue;
    const fileUrl = m[1];
    const file = `f${idx++}.woff2`;
    const res = await fetch(fileUrl, { headers: { "User-Agent": UA } });
    await pipeline(Readable.fromWeb(res.body), createWriteStream(path.join(outDir, file)));
    combined += block.replace(/url\(https:[^)]+\.woff2\)/, `url(./fonts/${file})`) + "\n\n";
  }
}

await writeFile(path.resolve("public/fonts.css"), combined, "utf8");
console.log(`Wrote ${idx} woff2 files and public/fonts.css`);
