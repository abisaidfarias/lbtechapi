# Reporte de Mantenimiento Mensual

**Plataforma:** LBOneTrack (LB Technology SPA)
**Cliente:** Camilo Espinoza — LBOneTrack
**Desarrollador / Mantenedor:** Abisaid Farías
**Período cubierto:** 1 – 30 de agosto de 2026
**Fecha de emisión:** 30 de agosto de 2026
**Repositorios:** `lbtechapi` (backend, Go) · `lbtechfront` (frontend, Angular)
**Ambientes:** `dev` (api.testkoi.com) · `prod` (api.lbonetrack.com)

---

## 1. Resumen ejecutivo

Agosto tuvo dos frentes bien diferenciados: el **desarrollo de un módulo nuevo completo** — el Reporte Automático de Certificación Multi-banda — y el **mantenimiento correctivo** del módulo de Control de Embarque puesto en producción el mes anterior.

El módulo nuevo es la pieza mayor del período: **4.317 líneas de código en 25 archivos**, con **36 pruebas automatizadas**, que sustituyen un proceso hoy manual. Hasta ahora, al terminar la fase de pruebas de un proyecto Multi-banda, el ingeniero debía **redactar el informe por fuera de la plataforma y subirlo como archivo**. A partir de este desarrollo, la plataforma le presenta un formulario, valida que no falte nada y **genera el PDF automáticamente**, con el formato oficial y la trazabilidad de cada resultado guardada en base de datos.

En paralelo se atendieron **tres defectos** detectados en el módulo de Control de Embarque tras su paso a producción, uno de ellos con **pérdida silenciosa de información** del usuario, y se ejecutaron cuatro mejoras funcionales solicitadas durante el mes.

En total, **7 entregas** (4 en backend, 3 en frontend), todas validadas en el ambiente de desarrollo antes de su publicación.

---

## 2. Contexto — antecedente del mes anterior

En julio se puso en producción el **certificado OABI del Control de Embarque**, migrando su generación desde un navegador headless (Chromium) a generación nativa, lo que redujo el tiempo de emisión de minutos —con fallos frecuentes— a aproximadamente 50 milisegundos.

Agosto continuó ese trabajo en dos direcciones:

1. **Estabilizar lo entregado.** Todo módulo recién puesto en producción revela detalles que solo aparecen con uso real. Se corrigieron tres defectos, detallados en la sección 3.2.

2. **Aplicar el mismo enfoque a un segundo documento.** La generación nativa de PDF demostró ser confiable con el certificado OABI, por lo que se reutilizó esa misma técnica para el nuevo Reporte Multi-banda, evitando reintroducir la dependencia de navegador que ya había causado un incidente de producción documentado.

---

## 3. Trabajo realizado en el período

### 3.1 Reporte Automático de Certificación Multi-banda (desarrollo nuevo)

**Situación previa.** Al completar la fase de pruebas de un proyecto Multi-banda, el ingeniero de laboratorio debía preparar el informe de evaluación **fuera de la plataforma** —en una planilla— y luego subirlo como documento adjunto. Esto implicaba: formato inconsistente entre informes, ningún control de que estuvieran todos los resultados, resultados que **no quedaban registrados como datos** (solo dentro de un PDF), y ninguna posibilidad de auditar, corregir o reemitir sin rehacer el archivo completo.

**Qué se construyó.** Un módulo completo que reemplaza ese paso manual. La plataforma ahora:

- Abre un formulario al terminar la fase de pruebas.
- **Precarga** los datos que ya conoce del equipo (fabricante, modelos, versiones, valor SAR, sistema operativo).
- Solicita al ingeniero únicamente lo que la plataforma no puede saber: los resultados de las pruebas y la evidencia fotográfica.
- **Valida que no falte nada** antes de permitir la emisión.
- **Genera el PDF** con el formato oficial y lo almacena en el proyecto.

#### Alcance funcional cubierto

