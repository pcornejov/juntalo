import { Link } from 'react-router-dom'

// Aviso mínimo de privacidad (Ley 21.719) para aportantes sin cuenta —
// Etapa 1 §1 riesgo 7. No reemplaza asesoría legal; cubre lo esencial:
// qué datos se piden, para qué se usan y con quién no se comparten.
export function PrivacyPage() {
  return (
    <div className="mx-auto max-w-md space-y-4 p-6 text-text-primary">
      <h1 className="text-xl font-semibold">Política de privacidad</h1>

      <p className="text-sm text-text-secondary">
        Última actualización: julio de 2026.
      </p>

      <section className="space-y-2">
        <h2 className="font-medium">¿Qué datos recolectamos?</h2>
        <p className="text-sm text-text-secondary">
          Cuando creas una cuenta: tu nombre, correo y contraseña. Cuando aportas a una campaña
          como participante: tu nombre, el monto aportado y, opcionalmente, un mensaje. No te
          pedimos datos de tu método de pago — eso lo procesa directamente la pasarela de pago.
        </p>
      </section>

      <section className="space-y-2">
        <h2 className="font-medium">¿Para qué los usamos?</h2>
        <p className="text-sm text-text-secondary">
          Solo para procesar tu aporte, mostrarlo en la campaña (salvo que elijas aportar de
          forma anónima) y contactarte si es estrictamente necesario para resolver un problema
          con tu aporte.
        </p>
      </section>

      <section className="space-y-2">
        <h2 className="font-medium">¿Con quién los compartimos?</h2>
        <p className="text-sm text-text-secondary">
          No vendemos ni compartimos tus datos con terceros para fines de marketing. El
          organizador de la campaña puede ver el detalle de los aportes, incluso los anónimos,
          para gestionar su campaña.
        </p>
      </section>

      <section className="space-y-2">
        <h2 className="font-medium">Tus derechos</h2>
        <p className="text-sm text-text-secondary">
          Puedes pedir que corrijamos o eliminemos tus datos escribiendo a{' '}
          <a href="mailto:hola@juntalo.cl" className="text-brand underline">
            hola@juntalo.cl
          </a>
          .
        </p>
      </section>

      <Link to="/" className="inline-block text-sm text-brand underline">
        Volver al inicio
      </Link>
    </div>
  )
}
