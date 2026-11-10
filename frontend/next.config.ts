import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  // Emit .next/standalone: a minimal server.js + only the node_modules it
  // actually imports, so the Docker image doesn't ship the full dependency tree.
  output: "standalone",
};

export default nextConfig;