| Bloque | Qué registra | Volumen |
|---|---|---|
| **Información del dispositivo** | 7 campos precargados + 7 que completa el ingeniero, más el sello de certificación y los operadores probados | 14 campos |
| **Evidencia fotográfica** | 3 capturas obligatorias (versión de software, notificación emergente, historial de alertas), con su escenario y operador | 3 imágenes + metadatos |
| **SIMLOCK** | 2 verificaciones contra cada operadora probada | Hasta 14 resultados |
| **Multi-banda** | Resultado consolidado por cada banda de frecuencia, en las 4 tecnologías (2G, 3G, 4G y 5G) | 17 bandas |
| **SAE** | 11 pruebas × 4 operadoras × 2 canales de radio × hasta 2 escenarios | Hasta 176 resultados |
| **FM Radio** | Declaración de soporte y prueba condicional | 2 campos |

En su configuración más exigente, un informe registra **hasta 227 datos individuales**, cada uno almacenado por separado en base de datos —no dentro de un PDF— lo que permite auditarlos, corregirlos y reemitir el documento sin rehacer nada.

#### Decisiones técnicas relevantes

**Guardado parcial.** El proceso de laboratorio no se completa en una sola sesión. El módulo permite **guardar a medias y retomar otro día**, sin exigir que el informe esté completo. El PDF se genera **solo cuando se solicita**, y puede reemitirse cuantas veces sea necesario.

**Validación en el servidor, no solo en pantalla.** Las reglas —qué campos son obligatorios, qué resultados exigen justificación, qué bloques aplican según el tipo de proceso— se verifican en el servidor. Una validación que vive únicamente en la interfaz puede evadirse y tiende a desincronizarse; ubicarla en el servidor garantiza que ningún informe incompleto llegue a emitirse.

**Justificación obligatoria de fallos.** Cada resultado FAIL o NO OK **exige un comentario**. Con hasta 176 celdas, un comentario olvidado dentro de una matriz es prácticamente invisible: por eso el sistema los reúne en un panel aparte que enumera exactamente cuáles faltan.

**Indicador de avance calculado por el servidor.** La pantalla muestra el porcentaje completado, qué falta por sección y qué está impidiendo emitir el documento. Este cálculo se resolvió en el servidor —junto a las reglas de validación— en lugar de reimplementarlo en la interfaz, evitando dos definiciones distintas de "qué es obligatorio" que inevitablemente divergen.

**Catálogos configurables.** Las operadoras, bandas, pruebas y sellos se almacenan como datos de catálogo, no incrustados en el código. Un cambio normativo futuro —una banda nueva, una operadora que entra o sale— se resuelve editando datos, sin rediseñar el formulario ni el documento.

**Generación nativa del PDF.** Se aplicó la misma técnica validada en julio con el certificado OABI: sin navegador headless. El documento resultante tiene 4 páginas, con encabezados que se repiten al cortar página, comentarios que se ajustan sin truncarse y resaltado de los resultados fallidos.

**Sin veredicto general.** Por requerimiento explícito de la especificación funcional aprobada, el documento **no calcula ni muestra un resultado global** de aprobación o rechazo: presenta los resultados individuales, que constituyen el resultado completo.

#### Cobertura de pruebas

Se implementaron **36 pruebas automatizadas** que verifican las reglas de negocio críticas: que un borrador acepte datos incompletos pero rechace valores inválidos; que la emisión exija cada campo obligatorio; que solo se pidan resultados de las operadoras efectivamente probadas; que las pruebas de tecnologías no soportadas se marquen automáticamente como no aplicables; y que el indicador de avance calcule correctamente en cada escenario.

### 3.2 Mantenimiento correctivo — Control de Embarque

Tres defectos detectados tras la puesta en producción de julio:

