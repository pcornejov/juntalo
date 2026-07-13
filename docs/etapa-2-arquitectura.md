# Juntalo — Etapa 2: Arquitectura

> Estado: **propuesta v2, pendiente de aprobación**. Depende de la Etapa 1 aprobada (`etapa-1-analisis-producto.md`). Decisiones confirmadas: modelo de fondos **split payment**; backend en **NestJS (TypeScript)** — el usuario trabaja TS a diario y Go sería aprendizaje; para 1 dev, un solo lenguaje + tipos compartidos end-to-end vale más que las ventajas operacionales de Go en esta etapa.

---

## 1. Visión general

Monorepo TypeScript con dos aplicaciones y un paquete compartido:

```
juntalo/
├── apps/
│   ├── web/               # SPA React + TS (Vite)
│   └── api/               # NestJS
├── packages/
│   └── shared/            # ← tipos compartidos: DTOs, códigos de error, tipos de campaña, helpers CLP
├── docs/
├── docker/
└── README.md
```

- **pnpm workspaces** (+ scripts npm coordinados; Turborepo solo si el build se vuelve lento — no de entrada). `packages/shared` es la ventaja estructural de ir full-TS: el contrato API vive en un solo lugar, tipado, importado por ambas apps. Cambias un DTO y el compilador te muestra todo lo que rompe, en frontend y backend.
- **Un solo deployable de backend** (monolito modular NestJS). Sin microservicios: a la escala del MVP un monolito es más rápido de desarrollar, desplegar y depurar; los módulos NestJS dan la modularidad interna para extraer servicios después *si los datos lo justifican*.
- **SPA + render OG en el API**: la página pública `/c/:slug` necesita meta tags Open Graph server-rendered (riesgo 14, Etapa 1 — la preview en WhatsApp es la conversión). En vez de migrar a Next.js, un controller NestJS sirve HTML mínimo con OG tags para bots/previews y la SPA para navegadores. Si más adelante el SEO público se vuelve crítico, se evalúa mover solo la página pública a SSR.

### Alternativas descartadas
- **Go + Fiber** (propuesta v1): mejor huella de RAM y binario único, pero el usuario no trabaja Go; el costo de aprendizaje + mantener el contrato API a mano supera esas ventajas para validar un MVP en solitario. El diseño de dominio se tradujo 1:1 — nada se perdió.
- **Next.js full-stack (API routes/server actions)**: acopla frontend y backend en un solo runtime y limita la evolución del API (webhooks de pasarelas, workers, rate limiting fino). NestJS da estructura de backend real que crecerá con el producto.
- **Microservicios / serverless**: sin equipo ni carga que lo justifique.

---

## 2. Backend (NestJS)

### 2.1 Módulos y capas

Módulos NestJS por vertical de dominio (misma partición que la Etapa 1), con separación interna dominio / aplicación / infraestructura donde aporta:

```
apps/api/src/
├── main.ts                       # bootstrap: helmet, CORS, pipes de validación globales
├── app.module.ts
├── modules/
│   ├── auth/                     # register, login, refresh; guards JWT; argon2id
│   ├── identity/                 # User, Organization (1:1 personal en MVP), user_identities
│   ├── campaigns/                # CRUD + máquina de estados + registro de tipos
│   │   ├── domain/               #    entidades puras, transiciones, campaign-types.registry.ts
│   │   ├── campaigns.service.ts  #    casos de uso
│   │   ├── campaigns.controller.ts
│   │   └── campaigns.repository.ts
│   ├── contributions/            # Contributor, Contribution, flujo de aporte
│   ├── payments/                 # Payment, máquina de estados, comisión
│   │   ├── domain/               #    payment.entity.ts, money.ts (CLP), transitions.ts
│   │   ├── provider/             #    payment-provider.interface.ts + mock/ (luego webpay/, mercadopago/)
│   │   └── webhooks.controller.ts
│   ├── files/                    # FileStorage: local hoy, R2 mañana (misma interfaz S3)
│   ├── public/                   # GET /c/:slug JSON + render OG HTML para bots
│   ├── dashboard/                # stats, participantes, export CSV
│   └── audit/                    # AuditLogs (interceptor + service)
├── common/                       # filtros de excepciones, interceptors, decorators, request-id
└── database/                     # drizzle: schema, migraciones, seeds, tx helper
```

