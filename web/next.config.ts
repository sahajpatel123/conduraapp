import type { NextConfig } from "next";
import path from "node:path";

const nextConfig: NextConfig = {
  turbopack: {
    // Keep manifests scoped to web/ when the monorepo has other lockfiles.
    root: path.resolve(process.cwd()),
  },
};

export default nextConfig;
