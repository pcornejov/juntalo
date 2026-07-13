import { Link } from 'react-router-dom'
import { Button } from '../shared/ui'

export function NotFoundPage() {
  return (
    <div className="mx-auto flex min-h-screen max-w-sm flex-col items-center justify-center gap-4 p-6 text-center">
      <h1 className="text-xl font-semibold">Esta página no existe</h1>
      <p className="text-text-secondary">
        Revisa el link o vuelve al inicio.
      </p>
      <Link to="/">
        <Button>Ir al inicio</Button>
      </Link>
    </div>
  )
}
