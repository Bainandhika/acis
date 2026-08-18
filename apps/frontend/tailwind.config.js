/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{vue,js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      fontFamily: {
        sans: ['"Plus Jakarta Sans"', 'Inter', 'system-ui', 'sans-serif'],
      },
      colors: {
        brand: {
          50: '#f0f7f4',
          100: '#d8ece1',
          200: '#b3d7c5',
          300: '#82bca0',
          400: '#589c7c',
          500: '#3d8364',
          600: '#2f6a50',
          700: '#265440',
          800: '#1f4233',
          900: '#173327',
          DEFAULT: '#3d8364',
        },
        charcoal: {
          50: '#f8fafc',
          100: '#f1f5f9',
          200: '#e2e8f0',
          300: '#cbd5e1',
          400: '#94a3b8',
          500: '#64748b',
          600: '#475569',
          700: '#334155',
          800: '#1e293b',
          900: '#0f172a',
          950: '#090d16',
        },
        surface: {
          50: '#1e293b',
          100: '#0f172a',
          200: '#090d16',
          card: '#0f172a',
        },
        tealBrand: {
          50: '#f0fdfa',
          100: '#ccfbf1',
          200: '#99f6e4',
          300: '#5eead4',
          400: '#2dd4bf',
          500: '#14b8a6',
          600: '#0d9488',
          700: '#0f766e',
          800: '#115e59',
          900: '#134e4a',
          DEFAULT: '#0f766e',
        }
      },
      boxShadow: {
        'card': '0 2px 14px 0 rgba(0, 0, 0, 0.4), 0 0 1px 0 rgba(255, 255, 255, 0.05)',
        'card-hover': '0 10px 25px -3px rgba(0, 0, 0, 0.5), 0 4px 6px -4px rgba(0, 0, 0, 0.3)',
        'float': '0 20px 40px -15px rgba(16, 185, 129, 0.2)',
      },
      borderRadius: {
        '2.5xl': '1.25rem',
        '3xl': '1.5rem',
        '4xl': '2rem',
      }
    },
  },
  plugins: [require("daisyui")],
  daisyui: {
    themes: [
      {
        dark: {
          "primary": "#10b981",
          "primary-content": "#022c22",
          "secondary": "#334155",
          "secondary-content": "#f8fafc",
          "accent": "#589c7c",
          "accent-content": "#f8fafc",
          "neutral": "#1e293b",
          "neutral-content": "#f8fafc",
          "base-100": "#0f172a",
          "base-200": "#090d16",
          "base-300": "#04070d",
          "base-content": "#f8fafc",
          "info": "#38bdf8",
          "success": "#10b981",
          "warning": "#f59e0b",
          "error": "#ef4444",
        }
      }
    ],
  },
}