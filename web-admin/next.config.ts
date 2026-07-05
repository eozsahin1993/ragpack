import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  output: "standalone",
  async redirects() {
    return [
      { source: "/playground", destination: "/playground/search", permanent: false },
    ];
  },
};

export default nextConfig;
