# Juntalo — Etapa 5: Estructura de carpetas

> Estado: **propuesta, pendiente de aprobación**. Consolida la estructura esbozada en Etapa 2 al nivel de archivos, coherente con el modelo (Etapa 3) y los endpoints (Etapa 4). Es el esqueleto exacto que se creará en la Etapa 7.

---

## 1. Raíz del monorepo

```
juntalo/
├── frontend/
├── backend/
├── docs/
│   ├── etapa-1-analisis-producto.md ... etapa-6-roadmap.md
│   └── decisiones/            # ADRs futuros (una decisión = un archivo, cuando amerite)
├── docker/
│   ├── backend.Dockerfile
│   ├── frontend.Dockerfile
│   ├── docker-compose.yml         # desarrollo
│   ├── docker-compose.prod.yml    # VPS
│   └── caddy/Caddyfile
├── .github/workflows/ci.yml
├── .gitignore · .editorconfig
├── Makefile                   # raíz: delega a backend/frontend (make dev, check, build)
└── README.md
```

---

## 2. Backend (Go)

```
backend/
├── cmd/
│   └── api/
│       └── main.go                       # carga config → conecta PG → wiring → Fiber listen
├── internal/
│   ├── domain/                           # ═══ núcleo puro: sin imports de fiber/pgx/nada ═══
│   │   ├── money/
│   │   │   ├── money.go                  # Money (int64 CLP + currency), operaciones seguras
│   │   │   └── money_test.go
│   │   ├── campaign/
│   │   │   ├── campaign.go               # entidad + invariantes (título, fechas, meta)
│   │   │   ├── status.go                 # draft/active/paused/finished/suspended + transiciones
│   │   │   ├── types.go                  # TypeDefinition, Registry (collection enabled, resto declarado)
│   │   │   ├── slug.go                   # generación slugify + sufijo anti-colisión
│   │   │   └── *_test.go
│   │   ├── payment/
│   │   │   ├── payment.go                # entidad: gross/commission(snapshot)/net
│   │   │   ├── status.go                 # máquina de estados pending→confirmed→refunded…
│   │   │   ├── commission.go             # cálculo + redondeo (una sola función, muy testeada)
│   │   │   ├── provider.go               # interface Provider + IntentRequest (Etapa 2 §2.3)
│   │   │   └── *_test.go
│   │   ├── contribution/
│   │   │   ├── contribution.go           # entidad + Contributor
│   │   │   └── contribution_test.go
│   │   ├── identity/
│   │   │   ├── user.go · organization.go # entidades
│   │   │   └── password.go               # reglas de password (el hash vive en infra)
│   │   └── audit/
│   │       └── entry.go                  # AuditEntry + catálogo de actions
│   │
│   ├── app/                              # ═══ casos de uso: orquestan dominio + puertos ═══
│   │   ├── ports.go                      # TODAS las interfaces de repos + TxManager en un archivo
│   │   ├── auth/
│   │   │   ├── register.go               # tx: user+identity+org+membership (Etapa 3 §2)
│   │   │   ├── login.go · refresh.go · logout.go
│   │   │   └── *_test.go                 # con repos fake (map en memoria)
│   │   ├── campaigns/
│   │   │   ├── create.go · update.go · get.go · list.go
│   │   │   ├── transition.go             # publish/pause/resume/finish (validan vía domain)
│   │   │   ├── delete.go
│   │   │   └── *_test.go
│   │   ├── contributions/
│   │   │   ├── start.go                  # ★ el caso de uso central: idempotencia + snapshot comisión + CreateIntent
│   │   │   ├── confirm.go                # ★ invocado por webhook: transición payment+contribution en una tx
│   │   │   ├── status.go · list.go
│   │   │   └── *_test.go                 # duplicados, carreras, transiciones ilegales
│   │   ├── dashboard/
│   │   │   └── export_csv.go             # streaming del CSV (Etapa 4 §3)
│   │   └── files/
│   │       └── upload.go                 # validación mime/size + storage + registro
│   │
│   ├── infra/                            # ═══ adaptadores ═══
│   │   ├── postgres/
│   │   │   ├── db.go                     # pool pgx + TxManager
│   │   │   ├── sqlc/                     # ← código GENERADO (no se edita)
│   │   │   └── repos/                    # implementan app/ports.go usando sqlc
│   │   │       ├── users.go · campaigns.go · contributions.go · payments.go · files.go · audit.go
│   │   │       └── *_integration_test.go # contra Postgres real (compose)
│   │   ├── payments/
│   │   │   └── mock/
│   │   │       ├── provider.go           # modos: instant/deferred/fail; misma key ⇒ mismo intent
│   │   │       ├── webhook.go            # se auto-invoca firmado (ensaya el pipeline real)
│   │   │       └── provider_test.go
│   │   ├── storage/
│   │   │   └── local/local.go            # FileStorage a disco (r2/ futuro, misma interfaz)
│   │   ├── auth/
│   │   │   ├── argon2.go                 # hash/verify
│   │   │   └── jwt.go                    # firma/parseo access tokens
│   │   └── ratelimit/
│   │       └── memory.go                 # sliding window in-memory (Redis futuro)
│   │
│   └── api/                              # ═══ HTTP ═══
│       ├── server.go                     # Fiber app: middleware stack + mount de rutas
│       ├── router.go                     # tabla de rutas ↔ handlers (Etapa 4 completa)
│       ├── middleware/
│       │   ├── auth.go                   # Bearer → user/org en contexto
│       │   ├── ratelimit.go · requestid.go · recover.go · logging.go
│       ├── handlers/
│       │   ├── auth.go · campaigns.go · contributions.go · public.go · webhooks.go · files.go · meta.go
│       ├── dto/
│       │   ├── requests.go               # structs + tags validator (borde, Etapa 2 §2.5)
│       │   ├── responses.go
│       │   └── errors.go                 # DomainError → envelope {error:{code,…}} + mapa HTTP
│       └── public/
│       │   └── og.go                     # GET /c/:slug → HTML OG para bots / SPA para humanos
│       └── openapi.yaml                  # ★ contrato: fuente de los tipos TS del frontend
├── db/
│   ├── migrations/
│   │   └── 0001_init.up.sql / .down.sql  # todo el schema de Etapa 3 + vista campaign_totals
│   ├── queries/                          # SQL fuente para sqlc, por entidad
│   │   ├── users.sql · campaigns.sql · contributions.sql · payments.sql · files.sql · audit.sql
│   └── seeds/
│       └── dev.sql                       # usuario demo + campañas de prueba
├── sqlc.yaml · .golangci.yml
├── go.mod / go.sum
└── Makefile                              # run, test, lint, migrate, seed, sqlc, openapi-check
```

