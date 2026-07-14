# Juntalo

SaaS chileno para crear y compartir campañas de recaudación/venta online (colectas, "vaquitas"), con página pública, aportes con pago simulado, reembolsos, dashboard de organizador y auto-publicación programada. Ver el diseño completo en [`docs/`](./docs):

1. [Análisis de producto](./docs/etapa-1-analisis-producto.md)
2. [Arquitectura](./docs/etapa-2-arquitectura.md)
3. [Modelo de datos](./docs/etapa-3-modelo-datos.md)
4. [Endpoints REST](./docs/etapa-4-endpoints.md)
5. [Estructura de carpetas](./docs/etapa-5-estructura.md)
6. [Roadmap técnico](./docs/etapa-6-roadmap.md)

Para convenciones de desarrollo y cómo verificar cambios (patrones de testing, gotchas conocidos), ver [`CLAUDE.md`](./CLAUDE.md).

## Funcionalidades

- Registro/login con verificación de email y recuperación de contraseña.
- Crear, editar, publicar/pausar/finalizar campañas; auto-publicar en una fecha programada; clonar una campaña existente.
- Página pública con Open Graph, código QR y botones para compartir por WhatsApp.
- Aporte público sin necesidad de cuenta, con pago simulado (`MockPaymentProvider`).
- Dashboard de organizador: participantes con buscador/filtro y paginación, export a CSV, totales bruto/neto, reembolsos parciales o totales.
- Notificaciones por email al organizador ante cada aporte confirmado.
- Modo oscuro y monitoreo de errores con Sentry.

## Stack

- **Backend**: Go + Fiber + sqlc + PostgreSQL
- **Frontend**: React + TypeScript + Vite + Tailwind CSS + React Query
- **Deploy**: Render.com (`render.yaml`); Docker Compose para desarrollo local

## Probar sin instalar nada

¿Quieres ver el MVP funcionando (incluida la preview de WhatsApp desde tu celular) sin levantar nada localmente? Ver [`docs/deploy-render.md`](./docs/deploy-render.md) — deploy gratis en Render en ~10 minutos.

## Desarrollo

Requisitos: Docker y Docker Compose.

```bash
cp .env.example .env
make dev          # levanta postgres + backend (hot-reload) + frontend
make migrate       # aplica las migraciones
make seed          # datos de desarrollo
```

- Backend: http://localhost:8080 (`/healthz`)
- Frontend: http://localhost:5173

## Comandos útiles

```bash
make check   # lint + test + typecheck de ambas apps
make build   # build de producción de ambas apps
```

Cada app tiene su propio `Makefile`/`package.json` con comandos específicos — ver `backend/Makefile` y `frontend/package.json`.
