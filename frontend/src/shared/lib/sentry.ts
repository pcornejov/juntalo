import * as Sentry from '@sentry/react'

// DSN vacío = Sentry queda sin inicializar y captureException/ErrorBoundary
// no hacen nada — mismo patrón "no-op sin configurar" que el backend
// (RESEND_API_KEY / SENTRY_DSN).
const dsn = import.meta.env.VITE_SENTRY_DSN

export function initSentry() {
  if (!dsn) return
  Sentry.init({
    dsn,
    environment: import.meta.env.MODE,
    tracesSampleRate: 0, // solo error tracking por ahora, no performance tracing
  })
}
