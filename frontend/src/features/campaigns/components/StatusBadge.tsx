import { Badge } from '../../../shared/ui'
import type { Campaign } from '../api'

const toneByStatus: Record<Campaign['status'], 'neutral' | 'success' | 'warning' | 'danger'> = {
  draft: 'neutral',
  active: 'success',
  paused: 'warning',
  finished: 'neutral',
  suspended: 'danger',
}

const labelByStatus: Record<Campaign['status'], string> = {
  draft: 'Borrador',
  active: 'Activa',
  paused: 'Pausada',
  finished: 'Finalizada',
  suspended: 'Suspendida',
}

export function StatusBadge({ status }: { status: Campaign['status'] }) {
  return <Badge tone={toneByStatus[status]}>{labelByStatus[status]}</Badge>
}
