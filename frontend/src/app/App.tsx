import { RouterProvider } from 'react-router-dom'
import * as Sentry from '@sentry/react'
import { router } from './router'
import { AuthProvider } from '../features/auth/hooks/useAuth'

function ErrorFallback() {
  return (
    <div className="mx-auto flex min-h-screen max-w-sm flex-col items-center justify-center gap-4 p-6 text-center">
      <h1 className="text-xl font-semibold">Algo salió mal</h1>
      <p className="text-text-secondary">Ya nos enteramos del problema. Intenta recargar la página.</p>
      <button
        type="button"
        onClick={() => window.location.reload()}
        className="rounded-lg bg-brand-hover px-4 py-2 font-medium text-white"
      >
        Recargar
      </button>
    </div>
  )
}

export function App() {
  return (
    <Sentry.ErrorBoundary fallback={<ErrorFallback />}>
      <AuthProvider>
        <RouterProvider router={router} />
      </AuthProvider>
    </Sentry.ErrorBoundary>
  )
}
