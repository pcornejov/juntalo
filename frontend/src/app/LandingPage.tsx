import { Navigate, Link } from 'react-router-dom'
import { MessageCircleHeart, Rocket, ShieldCheck, Sparkles, ArrowRight, QrCode, Users } from 'lucide-react'
import { useAuth } from '../features/auth/hooks/useAuth'
import { Button, Progress, Footer, ThemeToggle } from '../shared/ui'

// Clases completas y literales a propósito: Tailwind arma su CSS
// escaneando el código fuente en busca de nombres de clase exactos, así
// que construir "bg-gradient-to-" + variable en runtime no generaría el
// estilo en el build de producción.
const gradientClasses = {
  br: 'bg-gradient-to-br',
  tr: 'bg-gradient-to-tr',
  b: 'bg-gradient-to-b',
} as const

const showcaseCards = [
  { label: 'Colecta', title: 'Techo nuevo para la sede vecinal', amount: 340000, goal: 500000, people: 28, angle: 'br' },
  { label: 'Venta', title: 'Empanadas para el viaje de curso', amount: 210000, goal: 300000, people: 41, angle: 'tr' },
  { label: 'Evento', title: 'Bono para la fiesta de fin de año', amount: 180000, goal: 250000, people: 19, angle: 'b' },
] as const satisfies { label: string; title: string; amount: number; goal: number; people: number; angle: keyof typeof gradientClasses }[]

const avatarInitials = ['MJ', 'PC', 'FS', 'AV', 'RT']

const steps = [
  {
    icon: Rocket,
    title: 'Crea tu campaña',
    text: 'Título, foto y meta. Menos de 2 minutos y ya está publicada.',
  },
  {
    icon: MessageCircleHeart,
    title: 'Comparte por WhatsApp',
    text: 'Un link con vista previa bonita — se ve como una tarjeta, no como texto plano.',
  },
  {
    icon: Users,
    title: 'Recibe aportes al toque',
    text: 'Tu gente aporta desde el celular sin crear cuenta. Tú ves todo en un panel simple.',
  },
]

const highlights = [
  {
    icon: ShieldCheck,
    title: 'Transparente',
    text: 'Cada aporte, cada comisión y cada reembolso quedan a la vista — sin sorpresas al final.',
  },
  {
    icon: QrCode,
    title: 'Pensado para compartir',
    text: 'Link corto, preview con imagen y QR generado al vuelo para cuando el link no llega solo.',
  },
  {
    icon: Sparkles,
    title: 'Hecho para Chile',
    text: 'Montos en pesos, sin decimales raros, y una experiencia que funciona bien en cualquier celular.',
  },
]

