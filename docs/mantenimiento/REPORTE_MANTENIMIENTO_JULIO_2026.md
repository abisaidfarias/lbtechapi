# Reporte de Mantenimiento Mensual

**Plataforma:** LBOneTrack (LB Technology SPA)
**Cliente:** Camilo Espinoza — LBOneTrack
**Desarrollador / Mantenedor:** Abisaid Farías
**Período cubierto:** 1 – 28 de julio de 2026
**Fecha de emisión:** 28 de julio de 2026
**Repositorios:** `lbtechapi` (backend, Go) · `lbtechfront` (frontend, Angular)
**Ambientes:** `dev` (api.testkoi.com) · `prod` (api.lbonetrack.com)

---

## 1. Resumen ejecutivo

Durante julio de 2026 se registraron **21 commits de mantenimiento y mejora** entre los dos repositorios que componen la plataforma (10 en backend, 11 en frontend), cubriendo desde correcciones de despliegue hasta un rediseño completo del módulo de generación de certificados OABI, que presentaba fallos recurrentes de generación tanto en ambiente de desarrollo como en producción.

El hallazgo y la corrección más relevantes del mes fueron sobre el **certificado de Control de Embarque (OABI)**: el documento no se lograba generar de forma confiable en los ambientes desplegados (dev/prod), quedando la interfaz mostrando "Certificate generation is taking longer than expected" de forma indefinida. Se diagnosticó la causa raíz, se corrigió en una primera fase, y finalmente se **rediseñó el mecanismo de generación para eliminar por completo la dependencia de un navegador headless (Chromium)**, pasando a generación nativa en el lenguaje del backend. El resultado medido: de un proceso que tardaba minutos y frecuentemente fallaba por timeout, a uno que **genera el documento en ~50 milisegundos**, de forma idéntica en ambiente local, desarrollo y producción, sin importar el tamaño del servidor.

Adicionalmente se revisó y validó el flujo de despliegue continuo (CI/CD) existente — GitHub Actions → Docker Hub → AWS Elastic Beanstalk, con control de aprobación manual antes de tocar producción — confirmando que sigue operando correctamente tras los cambios, y se corrigieron varios problemas puntuales del pipeline de despliegue del frontend (invalidación de caché, YAML de integración continua).

Este reporte también documenta, con evidencia concreta del repositorio y de las herramientas de seguridad de GitHub, los **riesgos activos de la plataforma** que no fueron parte del alcance de este mes pero que representan exposición real si no se atienden — en particular, vulnerabilidades de dependencias sin remediar y un patrón de riesgo de infraestructura (generación de PDF pesada compartiendo servidor con la API) que aún persiste en otro módulo del sistema.

---

## 2. Contexto — antecedente del mes anterior

En el ciclo de mantenimiento previo (documentado en `REPORTE_PROD_INVESTIGACION.md` y `docs/mantenimiento/BORRADOR_REPORTE_CLIENTE.md`) se había diagnosticado un incidente de producción: caídas intermitentes (**504**) y lentitud severa de la API, causadas por la generación de reportes PDF de movimiento (Chromium headless) compitiendo por CPU y memoria con el propio servidor de la API, en una instancia EC2 de tamaño reducido. Se documentaron 8 intentos de generación abandonados por timeout durante un solo incidente.

La mitigación aplicada en ese momento fue de infraestructura (aumento de tamaño de instancia). Quedó pendiente, como recomendación de mediano plazo, **aislar o eliminar la dependencia de Chromium** del flujo de generación de documentos. El trabajo de este mes resuelve exactamente ese pendiente para el módulo de certificados OABI (ver sección 3.1); el mismo patrón de riesgo permanece abierto para los reportes de movimiento (ver sección 6).

---

## 3. Trabajo realizado en el período

### 3.1 Certificado OABI — diagnóstico, corrección y rediseño (backend + frontend)

Este fue el trabajo principal del mes, ejecutado en tres etapas sobre la rama `dev`, validado, y promovido a `main` (producción):

