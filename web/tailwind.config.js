/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./web/templates/**/*.templ",
  ],
  theme: {
    extend: {},
  },
  plugins: [require("daisyui")],
  daisyui: {
    themes: ["night", "light"],
  },
};