**Reglas de oro** (el equivalente pragmático de Clean Architecture aquí):
- La **lógica de negocio pura** (máquina de estados de pagos y campañas, cálculo de comisión, registro de tipos) vive en `domain/` como clases/funciones TS **sin decoradores NestJS ni imports de infraestructura** → testeable con Vitest sin levantar nada.
- **Interfaces solo donde pagan**: `PaymentProvider` (mock hoy, pasarelas mañana — pedido explícito) y `FileStorage` (local hoy, R2 mañana), inyectadas por token de DI. Los repositorios son clases concretas sobre Drizzle — NestJS DI permite sustituirlos en tests sin necesidad de interfaz formal; no se paga ceremonia extra.
- Los **DTOs de request/response viven en `packages/shared`** (Zod schemas): el backend los usa para validar (pipe de Zod), el frontend para tipar las llamadas. Una sola fuente de verdad del contrato.

### 2.2 ORM: Drizzle vs Prisma → **Drizzle**

El argumento de la v1 para sqlc (dominio financiero ⇒ SQL visible y auditable) se traduce directamente:

| Criterio | Drizzle | Prisma |
|---|---|---|
| Control del SQL | Queries que son SQL tipado 1:1; `sum().filterWhere(...)`, `FOR UPDATE`, CTEs — todo expresable y visible | Query engine intermedio; agregaciones financieras y locks acaban en `$queryRaw` sin tipos |
| Dominio financiero | El SQL *es* la especificación: auditar un descuadre de comisiones = leer la query | El SQL real queda oculto tras el cliente |
| Runtime | Ligero, sin binario extra | Query engine aparte (mejoró con la versión TS, pero sigue siendo una capa más) |
| Migraciones | `drizzle-kit generate` → SQL versionado en el repo, revisable en PR | `prisma migrate` similar, buen tooling |
| DX/madurez | Muy buena, ecosistema más joven | Excelente DX para CRUD, más madura |

**Decisión: Drizzle + node-postgres, migraciones SQL generadas por drizzle-kit y versionadas en el repo.** Donde `Payment` es la entidad central de reporting (Etapa 1 §4), controlar cada query financiera no es purismo — es auditabilidad. Prisma sería perfectamente válido para un CRUD genérico; aquí Drizzle encaja mejor con lo que el producto es.

### 2.3 PaymentProvider (diseño para split payment)

```ts
// packages/shared → tipos; apps/api/src/modules/payments/provider/payment-provider.interface.ts
export interface PaymentProvider {
  /** Inicia un intento de pago. idempotencyKey garantiza que reintentos
   *  del mismo intento no dupliquen cargos (riesgo 8, Etapa 1). */
  createIntent(req: IntentRequest): Promise<PaymentIntent>;
  /** Consulta el estado real en la pasarela (fuente de verdad externa). */
  getIntent(providerRef: string): Promise<PaymentIntent>;
  /** Reembolso total o parcial de un pago confirmado. */
  refund(providerRef: string, amount: Money): Promise<RefundResult>;
}

export interface IntentRequest {
  idempotencyKey: string;
  amount: Money;            // CLP, entero
  commission: Money;        // snapshot: lo que Juntalo retiene (split)
  payeeAccount: PayeeRef;   // cuenta del organizador (split payment)
  metadata?: Record<string, string>;
}
```

Puntos de diseño (idénticos a v1 — son independientes del lenguaje):
- **La máquina de estados vive en el dominio, no en el provider**: `pending → confirmed → refunded`, `pending → failed`, con transiciones validadas en `payments/domain/transitions.ts`. El provider reporta hechos externos; el dominio decide si la transición es legal. Cambiar de pasarela jamás toca lógica de negocio.
- **Split explícito en la interfaz**: `commission` y `payeeAccount` van en el request desde el día 1 — en split payment la comisión se declara *al crear el intento*.
- **`MockPaymentProvider`** implementa el ciclo completo con modos configurables: confirmación inmediata, diferida, fallo, y **reintentos/duplicados** (mismo `idempotencyKey` dos veces ⇒ mismo intent — la idempotencia llega probada a la pasarela real).
- **Webhooks-ready**: `POST /webhooks/payments/:provider` existe desde el MVP; el mock lo invoca para simular confirmación asíncrona, ensayando el flujo que Webpay/Mercado Pago usarán de verdad.

