/** @type {import('tailwindcss').Config} */
export default {
  darkMode: 'class',
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx,mdx}",
  ],
  theme: {
    extend: {
      boxShadow:{
        'button': '0px 1px 4px rgba(0, 0, 0, 0.05)',
        'card': [
          '0px 2px 1px rgba(0, 0, 0, 0.05)',
          '0px 0px 1px rgba(0, 0, 0, 0.25)'
        ],
        'popover': [
          '0px 0px 2px rgba(0, 0, 0, 0.2)',
          '0px 2px 10px rgba(0, 0, 0, 0.1)'
        ],
        'deep': [
          '0px 0px 0px 1px rgba(6, 44, 82, 0.1)',
          '0px 2px 16px rgba(33, 43, 54, 0.08);'
        ],
        'modal': [
          '0px 0px 1px rgba(0, 0, 0, 0.2)',
          '0px 26px 80px rgba(0, 0, 0, 0.2)'
        ],
        'base': [
          '0px 1px 3px rgba(63, 63, 68, 0.15)',
          '0px 0px 0px 1px rgba(63, 63, 68, 0.05)'
        ],
      },
    },
    fontFamily:{
      sans: ['Inter', 'sans-serif'],
    },
    screens:{
      'xs': {max:'489px'},
      // => @media (min-width: 490px) { ... }
      'sm': {min:'490px', max:'767px'},
      // => @media (min-width: 768px) { ... }
      'md': {min:'768px', max:'1039px'},
      // => @media (min-width: 1040px) { ... }
      'lg': {min:'1440px'},
      // => @media (min-width: 1440px) { ... }
    },
    spacing: {
      px: '1px',
      0: '0',
      0.25: '0.0625rem',
      0.5: '0.125rem',
      1: '0.25rem',
      2: '0.5rem',
      3: '0.75rem',
      4: '1rem',
      5: '1.25rem',
      6: '1.5rem',
      8: '2rem',
      10: '2.5rem',
      12: '3rem',
      16: '4rem',
      20: '5rem',
      24: '6rem',
      32: '8rem',
    },
    colors:{

      // Base Colors
      "surface":"#FAFAFA",
      "primary":"#3B82F6",
      "secondary":"#14B8A6",
      "critical":"#EF4444",
      "warning":"#F59E0B",
      "success":"#22C55E",

      // Surface Colors
      "surface-subdued": "#F4F4F5",
      "surface-hovered":"#e4e4e7",
      "surface-pressed":"#d4d4d8",
      "surface-input":"#FAFAFA",

      "surface-secondary-default":"#14B8A6",
      "surface-secondary-subdued":"#ccfbf1",
      "surface-secondary-hovered":"#0D9488",
      "surface-secondary-pressed":"#0F766E",
      "surface-secondary-selected":"#5eead4",

      "surface-primary-default":"#3B82F6",
      "surface-primary-subdued":"#dbeafe",
      "surface-primary-hovered":"#2563eb",
      "surface-primary-pressed":"#1D4ED8",
      "surface-primary-selected":"#93c5fd",

      "surface-critical-default":"#EF4444",
      "surface-critical-subdued":"#fee2e2",
      "surface-critical-hovered":"#dc2626",
      "surface-critical-pressed":"#B91C1C",

      "surface-warning-default":"#F59E0B",
      "surface-warning-subdued":"#fef3c7",
      "surface-warning-hovered":"#d97706",
      "surface-warning-pressed":"#B45309",

      "surface-success-default":"#22C55E",
      "surface-success-subdued":"#dcfce7",
      "surface-success-hovered":"#16a34a",
      "surface-success-pressed":"#15803D",

      // Text Colors
      "text-default":"#111827",
      "text-soft":"#1f2937",
      "text-strong":"#374151",
      "text-disabled":"#9ca3af",
      "text-critical":"#b91c1c",
      "text-warning":"#b45309",
      "text-success":"#15803d",
      "text-primary":"#1d4ed8",
      "text-secondary":"#0F766E",
      "text-on-primary":"#F9FAFB",
      "text-on-secondary":"#F9FAFB",

      // Border Colors
      "border-default":"#D4D4D8",
      "border-primary":"#2563EB",
      "border-secondary":"#0d9488",
      "border-success":"#16A34A",
      "border-critical":"#DC2626",
      "border-warning":"#D97706",
      "border-disabled":"#E4E4E7",
    }
  },
  plugins: [],
}

