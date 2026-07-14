import { useState } from 'react'
import { MailWarning } from 'lucide-react'
import * as authApi from '../api'
import { errorMessage } from '../../../shared/api/errors'

export function VerifyEmailBanner() {
  const [status, setStatus] = useState<'idle' | 'sending' | 'sent'>('idle')
  const [error, setError] = useState<string | null>(null)

  async function handleResend() {
    setStatus('sending')
    setError(null)
    try {
      await authApi.resendVerification()
      setStatus('sent')
    } catch (err) {
      setStatus('idle')
      setError(errorMessage(err))
    }
  }

  return (
    <div className="mb-6 flex flex-wrap items-center justify-between gap-3 rounded-xl border border-amber-300 bg-amber-50 px-4 py-3 text-sm dark:border-amber-900 dark:bg-amber-950">
      <div className="flex items-center gap-2 text-amber-900 dark:text-amber-200">
        <MailWarning className="h-4 w-4 flex-none" strokeWidth={1.75} />
        <span>
          {status === 'sent'
            ? 'Te enviamos un nuevo link de verificación — revisa tu correo.'
            : 'Todavía no verificas tu email.'}
        </span>
      </div>
      {status !== 'sent' && (
        <button
          type="button"
          onClick={handleResend}
          disabled={status === 'sending'}
          className="font-medium text-amber-900 underline hover:no-underline disabled:opacity-50 dark:text-amber-200"
        >
          {status === 'sending' ? 'Enviando…' : 'Reenviar verificación'}
        </button>
      )}
      {error && <p className="w-full text-xs text-danger">{error}</p>}
    </div>
  )
}
