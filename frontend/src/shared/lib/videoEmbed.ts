export interface VideoEmbed {
  platform: 'youtube' | 'vimeo'
  embedUrl: string
  thumbnailUrl?: string
}

// Juntalo no aloja video propio — solo linkea a YouTube/Vimeo, así que basta
// resolver el link público a una URL de embed (y, para YouTube, a su
// thumbnail público) sin llamar a ninguna API.
export function parseVideoEmbed(raw: string): VideoEmbed | null {
  let u: URL
  try {
    u = new URL(raw)
  } catch {
    return null
  }
  const host = u.hostname.replace(/^www\.|^m\./, '')

  if (host === 'youtube.com') {
    const id = u.pathname.startsWith('/embed/')
      ? u.pathname.split('/')[2]
      : u.searchParams.get('v')
    if (!id) return null
    return {
      platform: 'youtube',
      embedUrl: `https://www.youtube.com/embed/${id}`,
      thumbnailUrl: `https://img.youtube.com/vi/${id}/hqdefault.jpg`,
    }
  }
  if (host === 'youtu.be') {
    const id = u.pathname.slice(1)
    if (!id) return null
    return {
      platform: 'youtube',
      embedUrl: `https://www.youtube.com/embed/${id}`,
      thumbnailUrl: `https://img.youtube.com/vi/${id}/hqdefault.jpg`,
    }
  }
  if (host === 'vimeo.com') {
    const id = u.pathname.split('/').filter(Boolean)[0]
    if (!id || !/^\d+$/.test(id)) return null
    return { platform: 'vimeo', embedUrl: `https://player.vimeo.com/video/${id}` }
  }
  return null
}

export function isValidVideoUrl(raw: string): boolean {
  return parseVideoEmbed(raw) !== null
}
