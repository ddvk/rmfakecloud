(function () {
  "use strict";

  var XSL_URL = "/ui/api/themes/assets/theme-to-css.xsl";
  var DEBOUNCE_MS = 400;
  var cssXsl = null;
  var timer = null;
  var syncing = false;

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

  function colorInputs() {
    return Array.prototype.slice.call(document.querySelectorAll(".theme-color-input"));
  }

  function syncPickersFromXML() {
    var ta = document.getElementById("theme-xml");
    if (!ta) return;
    try {
      var doc = parseXMLString(ta.value);
      if (doc.querySelector("parsererror")) return;
      var colors = doc.querySelector("colors");
      if (!colors) return;
      syncing = true;
      colorInputs().forEach(function (input) {
        var key = input.getAttribute("data-color-key");
        if (!key) return;
        var val = colors.getAttribute(key);
        if (val && /^#[0-9A-Fa-f]{3,8}$/.test(val)) {
          input.value = val.length === 4
            ? "#" + val[1] + val[1] + val[2] + val[2] + val[3] + val[3]
            : val;
        }
      });
      syncing = false;
    } catch (_) {
      syncing = false;
    }
  }

  function applyPickersToXML() {
    var ta = document.getElementById("theme-xml");
    if (!ta) return;
    try {
      var doc = parseXMLString(ta.value);
      if (doc.querySelector("parsererror")) return;
      var colors = doc.querySelector("colors");
      if (!colors) return;
      colorInputs().forEach(function (input) {
        var key = input.getAttribute("data-color-key");
        if (key && input.value) colors.setAttribute(key, input.value);
      });
      var serialized = new XMLSerializer().serializeToString(doc);
      if (serialized && serialized !== ta.value) {
        ta.value = serialized;
      }
    } catch (e) {
      console.warn("[theme-studio] color sync failed", e);
    }
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

    syncPickersFromXML();
    colorInputs().forEach(function (input) {
      input.addEventListener("input", function () {
        if (syncing) return;
        applyPickersToXML();
        schedulePreview();
      });
    });
    ta.addEventListener("input", function () {
      syncPickersFromXML();
      schedulePreview();
    });
    var form = document.getElementById("theme-studio-form");
    if (form) {
      form.addEventListener("submit", function () {
        applyPickersToXML();
      });
    }
    previewFromTextarea();
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", init);
  } else {
    init();
  }
})();