| Etapa | Commit | Descripción |
|---|---|---|
| Fase 1 — estabilización | `1d0d247` | Timeouts de renderizado recalibrados (el valor previo, 90s, estaba dimensionado para una máquina local, no para el servidor de despliegue); propagación correcta de cancelación al proceso de Chrome para evitar procesos huérfanos; se empezó a persistir la causa real del fallo (timeout / Chrome no disponible / error de subida) en lugar de un mensaje genérico; compresión de una imagen de logo embebida que pesaba 1.34 MB (33.211 × 7.000 px para mostrarse a 220 px) a 38 KB. |
| Fase 2 — rediseño de fondo | `eaa4b07` | El certificado es un documento A4 de formato fijo donde solo ~15 campos de texto varían. Se reemplazó el motor de generación (HTML + Chromium headless) por **generación nativa de PDF en Go** (librería `go-pdf/fpdf`), eliminando por completo la dependencia de un navegador para este flujo. |
| Fase 3 — pulido de datos | `a00348f` | Corrección de una columna de la tabla de guías que se superponía visualmente con datos largos (número de rework); renombrado de columna para mayor claridad ("Imeis registrados"); corrección de un número de control de certificado que se corrompía progresivamente por una codificación de URL mal manejada (autorreparación: el sistema ahora reconstruye el número desde el identificador de cliente cuando detecta un valor previamente corrompido). |

**Resultado medido:**

| Métrica | Antes | Después |
|---|---|---|
| Tiempo de generación | Minutos; timeout frecuente (nunca completaba en dev/prod) | ~0.05 segundos |
| Dependencia de navegador Chromium | Sí (crítica) | No |
| Tamaño de imagen embebida (logo) | 1.34 MB | 38 KB |
| Comportamiento local vs. servidor | Distinto (solo funcionaba en local) | Idéntico |
| Visibilidad del error real ante fallo | Mensaje genérico | Causa específica (timeout / navegador / almacenamiento) |

En el frontend (`lbtechfront`), en paralelo se implementó el flujo de generación asíncrona con sondeo de estado (`78c1313`) y se corrigió el comportamiento posterior a la generación exitosa, que abría automáticamente el PDF en una pestaña nueva del navegador — ahora el certificado simplemente queda marcado como adjunto dentro del mismo modal, y el usuario lo abre solo si lo desea (`6061a30`).

### 3.2 Otros cambios de backend en el período

| Fecha | Commit | Descripción |
|---|---|---|
| 03-07 | `5b8d1bd` | Migración del envío de correo a Amazon SES vía SMTP; corrección de plantillas de notificación. |
| 07-07 | `c0debc4` | Mejoras al correo de control de embarque; switch opcional entre SES y SMTP tradicional vía variable de entorno. |
| 15-07 | `5548ae9` | Corrección de una validación de unicidad de multibanda que se ejecutaba innecesariamente al actualizar un control de embarque sin cambiar su asociación. |
| 17-07 | `c8c9249` / `e551c19` | Base inicial del módulo de certificado OABI (antes del rediseño de la sección 3.1); sello oficial embebido en el documento. |
| 18-07 | `c202d82` / `6cadceb` | Generación asíncrona con pool de Chrome (versión inicial, posteriormente reemplazada); corrección de arranque perezoso del motor PDF para no bloquear el despliegue en dev. |

### 3.3 Cambios de frontend en el período

| Fecha | Commit | Descripción |
|---|---|---|
| 03-07 | `4f0caa7` | Corrección de despliegue: servir `index.html` actualizado tras publicar en S3 (evitaba que quedara una versión vieja cacheada). |
| 07-07 | `965f116` | Alineación de layout de formularios y columnas de tablas masivas (multibanda / control de embarque). |
| 15-07 | `8fed469` | Corrección de despliegue: evitar cacheo prolongado de archivos de idioma (i18n) para que producción reciba traducciones nuevas sin esperar. |
| 15-07 | `85ab5ac` | Corrección de orden de imports que impedía a Angular resolver un servicio compartido correctamente. |
| 17-07 | `c82a80b` | Flujo de "Under Revision", campos de company, versión de sistema operativo — base necesaria para los datos que hoy alimentan el certificado OABI. |
| 17-07 | `cc91b0e` | Corrección: se perdía el número de certificado SUBTEL al reemplazar un informe de prueba en multibanda. |
| 17-07 | `40f9e3f` | Corrección de invalidación de caché de traducciones en despliegue, evitando etiquetas obsoletas en producción. |
| 17-07 | `94e26c2` / `f01684f` | Corrección de la configuración YAML de integración continua (pipeline de despliegue), rota por un reemplazo de texto mal aplicado. |
| 18-07 | `78c1313` | Generación asíncrona del certificado OABI con sondeo de estado desde la interfaz. |
| 24-07 | `6061a30` | El certificado ya no se abre automáticamente en una pestaña nueva tras generarse. |

### 3.4 Validación técnica y despliegue

