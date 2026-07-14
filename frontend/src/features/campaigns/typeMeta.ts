import { HeartHandshake, ShoppingBag, CalendarDays, Ticket, type LucideIcon } from 'lucide-react'

// El backend no manda ícono ni label corto por tipo (solo key/name/cta/unit
// vía /meta/campaign-types) — son pocos y estables, así que el mapeo visual
// vive acá en vez de ida y vuelta a la API para decorar un badge.
export const typeIcons: Record<string, LucideIcon> = {
  collection: HeartHandshake,
  sale: ShoppingBag,
  event: CalendarDays,
  raffle: Ticket,
}

export const typeLabels: Record<string, string> = {
  collection: 'Colecta',
  sale: 'Venta',
  event: 'Evento',
  raffle: 'Rifa',
  course: 'Curso',
  presale: 'Preventa',
}
