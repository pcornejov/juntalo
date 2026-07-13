# Juntalo — Etapa 6: Roadmap técnico

> Estado: **propuesta, pendiente de aprobación**. Última etapa de diseño; al aprobarla comienza la implementación (Etapa 7). Define el orden de construcción, hitos verificables y el criterio objetivo de "MVP terminado".

---

## Principios del roadmap

1. **Vertical, no horizontal**: no se construye "todo el backend y luego todo el frontend" — cada hito atraviesa el stack y termina en algo **demostrable en el navegador**. Es la única forma de detectar errores de integración temprano y de mantener motivación siendo 1 dev.
2. **El riesgo primero**: el flujo financiero (aporte → pago → confirmación) es el corazón y lo más difícil; se construye en el hito 3, no al final.
3. **Cada hito deja `main` desplegable**: `make check` verde + `docker compose up` funcional al cierre de cada hito. Sin ramas de larga vida.
4. Sin estimaciones de tiempo en días: siendo 1 dev sin apuro, el roadmap ordena y define "terminado", no promete fechas.

---

## Hito 0 — Esqueleto y andamiaje

**Objetivo**: el monorepo existe, compila, y el pipeline completo funciona en local y CI *antes* de escribir lógica.

- Estructura de Etapa 5 creada; Go module + Vite app inicializados.
- Docker Compose dev: Postgres + backend (hot-reload con `air`) + frontend (vite dev).
- Migración `0001_init` con el schema completo de Etapa 3 + seeds dev.
- sqlc configurado y generando; `/healthz` respondiendo.
- CI (GitHub Actions): lint + test + build de ambos lados.
- Design tokens Tailwind + los ~10 componentes base de `shared/ui` con una página de muestra.

**Demo del hito**: `make dev` levanta todo; `/healthz` OK; página de componentes visible.

## Hito 1 — Identidad

**Objetivo**: crear cuenta y sesión persistente.

- Backend: register (tx completa: user+identity+org+membership), login, refresh con rotación, logout, `/auth/me`; argon2id; middleware auth; rate limits de auth.
- Frontend: páginas login/registro; `useAuth` con refresh silencioso; guard de rutas; layout base del dashboard (vacío).
- OpenAPI de auth + primer `types.gen.ts`.

**Demo**: registrarse, cerrar el navegador, volver, seguir logueado. Tests: unitarios de register/login + integración de repos.

## Hito 2 — Campañas (sin dinero)

**Objetivo**: el organizador crea y administra campañas; existe página pública.

- Backend: CRUD + transiciones (publish/pause/resume/finish) + registro de tipos (`collection` enabled) + slugs + `GET /public/campaigns/:slug` + render OG `/c/:slug` + upload de imagen (storage local).
- Frontend: form de creación (<2 min: título+meta+descripción y publicar de una vez), dashboard con lista y detalle, página pública móvil-first con barra de progreso (en $0), botones compartir (wa.me, copy, QR).
- Cache 30s en la pública.

**Demo**: crear campaña, publicarla, abrir el link en el teléfono, ver preview OG al pegarlo en WhatsApp. Tests: transiciones de estado, slugs, tipos.

## Hito 3 — Dinero (el corazón) ★

**Objetivo**: el flujo completo de aporte con MockPaymentProvider, diseñado como si fuera real.

- Backend: `contributions/start` (idempotencia + snapshot comisión + CreateIntent), `MockPaymentProvider` (modos instant/deferred/fail; misma key ⇒ mismo intent), webhook firmado + `contributions/confirm` (tx payment+contribution), vista `campaign_totals`, polling de estado, `payment_refunds` (tabla operativa vía SQL, sin endpoint).
- Frontend: `ContributeSheet` (bottom-sheet), pantalla de confirmación con polling, progreso y lista de participantes en vivo (React Query invalidation), manejo de anónimos.
- Rate limiting agresivo en el endpoint público.

**Demo**: dos personas aportan desde sus teléfonos a la vez; la barra sube; un aporte "fallido" (modo fail del mock) no suma; repetir el submit no duplica. **Tests: los más exhaustivos del repo** — duplicados, carreras, transiciones ilegales, redondeo de comisión, webhook repetido.

## Hito 4 — Panel completo y cierre del MVP

**Objetivo**: todo lo que le falta al organizador para operar.

- Backend: participantes con filtros, export CSV streaming (con reembolsos restados), totales bruto/neto en dashboard, AuditLogs en las acciones de negocio.
- Frontend: tabla de participantes, export, panel de totales, estados vacíos/errores pulidos, revisión completa de UX móvil.
- Aviso de privacidad (Ley 21.719) en el form de aporte; términos mínimos en el footer.

**Demo**: el flujo de validación completo de la Etapa 1, de punta a punta, en un teléfono real.

## Hito 5 — Producción

**Objetivo**: Juntalo corriendo en un VPS con dominio real.

- `docker-compose.prod.yml` + Caddy (TLS automático) + Cloudflare delante.
- Migraciones como paso explícito de deploy; `pg_dump` diario con retención; monitoreo básico (uptime + logs).
- Checklist de seguridad: headers, CORS, cookies Secure, secrets fuera del repo, backups probados con un restore real.

**Demo**: URL pública real compartida por WhatsApp a un usuario de prueba que aporta desde su teléfono.

---

## Criterio de "MVP terminado" (checklist de la Etapa 1)

- [ ] Una persona se registra y crea una campaña en **menos de 2 minutos** (cronometrado, en móvil).
- [ ] El link compartido por WhatsApp muestra preview con imagen y título (OG).
- [ ] Un tercero aporta desde el in-app browser de WhatsApp sin crear cuenta.
- [ ] Repetir/reintentar un aporte no duplica cargos (idempotencia demostrada).
- [ ] El organizador ve totales bruto/neto correctos (incluyendo un reembolso de prueba) y exporta CSV.
- [ ] `make check` verde; deploy reproducible desde cero en un VPS limpio.

## Después del MVP (backlog ordenado, NO se construye ahora)

1. **Webpay Plus o Mercado Pago** (la decisión split payment ya está tomada; es implementar un `Provider` real + payout onboarding con KYC).
2. Verificación de email + recuperación de contraseña.
3. Segundo tipo de campaña (`sale` — valida el registro declarativo).
4. Comentarios/actualizaciones en la página pública.
5. Google OAuth · equipos/multi-admin · dominio personalizado · dark mode — en ese orden de demanda esperada.

---

## Siguiente paso

Con este roadmap aprobado, comienza la **Etapa 7: implementación**, partiendo por el Hito 0.
