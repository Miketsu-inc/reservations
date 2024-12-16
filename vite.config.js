import { TanStackRouterVite } from "@tanstack/router-plugin/vite";
import react from "@vitejs/plugin-react";
import { defineConfig } from "vite";

export default defineConfig(({ mode }) => {
  return {
    plugins: [react(), TanStackRouterVite()],
    root: "frontend",
    esbuild:
      mode === "production"
        ? {
            pure: ["console.log", "console.warn"],
          }
        : {},
    server:
      mode === "production"
        ? {}
        : {
            proxy: {
              "/api": {
                target: "http://localhost:8080/",
                changeOrigin: true,
                secure: false,
              },
            },
          },
    resolve: {
      alias: {
        "@components": "/src/components",
        "@icons": "/src/assets/icons",
        "@lib": "/src/lib",
      },
    },
  };
});
