// Fecha relativa corta en español para el chip "Publicada hace N días" de la
// página pública de campaña. Intencionalmente simple: solo cubre el rango
// típico de vida de una colecta (horas a meses), no fechas futuras.
export function relativeDate(isoDate: string): string {
  const date = new Date(isoDate)
  if (Number.isNaN(date.getTime())) return ''

  const diffMs = Date.now() - date.getTime()
  const diffHours = Math.floor(diffMs / (1000 * 60 * 60))
  const diffDays = Math.floor(diffHours / 24)

  if (diffHours < 1) return 'Publicada hace instantes'
  if (diffHours < 24) return `Publicada hace ${diffHours} ${diffHours === 1 ? 'hora' : 'horas'}`
  if (diffDays < 30) return `Publicada hace ${diffDays} ${diffDays === 1 ? 'día' : 'días'}`

  const months = Math.floor(diffDays / 30)
  return `Publicada hace ${months} ${months === 1 ? 'mes' : 'meses'}`
}
