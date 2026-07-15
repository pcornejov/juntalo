import { Link } from 'react-router-dom'

// Bases del sitio: cubre lo esencial que un organizador necesita saber antes
// de crear una campaña — comisión y cuándo recibe la plata. No reemplaza
// asesoría legal (mismo criterio que PrivacyPage), pero deja por escrito lo
// que hoy solo vivía como una conversación de producto.
export function TermsPage() {
  return (
    <div className="mx-auto max-w-md space-y-4 p-6 text-text-primary">
      <h1 className="text-xl font-semibold">Bases del sitio</h1>

      <p className="text-sm text-text-secondary">Última actualización: julio de 2026.</p>

      <section className="space-y-2">
        <h2 className="font-medium">¿Cómo funciona Juntalo?</h2>
        <p className="text-sm text-text-secondary">
          Juntalo es una plataforma de software: no organiza las campañas, y el organizador de
          cada una es responsable de su contenido y del uso del dinero que recibe. En cuanto al
          dinero: los aportes se procesan a través de Webpay hacia una cuenta operada por Juntalo,
          que retiene los fondos temporalmente y los transfiere al organizador en cortes
          periódicos, descontando la comisión (ver detalle abajo). Juntalo no es dueño de esos
          fondos ni decide su destino final — solo los custodia brevemente como parte del
          procesamiento del pago.
        </p>
      </section>

      <section className="space-y-2">
        <h2 className="font-medium">Comisión</h2>
        <p className="text-sm text-text-secondary">
          Juntalo cobra una comisión del <strong className="text-text-primary">5%</strong> sobre
          cada aporte confirmado. Esta comisión cubre el costo de operar la plataforma y procesar
          los pagos. El monto neto que recibe el organizador (aporte menos comisión) siempre queda
          visible en su panel, por separado del monto bruto recaudado.
        </p>
      </section>

      <section className="space-y-2">
        <h2 className="font-medium">Transferencias al organizador</h2>
        <p className="text-sm text-text-secondary">
          Los fondos recaudados se transfieren a la cuenta del organizador en cortes periódicos
          (aproximadamente cada 15 días), descontando la comisión de Juntalo. Los aportes se
          consideran definitivos al confirmarse el pago — Juntalo es una plataforma para aportar a
          causas, no gestiona reembolsos.
        </p>
      </section>

      <Link to="/" className="inline-block text-sm text-brand underline">
        Volver al inicio
      </Link>
    </div>
  )
}
