import { useRef, useState, type UIEvent } from 'react'
import { Play, Video } from 'lucide-react'
import { parseVideoEmbed } from '../../../shared/lib/videoEmbed'

interface MediaCarouselProps {
  images: string[]
  videoUrl?: string
  alt: string
  className?: string
}

// Extiende el carrusel de fotos con un slide de video (YouTube/Vimeo) al
// inicio — mismo scroll-snap táctil que el carrusel de solo-fotos, pero el
// primer slide reemplaza su contenido por un iframe embed al tocar play, en
// vez de navegar fuera de la página.
export function MediaCarousel({ images, videoUrl, alt, className }: MediaCarouselProps) {
  const trackRef = useRef<HTMLDivElement>(null)
  const [active, setActive] = useState(0)
  const [playing, setPlaying] = useState(false)
  const video = videoUrl ? parseVideoEmbed(videoUrl) : null
  const mediaClassName = className ?? 'h-56 w-full object-cover'

  const slideCount = (video ? 1 : 0) + images.length

  function handleScroll(e: UIEvent<HTMLDivElement>) {
    const el = e.currentTarget
    setActive(Math.round(el.scrollLeft / el.clientWidth))
  }

  function goTo(index: number) {
    const el = trackRef.current
    if (!el) return
    el.scrollTo({ left: index * el.clientWidth, behavior: 'smooth' })
  }

  if (slideCount <= 1) {
    if (video) {
      return <VideoSlide video={video} playing={playing} onPlay={() => setPlaying(true)} className={mediaClassName} />
    }
    return <img src={images[0]} alt={alt} className={mediaClassName} />
  }

  return (
    <div className="relative">
      <div
        ref={trackRef}
        onScroll={handleScroll}
        className="flex snap-x snap-mandatory overflow-x-auto scroll-smooth [-ms-overflow-style:none] [scrollbar-width:none] [&::-webkit-scrollbar]:hidden"
      >
        {video && (
          <div className="w-full flex-none snap-start">
            <VideoSlide video={video} playing={playing} onPlay={() => setPlaying(true)} className={mediaClassName} />
          </div>
        )}
        {images.map((src, i) => (
          <img
            key={src}
            src={src}
            alt={video || i > 0 ? `${alt} (${i + 1})` : alt}
            className={`w-full flex-none snap-start ${mediaClassName}`}
          />
        ))}
      </div>
      <div className="pointer-events-none absolute inset-x-0 bottom-3 z-10 flex justify-center gap-1.5">
        {Array.from({ length: slideCount }).map((_, i) => (
          <button
            key={i}
            type="button"
            onClick={() => goTo(i)}
            aria-label={video && i === 0 ? 'Ir al video' : `Ir a la foto ${video ? i : i + 1}`}
            className={`pointer-events-auto h-1.5 rounded-full transition-all ${
              i === active ? 'w-4 bg-white' : 'w-1.5 bg-white/50'
            }`}
          />
        ))}
      </div>
    </div>
  )
}

function VideoSlide({
  video,
  playing,
  onPlay,
  className,
}: {
  video: NonNullable<ReturnType<typeof parseVideoEmbed>>
  playing: boolean
  onPlay: () => void
  className: string
}) {
  if (playing) {
    return (
      <iframe
        src={`${video.embedUrl}?autoplay=1`}
        title="Video de la campaña"
        allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture"
        allowFullScreen
        className={className}
      />
    )
  }
  return (
    <button
      type="button"
      onClick={onPlay}
      className={`group relative flex w-full items-center justify-center bg-black ${className}`}
    >
      {video.thumbnailUrl ? (
        <img src={video.thumbnailUrl} alt="" className="absolute inset-0 h-full w-full object-cover opacity-80" />
      ) : (
        <div className="absolute inset-0 bg-gradient-to-br from-bg-surface via-accent-tint to-brand-hover/40" />
      )}
      <span className="relative flex h-14 w-14 items-center justify-center rounded-full bg-white shadow-lg transition-transform group-hover:scale-105">
        <Play className="ml-0.5 h-5 w-5 text-brand-hover" fill="currentColor" strokeWidth={0} />
      </span>
      <span className="absolute left-3 top-3 flex items-center gap-1.5 rounded-full bg-black/40 px-2.5 py-1 text-[11px] font-bold text-white backdrop-blur-sm">
        <Video className="h-3 w-3" strokeWidth={1.75} />
        Video
      </span>
    </button>
  )
}
