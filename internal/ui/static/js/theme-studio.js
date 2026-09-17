(function () {
  "use strict";

  var XSL_URL = "/ui/api/themes/assets/theme-to-css.xsl";
  var DEBOUNCE_MS = 400;
  var cssXsl = null;
  var timer = null;

  function parseXMLString(xmlString) {
    return new DOMParser().parseFromString(xmlString, "application/xml");
  }

  async function fetchXML(url) {
    var r = await fetch(url, { credentials: "same-origin" });
    if (!r.ok) throw new Error("Failed to fetch " + url);
    var text = await r.text();
    return parseXMLString(text);
  }

  function transformToText(xmlDoc, xslDoc) {
    if (typeof XSLTProcessor === "undefined") {
      throw new Error("XSLTProcessor not supported");
    }
    var proc = new XSLTProcessor();
    proc.importStylesheet(xslDoc);
    var frag = proc.transformToFragment(xmlDoc, document);
    return (frag && frag.textContent) || "";
  }

  function injectPreview(cssText) {
    var el = document.getElementById("theme-preview-style");
    if (!el) {
      el = document.createElement("style");
      el.id = "theme-preview-style";
      document.head.appendChild(el);
    }
    el.textContent = cssText || "";
  }

  function previewFromTextarea() {
    var ta = document.getElementById("theme-xml");
    if (!ta || !cssXsl) return;
    try {
      var doc = parseXMLString(ta.value);
      var parseErr = doc.querySelector("parsererror");
      if (parseErr) {
        console.warn("[theme-studio] XML parse error");
        return;
      }
      var css = transformToText(doc, cssXsl);
      injectPreview(css);
    } catch (e) {
      console.error("[theme-studio]", e);
    }
  }

  function schedulePreview() {
    if (timer) clearTimeout(timer);
    timer = setTimeout(previewFromTextarea, DEBOUNCE_MS);
  }

  async function init() {
    var ta = document.getElementById("theme-xml");
    if (!ta) return;

    try {
      cssXsl = await fetchXML(XSL_URL);
    } catch (e) {
      console.error("[theme-studio] failed to load XSL:", e);
      return;
    }

    ta.addEventListener("input", schedulePreview);
    previewFromTextarea();
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", init);
  } else {
    init();
  }
})();
