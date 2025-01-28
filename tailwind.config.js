/** @type {import('tailwindcss').Config} */
module.exports = {
  darkMode: 'class', // or 'media' if you prefer system-based dark mode
  content: [
    './pages/**/*.{js,ts,jsx,tsx,md,mdx,mdoc}', // Include .mdoc files
    './components/**/*.{js,ts,jsx,tsx}',
  ],
  darkMode: 'class',
  theme: {
    extend: {}
  },
  plugins: [require('@tailwindcss/typography')]

}
