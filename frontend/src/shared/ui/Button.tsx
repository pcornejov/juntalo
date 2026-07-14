import type { ButtonHTMLAttributes } from 'react'

type Variant = 'primary' | 'secondary' | 'danger' | 'cta'

const variantClasses: Record<Variant, string> = {
  primary: 'bg-brand text-white hover:bg-brand-hover',
  secondary: 'bg-bg-subtle text-text-primary border border-border-default hover:bg-border-default',
  danger: 'bg-danger text-white hover:opacity-90',
  // Reservado para la acción principal de la página (p.ej. "Aportar"): usa
  // directamente el tono fuerte del acento en vez de solo mostrarlo en
  // hover, para que destaque como el único llamado a la acción de verdad.
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
