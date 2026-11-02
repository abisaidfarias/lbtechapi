const puppeteer = require("puppeteer");
const fs = require("fs");
const path = require("path");

async function main() {
  const dir = __dirname;
  const htmlName =
    process.argv[2] || "REPORTE_MANTENIMIENTO_SEPTIEMBRE_2026_CLIENTE.html";
  const htmlPath = path.join(dir, htmlName);
  const cssPath = path.join(dir, "styles", "report-client.css");
  const pdfPath = htmlPath.replace(/\.html$/i, ".pdf");

  let html = fs.readFileSync(htmlPath, "utf8");
  const css = fs.readFileSync(cssPath, "utf8");
  html = html.replace(
    /<link rel="stylesheet" href="[^"]+"\s*\/>/,
    `<style>${css}</style>`
  );

  const browser = await puppeteer.launch({
    headless: true,
    args: ["--no-sandbox", "--disable-setuid-sandbox"],
  });
  const page = await browser.newPage();
  await page.setContent(html, { waitUntil: "networkidle0" });
  await page.pdf({
    path: pdfPath,
    format: "A4",
    printBackground: true,
    preferCSSPageSize: true,
    margin: { top: "0", right: "0", bottom: "0", left: "0" },
  });
  await browser.close();
  console.log("Wrote", pdfPath);
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});
