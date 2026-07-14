import { useEffect, useRef } from 'react'

// Site key vacía = sin captcha (dev/test, o un deploy que todavía no lo
// configuró) — mismo patrón "vacío = off" que VITE_SENTRY_DSN. El secret
// real vive solo en el backend (TURNSTILE_SECRET_KEY).
const siteKey = import.meta.env.VITE_TURNSTILE_SITE_KEY

const SCRIPT_SRC = 'https://challenges.cloudflare.com/turnstile/v0/api.js'

declare global {
  interface Window {
    turnstile?: {
      render: (
        container: HTMLElement,
        options: { sitekey: string; callback: (token: string) => void; 'expired-callback'?: () => void },
      ) => string
      remove: (widgetId: string) => void
    }
  }
}

let scriptPromise: Promise<void> | null = null

function loadScript(): Promise<void> {
  if (window.turnstile) return Promise.resolve()
  if (!scriptPromise) {
    scriptPromise = new Promise((resolve, reject) => {
      const script = document.createElement('script')
      script.src = SCRIPT_SRC
      script.async = true
      script.onload = () => resolve()
      script.onerror = () => reject(new Error('No se pudo cargar Turnstile'))
      document.head.appendChild(script)
    })
  }
  return scriptPromise
}

// isCaptchaEnabled deja que RegisterPage decida si exige un token antes de
// dejar enviar el formulario — sin site key, el widget ni se monta.
export function isCaptchaEnabled() {
  return Boolean(siteKey)
}

export function TurnstileWidget({ onVerify }: { onVerify: (token: string) => void }) {
  const containerRef = useRef<HTMLDivElement>(null)
  const widgetIdRef = useRef<string | null>(null)

  useEffect(() => {
    if (!siteKey || !containerRef.current) return
    let cancelled = false

    loadScript().then(() => {
      if (cancelled || !containerRef.current || !window.turnstile) return
      widgetIdRef.current = window.turnstile.render(containerRef.current, {
        sitekey: siteKey,
        callback: onVerify,
        'expired-callback': () => onVerify(''),
      })
    })

    return () => {
      cancelled = true
      if (widgetIdRef.current && window.turnstile) {
        window.turnstile.remove(widgetIdRef.current)
      }
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps -- onVerify se pasa fresh en cada render, no debe re-montar el widget
  }, [])

  if (!siteKey) return null

  return <div ref={containerRef} />
}
