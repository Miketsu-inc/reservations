import createBaseConfig from "../../vite.config.js";

export default createBaseConfig({
  root: import.meta.dirname,
  resolve: {
    alias: {
      "@reservations/jabulani/lib": "/src/lib",
      "@reservations/jabulani/components": "/src/components",
    },
  },
});
