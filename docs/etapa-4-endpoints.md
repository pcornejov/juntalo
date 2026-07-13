# Juntalo — Etapa 4: Endpoints REST

> Estado: **propuesta, pendiente de aprobación**. Depende de Etapas 1–3 aprobadas. Este documento es la fuente del spec OpenAPI (Etapa 2 §2.5: el contrato se formaliza con OpenAPI y genera los tipos TS del frontend).

---

## 1. Convenciones

- **Base path**: `/api/v1`. Versionado en la URL desde el día 1: barato ahora, imposible de retrofitear sin romper clientes.
- **Formato**: JSON; `snake_case` en payloads (espejo del schema SQL, sin capa de renombrado).
- **IDs**: uuid en rutas privadas; **slug** en rutas públicas (`/c/:slug` no expone uuids).
- **Auth**: `Authorization: Bearer <access_token>` (JWT 15 min). Refresh token en cookie httpOnly `jt_refresh` — el access token nunca se persiste en el navegador.
- **Errores** — envelope único en todo el API:
```json
{ "error": { "code": "campaign_not_active", "message": "…", "details": {} } }
```
  `code` es estable y documentado (el frontend traduce códigos, no parsea mensajes). HTTP: 400 validación · 401 sin/expirado token · 403 sin permiso · 404 no existe **o no es tuyo** (no filtrar existencia) · 409 conflicto de estado/idempotencia · 422 regla de negocio · 429 rate limit.
- **Paginación**: cursor-based (`?cursor=<uuidv7>&limit=20`) — uuidv7 es ordenable por tiempo, el cursor es el último id visto. Respuesta: `{ "items": [...], "next_cursor": "…" | null }`. Offset se descarta: se degrada con volumen y se rompe con inserciones concurrentes.
- **Idempotencia**: `POST /contributions` exige header `Idempotency-Key` (uuid generado por el cliente al montar el formulario). Repetir la key ⇒ misma respuesta, cero doble cargo (respaldado por el UNIQUE de Etapa 3 §5).

---

## 2. Auth — `/api/v1/auth`

| Método y ruta | Auth | Descripción |
|---|---|---|
| `POST /auth/register` | — | Crea user + identity + org personal + membership (una tx, Etapa 3 §2). Body: `email, password, full_name`. → `201 {user, access_token}` + cookie refresh. Password ≥ 8 chars validado con reglas simples (no ceremonias de símbolos). |
| `POST /auth/login` | — | Body: `email, password`. → `200 {user, access_token}` + cookie refresh. Error genérico `invalid_credentials` (no revelar si el email existe). |
| `POST /auth/refresh` | cookie | Rota el refresh token (el usado se revoca). → `200 {access_token}` + cookie nueva. Token revocado/expirado ⇒ 401 `session_expired`. |
| `POST /auth/logout` | cookie | Revoca el refresh token actual. → `204`. |
| `GET /auth/me` | Bearer | → `200 {user, organization}` (org personal del MVP). |

Rate limits: register `5/h por IP` · login `10/15min por IP` (+ backoff progresivo por email) · refresh `60/h`.

---

## 3. Campañas (privado, dueño) — `/api/v1/campaigns`

Toda ruta privada opera **scoped a la organización del usuario autenticado**: un uuid ajeno responde 404, nunca 403 (no filtrar existencia).

| Método y ruta | Descripción |
|---|---|
| `GET /campaigns` | Lista del dashboard. Filtros: `?status=`, paginación cursor. Cada item incluye `totals` (de la vista `campaign_totals`). |
| `POST /campaigns` | Crea en `draft`. Body: `type_key, title, description?, goal_amount?, starts_at?, ends_at?, settings?`. El slug **lo genera el servidor** (slugify + sufijo anti-colisión); el cliente no lo elige en MVP. → `201` con la campaña completa. |
| `GET /campaigns/:id` | Detalle + `totals`. |
| `PATCH /campaigns/:id` | Edición parcial de campos editables. Reglas por estado en dominio (p.ej. `goal_amount` no baja de lo ya recaudado ⇒ 422 `goal_below_raised`). |
| `POST /campaigns/:id/publish` | `draft → active`. Valida reglas del tipo (registro declarativo). Respuesta incluye `public_url` — es lo que el flujo "<2 min" muestra para compartir. |
| `POST /campaigns/:id/pause` · `/resume` · `/finish` | Transiciones `active→paused`, `paused→active`, `active|paused→finished`. Transición ilegal ⇒ 409 `invalid_status_transition` con `details.allowed`. `suspended` NO tiene endpoint: solo plataforma vía DB/admin futuro. |
| `DELETE /campaigns/:id` | Solo `draft` sin pagos ⇒ soft-delete. Cualquier otro estado ⇒ 422 `campaign_not_deletable` (se usa `finish`). |
| `GET /campaigns/:id/contributions` | Participantes para el dashboard (datos completos, incluidos anónimos — riesgo 15). Filtros `?status=`, cursor. |
| `GET /campaigns/:id/contributions/export` | CSV (`text/csv`, streaming). Columnas: fecha, nombre, email, teléfono, monto, estado, anónimo, monto_reembolsado. Sin paginación: stream directo de la query. |

