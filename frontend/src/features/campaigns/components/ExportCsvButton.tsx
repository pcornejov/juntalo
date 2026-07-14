import { useState } from 'react'
import { Download } from 'lucide-react'
import { Button } from '../../../shared/ui'
import { apiClient } from '../../../shared/api/client'

export function ExportCsvButton({ campaignId }: { campaignId: string }) {
  const [isDownloading, setIsDownloading] = useState(false)
  const [error, setError] = useState(false)

  async function handleExport() {
    setIsDownloading(true)
    setError(false)
    try {
      const blob = await apiClient.download(`/campaigns/${campaignId}/contributions/export`)
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = 'participantes.csv'
      a.click()
      URL.revokeObjectURL(url)
    } catch {
      setError(true)
    } finally {
      setIsDownloading(false)
    }
  }

  return (
    <div className="flex items-center gap-2">
      <Button
        variant="secondary"
        className="flex items-center gap-1.5"
        onClick={handleExport}
        disabled={isDownloading}
      >
        <Download className="h-3.5 w-3.5" strokeWidth={1.75} />
        {isDownloading ? 'Descargando…' : 'Exportar CSV'}
      </Button>
      {error && <span className="text-xs text-danger">No se pudo descargar el CSV.</span>}
    </div>
  )
}
