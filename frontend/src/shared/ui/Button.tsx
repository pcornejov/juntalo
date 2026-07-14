import type { ButtonHTMLAttributes } from 'react'

type Variant = 'primary' | 'secondary' | 'danger' | 'cta'

const variantClasses: Record<Variant, string> = {
  // bg-brand-hover, no bg-brand: --color-brand queda reservado para
  // progreso/cifras/badges (ver index.css) — usarlo de fondo de botón
  // rompía esa regla y bajaba el contraste con el texto blanco.
  primary: 'bg-brand-hover text-white hover:opacity-90',
  secondary: 'bg-bg-subtle text-text-primary border border-border-default hover:bg-border-default',
  danger: 'bg-danger text-white hover:opacity-90',
  // Reservado para la acción principal de la página (p.ej. "Aportar"): suma
  // tipografía Sora y sombra para destacar como el único llamado a la
  // acción de verdad, por sobre un primary común.
  cta: 'bg-brand-hover text-white font-display shadow-lg shadow-brand-hover/25 hover:opacity-90',
}

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: Variant
}

export function Button({ variant = 'primary', className = '', ...props }: ButtonProps) {
  return (
    <button
      className={`rounded-lg px-4 py-2 font-medium transition-colors disabled:opacity-50 disabled:cursor-not-allowed ${variantClasses[variant]} ${className}`}
      {...props}
    />
  )
}