**Transiciones como sub-recursos POST y no `PATCH {status}`**: cada transición tiene validaciones y efectos propios (publish valida el tipo y devuelve la URL; finish cierra aportes). Un PATCH genérico de status obligaría a un switch de reglas oculto — las acciones explícitas son el API del dominio.

---

## 4. Público — sin auth

| Método y ruta | Descripción |
|---|---|
| `GET /api/v1/public/campaigns/:slug` | Página pública. → `200 {campaign_public, totals, recent_contributions}`. `campaign_public` excluye datos internos; `recent_contributions` (últimas 20 confirmadas) muestra `full_name` o `"Anónimo"` según `is_anonymous`, nunca email/teléfono. Solo `active/paused/finished` (draft/suspended ⇒ 404). Cache-Control: `public, max-age=30` — la defensa ante viralidad (Etapa 2: la viralidad se resuelve con caché, no con el runtime). |
| `GET /c/:slug` | **No es JSON**: HTML con meta tags Open Graph para bots/previews de WhatsApp y redirect/hydrate a la SPA para navegadores (Etapa 2 §1). |
| `POST /api/v1/public/campaigns/:slug/contributions` | **El endpoint más importante del producto.** Header `Idempotency-Key` obligatorio. Body: `full_name, email?, phone?, amount, is_anonymous?, message?`. Flujo: valida campaña `active` y ventana de fechas → crea contributor + contribution(`pending`) + payment(`pending`) con snapshot de comisión → `Provider.CreateIntent` → `201 {contribution_id, payment: {status, redirect_url?}}`. Con el mock la confirmación llega vía webhook simulado; con pasarela real, `redirect_url` lleva al checkout. Errores: 422 `campaign_not_active` / `amount_out_of_range` · 409 `duplicate_contribution` (misma key ⇒ devuelve la original, 200). Rate limit agresivo: `10/min por IP`. |
| `GET /api/v1/public/contributions/:id/status` | Polling post-pago para la pantalla de confirmación: → `{status}` (`pending/confirmed/failed`). El id (uuid) es el capability token: solo quien hizo el aporte lo tiene. `2/s por IP`. |

---

## 5. Webhooks — `/api/v1/webhooks`

| Método y ruta | Descripción |
|---|---|
| `POST /webhooks/payments/:provider` | Confirmación asíncrona de pagos (Etapa 2 §2.3: existe desde el MVP; el mock lo invoca a sí mismo). Flujo: valida firma del provider (el mock firma con secret local — el pipeline de verificación queda ensayado) → busca payment por `provider_ref` → transición de estado en dominio dentro de una tx (payment + contribution) → `200`. **Idempotente**: evento repetido ⇒ 200 sin efecto. Evento desconocido ⇒ 200 y log (no 4xx: las pasarelas reintentan ante errores y amplifican el problema). |

---

## 6. Archivos — `/api/v1/files`

| Método y ruta | Auth | Descripción |
|---|---|---|
| `POST /files` | Bearer | `multipart/form-data`: `file` + `kind=campaign_cover`. Valida allowlist mime (jpeg/png/webp) + ≤ 5MB. → `201 {id, url}`. El cliente luego asocia con `PATCH /campaigns/:id {cover_file_id}`. Sube primero, asocia después: el form de creación no se bloquea por la imagen (riesgo 11: opcional, <2 min). `20/h por usuario`. |

(Con R2 en el futuro, este endpoint puede evolucionar a presigned URLs sin cambiar el contrato: `POST /files` pasaría a devolver la URL de subida — mismo recurso.)

---

## 7. Metadatos — `/api/v1/meta`

| Método y ruta | Auth | Descripción |
|---|---|---|
| `GET /meta/campaign-types` | Bearer | Registro declarativo serializado (solo `enabled`): labels, fields, rules. El form de creación se renderiza desde aquí — agregar "Venta" en el backend actualiza el frontend sin deploy coordinado. |
| `GET /healthz` | — | Liveness + ping a Postgres. Para Caddy/monitoring. Fuera de `/api/v1`. |

---

## 8. Resumen de superficie

**18 endpoints** — deliberadamente pocos:

```
Auth        POST   /auth/register · /auth/login · /auth/refresh · /auth/logout
            GET    /auth/me
Campañas    GET    /campaigns · /campaigns/:id · /campaigns/:id/contributions
            GET    /campaigns/:id/contributions/export
            POST   /campaigns · /campaigns/:id/publish · pause · resume · finish
            PATCH  /campaigns/:id
            DELETE /campaigns/:id
Público     GET    /public/campaigns/:slug · /public/contributions/:id/status
            GET    /c/:slug (HTML OG)
            POST   /public/campaigns/:slug/contributions
Webhooks    POST   /webhooks/payments/:provider
Files       POST   /files
Meta        GET    /meta/campaign-types · /healthz
```

Lo que NO existe y por qué: endpoints de organizaciones (org personal implícita en MVP), de usuarios/perfil (solo `me`), de refunds (operación de soporte vía admin futuro, el modelo lo soporta), de listado público de campañas (Juntalo distribuye por link compartido, no es marketplace — decisión de producto, no técnica).

---

## Siguiente paso

Aprobados los endpoints, la **Etapa 5** define la estructura de carpetas completa del monorepo (ya esbozada en Etapa 2, se consolida archivo por archivo) y la **Etapa 6** el roadmap técnico de implementación.