- Tras incorporar la librería `go-pdf/fpdf` (necesaria para la generación nativa de PDF), se ejecutó `go mod tidy` y se confirmó que **todo el proyecto compila sin errores** (`go build ./...`), pasa análisis estático (`go vet`) y las pruebas automatizadas existentes (`go test`) sin fallos nuevos.
- Los cambios se probaron primero en el ambiente `dev`, incluyendo pruebas manuales del flujo completo (generación, estado, visualización del PDF resultante) antes de promoverlos.
- La promoción a `main` (producción) se realizó a través del **pipeline de despliegue continuo ya existente** (GitHub Actions → Docker Hub → AWS Elastic Beanstalk), el cual quedó **validado end-to-end este mes**: build de imagen, publicación, espera de disponibilidad del ambiente, despliegue y verificación de salud (`/health`) automática — con el control de **aprobación manual obligatoria antes de tocar producción**, que operó correctamente durante esta promoción.

---

## 4. Alcance del mantenimiento recurrente (mensual)

Independientemente de si el mes presenta un hallazgo mayor (como el de este período), el servicio de mantenimiento mensual de la plataforma LBOneTrack contempla de forma recurrente:

- **Revisión de logs de aplicación e infraestructura** (Elastic Beanstalk, nginx, Docker) en busca de errores recurrentes o patrones anómalos.
- **Monitoreo de reportes de error** generados por los usuarios o detectados en los ambientes de desarrollo y producción.
- **Auditoría de dependencias y librerías** (GitHub Dependabot) para identificar vulnerabilidades conocidas antes de que sean explotadas.
- **Verificación de integridad del build**: confirmar que el proyecto compila, pasa análisis estático y pruebas automatizadas tras cada cambio relevante.
- **Gestión de despliegues** a los ambientes de desarrollo y producción, incluyendo la coordinación de ventanas de mantenimiento cuando el cambio lo amerita.
- **Documentación de incidentes y hallazgos técnicos**, para mantener trazabilidad y evitar que un mismo problema se re-investigue desde cero en el futuro.
- **Ajustes de rendimiento y estabilidad** ante incidentes reportados por el cliente o detectados proactivamente.

---

## 5. Riesgos identificados

> Riesgos documentados con evidencia directa del repositorio, del historial de commits y de las herramientas de seguridad de GitHub al 28-07-2026. Se incluyen tanto para justificar la necesidad de mantenimiento continuo como para dimensionar trabajo adicional recomendado.

### 5.1 Seguridad

| Riesgo | Evidencia | Impacto potencial |
|---|---|---|
| **19 vulnerabilidades de dependencias sin remediar** en el repositorio backend (7 críticas, 4 altas, 8 moderadas), reportadas por GitHub Dependabot al 28-07-2026. | Alerta automática de GitHub en cada push a `dev`/`main`. Una auditoría previa (`MEJORAS_FUTURAS.md`, sección 5.2) ya había registrado 24 vulnerabilidades en diciembre — la cifra cambia con cada release de dependencias de terceros y requiere revisión activa, no puntual. | Exposición a fallos de seguridad conocidos (CVEs) en librerías de terceros; severidad crítica/alta implica posible vector explotable remotamente según el paquete afectado. |
| **Plataforma Elastic Beanstalk (Docker sobre Amazon Linux 2) marcada como obsoleta** ("Deprecated") por la consola de AWS. | Documentado en `REPORTE_PROD_INVESTIGACION.md`. | Fin de soporte de parches de seguridad a nivel de sistema operativo/runtime base. |
| **Manejo de secretos de desarrollo en texto plano dentro de la instancia.** El script en `.ebextensions/01-environment.config` obtiene credenciales de AWS SSM Parameter Store y las escribe en un archivo `.env` sin cifrar en el disco de la instancia EC2. El equivalente para producción no está confirmado en el repositorio (`SECRETS_MANAGEMENT.md`, pendiente). | Lectura directa del archivo de configuración. | Si un atacante obtiene acceso a la instancia (o a una imagen/snapshot de ella), las credenciales quedan expuestas directamente. |
| **CORS sin política restrictiva confirmada en producción.** | `MEJORAS_FUTURAS.md`, sección 5.3 — abierto desde diciembre. | Orígenes no controlados podrían interactuar con la API si no se restringe explícitamente. |
| **Sin evidencia en el repositorio de rate limiting o WAF** delante del balanceador de carga. | No se encontró configuración de este tipo en el código ni en `.ebextensions`. Se recomienda confirmar directamente en la consola de AWS si existe a nivel de infraestructura. | Sin esta capa, la API queda más expuesta a abuso por volumen de tráfico o intentos automatizados. |

