# Juntalo — Etapa 2: Arquitectura

> Estado: **propuesta, pendiente de aprobación**. Depende de la Etapa 1 aprobada (`etapa-1-analisis-producto.md`). Decisión confirmada: modelo de fondos **split payment** (el dinero va directo al organizador vía la pasarela; Juntalo solo cobra comisión).

---

## 1. Visión general

Monorepo con dos aplicaciones desplegables y una infraestructura común:

```
juntalo/
├── frontend/          # SPA React + TS (Vite)
├── backend/           # API Go (Fiber) + workers ligeros
├── docs/              # Documentos de producto y arquitectura
├── docker/            # Dockerfiles + compose + config nginx/caddy
└── README.md
```

- **Un solo binario Go** sirve la API. No hay microservicios: a la escala del MVP (y de los primeros miles de usuarios) un monolito modular es más rápido de desarrollar, desplegar y depurar. La modularidad interna (ver §3) permite extraer servicios después *si algún módulo lo justifica con datos*, no por anticipación.
- **SPA + endpoint SSR mínimo**: la app es una SPA (Vite), pero la página pública de campaña necesita meta tags Open Graph server-rendered (riesgo 14 de Etapa 1). En vez de meter Next.js o SSR completo, el backend Go renderiza un HTML mínimo (plantilla con OG tags + redirect/hydrate a la SPA) para las rutas `/c/:slug` cuando el user-agent es un bot/preview, y sirve la SPA normal para navegadores. Un middleware de detección simple es suficiente y evita duplicar el stack de frontend.

### Alternativas descartadas
- **Next.js/Remix (SSR completo)**: resuelve OG tags "gratis" pero duplica la complejidad operacional (servidor Node + servidor Go) para 1 dev. Se descarta; si el SEO de páginas públicas se vuelve crítico, se revisa.
- **Microservicios / serverless**: sin equipo ni carga que lo justifique. El monolito modular es la decisión correcta hoy y no cierra puertas.

---

## 2. Backend (Go + Fiber)

### 2.1 Capas

Clean Architecture pragmática, tres capas + API. La regla de dependencia es estricta: las flechas apuntan hacia adentro.

```
backend/
├── cmd/api/main.go                  # composición: wiring de dependencias, arranque
├── internal/
│   ├── domain/                      # ← núcleo: entidades, invariantes, errores de negocio
│   │   ├── campaign/                #    Campaign, CampaignType (registro declarativo), estados
│   │   ├── payment/                 #    Payment, máquina de estados, comisión, Money (CLP)
│   │   ├── contribution/            #    Contributor, Contribution
│   │   ├── identity/                #    User, Organization, credenciales
│   │   └── audit/                   #    AuditLog
│   ├── app/                         # ← casos de uso: orquestan dominio + puertos
│   │   ├── campaigns/               #    CreateCampaign, PublishCampaign, FinishCampaign...
│   │   ├── contributions/           #    StartContribution, ConfirmContribution...
│   │   ├── auth/                    #    Register, Login, RefreshToken
│   │   └── dashboard/               #    ListCampaigns, ExportCSV, CampaignStats
│   ├── infra/                       # ← adaptadores: implementan los puertos
│   │   ├── postgres/                #    repositorios (sqlc), migraciones, tx manager
│   │   ├── payments/                #    PaymentProvider: mock/ (luego webpay/, mercadopago/)
│   │   ├── storage/                 #    FileStorage: local/ (luego r2/)
│   │   └── ratelimit/               #    limiter (in-memory MVP; Redis-ready)
│   └── api/                         # ← HTTP: handlers Fiber, middleware, DTOs, validación
│       ├── handlers/
│       ├── middleware/              #    auth JWT, rate limit, request-id, recover, logging
│       └── public/                  #    rutas públicas /c/:slug + render OG
├── db/
│   ├── migrations/                  # golang-migrate, SQL puro, versionadas
│   ├── queries/                     # SQL fuente para sqlc
│   └── seeds/
└── Makefile
```

