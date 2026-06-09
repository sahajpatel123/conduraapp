import type { MetadataRoute } from "next";
import { NAV_LINKS, SITE } from "@/lib/site";

export default function sitemap(): MetadataRoute.Sitemap {
  return ["/", ...NAV_LINKS.map((l) => l.href)].map((path) => ({
    url: `${SITE.url}${path === "/" ? "" : path}`,
    changeFrequency: "weekly",
    priority: path === "/" ? 1 : 0.7,
  }));
}
