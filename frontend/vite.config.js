import { tanstackRouter } from "@tanstack/router-plugin/vite";
import react from "@vitejs/plugin-react";
import path from "path";
import { defineConfig } from "vite";

/** @type {import('vite').UserConfig} */
const createBaseConfig = ({ appRoot }) =>
  defineConfig(({ mode }) => {
    return {
      envDir: path.resolve(__dirname, ".."),
      plugins: [
        tanstackRouter({
          target: "react",
          quoteStyle: "double",
          semicolons: true,
          disableTypes: true,
          autoCodeSplitting: false,
        }),
        react(),
      ],
      root: appRoot,
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
              allowedHosts: [".reservations.local"],
            },
      // Keeping for debug purposes
      // build: {
      //   rollupOptions: {
      //     output: {
      //       manualChunks(id) {
      //         if (id.includes("packages/components/")) return "components";
      //         if (id.includes("packages/lib/")) return "lib";
      //         if (id.includes("packages/assets/")) return "assets";
      //       },
      //     },
      //   },
      // },
    };
  });

export default createBaseConfig;