### 2.4 Tipos de campaña (registro declarativo)

Conforme a Etapa 1 (riesgo 1): tipos como código, no motor genérico de JSON schema.

```ts
// packages/shared/src/campaign-types.ts  ← compartido: el frontend renderiza desde aquí
export interface CampaignTypeDefinition {
  key: CampaignTypeKey;        // 'collection' | 'sale' | 'event' | ... | 'raffle' (disabled)
  enabled: boolean;
  labels: TypeLabels;          // CTA ("Aportar"/"Comprar"/"Inscribirse"), unidades, textos
  fields: FieldSpec[];         // campos extra del tipo (persisten en settings JSONB)
  rules: TypeRules;            // ¿meta económica requerida?, ¿monto libre?, montos sugeridos
}
export const CAMPAIGN_TYPE_REGISTRY: Record<CampaignTypeKey, CampaignTypeDefinition> = { ... };
// MVP: solo 'collection' enabled
```

Ventaja extra sobre la v1: al vivir en `packages/shared`, el frontend **importa el registro directamente** (tipado) en lugar de consumir `GET /campaign-types`. Agregar "Venta" = una entrada nueva + sus campos; cero refactor.

### 2.5 Transversales

- **Auth**: JWT access corto (15 min, `@nestjs/jwt` + guard global con `@Public()` para rutas abiertas) + refresh token opaco en cookie httpOnly, rotado y revocable en DB. Passwords con **argon2id**. Tabla `user_identities` separada de `users` → Google OAuth futuro es una fila por proveedor, no una migración.
- **Rate limiting**: `@nestjs/throttler` con límites agresivos en `POST /contributions` y auth; storage in-memory (1 VPS), Redis cuando haya más instancias.
- **Validación**: Zod schemas de `packages/shared` aplicados en un pipe global — la misma definición valida en el borde y tipa el cliente. Invariantes de negocio re-validados en dominio (defensa en profundidad).
- **Errores**: `DomainError` con código estable (`campaign_not_active`, `payment_duplicate`, ...) → exception filter único los mapea a HTTP; el frontend traduce códigos (compartidos) a mensajes.
- **Observabilidad**: `pino` (nestjs-pino) JSON estructurado + request-id; `AuditLogs` en Postgres para acciones de negocio (cambios de estado de campaña, futuros payouts).
- **Config**: `@nestjs/config` + schema Zod de env vars, fail-fast al arrancar si falta algo. `.env` solo en desarrollo.
- **Seguridad HTTP**: helmet, CORS estricto, cookies `Secure`/`SameSite`, límites de tamaño de payload.

---

## 3. Frontend (React + TS + Vite)

```
apps/web/src/
├── app/                # router, providers (QueryClient, Auth), layout raíz
├── features/           # verticales por dominio — espejo del backend
│   ├── auth/
│   ├── campaigns/      #    crear/editar/listar (dashboard)
│   ├── public-campaign/#    página pública /c/:slug + flujo de aporte
│   └── dashboard/      #    stats, participantes, export CSV
├── shared/
│   ├── ui/             # Design System: Button, Input, Card, Progress, Dialog...
│   ├── api/            # cliente HTTP tipado con los DTOs de packages/shared
│   ├── hooks/
│   └── lib/            # share/WhatsApp/QR helpers (formato CLP viene de packages/shared)
└── styles/             # tokens Tailwind (paleta, tipografía, dark-ready)
```

- **Feature-folders, no capas técnicas globales**: cada feature contiene páginas, hooks de React Query y llamadas API. Escala mejor para 1 dev que `components/`+`services/` globales.
- **React Query** como única capa de estado servidor (sin Redux): cache, invalidación tras mutaciones, estados de carga. Estado local con `useState`/`useReducer`.
- **Cliente HTTP tipado**: `shared/api` usa los Zod schemas de `packages/shared` para tipar (y opcionalmente validar) las respuestas — errores de contrato se detectan en compilación o en el borde, nunca en medio de la UI.
- **Design System mínimo pero real**: ~10 componentes sobre tokens Tailwind (tipografía Inter, espaciado consistente). Dark mode preparado vía CSS variables + `data-theme` (tokens semánticos `bg-surface`, `text-primary`), sin toggle en MVP.
- **Mobile-first estricto**: página pública y flujo de aporte diseñados primero a 375px (in-app browser de WhatsApp); el dashboard puede ser desktop-first.
- **Rutas**: `/c/:slug` (pública), `/login`, `/register`, `/dashboard`, `/dashboard/campaigns/new`, `/dashboard/campaigns/:id`. Dashboard protegido por guard de auth.

