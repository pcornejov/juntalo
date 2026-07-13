import { QRCodeSVG } from 'qrcode.react'

export function QrCode({ url }: { url: string }) {
  return (
    <div className="inline-block rounded-lg border border-border-default bg-white p-3">
      <QRCodeSVG value={url} size={128} />
    </div>
  )
}
