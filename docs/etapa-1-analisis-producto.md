# Juntalo — Etapa 1: Análisis de producto

## Contexto

El repositorio `pcornejov/juntalo` está vacío (sin commits). El usuario quiere construir un SaaS chileno para crear campañas de recaudación/venta online, siguiendo un proceso por etapas donde cada una requiere aprobación antes de avanzar. Este documento es el entregable de la **Etapa 1**: análisis del producto, riesgos, y propuesta de alcance del MVP. **No se escribe código en esta etapa.**

Parámetros confirmados con el usuario:
- **Monetización**: comisión porcentual sobre cada aporte (variable por campaña/tipo). El modelo de `Payment` debe guardar monto bruto, % comisión aplicado y monto neto por separado.
- **Recursos**: una persona, sin apuro (horizonte de meses). Se puede invertir algo más en una arquitectura limpia desde el día 1 sin sacrificar velocidad de validación — pero sin sobre-ingeniería.
- **Auth MVP**: solo email + contraseña. Google OAuth queda como extensión futura sin romper el diseño (se modela `AuthProvider`/`identity` pensando en esto, pero no se implementa).

---

## 0. Modelo de flujo de fondos y marco regulatorio (Chile) — a decidir antes de aprobar

Esta sección se agregó tras una revisión cruzada del documento. Recaudar dinero de terceros y liquidarlo al organizador convierte a Juntalo, de facto, en un intermediario de pagos. Esto no se puede dejar implícito porque cambia el modelo de datos y la exposición legal del negocio:

- **Decisión pendiente — flujo de fondos**: ¿el dinero pasa por una cuenta/entidad de Juntalo antes de liquidarse al organizador ("merchant of record"), o se usa **split payment** de la pasarela (Webpay/Mercado Pago/Stripe Connect) para que el dinero vaya directo al organizador y Juntalo solo cobre su comisión? La segunda opción reduce fuertemente la exposición regulatoria de Juntalo (no custodia fondos de terceros) y es la recomendada por defecto para el diseño, pero se deja como decisión explícita a confirmar con el usuario antes de Etapa 2, no asumida.
- **Boleta/factura**: cada aporte relevante puede requerir que el organizador emita boleta (SII) y que Juntalo facture su comisión. No se implementa en el MVP, pero el modelo de `Payment`/`Organization` debe dejar espacio para RUT y datos tributarios.
- **Ley 19.885 (donaciones con beneficio tributario)**: solo aplica a ciertas entidades receptoras; Juntalo NO debe presentar campañas genéricas como "donación deducible de impuestos" salvo que el organizador califique. Se documentará como advertencia de producto/legal, no como feature técnica del MVP.
- **UAF / Ley 19.913 (lavado de activos)**: relevante solo si Juntalo custodia fondos (opción "merchant of record"). Refuerza la preferencia por split payment.
- **Ley 21.719 (protección de datos personales)**: aportantes sin cuenta igual entregan datos personales (nombre/email/contacto) — se requiere aviso de privacidad/consentimiento básico en el formulario de aporte, aunque sea texto estático en el MVP.
- **KYC/onboarding del organizador**: aunque el pago sea mock en el MVP, el modelo de datos del organizador debe dejar espacio para RUT, cuenta bancaria y datos de payout, para no migrar dolorosamente en Etapa 3.

Estas decisiones no bloquean seguir con el resto de la Etapa 1, pero **sí deben resolverse antes de aprobar el modelo de datos en Etapa 3**, y afectan directamente el diseño de `PaymentProvider`.

## 1. Riesgos del producto

