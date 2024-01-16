/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./frontend/views/*.{templ, html}"],
  theme: {
    extend: {},
  },
  plugins: [require("daisyui")],
};