**Dónde SÍ hay interfaces (puertos)** — solo donde existe o existirá más de una implementación, o donde aísla I/O para testear:
- `payment.Provider` (mock hoy; Webpay/Mercado Pago/Stripe/Khipu mañana) — pedido explícito.
- `storage.FileStorage` (disco local hoy; Cloudflare R2 mañana).
- Repositorios por agregado (`CampaignRepository`, `PaymentRepository`, ...) — una sola implementación (Postgres), pero la interfaz vive en `app/` y permite tests de casos de uso sin base de datos. Es el único "costo Clean Architecture" que se paga sin segunda implementación, y se paga porque los casos de uso de pagos/comisión son exactamente lo que más vale la pena testear aislado.

**Dónde NO hay interfaces**: logging (se usa `slog` directo), validación, render de templates, generación de QR. Abstraer eso es ceremonia sin retorno.

### 2.2 sqlc vs GORM → **sqlc**

| Criterio | sqlc | GORM |
|---|---|---|
| Tipado | Genera structs y funciones tipadas desde SQL real; errores en compile-time | Reflexión en runtime; errores de mapping en producción |
| Dominio financiero | SQL explícito: sumas de pagos, snapshots de comisión y locks (`SELECT ... FOR UPDATE`) se ven y auditan en el código | El SQL generado queda oculto; expresiones como `SUM(...) FILTER (WHERE ...)` obligan a raw SQL igual |
| Rendimiento | Sin overhead de reflexión | Aceptable, pero N+1 fácil de introducir sin notar |
| Curva | Hay que escribir SQL (para este dominio, es una ventaja: el SQL *es* la especificación) | Más rápido para CRUD trivial |
| Migraciones | Ninguna de las dos las resuelve; se usa **golang-migrate** con SQL puro en ambos casos | AutoMigrate existe pero es peligroso en producción (cambios implícitos de schema) |

**Decisión: sqlc + golang-migrate + pgx.** En una plataforma donde `Payment` es la entidad central de reporting (Etapa 1, §4), querer ver y controlar cada query financiera no es purismo — es la diferencia entre poder auditar un descuadre de comisiones o no. El costo (escribir SQL) es bajo para 1 dev que además define el schema.

### 2.3 PaymentProvider (diseño para split payment)

```go
// domain/payment/provider.go
type Provider interface {
    // Inicia un intento de pago. La idempotencyKey garantiza que reintentos
    // del mismo intento no dupliquen cargos (riesgo 8, Etapa 1).
    CreateIntent(ctx context.Context, req IntentRequest) (Intent, error)
    // Consulta/confirma el estado real en la pasarela (fuente de verdad externa).
    GetIntent(ctx context.Context, providerRef string) (Intent, error)
    // Solicita reembolso total o parcial de un pago confirmado.
    Refund(ctx context.Context, providerRef string, amount Money) (Refund, error)
}

type IntentRequest struct {
    IdempotencyKey string
    Amount         Money        // CLP, entero
    Commission     Money        // snapshot: lo que Juntalo retiene (split)
    PayeeAccount   PayeeRef     // cuenta del organizador (split payment)
    Metadata       map[string]string
}
```

Puntos de diseño:
- **La máquina de estados vive en el dominio, no en el provider**: `pending → confirmed → refunded` y `pending → failed`, con transiciones validadas en `domain/payment`. El provider solo reporta hechos externos; el dominio decide si la transición es legal. Así, cambiar de pasarela jamás toca la lógica de negocio.
- **Split explícito en la interfaz**: `Commission` y `PayeeAccount` están en el request desde el día 1, porque en split payment la comisión se declara *al crear el intento*, no después. El `MockPaymentProvider` los acepta y los registra; Webpay/Mercado Pago los usarán de verdad.
- **`MockPaymentProvider`** implementa el ciclo completo, con modos configurables para simular: confirmación inmediata, confirmación diferida, fallo, y **reintentos/duplicados** (mismo `IdempotencyKey` dos veces debe devolver el mismo Intent — así la idempotencia llega probada a la pasarela real, Etapa 1 §2).
- **Webhooks-ready**: el endpoint `POST /webhooks/payments/:provider` existe desde el MVP (el mock lo invoca a sí mismo para simular confirmación asíncrona). Las pasarelas reales confirman por webhook; ensayar ese flujo con el mock evita rediseñar el flujo de confirmación después.

