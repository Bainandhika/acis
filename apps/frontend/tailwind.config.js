/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{vue,js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        primary: '#22C55E', // hijau
        background: '#F9FAFB', // abu-abu terang
        textMain: '#111827', // teks utama
        warning: '#F59E0B', // peringatan/pending
        danger: '#EF4444', // bahaya/pengeluaran
      },
      fontFamily: {
        sans: ['Inter', 'sans-serif'],
      },
    },
  },
  plugins: [],
}
