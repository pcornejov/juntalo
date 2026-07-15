import { Link } from 'react-router-dom'

export function Footer() {
  return (
    <footer className="border-t border-border-default p-4 text-center text-xs text-text-secondary">
      <p>
        Juntalo no organiza las campañas publicadas por sus usuarios: cada organizador es
        responsable de su contenido y del uso de lo recaudado. Los aportes se procesan a través de
        Webpay y Juntalo los retiene temporalmente antes de transferirlos al organizador en cortes
        periódicos, descontando su comisión. Más detalles en{' '}
        <Link to="/bases" className="underline">
          bases del sitio
        </Link>
        .
      </p>
      <p className="mt-2 flex items-center justify-center gap-3">
        <Link to="/explorar" className="underline">
          Explorar campañas
        </Link>
        <Link to="/como-funciona" className="underline">
          Cómo funciona
        </Link>
        <Link to="/consejos" className="underline">
          Consejos
        </Link>
        <Link to="/bases" className="underline">
          Bases del sitio
        </Link>
        <Link to="/privacidad" className="underline">
          Política de privacidad
        </Link>
      </p>
    </footer>
  )
}
