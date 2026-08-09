/** @type {import('tailwindcss').Config} */
export default {
  // class 策略：html.dark 控制夜间模式（默认夜间）
  darkMode: 'class',
  content: ['./index.html', './src/**/*.{vue,js}'],
  theme: {
    extend: {},
  },
  plugins: [],
}
