import createBaseConfig from "../../vite.config";

export default createBaseConfig({
  root: __dirname,
  resolve: {
    alias: {
      "@reservations/jabulani/lib": "/src/lib",
    },
  },
});
