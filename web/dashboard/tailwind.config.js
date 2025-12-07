// Krustron Dashboard - Tailwind Configuration
// Author: Anubhav Gain <anubhavg@infopercept.com>
// Design: Dark Mode OLED + Glassmorphism + Tech Startup Typography
// Based on UI/UX Pro Max skill recommendations

/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        // Primary - Trust Blue (#2563EB)
        primary: {
          50: '#eff6ff',
          100: '#dbeafe',
          200: '#bfdbfe',
          300: '#93c5fd',
          400: '#60a5fa',
          500: '#2563eb',
          600: '#1d4ed8',
          700: '#1e40af',
          800: '#1e3a8a',
          900: '#1e3a8a',
          950: '#172554',
        },
        // Accent - Energy Orange (#F97316)
        accent: {
          50: '#fff7ed',
          100: '#ffedd5',
          200: '#fed7aa',
          300: '#fdba74',
          400: '#fb923c',
          500: '#f97316',
          600: '#ea580c',
          700: '#c2410c',
          800: '#9a3412',
          900: '#7c2d12',
          950: '#431407',
        },
        // Secondary - Cool Cyan
        secondary: {
          50: '#ecfeff',
          100: '#cffafe',
          200: '#a5f3fc',
          300: '#67e8f9',
          400: '#22d3ee',
          500: '#06b6d4',
          600: '#0891b2',
          700: '#0e7490',
          800: '#155e75',
          900: '#164e63',
          950: '#083344',
        },
        // Dark Mode OLED Surfaces
        surface: {
          DEFAULT: '#0A0E27',
          50: '#1a1f3d',
          100: '#161b36',
          200: '#121630',
          300: '#0e1229',
          400: '#0A0E27',
          500: '#080b1f',
          600: '#060817',
          700: '#04050f',
          800: '#020308',
          900: '#000000',
        },
        // Enhanced Glass effects
        glass: {
          light: 'rgba(255, 255, 255, 0.03)',
          medium: 'rgba(255, 255, 255, 0.06)',
          heavy: 'rgba(255, 255, 255, 0.1)',
          ultra: 'rgba(255, 255, 255, 0.15)',
          border: 'rgba(255, 255, 255, 0.08)',
          'border-hover': 'rgba(255, 255, 255, 0.15)',
        },
        // Status Colors with glow variants
        status: {
          healthy: '#22c55e',
          'healthy-glow': 'rgba(34, 197, 94, 0.4)',
          warning: '#eab308',
          'warning-glow': 'rgba(234, 179, 8, 0.4)',
          error: '#ef4444',
          'error-glow': 'rgba(239, 68, 68, 0.4)',
          info: '#3b82f6',
          'info-glow': 'rgba(59, 130, 246, 0.4)',
          synced: '#22c55e',
          outOfSync: '#f97316',
          progressing: '#3b82f6',
          degraded: '#ef4444',
          unknown: '#6b7280',
        },
        // Neon accents for cyberpunk touches
        neon: {
          cyan: '#00FFFF',
          magenta: '#FF00FF',
          green: '#00FF00',
          purple: '#8B5CF6',
        },
      },
      // Tech Startup Typography: Space Grotesk + DM Sans
      fontFamily: {
        heading: ['Space Grotesk', 'system-ui', 'sans-serif'],
        sans: ['DM Sans', 'system-ui', 'sans-serif'],
        mono: ['JetBrains Mono', 'Fira Code', 'monospace'],
      },
      fontSize: {
        '2xs': ['0.625rem', { lineHeight: '0.875rem' }],
        '3xl': ['1.875rem', { lineHeight: '2.25rem', letterSpacing: '-0.02em' }],
        '4xl': ['2.25rem', { lineHeight: '2.5rem', letterSpacing: '-0.02em' }],
        '5xl': ['3rem', { lineHeight: '1.1', letterSpacing: '-0.02em' }],
        '6xl': ['3.75rem', { lineHeight: '1', letterSpacing: '-0.02em' }],
      },
      letterSpacing: {
        tightest: '-0.04em',
        tighter: '-0.02em',
      },
      backdropBlur: {
        xs: '2px',
        '2xl': '40px',
        '3xl': '64px',
      },
      animation: {
        // Smooth fade animations
        'fade-in': 'fadeIn 0.4s ease-out',
        'fade-out': 'fadeOut 0.3s ease-in',
        // Slide animations
        'slide-up': 'slideUp 0.4s cubic-bezier(0.16, 1, 0.3, 1)',
        'slide-down': 'slideDown 0.4s cubic-bezier(0.16, 1, 0.3, 1)',
        'slide-left': 'slideLeft 0.4s cubic-bezier(0.16, 1, 0.3, 1)',
        'slide-right': 'slideRight 0.4s cubic-bezier(0.16, 1, 0.3, 1)',
        // Scale animations
        'scale-in': 'scaleIn 0.3s cubic-bezier(0.16, 1, 0.3, 1)',
        'scale-out': 'scaleOut 0.2s ease-in',
        // Glow & pulse
        'pulse-slow': 'pulse 3s cubic-bezier(0.4, 0, 0.6, 1) infinite',
        'pulse-glow': 'pulseGlow 2s ease-in-out infinite',
        'glow': 'glow 2s ease-in-out infinite alternate',
        // Loading states
        'shimmer': 'shimmer 2s linear infinite',
        'spin-slow': 'spin 3s linear infinite',
        // Float effect
        'float': 'float 6s ease-in-out infinite',
        // Gradient shift
        'gradient': 'gradient 8s ease infinite',
        // Border glow
        'border-glow': 'borderGlow 2s ease-in-out infinite',
      },
      keyframes: {
        fadeIn: {
          '0%': { opacity: '0' },
          '100%': { opacity: '1' },
        },
        fadeOut: {
          '0%': { opacity: '1' },
          '100%': { opacity: '0' },
        },
        slideUp: {
          '0%': { opacity: '0', transform: 'translateY(16px)' },
          '100%': { opacity: '1', transform: 'translateY(0)' },
        },
        slideDown: {
          '0%': { opacity: '0', transform: 'translateY(-16px)' },
          '100%': { opacity: '1', transform: 'translateY(0)' },
        },
        slideLeft: {
          '0%': { opacity: '0', transform: 'translateX(16px)' },
          '100%': { opacity: '1', transform: 'translateX(0)' },
        },
        slideRight: {
          '0%': { opacity: '0', transform: 'translateX(-16px)' },
          '100%': { opacity: '1', transform: 'translateX(0)' },
        },
        scaleIn: {
          '0%': { opacity: '0', transform: 'scale(0.9)' },
          '100%': { opacity: '1', transform: 'scale(1)' },
        },
        scaleOut: {
          '0%': { opacity: '1', transform: 'scale(1)' },
          '100%': { opacity: '0', transform: 'scale(0.9)' },
        },
        pulseGlow: {
          '0%, 100%': { opacity: '1' },
          '50%': { opacity: '0.5' },
        },
        glow: {
          '0%': { boxShadow: '0 0 5px rgba(37, 99, 235, 0.3), 0 0 10px rgba(37, 99, 235, 0.2)' },
          '100%': { boxShadow: '0 0 20px rgba(37, 99, 235, 0.6), 0 0 40px rgba(37, 99, 235, 0.3)' },
        },
        shimmer: {
          '0%': { backgroundPosition: '-200% 0' },
          '100%': { backgroundPosition: '200% 0' },
        },
        float: {
          '0%, 100%': { transform: 'translateY(0)' },
          '50%': { transform: 'translateY(-20px)' },
        },
        gradient: {
          '0%, 100%': { backgroundPosition: '0% 50%' },
          '50%': { backgroundPosition: '100% 50%' },
        },
        borderGlow: {
          '0%, 100%': { borderColor: 'rgba(37, 99, 235, 0.5)' },
          '50%': { borderColor: 'rgba(249, 115, 22, 0.5)' },
        },
      },
      boxShadow: {
        // Glass shadows
        'glass': '0 4px 30px rgba(0, 0, 0, 0.1)',
        'glass-sm': '0 2px 15px rgba(0, 0, 0, 0.08)',
        'glass-lg': '0 8px 32px rgba(0, 0, 0, 0.2)',
        'glass-xl': '0 16px 48px rgba(0, 0, 0, 0.25)',
        // Glow shadows
        'glow-sm': '0 0 10px rgba(37, 99, 235, 0.3)',
        'glow-md': '0 0 20px rgba(37, 99, 235, 0.4)',
        'glow-lg': '0 0 30px rgba(37, 99, 235, 0.5)',
        'glow-xl': '0 0 50px rgba(37, 99, 235, 0.6)',
        'glow-accent': '0 0 20px rgba(249, 115, 22, 0.4)',
        'glow-accent-lg': '0 0 40px rgba(249, 115, 22, 0.5)',
        // Status glows
        'glow-success': '0 0 20px rgba(34, 197, 94, 0.4)',
        'glow-warning': '0 0 20px rgba(234, 179, 8, 0.4)',
        'glow-error': '0 0 20px rgba(239, 68, 68, 0.4)',
        // Neon shadows
        'neon-cyan': '0 0 20px #00FFFF, 0 0 40px #00FFFF',
        'neon-purple': '0 0 20px #8B5CF6, 0 0 40px #8B5CF6',
        // Inset shadows
        'inner-glow': 'inset 0 1px 1px rgba(255, 255, 255, 0.1)',
        'inner-dark': 'inset 0 2px 4px rgba(0, 0, 0, 0.2)',
      },
      backgroundImage: {
        // Gradient backgrounds
        'gradient-radial': 'radial-gradient(var(--tw-gradient-stops))',
        'gradient-conic': 'conic-gradient(from 180deg at 50% 50%, var(--tw-gradient-stops))',
        'gradient-mesh': 'radial-gradient(at 40% 20%, hsla(226, 68%, 54%, 0.3) 0px, transparent 50%), radial-gradient(at 80% 0%, hsla(28, 100%, 53%, 0.2) 0px, transparent 50%), radial-gradient(at 0% 50%, hsla(210, 100%, 50%, 0.2) 0px, transparent 50%)',
        // Glass gradients
        'glass-gradient': 'linear-gradient(135deg, rgba(255, 255, 255, 0.1), rgba(255, 255, 255, 0.05))',
        'glass-gradient-hover': 'linear-gradient(135deg, rgba(255, 255, 255, 0.15), rgba(255, 255, 255, 0.08))',
        // Shimmer effect
        'shimmer': 'linear-gradient(90deg, transparent, rgba(255, 255, 255, 0.1), transparent)',
        // Gradient borders
        'gradient-border': 'linear-gradient(135deg, rgba(37, 99, 235, 0.5), rgba(249, 115, 22, 0.5))',
        // Hero gradients
        'hero-glow': 'radial-gradient(ellipse 80% 50% at 50% -20%, rgba(37, 99, 235, 0.3), transparent)',
        'hero-glow-accent': 'radial-gradient(ellipse 80% 50% at 50% 120%, rgba(249, 115, 22, 0.2), transparent)',
      },
      borderRadius: {
        '4xl': '2rem',
        '5xl': '2.5rem',
      },
      transitionTimingFunction: {
        'bounce-in': 'cubic-bezier(0.68, -0.55, 0.265, 1.55)',
        'smooth': 'cubic-bezier(0.16, 1, 0.3, 1)',
      },
      transitionDuration: {
        '400': '400ms',
      },
      zIndex: {
        '60': '60',
        '70': '70',
        '80': '80',
        '90': '90',
        '100': '100',
      },
    },
  },
  plugins: [],
}
