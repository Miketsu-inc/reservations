/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./frontend/src/main.jsx",
    "./frontend/src/components/*.jsx",
    "./frontend/src/pages/**/*.jsx",
  ],
  theme: {
    fontFamily: {
      sans: ["Onest", "sans-serif"],
    },
    extend: {
      colors: {
        primary: "#b35666",
        customhvr1: "#a85362",
        secondary: "#d2a49a",
        customhvr2: "#c69d94",
        accent: "#c18676",
        custombg: "#f9f2f4",
        customtxt: "#0a0406",
      },
    },
  },
  plugins: [],
};