### 5.2 Infraestructura y disponibilidad

- **El mismo patrón de riesgo que causó el incidente de producción de este año persiste en otro módulo.** El certificado OABI ya no depende de Chromium (resuelto este mes), pero los **reportes de movimiento ("move reports") siguen generándose con Chromium headless en la misma instancia que atiende el tráfico de la API**. Un pico de generación de reportes de movimiento puede degradar la API completa, igual que ocurrió con los certificados.
- El **tipo de instancia EC2 de producción no está confirmado ni documentado** en el repositorio — checklist abierto desde el incidente anterior sin cerrar.
- **No hay alertas automáticas** de salud (`/health`) o de tasa de errores 5xx — la detección de incidentes depende de que alguien lo note manualmente o de un reporte del cliente.
- El ambiente de `dev` cumple también función de entorno compartido de desarrollo activo, sin un ambiente de *staging* claramente separado para validación previa a producción.

### 5.3 Código y mantenibilidad

- El pipeline de integración continua (`deploy-dev.yml`, `deploy-prod.yml`) **construye y despliega la imagen sin ejecutar la suite de pruebas automatizadas (`go test`) como parte del flujo**. Las pruebas se corren manualmente antes de cada push; un cambio que rompa un test podría llegar a producción sin que el pipeline lo detenga.
- La cobertura de pruebas automatizadas es puntual, no integral; no existen pruebas de integración end-to-end para flujos críticos de negocio (generación de certificados, reportes de movimiento, notificaciones).
- Los cambios de este mes tocaron áreas sensibles — manejo de secretos, envío de correo, generación de documentos — que se beneficiarían de pruebas de regresión dedicadas.

### 5.4 Continuidad operativa

- **Concentración de conocimiento en una sola persona** (bus factor): el diagnóstico de infraestructura AWS y el proceso de despliegue dependen de un único desarrollador/mantenedor.
- La documentación operativa está dispersa en múltiples archivos Markdown en la raíz del repositorio (`MEJORAS_FUTURAS.md`, `SECRETS_MANAGEMENT.md`, `DEPLOYMENT_ECONOMICO.md`, `AWS_DEPLOYMENT_GUIDE.md`, entre otros) sin un índice central que oriente por dónde empezar ante una eventualidad.

---

## 6. Recomendaciones priorizadas

### Corto plazo (impacto alto, esfuerzo bajo-medio)

1. Triage y remediación de las 19 vulnerabilidades reportadas por Dependabot, priorizando las 7 críticas.
2. Confirmar y documentar el tipo de instancia EC2 de producción; cerrar el checklist abierto desde el incidente previo.
3. Incorporar `go test ./...` como paso obligatorio del pipeline de CI antes de cada despliegue.
4. Configurar alertas básicas de salud (`/health`) y errores 5xx (CloudWatch o servicio externo tipo UptimeRobot).

### Mediano plazo

5. Aplicar a los reportes de movimiento la misma migración realizada este mes sobre el certificado OABI: generación nativa o, alternativamente, aislar Chromium en un servicio/worker separado de la instancia que sirve la API.
6. Migrar la plataforma Elastic Beanstalk (Docker AL2) a una versión con soporte activo (AL2023).
7. Confirmar y endurecer el manejo de secretos en producción (Parameter Store, sin archivos `.env` en texto plano en la instancia).
8. Revisar y restringir la política de CORS en producción.

### Largo plazo

9. Ampliar cobertura de pruebas de integración para los flujos de negocio críticos.
10. Observabilidad: logging estructurado y métricas/dashboards (CloudWatch o Prometheus/Grafana).
11. Plan de continuidad operativa: documentación centralizada y reducción de la concentración de conocimiento en una sola persona.

---

## 7. Conclusión

El mantenimiento de julio resolvió un problema activo y visible para el usuario final (generación de certificados OABI fallando de forma consistente), con una solución que no es un parche temporal sino un **cambio de arquitectura que elimina la clase de problema completa** para ese flujo — mismo patrón que ya había costado un incidente de producción documentado meses atrás. El pipeline de despliegue continuo fue ejercitado y validado de extremo a extremo, incluyendo su control de aprobación manual en producción.

Persisten riesgos reales y verificables — vulnerabilidades de dependencias sin remediar, el mismo patrón de riesgo de infraestructura en otro módulo, ausencia de alertas automáticas y de pruebas obligatorias en el pipeline — que no forman parte del alcance habitual de un mantenimiento mensual reactivo y que justifican destinar trabajo dedicado a cerrarlos antes de que deriven en un nuevo incidente.
