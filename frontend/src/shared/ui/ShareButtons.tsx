import { useState } from 'react'
import { Button } from './Button'
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
    <div className="flex flex-wrap gap-2">
      <Button onClick={handleShare}>Compartir por WhatsApp</Button>
      <Button variant="secondary" onClick={handleCopy}>
        {copied ? 'Copiado' : 'Copiar link'}
      </Button>
    </div>
  )
}
