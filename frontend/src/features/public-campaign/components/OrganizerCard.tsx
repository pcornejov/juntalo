import { ShieldCheck } from 'lucide-react'

interface OrganizerCardProps {
  name: string
  campaignCount?: number
  isVerified?: boolean
}

// Iniciales para el avatar circular (p.ej. "Comité de Vecinos Los Aromos" -> "CV").
function initials(name: string) {
  const words = name.trim().split(/\s+/).filter(Boolean)
  const first = words[0]?.[0] ?? ''
  const second = words.find((w) => w[0] === w[0]?.toUpperCase() && w !== words[0])?.[0] ?? words[1]?.[0] ?? ''
  return (first + second).toUpperCase()
}

export function OrganizerCard({ name, campaignCount, isVerified }: OrganizerCardProps) {
  return (
    <div className="flex items-center gap-3 rounded-xl border border-border-default bg-bg-surface p-3">
      <div className="flex h-8 w-8 flex-none items-center justify-center rounded-full bg-accent-tint font-display text-xs font-bold text-brand-hover">
        {initials(name)}
      </div>
      <div className="flex-1 text-sm">
        <p className="font-semibold text-text-primary">{name}</p>
        <p className="text-xs text-text-secondary">
          Organiza{campaignCount ? ` · ${campaignCount} campañas anteriores` : ' esta colecta'}
        </p>
      </div>
      {isVerified && (
        <span className="flex flex-none items-center gap-1 rounded-full bg-accent-tint px-2.5 py-1 text-[11px] font-bold text-brand-hover">
          <ShieldCheck className="h-3 w-3" strokeWidth={1.75} />
          Identidad OK
        </span>
      )}
    </div>
  )
}