### 2.4 Tipos de campaña (registro declarativo)

Conforme a Etapa 1 (riesgo 1): tipos como código, no motor genérico.

```go
// domain/campaign/types.go
type TypeDefinition struct {
    Key          TypeKey          // "collection", "sale", "event", ... "raffle" (disabled)
    Enabled      bool
    Labels       TypeLabels       // textos: CTA ("Aportar"/"Comprar"/"Inscribirse"), unidades
    Fields       []FieldSpec      // campos extra que este tipo agrega al form (van a settings JSONB)
    Rules        TypeRules        // requiere meta económica?, permite monto libre?, montos sugeridos?
}
var Registry = map[TypeKey]TypeDefinition{ ... } // MVP: solo "collection" Enabled
```

El frontend consume `GET /campaign-types` y renderiza según la definición. Agregar "Venta" en el futuro = agregar una entrada al registro + sus campos; cero refactor del flujo.

### 2.5 Transversales

- **Auth**: JWT access token corto (15 min) + refresh token opaco en cookie httpOnly (rotado, revocable en DB). Passwords con **argon2id**. Tabla `user_identities` separada de `users` desde el día 1 → agregar Google OAuth después es una fila más por proveedor, no una migración.
- **Rate limiting**: middleware por IP+ruta con límites agresivos en `POST /contributions` y auth. Implementación in-memory (el MVP corre en 1 VPS); la interfaz permite backend Redis cuando haya más de una instancia.
- **Validación**: en el borde (`api/`) con DTOs + `go-playground/validator`; los invariantes de negocio se re-validan en dominio (defensa en profundidad).
- **Errores**: tipo `DomainError` con código estable (`campaign_not_active`, `payment_duplicate`, ...) mapeado a HTTP en un solo lugar; el frontend traduce códigos a mensajes.
- **Observabilidad**: `slog` JSON estructurado + request-id propagado; `AuditLogs` en Postgres para acciones de negocio (cambios de estado de campaña, payouts futuros).
- **Config**: struct tipada cargada de env vars (`envconfig`), `.env` solo en desarrollo, fail-fast si falta algo en producción.

---

## 3. Frontend (React + TS + Vite)

```
frontend/
├── src/
│   ├── app/                # router, providers (QueryClient, Auth), layout raíz
│   ├── features/           # verticales por dominio — misma lógica que el backend
│   │   ├── auth/           #    páginas + hooks + api de registro/login
│   │   ├── campaigns/      #    crear/editar/listar (dashboard)
│   │   ├── public-campaign/#    página pública /c/:slug + flujo de aporte
│   │   └── dashboard/      #    stats, participantes, export CSV
│   ├── shared/
│   │   ├── ui/             # Design System: Button, Input, Card, Progress, Dialog...
│   │   ├── api/            # cliente HTTP (fetch tipado), manejo de errores/códigos
│   │   ├── hooks/
│   │   └── lib/            # formato CLP, fechas, helpers de share/WhatsApp/QR
│   └── styles/             # tokens Tailwind (tailwind.config: colores, tipografía, dark-ready)
└── ...
```

