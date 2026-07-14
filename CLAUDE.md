# Juntalo — guía para Claude Code

SaaS chileno para crear y compartir campañas de recaudación/venta online, estilo Notion/Stripe/Vercel/Linear. Este archivo documenta convenciones del repo para que una sesión nueva de Claude Code no tenga que re-derivarlas.

## Stack

- **Backend**: Go + Fiber, sqlc (queries "hand-written generated-style" — no hay binario `sqlc` instalado en muchos entornos, así que el código generado en `internal/infra/postgres/sqlc/` se edita a mano siguiendo el estilo exacto que sqlc produce), pgx/v5, PostgreSQL.
- **Frontend**: React + TypeScript + Vite + Tailwind CSS v4 + React Router + React Query (`useInfiniteQuery` para paginación).
- **Deploy**: Render.com (`render.yaml` en la raíz). Docker Compose solo para desarrollo local (`docker/`).

## Arquitectura backend

Capas dentro de `backend/internal/`:

- `domain/` — entidades y reglas de negocio puras (sin I/O): `campaign`, `contribution`, `payment`, `identity`, `money`, `audit`, `apperr`. Cada paquete de dominio suele tener su propio `validate.go`/`status.go` con funciones `Validate*`/`CanTransition` y tests unitarios sin mocks.
- `app/` — casos de uso, organizados por feature: `auth`, `campaigns`, `contributions`, `dashboard`, `files`. Un único archivo `app/ports.go` concentra **todas** las interfaces de repositorio/adaptador (`CampaignRepository`, `PaymentRepository`, `EmailSender`, `PaymentProvider`, etc.) — es el contrato de persistencia completo de un vistazo.
- `infra/` — implementaciones concretas: `postgres/repos` (implementa los `*Repository` de `ports.go`), `postgres/sqlc` (código estilo sqlc), `email` (Resend real + no-op), `payments/mock` (simulador de pasarela con modos `instant`/`deferred`), `storage/local`, `auth` (argon2id + JWT).
- `api/` — capa HTTP: `handlers/`, `dto/` (DTOs + `WriteError` que mapea `apperr` a códigos HTTP estables), `middleware/`, `router.go` (rutas bajo `/api/v1`), `server.go` (wiring completo: construye todos los repos/servicios/handlers y los conecta).

Patrones recurrentes a mantener:

- **Errores de dominio**: `apperr.New(code, message)` + `codeStatus` map en `dto/errors.go`. Un error no mapeado se reporta a Sentry y responde 500 genérico sin filtrar detalles internos.
- **Autorización "404 en vez de 403"**: un recurso ajeno a la organización del usuario responde `not_found`, nunca un error de permisos — no se confirma ni niega su existencia. Todos los casos de uso del dashboard filtran por `organization_id`.
- **Transacciones financieras**: `PaymentRepo.Refund` / `ConfirmByProviderRef` bloquean la fila con `FOR UPDATE` dentro de `pgx.BeginFunc` y transicionan pago + contribution atómicamente. Al escribir una query nueva que solo debe tocar `status`, usar una query dedicada (`UpdatePaymentStatusOnly`) — reusar la que también setea `confirmed_at`/`failed_at` los deja en NULL si no se pasan explícitamente (bug real que ya se cometió y corrigió una vez).
- **Idempotencia**: `Idempotency-Key` en aportes; `ConfirmByProviderRef` devuelve `(payment, transitioned bool, err)` para que un webhook reintentado no dispare notificaciones duplicadas.
- **Tokens de un solo uso**: mismo patrón para refresh tokens, password reset y email verification — hash aleatorio de 32 bytes, solo se guarda el hash, TTL, `used_at`.
- **Comisión**: se congela en el momento del pago (`commission_rate_applied`, `commission_amount`) — nunca se recalcula con la tarifa vigente después.

## Arquitectura frontend

