/** @type {import('tailwindcss').Config} */
module.exports = {
  mode: "jit",
  content: ["./frontend/views/*.{templ, html}"],
  theme: {
    extend: {},
  },
  plugins: [require("daisyui")],
};
