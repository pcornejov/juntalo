import { Link } from 'react-router-dom'
import { MessageCircleHeart, Rocket, ShieldCheck, Users, HelpCircle } from 'lucide-react'
import { Footer, ThemeToggle } from '../shared/ui'

const steps = [
  {
    icon: Rocket,
    title: 'Crea tu campaña',
    text: 'Título, foto y meta (opcional). Se publica en menos de 2 minutos, sin trámites.',
  },
  {
    icon: MessageCircleHeart,
    title: 'Comparte el link',
    text: 'Por WhatsApp, redes o donde quieras — se ve con imagen y descripción, no como texto plano.',
  },
  {
    icon: Users,
    title: 'Recibe aportes',
    text: 'Cualquiera puede aportar desde el celular, sin crear una cuenta.',
  },
  {
    icon: ShieldCheck,
    title: 'Administra todo',
    text: 'Ves cada aporte, exportas un CSV y puedes reembolsar si algo sale mal — todo desde tu panel.',
  },
]

const faqs = [
  {
    q: '¿Juntalo cobra comisión?',
    a: 'Sí, un porcentaje sobre cada aporte confirmado. El monto neto recaudado siempre se muestra por separado del bruto.',
  },
  {
    q: '¿Qué significa el ícono de "verificada"?',
    a: 'Que el organizador confirmó su correo electrónico al crear la cuenta. No es una revisión manual de la campaña ni una garantía sobre el uso de los fondos — es una señal básica de identidad, no más.',
  },
  {
    q: '¿Quién administra el dinero recaudado?',
    a: 'Juntalo es una plataforma de software: no organiza las campañas ni administra los fondos. El organizador es responsable del contenido de su campaña y del uso de lo recaudado.',
  },
  {
    q: '¿Necesito cuenta para aportar?',
    a: 'No. Solo tu nombre y el monto — puedes aportar de forma anónima si prefieres que tu nombre no aparezca públicamente.',
  },
]

// Página estática "Cómo funciona" (inspirado en el centro de ayuda de
// plataformas similares) — sin backend: solo explica el flujo y responde
// las dudas más obvias antes de que alguien cree o comparta una campaña.
export function HowItWorksPage() {
  return (
    <div className="min-h-screen bg-bg-subtle">
      <header className="mx-auto flex max-w-3xl items-center justify-between px-6 py-5">
        <Link to="/" className="font-display text-lg font-bold tracking-tight">
          Juntalo
        </Link>
        <div className="flex items-center gap-3">
          <Link to="/explorar" className="text-sm font-medium text-text-secondary hover:text-text-primary">
            Explorar campañas
          </Link>
          <ThemeToggle />
        </div>
      </header>

      <div className="mx-auto max-w-3xl px-6 pb-16">
        <h1 className="mb-2 font-display text-2xl font-bold tracking-tight">Cómo funciona Juntalo</h1>
        <p className="mb-10 text-text-secondary">
          Crear una campaña y recibir aportes toma minutos, no días. Así funciona de principio a fin.
        </p>

        <div className="mb-12 grid gap-4 sm:grid-cols-2">
          {steps.map((s, i) => (
            <div key={s.title} className="rounded-xl border border-border-default bg-bg-surface p-4">
              <div className="mb-3 flex items-center gap-2">
                <span className="flex h-7 w-7 flex-none items-center justify-center rounded-full bg-accent-tint text-xs font-bold text-brand-hover">
                  {i + 1}
                </span>
                <s.icon className="h-4 w-4 text-brand-hover" strokeWidth={1.75} />
              </div>
              <h3 className="mb-1 font-display text-sm font-bold">{s.title}</h3>
              <p className="text-sm text-text-secondary">{s.text}</p>
            </div>
          ))}
        </div>

        <h2 className="mb-4 flex items-center gap-2 font-display text-lg font-bold tracking-tight">
          <HelpCircle className="h-5 w-5 text-brand-hover" strokeWidth={1.75} />
          Preguntas frecuentes
        </h2>
        <div className="space-y-4">
          {faqs.map((f) => (
            <div key={f.q} className="rounded-xl border border-border-default bg-bg-surface p-4">
              <p className="mb-1 font-semibold text-text-primary">{f.q}</p>
              <p className="text-sm text-text-secondary">{f.a}</p>
            </div>
          ))}
        </div>
      </div>

      <Footer />
    </div>
  )
}