Notas:
- **`app/ports.go` único**: todas las interfaces de repositorio en un archivo — se ve el contrato de persistencia completo de un vistazo; se divide solo si crece de verdad.
- Los dos archivos marcados ★ (`start.go`, `confirm.go`) concentran el riesgo financiero del producto: son los más testeados del repo.
- `openapi.yaml` se mantiene a mano junto a los handlers (18 endpoints lo permiten); CI valida que compile y que los tipos TS generados estén al día (`openapi-check`).

---

## 3. Frontend (React + TS + Vite)

```
frontend/
├── src/
│   ├── app/
│   │   ├── main.tsx · App.tsx            # providers: QueryClient, AuthProvider, Router
│   │   ├── router.tsx                    # rutas de Etapa 2 §3 + guard
│   │   └── guard.tsx                     # RequireAuth → redirect /login
│   ├── features/
│   │   ├── auth/
│   │   │   ├── pages/LoginPage.tsx · RegisterPage.tsx
│   │   │   ├── hooks/useAuth.ts          # sesión, refresh silencioso, logout
│   │   │   └── api.ts
│   │   ├── campaigns/                    # dashboard del organizador
│   │   │   ├── pages/CampaignListPage.tsx · CampaignFormPage.tsx · CampaignDetailPage.tsx
│   │   │   ├── components/CampaignCard.tsx · StatusBadge.tsx · TotalsPanel.tsx
│   │   │   │              · ParticipantsTable.tsx · ExportCsvButton.tsx · SharePanel.tsx
│   │   │   ├── hooks/useCampaigns.ts · useCampaignMutations.ts   # React Query
│   │   │   └── api.ts
│   │   └── public-campaign/              # ★ la página que se comparte por WhatsApp
│   │       ├── pages/PublicCampaignPage.tsx · ContributeSuccessPage.tsx
│   │       ├── components/ProgressBar.tsx · ContributeSheet.tsx   # bottom-sheet móvil
│   │       │              · ContributorsList.tsx · ShareButtons.tsx · QrCode.tsx
│   │       ├── hooks/usePublicCampaign.ts · useContribute.ts      # genera Idempotency-Key al montar
│   │       └── api.ts
│   ├── shared/
│   │   ├── ui/                           # Design System (Etapa 2 §3)
│   │   │   ├── Button.tsx · Input.tsx · TextArea.tsx · Card.tsx · Dialog.tsx
│   │   │   ├── Progress.tsx · Badge.tsx · Skeleton.tsx · Toast.tsx · EmptyState.tsx
│   │   │   └── index.ts
│   │   ├── api/
│   │   │   ├── client.ts                 # fetch wrapper: base URL, Bearer, refresh en 401, envelope de error
│   │   │   ├── types.gen.ts              # ← GENERADO desde backend/api/openapi.yaml
│   │   │   └── errors.ts                 # code → mensaje es-CL
│   │   └── lib/
│   │       ├── clp.ts                    # formato $12.345 (es-CL) + parseo de input
│   │       ├── dates.ts · share.ts       # wa.me link, navigator.share, copy
│   ├── styles/
│   │   └── index.css                     # @theme tokens semánticos (bg-surface…) dark-ready
│   └── vite-env.d.ts
├── public/favicon.svg · robots.txt
├── index.html
├── vite.config.ts · tsconfig.json · eslint.config.js · .prettierrc
└── package.json                          # dev, build, check, generate:api
```

Notas:
- **`types.gen.ts` es la materialización del contrato** (Etapa 2 §2.5): `npm run generate:api` lee `backend/api/openapi.yaml`; CI falla si está desactualizado. El cliente HTTP y todos los hooks tipan contra él.
- `ContributeSheet` como bottom-sheet: el flujo de aporte ocurre **sin salir de la página** — crítico en el in-app browser de WhatsApp (riesgo 14).
- Tailwind v4 (config CSS-first en `styles/index.css` con `@theme`); tokens semánticos desde el día 1 para el dark mode futuro.

---

## 4. Reglas de dependencia (se validan en CI)

```
backend:   domain ← app ← {infra, api}     # domain no importa nada interno; app no importa infra/api
frontend:  shared ← features ← app         # features no se importan entre sí; shared no importa features
```
Backend: verificado con `depguard` (golangci-lint). Frontend: `eslint-plugin-boundaries`. La regla en CI evita que la disciplina se erosione con el tiempo — que es exactamente como mueren estas arquitecturas.

---

## Siguiente paso

Aprobada la estructura, la **Etapa 6** define el roadmap técnico: orden de construcción, hitos verificables y criterio de "MVP terminado".
