import {
  HeartPulse,
  Flame,
  PawPrint,
  Home,
  GraduationCap,
  Trophy,
  PartyPopper,
  Flower2,
  Package,
  type LucideIcon,
} from 'lucide-react'

export interface CampaignArchetype {
  id: string
  label: string
  icon: LucideIcon
  title: string
  description: string
}

// Plantillas de arranque rápido dentro del tipo de campaña "colecta" (el
// único habilitado hoy) — no son tipos de campaña nuevos, solo pre-llenan
// título y descripción con algo editable en vez de una hoja en blanco.
// Cubren los motivos más comunes para juntar plata en Chile.
export const campaignArchetypes: CampaignArchetype[] = [
  {
    id: 'salud',
    label: 'Salud / accidente',
    icon: HeartPulse,
    title: 'Ayuda con gastos médicos',
    description:
      'Estamos juntando plata para cubrir los gastos médicos de [nombre] tras [qué pasó]. Cualquier aporte ayuda a cubrir tratamiento, exámenes y recuperación.',
  },
  {
    id: 'emergencia',
    label: 'Emergencia / desastre',
    icon: Flame,
    title: 'Ayuda de emergencia',
    description:
      'Familias del sector fueron afectadas por [incendio/inundación/otro]. Con lo recaudado cubrimos artículos básicos y reposición de enseres mientras se normaliza la situación.',
  },
  {
    id: 'mascota',
    label: 'Mascotas',
    icon: PawPrint,
    title: 'Ayuda para nuestras mascotas',
    description:
      'Organizamos una jornada/tratamiento para perros y gatos del sector. Con tu aporte cubrimos vacunas, consultas veterinarias e insumos.',
  },
  {
    id: 'vecinos',
    label: 'Junta de vecinos',
    icon: Home,
    title: 'Mejoras para nuestro barrio',
    description:
      'Estamos juntando fondos para [arreglar/equipar] la sede/espacio común del barrio. Cada aporte se traduce directo en mejoras para todos los vecinos.',
  },
  {
    id: 'estudio',
    label: 'Viaje de curso',
    icon: GraduationCap,
    title: 'Viaje de estudios del curso',
    description:
      'Estamos juntando plata para financiar el viaje de estudios del curso. Cada aporte ayuda a bajar el costo por alumno.',
  },
  {
    id: 'deporte',
    label: 'Equipo deportivo',
    icon: Trophy,
    title: 'Apoyo para el equipo',
    description:
      'Necesitamos cubrir uniformes, arriendo de cancha e inscripción a torneo. Tu aporte ayuda a que el equipo pueda competir.',
  },
  {
    id: 'celebracion',
    label: 'Celebración',
    icon: PartyPopper,
    title: 'Aporta a la celebración',
    description:
      'Estamos organizando [cumpleaños/matrimonio/despedida] y juntando plata entre todos para hacerlo posible. ¡Todos los aportes suman!',
  },
  {
    id: 'funeral',
    label: 'Funeral',
    icon: Flower2,
    title: 'Apoyo para gastos funerarios',
    description:
      'Estamos juntando plata para cubrir los gastos funerarios de [nombre]. Agradecemos cualquier aporte en este momento difícil.',
  },
  {
    id: 'emprendimiento',
    label: 'Preventa / emprendimiento',
    icon: Package,
    title: 'Preventa de [producto]',
    description:
      'Estamos validando demanda antes de fabricar/producir. Tu aporte reserva tu unidad y nos ayuda a dar el siguiente paso.',
  },
]
