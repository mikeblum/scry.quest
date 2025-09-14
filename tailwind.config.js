module.exports = {
  content: [
    "./templates/**/*.{templ,go}",
    "./cmd/web/static/js/**/*.js"
  ],
  theme: {
    extend: {
      colors: {
        crystal: {
          400: '#c084fc',
          500: '#a855f7',
          600: '#9333ea',
        },
        primary: 'var(--text-primary)',
        secondary: 'var(--text-secondary)',
        muted: 'var(--text-muted)',
      },
    },
  },
  plugins: [],
}