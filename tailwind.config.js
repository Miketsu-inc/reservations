/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./frontend/index.html",
    "./frontend/src/main.jsx",
    "./frontend/src/assets/*.jsx",
    "./frontend/src/pages/**/*.jsx",
    "./frontend/src/components/*.jsx",
  ],
  darkMode: "class",
  theme: {
    fontFamily: {
      sans: ["Onest", "sans-serif"],
    },
    extend: {
      colors: {
        primary: "rgba(var(--primary))",
        secondary: "rgba(var(--secondary))",
        hvr_primary: "rgba(var(--hvr-primary))",
        hvr_secondary: "rgba(var(--hvr-secondary))",
        hvr_gray: "rgba(var(--hvr-gray))",
        accent: "rgba(var(--accent))",
        bg_color: "rgba(var(--bg-color))",
        layer_bg: "rgba(var(--layer-bg))",
        text_color: "rgba(var(--text-color))",
      },
    },
  },
  plugins: [],
};