---

## 4. Infraestructura y despliegue

```
docker/
├── api.Dockerfile          # multi-stage: pnpm build → node:22-slim, solo dist + prod deps
├── web.Dockerfile          # build Vite → estáticos servidos por Caddy
├── docker-compose.yml      # dev: postgres + api (watch) + web (vite dev)
└── docker-compose.prod.yml # prod VPS: caddy (TLS automático) + api + postgres
```

- **Caddy** como reverse proxy: TLS automático (Let's Encrypt), sirve estáticos del frontend, proxy `/api/*` y `/c/*` (render OG) al backend NestJS. **Cloudflare-ready** (DNS+proxy delante sin cambios).
- **Cloudflare R2-ready**: `FileStorage` con implementación local (disco servido por Caddy) en MVP; R2 es la misma interfaz vía SDK S3-compatible.
- **Migraciones** como paso explícito (`pnpm db:migrate` / job en compose), nunca automáticas al arrancar en producción.
- **Backups**: `pg_dump` diario vía cron desde el día 1 — la plataforma registra dinero, aunque sea mock.
- Nota operacional: Node consume más RAM que Go (~150-300MB vs ~40MB); irrelevante para el MVP en cualquier VPS de 2GB. Se acepta el trade-off conscientemente.

---

## 5. Calidad

- **Tests (Vitest en todo el monorepo)**: unitarios del dominio (máquina de estados de pagos y campañas, comisión, registro de tipos — todo `domain/` puro sin NestJS); tests de servicios con providers/repos sustituidos vía DI; integración de repositorios Drizzle contra Postgres real (testcontainers). Frontend: Vitest + Testing Library en flujos críticos (form de aporte, creación de campaña). Sin E2E en MVP; Playwright cuando el flujo se estabilice.
- **Lint/formato**: ESLint (config compartida en el monorepo) + Prettier; `pnpm check` corre lint + typecheck + tests de todo.
- **CI (GitHub Actions)**: lint + typecheck + tests + build de ambas apps en cada PR.
- **Scripts**: `pnpm dev` (compose + watch), `pnpm db:migrate`, `pnpm db:seed`, `pnpm check`, `pnpm build`.

---

## 6. Resumen de decisiones

| Decisión | Elección | Alternativa descartada y por qué |
|---|---|---|
| Lenguaje backend | TypeScript (NestJS) | Go: mejor huella operacional, pero curva de aprendizaje + contrato API manual pesan más para 1 dev TS |
| Monorepo | pnpm workspaces + `packages/shared` | Repos separados: se pierde el contrato tipado compartido |
| Topología | Monolito modular NestJS | Microservicios: sin equipo/carga que lo justifique |
| OG tags | Controller NestJS para `/c/:slug` | Next.js SSR: acopla frontend/backend y limita el API |
| ORM | Drizzle + migraciones SQL versionadas | Prisma: gran DX, pero SQL financiero oculto tras el query engine |
| Validación | Zod schemas compartidos front/back | class-validator: duplica el contrato en decoradores solo-backend |
| Fondos | Split payment en la interfaz del provider | Merchant of record: custodia = exposición UAF |
| Estado frontend | React Query + estado local | Redux: sin estado global real que gestionar |
| Proxy | Caddy | nginx: más config manual para TLS |
| Passwords | argon2id | bcrypt: válido, pero argon2id es el estándar actual |
| Auth futura | `user_identities` separada desde día 1 | Columnas OAuth en `users`: migración dolorosa |

---

## Siguiente paso

Con esta arquitectura aprobada, la **Etapa 3** produce el modelo entidad-relación completo (tablas, columnas, índices, constraints) coherente con estas decisiones.
