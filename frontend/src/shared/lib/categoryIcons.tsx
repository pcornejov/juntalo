import {
  TriangleAlert,
  HeartPulse,
  GraduationCap,
  PartyPopper,
  Users,
  PawPrint,
  Sparkles,
  type LucideIcon,
} from 'lucide-react'

// Mapeo explícito (no `import * as icons from 'lucide-react'`) para no
// depender de que el nombre que manda el backend calce 1:1 con un export de
// la librería — un typo en el registro del backend degrada a Sparkles en
// vez de romper el render.
const iconsByName: Record<string, LucideIcon> = {
  TriangleAlert,
  HeartPulse,
  GraduationCap,
  PartyPopper,
  Users,
  PawPrint,
  Sparkles,
}

export function categoryIcon(name: string): LucideIcon {
  return iconsByName[name] ?? Sparkles
}
