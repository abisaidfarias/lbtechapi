# Reporte de investigación — Producción LBTech API

**Fecha:** 31 mayo 2026  
**Environment:** `lbtechapi-prod` (us-east-1)  
**URL:** https://api.lbonetrack.com  
**Versión EB desplegada (al analizar):** `prod-6b98c8d…` (último merge a `main` ~20 jun 2026)

Documento para retomar mejoras de infraestructura, rendimiento y observabilidad.

---

## Resumen ejecutivo

| Síntoma | Causa probable | Severidad |
|---------|----------------|-----------|
| **504** en prod | Instancia reiniciándose / no bootstrapped; ALB sin targets sanos | Alta (downtime) |
| **Lentitud intermitente** | Generación de PDF con Chromium bloqueando recursos de la misma instancia | Alta (degradación) |
| **499** en nginx (health) | Timeouts del ELB mientras la app estaba saturada | Media (riesgo de marcar instancia unhealthy) |

No hubo crashes ni panics en logs de aplicación durante el incidente analizado.

---

## Infraestructura conocida

### Confirmado en repo / despliegue

| Componente | Valor |
|------------|-------|
| Elastic Beanstalk | `lbtechapi-prod` |
| Región | `us-east-1` |
| Imagen Docker | `abisaid/lbtechapi:prod-latest` |
| Plataforma EB | Docker on Amazon Linux 2 (versión obsoleta ~3.4.17 — aviso en consola) |
| Disco (`.ebextensions/02-root-volume.config`) | **gp3, 30 GB** |
| Bundle prod (GitHub Actions) | Incluye `02-root-volume.config` y `03-nginx.config`; **no** incluye `01-environment.config` (secrets dev) |
| Health check post-deploy | `curl -f http://lbtechapi-prod.eba-p26pfjvm.us-east-1.elasticbeanstalk.com/health` tras 90 s |

### No definido en el repositorio

| Dato | Notas |
|------|-------|
| **Tipo de instancia EC2** | Se configura en EB → Configuración → Capacidad. **No** está en `.ebextensions` ni en `deploy-prod.yml` |
| Documentación interna sugiere | `t2.micro` como objetivo económico (`DEPLOYMENT_ECONOMICO.md`, `MEJORAS_FUTURAS.md`) — **confirmar en consola** |

### Comportamiento observado en prod

- Environment **Load Balanced** (ALB).
- Health checks del ELB desde IPs internas (ej. `172.31.34.32`, `172.31.18.207`).
- Tras reinicio/bootstrap, `/health` respondió **200** con payload `lbtechprod`.
- `/health` en estado normal: ~**30 µs**; durante incidente PDF: **10+ segundos**.

### Cómo verificar tipo de instancia

**Consola:** EB → `lbtechapi-prod` → Configuración → Capacidad → tipo de instancia EC2.

**CLI:**

```powershell
aws ec2 describe-instances --region us-east-1 `
  --filters "Name=tag:elasticbeanstalk:environment-name,Values=lbtechapi-prod" `
  --query "Reservations[*].Instances[*].[InstanceType,State.Name]" `
  --output table
