/**
 * EPUB website viewer: keep the contents list in sync with the iframe.
 */
(function () {
  "use strict";

  function currentPath(frame) {
    try {
      if (frame.contentWindow && frame.contentWindow.location) {
        return frame.contentWindow.location.pathname;
      }
    } catch (_) {}
    try {
      return new URL(frame.getAttribute("src") || "", window.location.href).pathname;
    } catch (_) {
      return "";
    }
  }

  function markToc(frame, toc) {
    var path = currentPath(frame);
    toc.querySelectorAll("a[href]").forEach(function (a) {
      var hrefPath = "";
      try {
        hrefPath = new URL(a.getAttribute("href"), window.location.href).pathname;
      } catch (_) {}
      if (hrefPath && path && hrefPath === path) {
        a.setAttribute("aria-current", "page");
      } else {
        a.removeAttribute("aria-current");
      }
    });
  }

  function init() {
    var frame = document.getElementById("epub-frame");
    var toc = document.querySelector(".epub-toc");
    if (!frame || !toc) return;
    markToc(frame, toc);
    frame.addEventListener("load", function () {
      markToc(frame, toc);
    });
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", init);
  } else {
    init();
  }
})();
