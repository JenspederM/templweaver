/** @type {import('tailwindcss').Config} */
module.exports = {
  mode: "jit",
  content: ["./**/*.{templ, html}"],
  theme: {
    extend: {},
  },
  plugins: [require("daisyui")],
};
