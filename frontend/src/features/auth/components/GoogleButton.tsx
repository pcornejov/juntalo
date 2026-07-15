import { useEffect, useRef } from 'react'

// Client ID vacío = sin login con Google (dev/test, o un deploy que
// todavía no lo configuró) — mismo patrón "vacío = off" que
// VITE_TURNSTILE_SITE_KEY. El client ID es público (Google lo espera así,
// se usa desde el navegador), el secreto real nunca sale del backend.
const clientId = import.meta.env.VITE_GOOGLE_CLIENT_ID

const SCRIPT_SRC = 'https://accounts.google.com/gsi/client'

declare global {
  interface Window {
    google?: {
      accounts: {
        id: {
          initialize: (options: {
            client_id: string
            callback: (response: { credential: string }) => void
          }) => void
          renderButton: (
            container: HTMLElement,
            options: { theme: string; size: string; width?: number; text?: string },
          ) => void
        }
      }
    }
  }
}

let scriptPromise: Promise<void> | null = null

function loadScript(): Promise<void> {
  if (window.google?.accounts?.id) return Promise.resolve()
  if (!scriptPromise) {
    scriptPromise = new Promise((resolve, reject) => {
      const script = document.createElement('script')
      script.src = SCRIPT_SRC
      script.async = true
      script.defer = true
      script.onload = () => resolve()
      script.onerror = () => reject(new Error('No se pudo cargar Google Identity Services'))
      document.head.appendChild(script)
    })
  }
  return scriptPromise
}

export function isGoogleLoginEnabled() {
  return Boolean(clientId)
}

export function GoogleButton({ onCredential }: { onCredential: (credential: string) => void }) {
  const containerRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (!clientId || !containerRef.current) return
    let cancelled = false

    loadScript().then(() => {
      if (cancelled || !containerRef.current || !window.google) return
      window.google.accounts.id.initialize({
        client_id: clientId,
        callback: (response) => onCredential(response.credential),
      })
      window.google.accounts.id.renderButton(containerRef.current, {
        theme: 'outline',
        size: 'large',
        width: 320,
        text: 'continue_with',
      })
    })

    return () => {
      cancelled = true
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps -- onCredential se pasa fresh en cada render, no debe re-montar el botón
  }, [])

  if (!clientId) return null

  return <div ref={containerRef} className="flex justify-center" />
}
