/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./frontend/index.html",
    "./frontend/src/main.jsx",
    "./frontend/src/assets/icons/*.jsx",
    "./frontend/src/components/*.jsx",

    "./frontend/src/routes/*.jsx",
    "./frontend/src/routes/**/*.jsx",
    "./frontend/src/routes/**/**/*.jsx",
    "./frontend/src/routes/**/**/**/*.jsx",
    "./frontend/src/routes/**/**/**/**/*.jsx",
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
  plugins: [
    function ({ addBase }) {
      addBase({
        ".autofill:-webkit-autofill, .autofill:-webkit-autofill:hover, .autofill:-webkit-autofill:focus, .autofill-styles select:-webkit-autofill, .autofill-styles select:-webkit-autofill:hover, .autofill-styles select:-webkit-autofill:focus":
          {
            "-webkit-text-fill-color": "rgb(var(--text-color))",
            "-webkit-box-shadow": "0 0 0px 1000px rgba(255, 255, 255, 0) inset",
            transition: "background-color 5000s ease-in-out 0s",
          },
      });
    },
  ],
};
