import { useState } from 'react'
import { Check, Copy } from 'lucide-react'
import { Button } from './Button'
import { WhatsAppIcon } from './icons/WhatsAppIcon'
import { copyToClipboard, whatsappShareUrl } from '../lib/share'

export function ShareButtons({ url, title }: { url: string; title: string }) {
  const [copied, setCopied] = useState(false)

  async function handleCopy() {
    const ok = await copyToClipboard(url)
    setCopied(ok)
    setTimeout(() => setCopied(false), 2000)
  }

  async function handleShare() {
    if (navigator.share) {
      await navigator.share({ title, url }).catch(() => undefined)
      return
    }
    window.open(whatsappShareUrl(url, title), '_blank')
  }

  return (
    <div className="grid grid-cols-2 gap-2">
      {/*
        Fondo fijo (no usa el token bg-text-primary): ese token se invierte
        entre temas para seguir siendo legible como TEXTO, lo que lo vuelve
        casi invisible como fondo de botón en modo oscuro. El chip oscuro de
        WhatsApp es una elección de marca constante, no debe seguir el tema.
      */}
      <Button
        onClick={handleShare}
        className="flex items-center justify-center gap-2 border border-border-default bg-zinc-900 hover:opacity-90"
      >
        <WhatsAppIcon className="h-4 w-4" />
        WhatsApp
      </Button>
      <Button variant="secondary" onClick={handleCopy} className="flex items-center justify-center gap-2">
        {copied ? <Check className="h-4 w-4" strokeWidth={1.75} /> : <Copy className="h-4 w-4" strokeWidth={1.75} />}
        {copied ? 'Copiado' : 'Copiar link'}
      </Button>
    </div>
  )
}
