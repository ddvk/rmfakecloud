/**
 * Paginated notebook / annotated PDF·EPUB viewer.
 * Notebooks: SVG pages. Annotated docs: PNG composites (bg × .rm).
 */
(function () {
  "use strict";

  function clamp(n, lo, hi) {
    if (n < lo) return lo;
    if (n > hi) return hi;
    return n;
  }

  function init() {
    var viewer = document.getElementById("nb-viewer");
    if (!viewer) return;
    var img = document.getElementById("nb-page");
    var status = document.getElementById("nb-page-status");
    var prev = document.getElementById("nb-prev");
    var next = document.getElementById("nb-next");
    if (!img) return;

    var docid = viewer.getAttribute("data-doc-id") || "";
    var mode = (viewer.getAttribute("data-mode") || "svg").toLowerCase();
    var pages = parseInt(viewer.getAttribute("data-pages") || "1", 10);
    if (!pages || pages < 1) pages = 1;
    var page = parseInt(viewer.getAttribute("data-page") || "1", 10);
    page = clamp(page, 1, pages);

    function pageURL(n) {
      if (mode === "png") {
        return "/ui/api/documents/" + encodeURIComponent(docid) + "/page/" + n;
      }
      return "/documents/" + encodeURIComponent(docid) + "/page/" + n + "/svg";
    }

    function show(n) {
      page = clamp(n, 1, pages);
      img.src = pageURL(page);
      var label = "Page " + page + " of " + pages;
      img.alt = label;
      if (status) status.textContent = label;
      if (prev) prev.disabled = page <= 1;
      if (next) next.disabled = page >= pages;
      viewer.setAttribute("data-page", String(page));
      try {
        var url = new URL(window.location.href);
        url.searchParams.set("page", String(page));
        history.replaceState(null, "", url.pathname + url.search);
      } catch (_) {}
    }

    if (prev) {
      prev.addEventListener("click", function () {
        show(page - 1);
      });
    }
    if (next) {
      next.addEventListener("click", function () {
        show(page + 1);
      });
    }
    document.addEventListener("keydown", function (e) {
      if (e.defaultPrevented) return;
      var t = e.target;
      if (t && (t.tagName === "INPUT" || t.tagName === "TEXTAREA" || t.tagName === "SELECT" || t.isContentEditable)) {
        return;
      }
      if (e.key === "ArrowLeft" || e.key === "PageUp") {
        e.preventDefault();
        show(page - 1);
      } else if (e.key === "ArrowRight" || e.key === "PageDown") {
        e.preventDefault();
        show(page + 1);
      } else if (e.key === "Home") {
        e.preventDefault();
        show(1);
      } else if (e.key === "End") {
        e.preventDefault();
        show(pages);
      }
    });

    show(page);
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", init);
  } else {
    init();
  }
})();
