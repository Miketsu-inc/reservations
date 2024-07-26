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
