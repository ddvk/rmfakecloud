/** @type {import('tailwindcss').Config} */

module.exports = {
  content: ['./*.html', './src/**/*.{js,ts,jsx,tsx,css}'],
  theme: {
    extend: {
      keyframes: {
        'roll-down': {
          '0%': { 'max-height': '0' },
          '100%': { 'max-height': '100vh' }
        },
        fadein: {
          from: { opacity: 0 },
          to: { opacity: 1 }
        },
        'roll-up': {
          '0%': { opacity: 1, 'max-height': '100vh' },
          '100%': { opacity: 0, 'max-height': '0' }
        },
        'flip-x': {
          '0%': { transform: 'rotateX(180deg)' },
          '100%': { transform: 'rotateX(0deg)' }
        },
        'flip-x-reverse': {
          '0%': { transform: 'rotateX(-180deg)' },
          '100%': { transform: 'rotateX(0deg)' }
        }
      },
      animation: {
        'roll-down': 'roll-down 0.5s ease-in',
        'roll-up': 'roll-up 0.5s ease-out',
        fadein: 'fadein 0.5s ease-in',
        'flip-x': 'flip-x 0.3s ease-out',
        'flip-x-reverse': 'flip-x-reverse 0.3s ease-out'
      }
    },
    fontFamily: {
      sans: ['system-ui']
    }
  },
  plugins: []
}