| # | Defecto | Impacto | Causa |
|---|---|---|---|
| 1 | **Pérdida de datos al avanzar de fase** | Los campos ID y Validación ingresados por el usuario **se borraban solos** al mover el proyecto de fase | El formulario de cambio de fase no reenviaba esos campos, y el sistema los sobrescribía con vacío. Se agregó una protección que conserva el valor guardado cuando el formulario no lo envía. |
| 2 | **Error al guardar (código 400)** | No se podía guardar un control de embarque; el guardado fallaba sin explicación clara | Incompatibilidad de tipos: la interfaz enviaba un valor numérico donde el servidor esperaba texto. Se implementó un tipo tolerante que acepta ambos formatos. |
| 3 | **Número de certificado corrupto** | El número de control del certificado se degradaba en cada reemisión (`123` → `%2560123`), apareciendo así impreso en el documento oficial | El número se releía desde la dirección web del archivo almacenado, acumulando codificación en cada ciclo. Se implementó autocorrección: cuando el sistema detecta un número malformado, lo reconstruye desde el identificador de cliente. |

El defecto 1 es el de mayor gravedad: no producía error visible: el usuario guardaba, el sistema respondía correctamente, y la información desaparecía sin aviso.

### 3.3 Mejoras funcionales

| Mejora | Descripción |
|---|---|
| **Campos ID y Validación** | Incorporados al control de embarque en los tres flujos: alta individual, edición y carga masiva por planilla (reconociendo las columnas del archivo del cliente, con o sin acentos) |
| **Nombre compuesto del certificado** | El certificado OABI se descarga con un nombre que identifica el envío (`N° control - ID - rework - registro OABI`) en lugar de un código suelto, resuelto sin alterar la ruta de almacenamiento para no reintroducir el defecto 3 |
| **Correos de cambio de fase** | Incorporan los campos nuevos y se reorganizó la distribución de la tabla, que presentaba varios campos apilados en una sola columna |
| **Correo de alta de dispositivo** | Rediseñado por completo: presentaba sus ~50 campos como una lista plana sin formato. Ahora agrupa la información en 11 secciones con el mismo lenguaje visual del resto de las notificaciones |

### 3.4 Validación técnica y despliegue

- El proyecto **compila sin errores**, pasa el análisis estático y ejecuta su suite de pruebas sin fallos nuevos en cada entrega.
- Cada cambio se validó primero en el ambiente de desarrollo, incluyendo pruebas manuales del flujo completo.
- Las publicaciones se realizaron mediante el **pipeline de despliegue continuo** (GitHub Actions → Docker Hub → AWS Elastic Beanstalk), con verificación automática de disponibilidad posterior.

---

## 4. Alcance del mantenimiento recurrente (mensual)

Independientemente de los desarrollos del período, el servicio contempla de forma continua:

- **Revisión de logs** de aplicación e infraestructura en busca de errores recurrentes o patrones anómalos.
- **Atención de defectos** reportados por los usuarios o detectados de forma proactiva.
- **Auditoría de dependencias** (GitHub Dependabot) para identificar vulnerabilidades conocidas.
- **Verificación de integridad del build**: compilación, análisis estático y pruebas automatizadas tras cada cambio.
- **Gestión de despliegues** a desarrollo y producción.
- **Documentación de incidentes y decisiones técnicas**, para mantener trazabilidad y evitar reinvestigar lo ya resuelto.
- **Ajustes de rendimiento y estabilidad** ante incidentes.

---

## 5. Riesgos identificados

> Riesgos documentados con evidencia del repositorio y de las herramientas de seguridad de GitHub al 30-08-2026.

### 5.1 Seguridad

| Riesgo | Estado | Impacto potencial |
|---|---|---|
| **19 vulnerabilidades de dependencias sin remediar** (7 críticas, 4 altas, 8 moderadas) | **Sin cambios respecto a julio** — sigue abierto | Exposición a fallos de seguridad conocidos en librerías de terceros. La severidad crítica implica un posible vector explotable de forma remota. |
| **Plataforma Elastic Beanstalk marcada como obsoleta** por AWS | Sin cambios | Fin de soporte de parches de seguridad a nivel de sistema operativo |
| **Credenciales de desarrollo en texto plano** dentro de la instancia | Sin cambios | Si un atacante accede a la instancia o a una copia de ella, las credenciales quedan expuestas |
| **CORS sin política restrictiva confirmada** en producción | Sin cambios | Orígenes no controlados podrían interactuar con la API |

