/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./frontend/src/main.jsx",
    "./frontend/src/components/*.jsx",
    "./frontend/src/pages/**/*.jsx",
  ],
  theme: {
    fontFamily: {
      sans: ['Onest', 'sans-serif'],
    },
    extend: {},
  },
  plugins: [],
}

