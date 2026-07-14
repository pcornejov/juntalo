# Probar Juntalo gratis en Render

Esto despliega el MVP completo (backend + frontend + Postgres) en Render.com, gratis, sin tarjeta de crédito y sin instalar nada localmente. Es para **probar**, no para producción real — ver limitaciones al final.

## Pasos

1. Entra a [render.com](https://render.com) y crea una cuenta (puedes usar tu cuenta de GitHub).
2. En el dashboard, click **New** → **Blueprint**.
3. Conecta el repositorio `pcornejov/juntalo` y selecciona la rama `claude/juntalo-saas-mvp-m1hk0l` (o `main` si ya se fusionó).
4. Render detecta automáticamente el archivo `render.yaml` en la raíz del repo y te muestra 3 recursos a crear:
   - `juntalo-db` (Postgres, free)
   - `juntalo-api` (backend Go, free)
   - `juntalo-web` (frontend estático, free)
5. Click **Apply**. Render construye y despliega los tres servicios (toma unos 5-10 minutos la primera vez, principalmente compilando el backend en Go).
6. Cuando `juntalo-api` termine de desplegar, corre las migraciones automáticamente al arrancar (`RUN_MIGRATIONS_ON_BOOT=true` — ver nota abajo).
7. Abre la URL de `juntalo-web` (algo como `https://juntalo-web.onrender.com`) — ahí está la app.

No necesitas configurar nada más: `render.yaml` ya conecta el frontend con el backend y el backend con la base de datos, generando los secretos (`JWT_SECRET`, `MOCK_WEBHOOK_SECRET`) automáticamente.

## Probar el flujo completo

1. Regístrate en `juntalo-web`.
2. Crea una campaña (queda en `draft`).
3. Publícala.
4. Abre el link público (`/public/<slug>`) — o mejor, comparte el link `/c/<slug>` por WhatsApp desde tu celular para ver la preview con imagen y título funcionando de verdad (esto solo se puede probar con una URL pública real, por eso vale la pena este deploy).
5. Aporta desde el celular sin crear cuenta. El pago es simulado (`MockPaymentProvider` en modo `deferred`): queda "pendiente" ~1.5s y se confirma solo, vía un webhook real que el backend se manda a sí mismo.
6. Vuelve al dashboard y revisa participantes / exporta el CSV.

## Limitaciones del free tier (por eso es solo para probar)

- **El backend "duerme"** tras 15 minutos sin tráfico. El primer request después de dormir tarda ~30-50 segundos en responder — no es que esté roto, solo está despertando.
- **Las imágenes de portada son efímeras por defecto**: Render free no permite disco persistente en web services, así que cualquier imagen subida se pierde en el próximo redeploy o reinicio del servicio — salvo que se configure Cloudflare R2 (ver siguiente sección), que sí persiste entre deploys y es gratis dentro de un umbral generoso.
- **La base de datos gratis se borra sola a los ~30 días** de creada. Si sigues probando después de ese plazo, hay que recrear el Blueprint (o solo la base) y el schema se vuelve a crear solo gracias a `RUN_MIGRATIONS_ON_BOOT`.
- Los pagos son siempre simulados (`MockPaymentProvider`) — no hay dinero real involucrado en ningún punto.

## Pasarela de pago real (Webpay Plus)

Por defecto el deploy sigue usando `MockPaymentProvider` — no requiere nada extra y es la forma más rápida de probar el flujo completo. Para probar el pago real (con tarjetas de prueba, sin dinero de verdad) contra el ambiente de integración de Transbank:

1. En el dashboard de Render, en `juntalo-api` → **Environment**, agrega estas 3 variables (no vienen declaradas en `render.yaml` — Render permite agregar variables custom aunque el blueprint no las liste):
   - `WEBPAY_COMMERCE_CODE` = `597055555532`
   - `WEBPAY_API_KEY` = `579B532A7440BB0C9079DED94D31EA1615BACEB56610332264630D42D0A36B1C`
   - `WEBPAY_ENVIRONMENT` = `integration`

   Estas credenciales son **públicas**: Transbank las publica iguales para todos los que están integrando (no son un secreto tuyo) — por eso es seguro pegarlas directo, a diferencia de `RESEND_API_KEY` o las credenciales de R2.
2. Redeploy manual de `juntalo-api`.
3. Al aportar, en vez de confirmarse al toque (mock), el navegador te manda a la página real de Webpay. Usa una tarjeta de prueba (Transbank las publica en su documentación; para el formulario de autenticación con RUT y clave, el de prueba es RUT `11.111.111-1` / clave `123`).
4. Para pasar a producción real: reemplaza las 3 variables por las credenciales de tu comercio afiliado real en Transbank (`WEBPAY_ENVIRONMENT=production`), y cárgalas como secretas (edita `render.yaml` para agregarlas con `sync: false` en vez de dejarlas sueltas en el dashboard, para que quede documentado en el repo que existen sin exponer su valor).

## Storage persistente de imágenes (Cloudflare R2)

Por defecto las imágenes se guardan en el disco del contenedor de `juntalo-api`, que Render borra en cada redeploy o reinicio. Para que las imágenes persistan de verdad:

1. En el dashboard de Cloudflare, crea un bucket en **R2** (tiene capa gratis generosa, sin tarjeta).
2. Activa **Public access** en el bucket (o conecta un dominio custom) para obtener la URL pública base.
3. En **R2 → Manage R2 API Tokens**, crea un token con permisos de lectura/escritura sobre ese bucket — te da un Account ID, Access Key ID y Secret Access Key.
4. En el dashboard de Render, en el servicio `juntalo-api` → **Environment**, completa a mano estas 5 variables (ya están declaradas en `render.yaml` con `sync: false`, así que Render las pide pero no las genera ni las commitea):
   - `R2_ACCOUNT_ID`
   - `R2_ACCESS_KEY_ID`
   - `R2_SECRET_ACCESS_KEY`
   - `R2_BUCKET`
   - `R2_PUBLIC_URL` (la URL pública del paso 2, sin `/` al final)
5. Redeploy manual de `juntalo-api` para que tome las variables nuevas.

Si estas 5 variables quedan vacías, el backend sigue funcionando normal y cae automáticamente al disco local (efímero) — no hace falta configurarlas para que el deploy de prueba funcione, solo para que las imágenes persistan.

## Diferencia con el despliegue real (Etapa 5/6 del roadmap)

El diseño de producción (`docker/docker-compose.prod.yml` + Caddy + VPS) sirve frontend y backend bajo **un solo dominio**, con Cloudflare delante y backups automáticos — así se evita el tema de cookies/CORS cross-origin que sí existe en este deploy de prueba (frontend y backend viven en subdominios distintos de Render). Por eso el código ahora soporta ambos modos: en desarrollo y en el VPS real, todo es same-origin (cookie `SameSite=Lax`); en este deploy de prueba cross-origin, la cookie de sesión usa `SameSite=None` automáticamente cuando `APP_ENV=production` viene de un dominio distinto al frontend.
