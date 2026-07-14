import { Link } from 'react-router-dom'

export function Footer() {
  return (
    <footer className="border-t border-border-default p-4 text-center text-xs text-text-secondary">
      <p>
        Juntalo es una plataforma de software. No organiza ni administra las campañas publicadas
        por sus usuarios, quienes son responsables de su contenido y del uso de los fondos
        recaudados.
      </p>
      <p className="mt-2 flex items-center justify-center gap-3">
        <Link to="/explorar" className="underline">
          Explorar campañas
        </Link>
        <Link to="/como-funciona" className="underline">
          Cómo funciona
        </Link>
        <Link to="/privacidad" className="underline">
          Política de privacidad
        </Link>
      </p>
    </footer>
  )
}