// Landing pública en '/': invita a crear cuenta antes de pedir login. Los
// usuarios ya autenticados no deberían ver esto — se los manda directo al
// dashboard, igual que cualquier producto que usa la home como pitch de
// venta y no como pantalla de trabajo.
export function LandingPage() {
  const { user, isLoading } = useAuth()

  if (isLoading) {
    return <div className="p-6 text-center text-text-secondary">Cargando…</div>
  }
  if (user) {
    return <Navigate to="/dashboard" replace />
  }

  return (
    <div className="min-h-screen bg-bg-subtle">
      <header className="mx-auto flex max-w-5xl items-center justify-between px-6 py-5">
        <span className="font-display text-lg font-bold tracking-tight">Juntalo</span>
        <div className="flex items-center gap-3">
          <Link
            to="/explorar"
            className="text-sm font-medium text-text-secondary hover:text-text-primary"
          >
            Explorar campañas
          </Link>
          <Link
            to="/como-funciona"
            className="text-sm font-medium text-text-secondary hover:text-text-primary"
          >
            Cómo funciona
          </Link>
          <ThemeToggle />
          <Link to="/login" className="text-sm font-medium text-text-secondary hover:text-text-primary">
            Ingresar
          </Link>
        </div>
      </header>

      <section className="relative overflow-hidden">
        {/* Blobs decorativos + grilla de puntos: sin fotos random (probamos
            picsum y salían fotos sin relación, tipo cerros o escaleras) —
            esto se ve intencional en cualquier tema y no depende de bajar
            ni alojar imágenes de terceros. */}
        <div
          aria-hidden
          className="pointer-events-none absolute inset-0 opacity-[0.4]"
          style={{
            backgroundImage: 'radial-gradient(var(--color-border-default) 1px, transparent 1px)',
            backgroundSize: '28px 28px',
            maskImage: 'linear-gradient(to bottom, black, transparent)',
          }}
        />
        <div
          aria-hidden
          className="pointer-events-none absolute -left-24 -top-24 h-72 w-72 rounded-full bg-brand opacity-20 blur-3xl"
        />
        <div
          aria-hidden
          className="pointer-events-none absolute -right-16 top-20 h-64 w-64 rounded-full bg-brand-hover opacity-20 blur-3xl"
        />

        <div className="relative mx-auto max-w-3xl px-6 pb-10 pt-10 text-center sm:pt-16">
          <span className="mb-5 inline-flex items-center gap-1.5 rounded-full bg-accent-tint px-3 py-1 text-xs font-semibold text-brand-hover">
            <Sparkles className="h-3.5 w-3.5" strokeWidth={1.75} />
            Crea y comparte en minutos
          </span>
          <h1 className="text-balance font-display text-4xl font-bold leading-[1.1] tracking-tight sm:text-5xl">
            Junta plata con tu gente,
            <br className="hidden sm:block" /> sin vueltas
          </h1>
          <p className="mx-auto mt-5 max-w-lg text-balance text-base leading-relaxed text-text-secondary sm:text-lg">
            Colectas, ventas y rifas con un link que se ve bien en WhatsApp. Tú organizas, ellos
            aportan desde el celular — sin apps, sin cuentas, sin fricción.
          </p>
          <div className="mt-8 flex flex-col items-center justify-center gap-3 sm:flex-row">
            <Link to="/register" className="w-full sm:w-auto">
              <Button variant="cta" className="flex w-full items-center justify-center gap-2 sm:w-auto">
                Crear mi campaña gratis
                <ArrowRight className="h-4 w-4" strokeWidth={1.75} />
              </Button>
            </Link>
            <Link
              to="/login"
              className="text-sm font-medium text-text-secondary underline-offset-4 hover:text-text-primary hover:underline"
            >
              Ya tengo cuenta
            </Link>
          </div>

          <div className="mt-8 flex items-center justify-center gap-3">
            <div className="flex -space-x-2">
              {avatarInitials.map((initials, i) => (
                <div
                  key={initials}
                  className="flex h-8 w-8 items-center justify-center rounded-full border-2 border-bg-subtle text-[10px] font-bold text-white"
                  style={{ backgroundColor: i % 2 === 0 ? 'var(--color-brand)' : 'var(--color-brand-hover)' }}
                >
                  {initials}
                </div>
              ))}
            </div>
            <p className="text-xs text-text-secondary">
              Familias, cursos y juntas de vecinos ya están juntando plata así
            </p>
          </div>
        </div>
      </section>

      <section className="mx-auto max-w-5xl px-6 pb-16">
        <div className="grid gap-4 sm:grid-cols-3">
          {showcaseCards.map((c) => (
            <div
              key={c.label}
              className="overflow-hidden rounded-2xl border border-border-default bg-bg-surface shadow-sm"
            >
              <div className={`relative h-24 ${gradientClasses[c.angle]} from-brand to-brand-hover`}>
                <span className="absolute left-3 top-3 rounded-full bg-white/20 px-2 py-0.5 text-[10px] font-semibold uppercase tracking-wide text-white backdrop-blur-sm">
                  {c.label}
                </span>
              </div>
              <div className="space-y-3 p-4">
                <p className="font-display text-sm font-bold">{c.title}</p>
                <Progress value={c.amount} max={c.goal} />
                <div className="flex items-center justify-between text-xs text-text-secondary">
                  <span>
                    <strong className="text-text-primary">${c.amount.toLocaleString('es-CL')}</strong> de $
                    {c.goal.toLocaleString('es-CL')}
                  </span>
                  <span className="flex items-center gap-1">
                    <Users className="h-3 w-3" strokeWidth={1.75} />
                    {c.people}
                  </span>
                </div>
              </div>
            </div>
          ))}
        </div>
        <p className="mt-3 text-center text-xs text-text-secondary">
          Así se ven las campañas cuando las comparten — claras y directo al grano.
        </p>
      </section>

      <section className="mx-auto max-w-5xl px-6 pb-16">
        <h2 className="mb-8 text-center font-display text-2xl font-bold tracking-tight">
          Tres pasos, nada más
        </h2>
        <div className="grid gap-4 sm:grid-cols-3">
          {steps.map((step, i) => (
            <div key={step.title} className="rounded-2xl border border-border-default bg-bg-surface p-5">
              <div className="mb-3 flex h-9 w-9 items-center justify-center rounded-full bg-accent-tint text-sm font-bold text-brand-hover">
                {i + 1}
              </div>
              <step.icon className="mb-2 h-5 w-5 text-brand" strokeWidth={1.75} />
              <p className="mb-1 font-display text-sm font-bold">{step.title}</p>
              <p className="text-sm leading-relaxed text-text-secondary">{step.text}</p>
            </div>
          ))}
        </div>
      </section>

      <section className="mx-auto max-w-5xl px-6 pb-20">
        <div className="grid gap-4 sm:grid-cols-3">
          {highlights.map((h) => (
            <div key={h.title} className="space-y-2">
              <h.icon className="h-5 w-5 text-brand" strokeWidth={1.75} />
              <p className="font-display text-sm font-bold">{h.title}</p>
              <p className="text-sm leading-relaxed text-text-secondary">{h.text}</p>
            </div>
          ))}
        </div>
      </section>

      <section className="border-t border-border-default bg-bg-surface px-6 py-16 text-center">
        <h2 className="mb-3 font-display text-2xl font-bold tracking-tight">
          ¿Listo para juntar la tuya?
        </h2>
        <p className="mx-auto mb-6 max-w-md text-text-secondary">
          Crear una cuenta toma menos de un minuto y tu campaña queda publicada al tiro.
        </p>
        <Link to="/register">
          <Button variant="cta" className="inline-flex items-center gap-2">
            Crear mi campaña gratis
            <ArrowRight className="h-4 w-4" strokeWidth={1.75} />
          </Button>
        </Link>
      </section>

      <Footer />
    </div>
  )
}
