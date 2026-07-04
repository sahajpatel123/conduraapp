import { promises as fs } from "node:fs";
import path from "node:path";
import { marked } from "marked";

/**
 * Reads a markdown file from the repository root (the parent of the Next.js
 * project directory) at build time and converts it to an HTML string.
 *
 * Returns `null` when the file cannot be read so callers can render a fallback.
 */
export async function readRepoMarkdown(
  fileName: string,
): Promise<string | null> {
  try {
    const filePath = path.join(process.cwd(), "..", fileName);
    const raw = await fs.readFile(filePath, "utf8");
    const html = await marked.parse(raw, { gfm: true, breaks: false });
    return html;
  } catch {
    return null;
  }
}
