import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  output: "standalone",
  async redirects() {
    return [
      { source: "/playground", destination: "/playground/search", permanent: false },
      { source: "/collections/:slug/documents/:id", destination: "/collections/:slug/documents/:id/chunks", permanent: false },
    ];
  },
};

export default nextConfig;
