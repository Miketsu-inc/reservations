import { TanStackRouterVite } from "@tanstack/router-plugin/vite";
import react from "@vitejs/plugin-react";
import { defineConfig } from "vite";

export default defineConfig(({ mode }) => {
  if (mode === "production") {
    return {
      plugins: [react(), TanStackRouterVite()],
      root: "frontend",
      esbuild: {
        pure: ["console.log", "console.warn"],
      },
    };
  } else {
    return {
      plugins: [react(), TanStackRouterVite()],
      root: "frontend",
      server: {
        proxy: {
          "/api": {
            target: "http://localhost:8080/",
            changeOrigin: true,
            secure: false,
          },
        },
      },
    };
  }
});