`frontend/src/features/<feature>/` con `api.ts` (llamadas HTTP), `hooks/` (React Query), `pages/`, `components/`. `shared/ui/` tiene los componentes base (`Button`, `Input`, `Card`, `Badge`, etc. — sin librería de componentes externa). `shared/lib/` tiene utilidades puras (`clp.ts` formateo de moneda, `search.ts` normalización para búsqueda insensible a tildes).

## Testing — cómo verificar cambios en este repo

**Nunca dar un cambio por probado sin correrlo contra Postgres real.** El patrón usado en todas las sesiones:

```bash
sudo -u postgres psql -c "CREATE DATABASE juntalo_test_xyz;"
cd backend && nohup env \
  DATABASE_URL="postgres://postgres:postgres@localhost:5432/juntalo_test_xyz?sslmode=disable" \
  JWT_SECRET=testsecret RUN_MIGRATIONS_ON_BOOT=true \
  MOCK_PAYMENT_MODE=instant PORT=8080 \
  go run ./cmd/api > /tmp/server.log 2>&1 &
disown
```

- `MOCK_PAYMENT_MODE=instant` confirma pagos al toque (útil para probar flujos de aporte sin simular webhooks); `deferred` requiere `POST /api/v1/webhooks/payments/mock`.
- `EXPOSE_RESET_LINKS=true` **solo en local** — nunca en un deploy real, expone el link de reset de contraseña en la respuesta de la API en vez de mandarlo por email. Ya hubo un intento accidental de dejarlo en `true` en `render.yaml` apuntando al deploy público; el auto-mode classifier lo bloqueó.
- Frontend dev (`npm run dev -- --port 5173`) tiene un proxy `/api/v1 → localhost:8080` en `vite.config.ts` — el backend debe correr en el puerto 8080 para que funcione sin tocar config.
- **Gotcha de cookies**: sesión cross-port (5173↔8080) con `SameSite=Lax` bloquea que `page.goto()` con recarga completa mantenga la sesión en Playwright. Navegar autenticado dentro de la SPA con clicks (`page.click('text=...')`), no con `page.goto()` a rutas protegidas.
- Playwright: Chromium preinstalado en `/opt/pw-browsers/chromium`, paquete en `/opt/node22/lib/node_modules/playwright/index.js` (CommonJS: `import pkg from '...'; const { chromium } = pkg`). No correr `playwright install`.
- Siempre correr antes de dar algo por terminado: backend `gofmt -l .`, `go vet ./...`, `go test ./...`; frontend `npm run check` (lint + tsc), `npm run build`.
- Limpiar al terminar: matar el proceso del backend de prueba y `DROP DATABASE` de la base temporal.

## Deploy

Render.com, dos servicios (`juntalo-api`, `juntalo-web`) definidos en `render.yaml`. Variables sensibles (`JWT_SECRET`, `DATABASE_URL`, `RESEND_API_KEY`, `SENTRY_DSN`, `VITE_SENTRY_DSN`, `MOCK_WEBHOOK_SECRET`) van con `sync: false` — se cargan manualmente en el dashboard de Render, nunca se commitean. Verificación de deploy: pollear `juntalo-api.onrender.com/healthz` y el hash del bundle JS de `juntalo-web.onrender.com` hasta que cambie.

## Estado del producto (MVP + mejoras post-MVP)

Construido: registro/login, verificación de email, recuperación de contraseña, creación/edición/publicación de campañas (tipo "collection" habilitado), página pública con OG tags + QR + compartir WhatsApp, flujo de aporte público con pago mock, dashboard de organizador (participantes, export CSV, totales), notificaciones por email al organizador, reembolsos (parciales/totales), buscador/filtro (campañas y participantes), clonar campaña, auto-publicar en fecha programada (scheduler en background, tick cada minuto), dark mode, monitoreo de errores con Sentry.

Fuera de alcance del MVP (ver `docs/etapa-1-analisis-producto.md` para el detalle completo): pasarela de pago real, tipos de campaña adicionales (venta, evento, rifa), organizaciones multi-usuario, dominio personalizado, notificaciones automatizadas más allá de email transaccional.