1. **Ambigüedad de "campaña"**: colecta solidaria, venta de productos y evento no son la misma entidad de negocio (unas mueven dinero hacia una meta, otras venden ítems con stock/precio unitario). Modelar todo como una sola tabla `Campaigns` sin diferenciar comportamiento por tipo puede volverse un "God model". → Mitigación: **tipos como código con configuración declarativa** — un enum `campaign_type` + registro de tipos en Go (cada tipo declara sus campos, textos y reglas) + columna JSONB `settings` para atributos específicos del tipo. Se descartó un motor genérico de JSON schema configurable: para 1 dev implicaría form-rendering dinámico, validación dinámica y schemas versionados al servicio de un único tipo activo — sobre-ingeniería que no compra extensibilidad real frente al registro declarativo.
2. **Pagos simulados hoy, reales mañana**: si `MockPaymentProvider` no replica el ciclo de vida real (pendiente → confirmado → fallido → reembolsado), migrar a Webpay/Mercado Pago después obligará a rediseñar el dominio de `Payment`. → Mitigación: diseñar la máquina de estados de pago completa desde ya, aunque el mock la simplifique.
3. **Comisión variable**: si el % de comisión no se congela en el momento del pago (snapshot), cambios futuros de tarifa corromperán reportes históricos. → Mitigación: `Payment` guarda `commission_rate_applied` y `commission_amount`, no una referencia calculada en caliente.
4. **Multi-tenant tardío**: si `Users` es dueño directo de `Campaigns` sin capa de "organización", agregar equipos/permisos después implica migración de datos dolorosa. → Mitigación: introducir una entidad `Organizations` desde el modelo de datos (aunque en el MVP cada usuario tenga automáticamente una organización personal 1:1), sin construir UI de equipos todavía.
5. **Slug/página pública**: colisiones de slug, campañas indexadas por buscadores con contenido de terceros (riesgo de moderación/legal). → Mitigación: slugs únicos generados + validados, y un campo `status` que permita ocultar/despublicar campañas.
6. **Rate limiting y abuso**: página pública de "aportar" es un objetivo fácil de bots. Necesario desde el MVP, no como deuda técnica.
7. **Alcance del "participante"**: ¿un participante requiere cuenta? Si se exige registro para aportar, se pierde el objetivo de "compartir por WhatsApp y aportar en <2 min". → Decisión: el aporte público NO requiere cuenta (solo nombre + contacto opcional); se modela como `Contributor`/`Contribution` (no un `User` completo), con un aviso mínimo de privacidad de datos (Ley 21.719) en el formulario.
8. **Doble-submit / idempotencia de pagos**: el flujo de aporte debe incluir una `idempotency_key` por intento de pago desde el diseño del MVP (aunque el provider sea mock), para no retrofitear esto cuando se integre una pasarela real.
9. **Moneda**: el MVP fija **CLP** explícitamente (montos enteros, sin decimales) — se deja el campo `currency` en el modelo por prolijidad, sin soportar multi-moneda todavía.
10. **Reembolsos en el reporting**: el monto neto recaudado y el CSV exportado deben restar reembolsos desde el día 1 (aunque el mock casi no los genere), para que el reporting financiero no nazca incorrecto.
11. **Imágenes/archivos**: se necesita una entidad `File` mínima (URL, tipo, campaña asociada) desde el modelo de datos, aunque el storage real (Cloudflare R2) se conecte más adelante. La imagen de campaña es **opcional con placeholder decente** — si subir imagen bloquea publicar, se rompe el objetivo de "crear y compartir en <2 min".
12. **Consistencia del monto recaudado**: el "monto recaudado" NO será un campo denormalizado actualizado con read-modify-write (bug de carrera clásico con pagos concurrentes). Decisión: se calcula como `SUM(payments WHERE status='confirmed') - reembolsos` — correcta por construcción y trivial a la escala del MVP; si algún día se denormaliza por rendimiento, será con `UPDATE ... SET amount = amount + X` transaccional junto al cambio de estado del Payment.
13. **Estados de campaña**: "borrador/activa/finalizada" era insuficiente — mezclaba expiración temporal con decisión del organizador y no cubría moderación. Decisión: `draft / active / paused / finished / suspended` (paused = por el dueño; suspended = por moderación/plataforma; "expirada" se deriva de `end_date`, no es estado propio).
14. **Conversión desde WhatsApp**: la preview del link en WhatsApp ES la conversión. Requisito MVP explícito: meta tags **Open Graph** en la página pública (server-rendered o pre-renderizadas) y flujo de aporte que funcione dentro del in-app browser de WhatsApp (sin popups/redirects frágiles).
15. **Contribuyente anónimo**: "anónimo" oculta el nombre en la página pública pero SÍ cuenta en el total y en el número de participantes; el organizador siempre ve los datos completos en su dashboard y CSV.

