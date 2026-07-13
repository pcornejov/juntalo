# Juntalo

Plataforma para crear campañas de recaudación y venta online. Ver el diseño completo en [`docs/`](./docs):

1. [Análisis de producto](./docs/etapa-1-analisis-producto.md)
2. [Arquitectura](./docs/etapa-2-arquitectura.md)
3. [Modelo de datos](./docs/etapa-3-modelo-datos.md)
4. [Endpoints REST](./docs/etapa-4-endpoints.md)
5. [Estructura de carpetas](./docs/etapa-5-estructura.md)
6. [Roadmap técnico](./docs/etapa-6-roadmap.md)

## Stack

- **Backend**: Go + Fiber + sqlc + PostgreSQL
- **Frontend**: React + TypeScript + Vite + Tailwind CSS + React Query
- **Infra**: Docker Compose, Caddy, Cloudflare-ready

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
