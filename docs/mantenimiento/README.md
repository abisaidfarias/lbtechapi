# Reportes de mantenimiento LBOneTrack

Informes mensuales para el cliente (Markdown → PDF / Word).

**Proceso fijo (no perder entre sesiones):** [PROCESO_REPORTE_MENSUAL.md](./PROCESO_REPORTE_MENSUAL.md) · regla Cursor `.cursor/rules/reporte-mantenimiento-mensual.mdc`

| Período | Archivo |
|---------|---------|
| Julio 2026 | [REPORTE_MANTENIMIENTO_JULIO_2026.md](./REPORTE_MANTENIMIENTO_JULIO_2026.md) |
| Agosto 2026 | [REPORTE_MANTENIMIENTO_AGOSTO_2026.md](./REPORTE_MANTENIMIENTO_AGOSTO_2026.md) |
| Septiembre 2026 (técnico) | [REPORTE_MANTENIMIENTO_SEPTIEMBRE_2026.md](./REPORTE_MANTENIMIENTO_SEPTIEMBRE_2026.md) |
| Septiembre 2026 (cliente / PDF) | [REPORTE_MANTENIMIENTO_SEPTIEMBRE_2026_CLIENTE.html](./REPORTE_MANTENIMIENTO_SEPTIEMBRE_2026_CLIENTE.html) |

**Nuevo mes:** copiar [REPORTE_PLANTILLA.md](./REPORTE_PLANTILLA.md) → `REPORTE_MANTENIMIENTO_<MES>_<AÑO>.md`.

Borradores e investigación: [BORRADOR_REPORTE_CLIENTE.md](./BORRADOR_REPORTE_CLIENTE.md)

---

## Formato “más atractivo” (checklist)

Cada reporte puede incluir:

1. **En un vistazo** — tabla de indicadores + semáforo 🟢🟡🔴  
2. **Índice** — enlaces a secciones  
3. **Mapa visual del mes** — 1–2 diagramas [Mermaid](https://mermaid.js.org/) (flujo o secuencia)  
4. **Entregas del período** — tabla cronológica commit × fecha × repo  
5. **Antes / después** — tabla cuando haya métrica (tiempo, errores, pasos manuales eliminados)

Los diagramas Mermaid se ven en **GitHub** y en **VS Code** (Markdown Preview Mermaid Support).

---

## Exportar a PDF con marca LB

### Opción A — md-to-pdf (Node, funciona en Windows sin Pandoc)

```bash
cd docs/mantenimiento
npx --yes md-to-pdf REPORTE_MANTENIMIENTO_SEPTIEMBRE_2026.md --css styles/report-print.css
```

Genera `REPORTE_MANTENIMIENTO_SEPTIEMBRE_2026.pdf` junto al `.md`.

**Versión para cliente (lenguaje simple, portada LB):**

No uses `md-to-pdf` sobre el `.html` (imprime el código). Usa Puppeteer:

```bash
cd docs/mantenimiento
npm install puppeteer --no-save
node generate-client-pdf.cjs
```

Salida: `REPORTE_MANTENIMIENTO_SEPTIEMBRE_2026_CLIENTE.pdf` (HTML renderizado con estilos).

### Opción B — Pandoc

Requiere [Pandoc](https://pandoc.org/) instalado.

```bash
cd docs/mantenimiento
pandoc REPORTE_MANTENIMIENTO_SEPTIEMBRE_2026.md -o Reporte_Septiembre_2026.pdf \
  --css=styles/report-print.css \
  -V lang=es
```

Para **Word** (editable por el cliente):

```bash
pandoc REPORTE_MANTENIMIENTO_SEPTIEMBRE_2026.md -o Reporte_Septiembre_2026.docx
```

> Pandoc no renderiza Mermaid en PDF por defecto. Para PDF con diagramas: previsualizar en VS Code → **Imprimir → Guardar como PDF**, o exportar diagramas como PNG desde [mermaid.live](https://mermaid.live) e insertar imágenes en el `.md` antes del pandoc.

### Opción C — VS Code

1. Abrir el `.md` → vista previa (Ctrl+Shift+V).  
2. Extensión **Markdown PDF** o imprimir desde el navegador de preview.  
3. Opcional: pegar CSS de `styles/report-print.css` en la config de la extensión.

### Opción D — GitHub

Subir el `.md` al repo → el cliente ve tablas y Mermaid en el navegador → **Imprimir / PDF** desde el browser.

---

## Colores de marca (referencia)

| Uso | Color |
|-----|--------|
| Encabezado / títulos | `#0b1736` |
| Acento / barra | `#8DB319` |
| Texto | `#1f2937` |

Definidos también en correos HTML de la plataforma (`password_action.html`, notificaciones).