## 2. Funcionalidades propuestas para el MVP

Alineadas 1:1 con el flujo de validación pedido (crear cuenta → crear campaña → link público → compartir → aportar con pago simulado → administrar):

- Registro/login con email + contraseña (JWT).
- Crear campaña con un único tipo habilitado al inicio (ej. "Colecta/Recaudación" genérica), pero usando el mecanismo de `CampaignType` configurable (no hardcodeado) para que agregar "Venta" o "Evento" después sea configuración, no refactor.
- Campos de campaña: título, descripción, imagen (opcional, con placeholder), slug, meta económica, fecha inicio/término, estado (`draft/active/paused/finished/suspended`).
- Página pública: imagen, descripción, barra de progreso, meta, lista de participantes (nombre + monto, con opción de anónimo), botón "Aportar", botón compartir (link + WhatsApp), QR generado en el momento (no lo pre-generamos y guardamos, se genera on-the-fly desde el slug), y meta tags Open Graph para la preview en WhatsApp (riesgo 14).
- Flujo de aporte: formulario simple (nombre, monto, contacto opcional) → `MockPaymentProvider` → confirmación → actualiza monto recaudado.
- Dashboard organizador: listar/crear/editar/finalizar campañas, ver participantes, exportar CSV, ver monto recaudado (bruto/neto de comisión).
- Auditoría mínima: tabla `AuditLogs` enfocada en cambios de estado de campaña y (futuro) payouts — auditar transiciones de pagos mock aporta poco; se amplía al integrar la pasarela real. Sin UI de auditoría todavía.
- El `MockPaymentProvider` debe poder **simular reintentos y duplicados**, para que la idempotencia (riesgo 8) llegue probada a la integración con la pasarela real, no solo especificada.
- Publicar la campaña directo desde el formulario de creación (sin pasar obligatoriamente por el dashboard), para cumplir el objetivo de <2 minutos.

## 3. Funcionalidades explícitamente fuera del MVP (cortadas)

- Rifas (mencionadas como tipo futuro, pero deshabilitadas — no se implementa su lógica, solo se deja el `CampaignType` reservado).
- Comentarios, actualizaciones, galería, videos en la página pública (se deja el layout preparado, sin backend).
- Organizaciones/equipos con múltiples administradores y permisos granulares (modelo de datos lo soporta vía `Organizations`, pero sin UI ni invitaciones en el MVP).
- Dominio personalizado por campaña.
- Integración real de pago (Webpay/Mercado Pago/Stripe/Khipu) — solo la abstracción `PaymentProvider` + mock.
- Google OAuth u otros proveedores de identidad.
- Notificaciones (email/WhatsApp automatizado) más allá del link para compartir manualmente.
- Dark mode implementado (se deja preparado a nivel de design tokens, no se construye el toggle).
- Panel de administración de la plataforma (super-admin) — solo el dashboard del organizador.

## 4. Justificación de decisiones clave que arrastran a Etapa 2

- **Clean Architecture "cuando aporte valor"**: dado que es 1 desarrollador por meses, se recomienda una capa de dominio + aplicación + infraestructura simple en Go (sin excesivo desacoplamiento tipo hexagonal puro), evitando interfaces por cada repositorio si no hay más de una implementación real todavía — excepto en `PaymentProvider`, donde el desacoplamiento es explícitamente pedido y valioso.
- **sqlc vs GORM**: se recomendará sqlc en Etapa 2 (SQL explícito, tipado fuerte generado, mejor rendimiento y control sobre migraciones complejas de un dominio financiero) — se justificará con detalle comparativo en esa etapa, no aquí.
- **Comisión porcentual**: implica que `Payment` es la entidad central de reporting financiero desde el día 1, no un anexo.

---

## Siguiente paso

Este documento es la Etapa 1. Al aprobarla, la Etapa 2 (arquitectura completa: capas, módulos backend/frontend, decisión final sqlc vs GORM, estructura de `PaymentProvider`) se aborda en un turno separado — **no se escribirá código todavía**.
