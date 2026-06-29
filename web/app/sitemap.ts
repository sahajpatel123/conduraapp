import type { MetadataRoute } from "next";
import { SITE } from "@/lib/site";

/**
 * Sitemap for the Condura marketing site.
 *
 * Lists every public static route. Dynamic API routes
 * (/api/auth/*, /api/download/[platform]) are intentionally
 * excluded — they are server-rendered on demand, not
 * crawler-friendly destinations.
 */
export default function sitemap(): MetadataRoute.Sitemap {
  const lastModified = new Date();
  const routes = [
    "",
    "/manifesto",
    "/orchestration",
    "/ecosystem",
    "/security",
    "/changelog",
    "/download",
    "/legal",
    "/privacy",
  ];
  return routes.map((route) => ({
    url: `${SITE.url}${route}`,
    lastModified,
    changeFrequency: "weekly" as const,
    priority: route === "" ? 1 : 0.7,
  }));
}
