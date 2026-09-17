(async function () {
  const el = document.getElementById("page-xml");
  if (!el) return;
  const status = document.getElementById("xslt-status");
  try {
    const xmlText = el.textContent;
    const xml = new DOMParser().parseFromString(xmlText, "application/xml");
    if (xml.querySelector("parsererror")) {
      throw new Error("Invalid page XML");
    }
    const xslText = await fetch("/assets/xslt/shell.xsl", { credentials: "same-origin" }).then((r) => {
      if (!r.ok) throw new Error("Failed to load shell.xsl");
      return r.text();
    });
    const xsl = new DOMParser().parseFromString(xslText, "application/xml");
    if (xsl.querySelector("parsererror")) {
      throw new Error("Invalid shell.xsl");
    }
    const proc = new XSLTProcessor();
    proc.importStylesheet(xsl);
    const out = proc.transformToDocument(xml);
    let html = new XMLSerializer().serializeToString(out);
    html = html.replace(/<\?xml[^?]*\?>\s*/i, "");
    if (!/^\s*<!DOCTYPE/i.test(html)) {
      html = "<!DOCTYPE html>\n" + html;
    }
    document.open();
    document.write(html);
    document.close();
  } catch (err) {
    if (status) {
      status.textContent = "UI render failed: " + (err && err.message ? err.message : String(err));
    }
    console.error("[xslt]", err);
  }
})();
