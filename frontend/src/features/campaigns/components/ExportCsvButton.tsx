import { useState } from 'react'
import { Button } from '../../../shared/ui'
import { getAccessToken } from '../../../shared/api/client'

export function ExportCsvButton({ campaignId }: { campaignId: string }) {
  const [isDownloading, setIsDownloading] = useState(false)

  async function handleExport() {
    setIsDownloading(true)
    try {
      const res = await fetch(`/api/v1/campaigns/${campaignId}/contributions/export`, {
        credentials: 'include',
        headers: { Authorization: `Bearer ${getAccessToken()}` },
      })
      if (!res.ok) return
      const blob = await res.blob()
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = 'participantes.csv'
      a.click()
      URL.revokeObjectURL(url)
    } finally {
      setIsDownloading(false)
    }
  }

  return (
    <Button variant="secondary" onClick={handleExport} disabled={isDownloading}>
      {isDownloading ? 'Descargando…' : 'Exportar CSV'}
    </Button>
  )
}