```

---

## Incidente 504 (reinicio / bootstrap)

### Síntoma

- Clientes recibían **504 Gateway Timeout**.
- No era un problema de CORS en origen.

### Causa

- Instancia EC2 en proceso de reinicio o bootstrap.
- Logs indicaban estado tipo **"Instance has NOT been bootstrapped"** y actividad de `docker pull` al arrancar.
- ALB sin instancias sanas → 504.

### Resolución observada

- Tras completar bootstrap y pull de imagen, `/health` volvió a **200**.
- API operativa de nuevo sin redeploy del usuario (~1 mes sin subir código).

---

## Incidente lentitud intermitente (PDF / Chromium)

### Disparador

- **18:49:58** — `PUT /api/v1/device-tracking` desde `lbonetrack.com/tracking`.
- Disparó generación de **move report PDF** (`tracking_id=260626008`) vía Chromium headless en el mismo contenedor/proceso que la API.

### Evolución en logs

| Hora (aprox.) | Evento |
|---------------|--------|
| 18:49 | Inicio generación PDF |
| 18:51 – 21:46 | **8 intentos lentos**, todos **timeout** → estado `ABANDONED` |
| Durante el periodo | `/health` degradado (10+ s); nginx **499** en checks del ELB |
| Tras 21:46 | API normal (~30 µs en `/health`) |

### Archivos de log útiles (bundle EB)

1. `var/log/eb-docker/containers/eb-current-app/eb-*-stdouterr.log` — aplicación Go
2. `var/log/nginx/access.log` — 499 / 502 / 504
3. `var/log/eb-engine.log`, `var/log/docker-events.log`, `var/log/messages` — reinicios, OOM, Docker

**Bundle analizado:** `BundleLogs-1782518591490` (descargado desde consola EB).

### Conclusión técnica

Chromium + PDF en la **misma instancia pequeña** que sirve tráfico HTTP satura CPU/RAM y bloquea el event loop / workers de Gin. El ALB interpreta la instancia como lenta o unhealthy.

Si la instancia es `t2.micro` / `t3.micro` (1 vCPU, ~1 GB RAM), el comportamiento es esperable.

---

## Configuración relevante del código

| Archivo | Relevancia |
|---------|------------|
| `Dockerfile` | `golang:latest` + Chromium; `ENVIRONMENT=prod` |
| `.github/workflows/deploy-prod.yml` | Deploy a EB; health check único a los 90 s |
| `.ebextensions/02-root-volume.config` | Volumen root 30 GB gp3 |
| `.ebextensions/03-nginx.config` | Proxy nginx hacia la app |
| `MEJORAS_FUTURAS.md` | Backlog previo (SSM prod, CORS, etc.) |

---

## Recomendaciones priorizadas

### Corto plazo (impacto alto, esfuerzo bajo–medio)

1. **Confirmar tipo de instancia** en EB y documentarlo en este archivo o en `AWS_DEPLOYMENT_GUIDE.md`.
2. **Subir capacidad** si es micro: mínimo **`t3.small`** (2 vCPU, 2 GB) para convivir API + Chromium ocasional.
3. **Alertas en `/health`** (CloudWatch, UptimeRobot, etc.) con umbral de latencia (ej. > 2 s).
4. **`GIN_MODE=release`** en prod si no está ya configurado en variables de entorno EB.
5. **Regenerar PDF abandonado** `tracking_id=260626008` si el negocio lo requiere.

### Medio plazo

6. **Aislar generación PDF** del request path síncrono:
   - Cola (SQS) + worker en instancia separada o Lambda con layer Chromium.
   - O servicio dedicado solo para PDFs.
7. **Timeouts y límites** en `move_report_pdf`: cancelación, reintentos con backoff, no bloquear el pool HTTP.
8. **Actualizar plataforma EB** (Docker AL2 soportada actual) para parches de seguridad.

### Largo plazo

9. **Observabilidad:** métricas de latencia por ruta, logs estructurados, correlación `tracking_id`.
10. **Secrets prod en SSM** (`/lbtechapi/prod/`) — ver `SECRETS_MANAGEMENT.md` y `MEJORAS_FUTURAS.md`.
11. **Revisar CORS** en prod (orígenes restrictivos).

---

## Checklist para retomar mejoras

- [ ] Anotar aquí el **InstanceType** real de prod: `_______________`
- [ ] Anotar **número de instancias** (Min/Max en EB Capacity): `_______________`
- [ ] Decidir: **resize instancia** vs **worker PDF separado**
- [ ] Implementar alerta `/health`
- [ ] Revisar código de generación PDF (ruta, timeouts, async)
- [ ] Actualizar plataforma EB
- [ ] Opcional: añadir `InstanceType` a `.ebextensions` para infra as code

---

## Comandos y referencias

| Recurso | Ubicación |
|---------|-----------|
| Script lectura AWS | `check-aws-current-state.sh` |
| Guía deployment económico | `DEPLOYMENT_ECONOMICO.md` |
| Backlog mejoras | `MEJORAS_FUTURAS.md` |
| Secrets | `SECRETS_MANAGEMENT.md` |
| Workflow prod | `.github/workflows/deploy-prod.yml` |

---

## Notas de sesión

- Usuario **no desplegó** recientemente (~1 mes); último merge relevante `6b98c8d`.
- AWS CLI local sin credenciales válidas durante la investigación (`InvalidClientTokenId`) — verificación de instancia pendiente en consola.
- Tras reinicio de instancia, prod respondió **200** en `/health`; problema principal recurrente es **contención de recursos por PDF**, no bugs de crash en Go.

---

*Actualizar este documento cuando se confirme el tipo de instancia o se implementen cambios.*
