import { Button, Input, Card, Progress, Badge } from '../shared/ui'

export function ComponentsShowcasePage() {
  return (
    <div className="mx-auto max-w-xl space-y-6 p-6">
      <h1 className="text-2xl font-semibold">Juntalo — Design System</h1>

      <Card className="space-y-4">
        <div className="flex gap-2">
          <Button>Aportar</Button>
          <Button variant="secondary">Compartir</Button>
          <Button variant="danger">Eliminar</Button>
        </div>

        <Input placeholder="Nombre completo" />

        <div className="flex gap-2">
          <Badge tone="success">Activa</Badge>
          <Badge tone="warning">Pausada</Badge>
          <Badge tone="danger">Suspendida</Badge>
          <Badge>Borrador</Badge>
        </div>

        <Progress value={350000} max={1000000} />
      </Card>
    </div>
  )
}
