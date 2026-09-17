(function () {
  "use strict";

  var PDFJS_CDN = "https://cdnjs.cloudflare.com/ajax/libs/pdf.js/3.11.174/pdf.min.js";
  var WORKER_CDN = "https://cdnjs.cloudflare.com/ajax/libs/pdf.js/3.11.174/pdf.worker.min.js";

  function loadScript(src) {
    return new Promise(function (resolve, reject) {
      var existing = document.querySelector('script[src="' + src + '"]');
      if (existing && window.pdfjsLib) {
        resolve();
        return;
      }
      var s = document.createElement("script");
      s.src = src;
      s.async = true;
      s.onload = function () {
        resolve();
      };
      s.onerror = function () {
        reject(new Error("Failed to load " + src));
      };
      document.head.appendChild(s);
    });
  }

  async function ensurePdfJs() {
    if (window.pdfjsLib) return window.pdfjsLib;
    await loadScript(PDFJS_CDN);
    if (!window.pdfjsLib) throw new Error("pdfjsLib not available after CDN load");
    window.pdfjsLib.GlobalWorkerOptions.workerSrc = WORKER_CDN;
    return window.pdfjsLib;
  }

  async function renderPdf(url, container) {
    var pdfjsLib = await ensurePdfJs();
    container.textContent = "Loading PDF…";

    var loadingTask = pdfjsLib.getDocument({
      url: url,
      withCredentials: true,
    });
    var pdf = await loadingTask.promise;
    container.textContent = "";

    var scale = parseFloat(container.getAttribute("data-scale") || "1.25");
    if (!Number.isFinite(scale) || scale <= 0) scale = 1.25;

    for (var pageNum = 1; pageNum <= pdf.numPages; pageNum++) {
      var page = await pdf.getPage(pageNum);
      var viewport = page.getViewport({ scale: scale });

      var wrap = document.createElement("div");
      wrap.className = "pdf-page";
      wrap.setAttribute("data-page", String(pageNum));

      var canvas = document.createElement("canvas");
      canvas.width = viewport.width;
      canvas.height = viewport.height;
      canvas.setAttribute("aria-label", "Page " + pageNum);
      wrap.appendChild(canvas);
      container.appendChild(wrap);

      var ctx = canvas.getContext("2d");
      await page.render({ canvasContext: ctx, viewport: viewport }).promise;
    }
  }

  async function init() {
    var viewer = document.getElementById("pdf-viewer");
    if (!viewer) return;

    var url = viewer.getAttribute("data-doc-url");
    if (!url) {
      console.error("[pdfview] missing data-doc-url on #pdf-viewer");
      return;
    }

    var container = document.getElementById("pdf-canvas-container");
    if (!container) {
      container = document.createElement("div");
      container.id = "pdf-canvas-container";
      viewer.appendChild(container);
    }

    try {
      await renderPdf(url, container);
    } catch (e) {
      console.error("[pdfview]", e);
      container.textContent = "Failed to load PDF: " + (e.message || String(e));
    }
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", init);
  } else {
    init();
  }
})();
