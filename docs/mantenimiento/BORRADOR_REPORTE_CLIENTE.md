# Borrador — Reporte de mantenimiento (cliente)

**Proyecto:** LBTech API — producción  
**Environment:** `lbtechapi-prod` · https://api.lbonetrack.com  
**Región AWS:** us-east-1  
**Estado del borrador:** en progreso — completar cuando termine el cambio de instancia  

Documento interno para elaborar el informe de mantenimiento facturable al cliente. Detalle técnico ampliado en [`REPORTE_PROD_INVESTIGACION.md`](../../REPORTE_PROD_INVESTIGACION.md).

---

## Evidencia capturada — 28 jun 2026 (~19:29 UTC)

**Archivo:** [`eb-prod-degraded-cambio-instancia-2026-06-28.png`](./eb-prod-degraded-cambio-instancia-2026-06-28.png)

**Qué muestra la captura (consola AWS Elastic Beanstalk):**

| Elemento | Valor observado |
|----------|-----------------|
| Salud del environment | **Degraded** — 4 problemas de estado |
| Banner superior | Actualización del environment **en curso** |
| Aviso plataforma | **Plataforma obsoleta** (Docker AL2 3.4.17 — Deprecated) |
| Versión aplicación | `prod-6b98c8d538816006a64d07c240b494e2663e4c07` |
| Dominio EB | `lbtechapi-prod.eba-p26pfjvm.us-east-1.elasticbeanstalk.com` |

**Eventos visibles en la captura:**

| Hora (UTC) | Tipo | Descripción |
|------------|------|-------------|
| 19:29:02 | WARN | Salud pasó de Info a **Degraded**. Procesos ELB no sanos en **1 de 2** instancias. Actualización de configuración en curso (~2 min). |
| 19:29:02 | INFO | Nueva instancia agregada: `i-0244f547049f1dc85` |
| 19:28:00 | INFO | Esperando terminación de instancias EC2 |
| 19:28:00 | INFO | Eliminación de políticas del Auto Scaling group |

**Contexto:** captura tomada durante el **cambio de tipo de instancia** (`t2.micro` → `t3.small`) acordado para mitigar lentitud intermitente por generación de PDFs (Chromium) en la misma máquina que sirve la API.

> Completar después: captura final con salud **Ok** y tipo **t3.small** confirmado en Configuración → Escalado.

---

## Trabajo realizado / en curso (para factura)

Completar y marcar al cerrar el mantenimiento:

- [ ] Diagnóstico de incidentes **504** y lentitud intermitente en producción
- [ ] Análisis de logs EB (bundle CloudWatch / stdouterr) — correlación con PDF `move report`
- [ ] Revisión de métricas CloudWatch (CPU, créditos CPU, NetworkOut, ALB)
- [ ] Documentación de arquitectura actual y recomendaciones (`REPORTE_PROD_INVESTIGACION.md`)
- [ ] Upgrade instancia EC2 prod: **t2.micro → t3.small** (Elastic Beanstalk, sin cambio de código)
- [ ] Verificación post-cambio: `/health`, eventos EB, operación normal de API
- [ ] (Pendiente / futuro) Migración plataforma EB AL2 → AL2023 (Deprecated)
- [ ] (Pendiente / futuro) Worker o servicio externo para PDFs si persisten problemas

**Horas / ítems facturables:** _completar_

---

## Resumen ejecutivo (borrador para cliente — lenguaje no técnico)

_Para copiar/adaptar al informe final:_

Durante la revisión de producción de la API LBTech se identificaron periodos de **indisponibilidad temporal (error 504)** y **lentitud intermitente**, asociados a la generación automática de **reportes PDF** (movimientos de tracking) en un servidor con recursos limitados (**instancia t2.micro**: 1 GB RAM).

Se realizó análisis de logs y monitoreo en AWS, y se aplicó **mejora de capacidad** cambiando la instancia de producción a **t3.small** (2 GB RAM, mayor CPU) para reducir saturación sin modificar la aplicación ni el flujo de despliegue.

Queda documentado un aviso de **plataforma Elastic Beanstalk obsoleta**; se recomienda planificar su actualización en una ventana de mantenimiento posterior.

---

## Problema → causa → acción (tabla para el reporte)

| Problema reportado | Causa raíz identificada | Acción tomada |
|--------------------|-------------------------|---------------|
| 504 / API caída puntual | Reinicio/bootstrap instancia; ALB sin targets sanos | Monitoreo; upgrade instancia |
| Lentitud intermitente | PDF con Chromium compitiendo por CPU/RAM con la API | Upgrade **t2.micro → t3.small** |
| PDF no generado (casos aislados) | Timeouts tras múltiples reintentos en instancia pequeña | Misma mejora de instancia; seguimiento 2–4 semanas |
| Aviso "Deprecated" en consola | Plataforma Docker Amazon Linux 2 sin soporte activo | Pendiente migración AL2023 (no urgente) |

---

## Seguimiento post-mantenimiento

| Fecha | Acción | Resultado |
|-------|--------|-----------|
| 2026-06-28 | Captura EB Degraded durante cambio instancia | Evidencia guardada |
| 2026-06-28 | Inicio cambio t2.micro → t3.small | En curso en captura |
| | Environment Ok + t3.small confirmado | _pendiente_ |
| | /health estable 24–48 h | _pendiente_ |
| | Primer movimiento con PDF sin ABANDONED | _pendiente_ |

---

## Anexos

| Documento | Uso |
|-----------|-----|
| [`eb-prod-degraded-cambio-instancia-2026-06-28.png`](./eb-prod-degraded-cambio-instancia-2026-06-28.png) | Evidencia consola AWS (Degraded + update en curso) |
| [`REPORTE_PROD_INVESTIGACION.md`](../../REPORTE_PROD_INVESTIGACION.md) | Detalle técnico completo |
| Bundle logs (si aplica) | `BundleLogs-*` descargado de EB |

---

## Notas para facturación

- Trabajo de **investigación + infraestructura**; no incluye redeploy de código de aplicación.
- Cambio de instancia implica **costo AWS recurrente** adicional (~USD 7/mes vs micro) — informar al cliente si corresponde.
- Guardar captura final **Ok** + email/fecha de entrega del reporte al cliente.