- **Feature-folders, no capas técnicas globales**: cada feature contiene sus páginas, hooks de React Query y llamadas API. `shared/ui` es el único código verdaderamente transversal. Esto escala mejor para 1 dev que `components/`+`pages/`+`services/` globales.
- **React Query** como única capa de estado servidor (sin Redux): cache, invalidación tras mutaciones, estados de carga. Estado local con `useState`/`useReducer`; no se agrega gestor global salvo necesidad demostrada.
- **Design System mínimo pero real**: ~10 componentes en `shared/ui` construidos sobre tokens Tailwind (paleta definida en `tailwind.config`, tipografía Inter, espaciado consistente). Dark mode preparado vía CSS variables + `data-theme` (tokens semánticos: `bg-surface`, `text-primary`), sin toggle en MVP.
- **Mobile-first estricto**: la página pública y el flujo de aporte se diseñan primero a 375px (in-app browser de WhatsApp), el dashboard puede ser desktop-first.
- **Rutas**: `/c/:slug` (pública), `/login`, `/register`, `/dashboard`, `/dashboard/campaigns/new`, `/dashboard/campaigns/:id`. Rutas de dashboard protegidas por guard de auth.

---

## 4. Infraestructura y despliegue

```
docker/
├── backend.Dockerfile      # multi-stage: build Go → imagen distroless/alpine
├── frontend.Dockerfile     # build Vite → estáticos servidos por Caddy
├── docker-compose.yml      # dev: postgres + backend (air hot-reload) + frontend (vite dev)
└── docker-compose.prod.yml # prod VPS: caddy (TLS automático) + backend + postgres
```

- **Caddy** como reverse proxy en el VPS: TLS automático (Let's Encrypt), sirve los estáticos del frontend, proxy `/api/*` y `/c/*` (OG render) al backend. Más simple de operar que nginx para 1 dev; **Cloudflare-ready** (DNS+proxy delante, sin cambios).
- **Cloudflare R2-ready**: `FileStorage` con implementación `local/` (disco + servido por Caddy) en MVP; `r2/` es la misma interfaz con SDK S3-compatible.
- **Migraciones** corren como paso explícito (`make migrate` / job en compose), nunca automáticas al arrancar el binario en producción.
- **Backups**: `pg_dump` diario vía cron en el VPS desde el día 1 — es una plataforma que registra dinero, aunque sea mock.

---

## 5. Calidad

- **Tests**: unitarios de dominio (máquina de estados de pago, comisión, transiciones de campaña, tipos) y de casos de uso con repos fake; tests de integración de repositorios sqlc contra Postgres real (testcontainers o compose). Frontend: Vitest + Testing Library en flujos críticos (form de aporte, creación de campaña). Sin E2E en MVP (costo alto, se agrega con Playwright cuando el flujo se estabilice).
- **Lint/formato**: `golangci-lint` + `gofumpt`; ESLint + Prettier. `make check` corre todo.
- **CI (GitHub Actions)**: lint + tests + build de ambas apps en cada PR.
- **Scripts**: `make dev` (compose up con hot-reload), `make migrate`, `make seed`, `make check`, `make build`.

---

## 6. Resumen de decisiones

| Decisión | Elección | Alternativa descartada y por qué |
|---|---|---|
| Topología | Monolito modular Go | Microservicios: sin equipo/carga que lo justifique |
| OG tags | Render mínimo en Go para `/c/:slug` | Next.js SSR: duplica stack operacional |
| ORM | sqlc + golang-migrate + pgx | GORM: SQL oculto en dominio financiero |
| Fondos | Split payment en la interfaz del provider | Merchant of record: custodia = exposición UAF |
| Estado frontend | React Query + estado local | Redux: sin estado global real que gestionar |
| Estructura frontend | Feature-folders | Capas técnicas globales: se degradan al crecer |
| Proxy | Caddy | nginx: más config manual para TLS |
| Passwords | argon2id | bcrypt: válido, pero argon2id es el estándar actual |
| Auth futura | `user_identities` separada desde día 1 | Columnas OAuth en `users`: migración dolorosa |

---

## Siguiente paso

Con esta arquitectura aprobada, la **Etapa 3** produce el modelo entidad-relación completo (tablas, columnas, índices, constraints) coherente con estas decisiones.
