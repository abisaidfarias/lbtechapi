# Proceso — reporte de mantenimiento mensual

Regla de Cursor asociada: `.cursor/rules/reporte-mantenimiento-mensual.mdc` (se aplica al trabajar en `docs/mantenimiento/` o cuando pidas “reporte mensual”).

## Objetivo

Entregar cada mes un informe **legible para el cliente** (PDF) y uno **técnico** (Markdown en el repo) sobre el trabajo en **LBOneTrack** (`lbtechapi` + `lbtechfront`).

## Archivos de referencia

| Uso | Ruta |
|-----|------|
| Índice de reportes | [README.md](./README.md) |
| Plantilla técnica | [REPORTE_PLANTILLA.md](./REPORTE_PLANTILLA.md) |
| Plantilla visual cliente | [REPORTE_MANTENIMIENTO_SEPTIEMBRE_2026_CLIENTE.html](./REPORTE_MANTENIMIENTO_SEPTIEMBRE_2026_CLIENTE.html) (copiar y renombrar) |
| Estilos PDF cliente | [styles/report-client.css](./styles/report-client.css) |
| Generador PDF | [generate-client-pdf.cjs](./generate-client-pdf.cjs) |
| Ejemplo técnico reciente | [REPORTE_MANTENIMIENTO_SEPTIEMBRE_2026.md](./REPORTE_MANTENIMIENTO_SEPTIEMBRE_2026.md) |

## Pasos cada mes

1. **Recopilar trabajo:** `git log` en api y front desde el mes anterior; notas de incidentes, despliegues, módulos tocados.
2. **Crear Markdown técnico:** `REPORTE_MANTENIMIENTO_<MES>_<AÑO>.md` desde plantilla.
3. **Crear HTML cliente:** copiar el `_CLIENTE.html` del mes previo, actualizar textos y tablas (lenguaje simple, beneficios para el usuario).
4. **Generar PDF:**
   ```bash
   cd docs/mantenimiento
   npm install puppeteer --no-save
   node generate-client-pdf.cjs REPORTE_MANTENIMIENTO_<MES>_<AÑO>_CLIENTE.html
   ```
5. **Revisar PDF:** sin páginas en blanco, sin código HTML visible, portada en una sola hoja.
6. **Actualizar** [README.md](./README.md) con el nuevo período.

## Reglas de redacción (cliente)

- Explicar **qué cambió para el usuario**, no cómo está implementado.
- Evitar: nombres de endpoints, commits, “backend/front”, SSM, MongoDB.
- Incluir flujos: “nuevo usuario”, “olvidé contraseña”, módulos de negocio (Multi-banda, embarque, etc.).
- Trabajo terminado a principios del mes calendario siguiente → **incluirlo en el reporte del mes que se está cerrando** si fue parte del mismo ciclo de facturación.

## Reglas de redacción (técnico)

- Commits, archivos, riesgos Dependabot, EB, CI, deuda técnica.
- Mermaid opcional en Markdown (GitHub); no depender de Mermaid en el PDF cliente.

## Errores conocidos

| Error | Solución |
|-------|----------|
| PDF muestra `<div>` y tags | No usar `md-to-pdf` sobre `.html`; usar `generate-client-pdf.cjs` |
| Página casi vacía | No usar saltos de página manuales; portada `height: 297mm` en CSS |
| Mermaid no en PDF | Solo en `.md`; en cliente usar listas numeradas |

## Marca

- Azul `#0b1736`, acento `#8DB319` (alineado a correos LB OneTrack).
