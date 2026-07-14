import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { App } from './App'
import { initTheme } from '../shared/lib/theme'
import { initSentry } from '../shared/lib/sentry'
import '../styles/index.css'

// Se aplica antes del primer render para evitar el flash de tema
// incorrecto (arrancar en claro y saltar a oscuro un instante después).
initTheme()
initSentry()

const queryClient = new QueryClient()

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <QueryClientProvider client={queryClient}>
      <App />
    </QueryClientProvider>
  </StrictMode>,
)
