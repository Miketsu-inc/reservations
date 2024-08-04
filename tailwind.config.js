/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./frontend/index.html",
    "./frontend/src/main.jsx",
    "./frontend/src/assets/*.jsx",
    "./frontend/src/pages/**/*.jsx",
    "./frontend/src/components/*.jsx",
  ],
  theme: {
    fontFamily: {
      sans: ["Onest", "sans-serif"],
    },
    extend: {
      colors: {
        primary: "#6454b1",
        customhvr1: "#6d5fb3",
        secondary: "#873984",
        customhvr2: "#B347AF",
        accent: "#6666c0",
        custombg: "#131217",
        customtxt: "#EDEDEF",
      },
    },
  },
  plugins: [],
};
