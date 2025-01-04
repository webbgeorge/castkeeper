/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["../**/*.{html,templ}"],
  theme: {
    extend: {},
  },
  plugins: [require('@tailwindcss/typography'), require("daisyui")],
  daisyui: {
    themes: [
      {
        mytheme: {
          "primary": "#223257",
          "secondary": "#be3434",
          "accent": "#1976d2",
          "neutral": "#3d4451",
          "base-100": "#ffffff",
        },
      },
    ],
  },
}