Los cuatro riesgos de seguridad **permanecen sin atender desde el reporte anterior**. No formaron parte del alcance de agosto, dedicado al desarrollo del módulo nuevo y a la corrección de defectos en producción.

### 5.2 Infraestructura y disponibilidad

- **El patrón de riesgo que causó el incidente de producción persiste** en los reportes de movimiento, que siguen generándose con navegador headless en la misma instancia que atiende la API. Tanto el certificado OABI (julio) como el nuevo reporte Multi-banda (agosto) ya no lo hacen; **los reportes de movimiento son el último módulo que falta migrar**.
- El **tipo de instancia de producción** sigue sin confirmarse ni documentarse.
- **No existen alertas automáticas** de salud o de tasa de errores: la detección de incidentes depende de que alguien lo note.

### 5.3 Código y mantenibilidad

- El pipeline de integración continua **construye y publica sin ejecutar las pruebas automatizadas** como paso obligatorio. Se ejecutan manualmente antes de cada entrega. Con 36 pruebas nuevas incorporadas este mes, el valor de automatizarlas en el pipeline aumentó considerablemente.
- **Nuevo:** el módulo de reporte Multi-banda incorpora **imágenes de sello provisionales**. Las cinco imágenes definitivas están pendientes de entrega por parte del cliente. Los archivos actuales aproximan el diseño y son reemplazables sin modificar código, pero **el documento no debe emitirse a un cliente final hasta que se incorporen las definitivas**.

### 5.4 Continuidad operativa

- **Concentración de conocimiento en una sola persona**: el diagnóstico de infraestructura y el proceso de despliegue dependen de un único mantenedor.
- Documentación operativa dispersa en múltiples archivos sin un índice central.

---

## 6. Recomendaciones priorizadas

### Corto plazo

1. **Entregar las cinco imágenes de sello definitivas** para habilitar la emisión del nuevo reporte a clientes finales. Es el único bloqueante funcional del módulo.
2. **Triage y remediación de las 19 vulnerabilidades** reportadas por Dependabot, priorizando las 7 críticas. Pendiente desde julio.
3. **Incorporar la ejecución de pruebas al pipeline** antes de cada publicación.
4. **Configurar alertas** de salud y de errores 5xx.

### Mediano plazo

5. **Migrar los reportes de movimiento** a generación nativa, cerrando el último módulo que mantiene la dependencia de navegador que ya causó un incidente de producción.
6. **Actualizar la plataforma Elastic Beanstalk** a una versión con soporte activo.
7. **Endurecer el manejo de credenciales** en producción.
8. **Revisar y restringir la política de CORS**.

### Largo plazo

9. Ampliar la cobertura de pruebas de integración a los flujos críticos restantes.
10. Observabilidad: registro estructurado y métricas.
11. Plan de continuidad operativa y reducción de la concentración de conocimiento.

---

## 7. Conclusión

Agosto entregó un **módulo nuevo completo** que elimina un proceso manual del flujo de certificación Multi-banda: el informe de evaluación deja de redactarse fuera de la plataforma y pasa a generarse automáticamente, con sus resultados almacenados como datos auditables en lugar de quedar encerrados en un archivo. Son 4.317 líneas de código nuevas respaldadas por 36 pruebas automatizadas.

En paralelo se corrigieron tres defectos del módulo puesto en producción el mes anterior, incluido uno que **borraba información del usuario sin dar aviso** —el tipo de defecto que erosiona la confianza en una herramienta más rápido que una caída visible.

Persisten los riesgos de seguridad e infraestructura señalados en el reporte anterior, **sin variación**: las 19 vulnerabilidades de dependencias, la plataforma obsoleta y la ausencia de alertas automáticas. No formaron parte del alcance de este mes y requieren tiempo dedicado; conviene planificarlos antes de que deriven en un incidente, en lugar de atenderlos como respuesta a uno.
