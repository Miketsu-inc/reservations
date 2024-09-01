import react from "@vitejs/plugin-react";
import { defineConfig } from "vite";

export default defineConfig(({ mode }) => {
  if (mode === "production") {
    return {
      plugins: [react()],
      root: "frontend",
    };
  } else {
    return {
      plugins: [react()],
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
