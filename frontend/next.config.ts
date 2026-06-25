import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  reactCompiler: true,

  /*
   * 本番Dockerでは必要な実行ファイルだけを
   * .next/standalone に出力する。
   */
  output: "standalone",
};

export default nextConfig;
