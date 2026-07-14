import { QRCodeSVG } from 'qrcode.react'
import { QrCode as QrCodeIcon } from 'lucide-react'

export function QrCode({ url }: { url: string }) {
  return (
    <div className="flex items-center gap-3 rounded-lg border border-border-default bg-bg-surface p-3">
      <div className="rounded-md border border-border-default bg-white p-2">
        <QRCodeSVG value={url} size={56} />
      </div>
      <div className="text-left">
        <p className="mb-0.5 flex items-center gap-1.5 text-xs font-semibold text-text-primary">
          <QrCodeIcon className="h-3.5 w-3.5" strokeWidth={1.75} />
          Código QR
        </p>
        <p className="text-xs text-text-secondary">Escanea para donar desde otro teléfono.</p>
      </div>
    </div>
  )
}
