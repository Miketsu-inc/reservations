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
        primary: "#9D90DD",
        customhvr1: "#9589CF",
        secondary: "#41386B",
        customhvr2: "#9589CF",
        accent: "#6A5AB8",
        custombg: "#131217",
        customtxt: "#EDEDEF",
      },
    },
  },
  plugins: [],
};
