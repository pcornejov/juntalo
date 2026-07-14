import { useRef, useState, type UIEvent } from 'react'

interface CarouselProps {
  images: string[]
  alt: string
  className?: string
}

// Carrusel con scroll-snap nativo (sin librería): funciona con swipe táctil
// real y no le agrega peso al bundle — clave en una página que se abre
// desde el navegador in-app de WhatsApp.
export function Carousel({ images, alt, className }: CarouselProps) {
  const trackRef = useRef<HTMLDivElement>(null)
  const [active, setActive] = useState(0)

  function handleScroll(e: UIEvent<HTMLDivElement>) {
    const el = e.currentTarget
    const index = Math.round(el.scrollLeft / el.clientWidth)
    setActive(index)
  }

  function goTo(index: number) {
    const el = trackRef.current
    if (!el) return
    el.scrollTo({ left: index * el.clientWidth, behavior: 'smooth' })
  }

  if (images.length <= 1) {
    return (
      <img
        src={images[0]}
        alt={alt}
        className={className ?? 'h-56 w-full object-cover'}
      />
    )
  }

  return (
    <div className="relative">
      <div
        ref={trackRef}
        onScroll={handleScroll}
        className="flex snap-x snap-mandatory overflow-x-auto scroll-smooth [-ms-overflow-style:none] [scrollbar-width:none] [&::-webkit-scrollbar]:hidden"
      >
        {images.map((src, i) => (
          <img
            key={src}
            src={src}
            alt={i === 0 ? alt : `${alt} (${i + 1})`}
            className={`w-full flex-none snap-start ${className ?? 'h-56 object-cover'}`}
          />
        ))}
      </div>
      <div className="pointer-events-none absolute inset-x-0 bottom-3 z-10 flex justify-center gap-1.5">
        {images.map((src, i) => (
          <button
            key={src}
            type="button"
            onClick={() => goTo(i)}
            aria-label={`Ir a la foto ${i + 1}`}
            className={`pointer-events-auto h-1.5 rounded-full transition-all ${
              i === active ? 'w-4 bg-white' : 'w-1.5 bg-white/50'
            }`}
          />
        ))}
      </div>
    </div>
  )
}
