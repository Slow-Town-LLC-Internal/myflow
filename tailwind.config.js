/** @type {import('tailwindcss').Config} */
module.exports = {
  darkMode: 'class', // or 'media' if you prefer system-based dark mode
  content: [
    './pages/**/*.{js,ts,jsx,tsx,md,mdx,mdoc}', // Include .mdoc files
    './components/**/*.{js,ts,jsx,tsx}',
    './app/**/*.{js,ts,jsx,tsx}',
  ],
  theme: {
    extend: {
      colors: {
        'morandi-light': {
          background: '#F4F5F0',
          text: '#333333',
          primary: '#A3B18A',
          secondary: '#CAD2C5',
          border: '#E5E7EB',
        },
        'morandi-dark': {
          background: '#1E293B',
          text: '#E2E8F0',
          primary: '#A3B18A',
          secondary: '#334155',
          border: '#4B5563',
        },
      },
      fontFamily: {
        'sans': ['Inter', 'var(--font-inter)', ...require('tailwindcss/defaultTheme').fontFamily.sans], // Inter for general text
        'mono': ['Fira Code', 'var(--font-fira-code)', ...require('tailwindcss/defaultTheme').fontFamily.mono], // Fira Code for code
      },
    },
  },
  plugins: [],
}