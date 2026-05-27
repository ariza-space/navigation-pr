import type { Config } from 'tailwindcss'

export default {
  darkMode: ['class'],
  content: ['./index.html', './src/**/*.{vue,ts}'],
  theme: {
    extend: {
      colors: {
        border: 'var(--border)',
        input: 'var(--surface-input)',
        ring: 'var(--focus)',
        background: 'var(--page-bg-solid)',
        foreground: 'var(--page-text)',
        primary: {
          DEFAULT: 'var(--accent)',
          foreground: 'var(--accent-text)',
        },
        muted: {
          DEFAULT: 'var(--surface)',
          foreground: 'var(--page-muted)',
        },
        destructive: {
          DEFAULT: 'var(--danger-bg)',
          foreground: 'var(--danger-text)',
        },
      },
      boxShadow: {
        card: '0 18px 52px var(--shadow)',
        dialog: '0 30px 90px var(--shadow-strong)',
      },
      borderRadius: {
        sm: '8px',
        md: '10px',
        lg: '14px',
      },
    },
  },
  plugins: [],
} satisfies Config
