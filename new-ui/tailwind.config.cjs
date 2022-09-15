/** @type {import('tailwindcss').Config} */

module.exports = {
  content: ['./*.html', './src/**/*.{js,ts,jsx,tsx,css}'],
  theme: {
    extend: {
      keyframes: {
        'roll-down': {
          '0%': { margin: '0 0', opacity: 0 },
          '50%': { margin: '10% 0' },
          '100%': { margin: '0 0', opacity: 1 }
        },
        fadein: {
          from: { opacity: 0 },
          to: { opacity: 1 }
        },
        'roll-up': {
          '0%': { opacity: 1, margin: '0 0' },
          '10%': { opacity: 1, margin: '10% 0' },
          '100%': { opacity: 0, margin: '-27% 0' }
        }
      },
      animation: {
        'roll-down': 'roll-down 0.5s ease-in-out',
        'roll-up': 'roll-up 0.5s ease-out',
        fadein: 'fadein 0.5s ease-in'
      }
    },
    fontFamily: {
      sans: ['system-ui']
    }
  },
  plugins: []
}
