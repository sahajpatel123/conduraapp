import type { MetadataRoute } from "next";
import { SITE } from "@/lib/site";

/**
 * Robots policy for the Condura marketing site.
 *
 * Allow crawling of the marketing surface; disallow the
 * dynamic auth and download APIs (no crawler value, no
 * indexing value, and the download API is a binary proxy
 * that should not appear in search results).
 */
export default function robots(): MetadataRoute.Robots {
  return {
    rules: [
      {
        userAgent: "*",
        allow: ["/"],
        disallow: ["/api/"],
      },
    ],
    sitemap: `${SITE.url}/sitemap.xml`,
  };
}
