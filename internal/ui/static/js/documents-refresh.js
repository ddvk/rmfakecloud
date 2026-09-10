/**
 * reMarkable-style documents browser: folder cluster, framed thumbs, dock.
 */
(function () {
  "use strict";

  var INTERVAL_MS = 15000;
  var lastSnap = "";
  var PDFJS_CDN = "https://cdnjs.cloudflare.com/ajax/libs/pdf.js/3.11.174/pdf.min.js";
  var WORKER_CDN = "https://cdnjs.cloudflare.com/ajax/libs/pdf.js/3.11.174/pdf.worker.min.js";
  var pdfjsReady = null;
  var thumbObserver = null;

  function activeFolderId() {
    var params = new URLSearchParams(window.location.search);
    return params.get("folder") || "";
  }

  function findFolder(entries, id) {
    if (!id) return entries || [];
    entries = entries || [];
    for (var i = 0; i < entries.length; i++) {
      var e = entries[i];
      if (e.isFolder && e.id === id) return e.children || [];
      if (e.isFolder && e.children) {
        var found = findFolder(e.children, id);
        if (found) return found;
      }
    }
    return null;
  }

  function walkSnap(entries, acc) {
    (entries || []).forEach(function (e) {
      acc.push((e.isFolder ? "F:" : "D:") + e.id + ":" + (e.name || "") + ":" + (e.size || 0));
      if (e.isFolder && e.children) walkSnap(e.children, acc);
    });
  }

  function snapshot(tree) {
    var acc = [];
    var entries = (tree && (tree.Entries || tree.entries)) || [];
    var trash = (tree && (tree.Trash || tree.trash)) || [];
    walkSnap(entries, acc);
    walkSnap(trash, acc);
    return acc.join("|");
  }

  function docType(e) {
    var t = String(e.type || e.DocumentType || "").toLowerCase().replace(/^\./, "");
    if (t === "pdf" || t === "application/pdf") return "pdf";
    if (t === "epub" || t === "application/epub+zip") return "epub";
    return "notebook";
  }

  function thumbPage(e) {
    var pages = e.pageCount || e.PageCount || 0;
    var page = e.currentPage || e.CurrentPage || 0;
    var n = page + 1;
    if (n < 1) n = 1;
    if (pages > 0 && n > pages) n = pages;
    return n;
  }

  function pageLabel(e) {
    var pages = e.pageCount || e.PageCount || 0;
    var n = thumbPage(e);
    if (pages > 0) return "Page " + n + " of " + pages;
    var t = docType(e);
    if (t === "pdf") return "PDF";
    if (t === "epub") return "EPUB";
    return "Notebook";
  }

  function renderFolders(ul, folders) {
    ul.innerHTML = "";
    folders.forEach(function (e) {
      var li = document.createElement("li");
      li.className = "rm-folder-item";
      li.setAttribute("data-name", e.name || "");
      li.setAttribute("data-modified", e.lastModified || "");
      var a = document.createElement("a");
      a.href = "/documents?folder=" + encodeURIComponent(e.id);
      a.innerHTML =
        '<span class="rm-folder-icon" aria-hidden="true"></span>' +
        '<span class="rm-folder-name"></span>';
      a.querySelector(".rm-folder-name").textContent = e.name || e.id;
      if (e.pinned) {
        var star = document.createElement("span");
        star.className = "rm-star";
        star.setAttribute("aria-label", "Favorite");
        star.textContent = "★";
        a.appendChild(star);
      }
      li.appendChild(a);
      ul.appendChild(li);
    });
  }

  function renderFiles(ul, files) {
    ul.innerHTML = "";
    files.forEach(function (e) {
      var t = docType(e);
      var li = document.createElement("li");
      li.className = "rm-file-item";
      li.setAttribute("data-name", e.name || "");
      li.setAttribute("data-modified", e.lastModified || "");
      li.setAttribute("data-type", t);
      var a = document.createElement("a");
      a.href = "/documents/" + encodeURIComponent(e.id);
      a.className = "rm-file-link";
      var frame = document.createElement("span");
      frame.className = "rm-page-frame is-" + t;
      var page = thumbPage(e);
      if (t === "pdf") {
        var canvas = document.createElement("canvas");
        canvas.className = "rm-thumb-canvas";
        canvas.width = 180;
        canvas.height = 240;
        canvas.setAttribute("data-pdf-url", "/ui/api/documents/" + encodeURIComponent(e.id) + "?type=pdf");
        canvas.setAttribute("data-pdf-page", String(page));
        canvas.setAttribute("aria-hidden", "true");
        frame.appendChild(canvas);
      } else if (t === "notebook") {
        var img = document.createElement("img");
        img.className = "rm-thumb-img";
        img.width = 180;
        img.height = 240;
        img.alt = "";
        img.loading = "lazy";
        img.decoding = "async";
        img.src = "/ui/api/documents/" + encodeURIComponent(e.id) + "/page/" + page;
        img.addEventListener("load", function () {
          frame.classList.add("has-preview");
        });
        frame.appendChild(img);
      } else if (t === "epub") {
        var epubImg = document.createElement("img");
        epubImg.className = "rm-thumb-img";
        epubImg.width = 180;
        epubImg.height = 240;
        epubImg.alt = "";
        epubImg.loading = "lazy";
        epubImg.decoding = "async";
        epubImg.src = "/ui/api/documents/" + encodeURIComponent(e.id) + "/epub/thumb";
        epubImg.addEventListener("load", function () {
          frame.classList.add("has-preview");
        });
        frame.appendChild(epubImg);
      } else {
        var ph = document.createElement("span");
        ph.className = "rm-page-placeholder";
        ph.setAttribute("aria-hidden", "true");
        frame.appendChild(ph);
      }
      if (t !== "epub") {
        var ear = document.createElement("span");
        ear.className = "rm-page-ear";
        ear.setAttribute("aria-hidden", "true");
        frame.appendChild(ear);
      }
      var meta = document.createElement("span");
      meta.className = "rm-file-meta";
      var name = document.createElement("span");
      name.className = "rm-file-name";
      name.textContent = e.name || e.id;
      if (e.pinned) {
        var star = document.createElement("span");
        star.className = "rm-star";
        star.setAttribute("aria-label", "Favorite");
        star.textContent = " ★";
        name.appendChild(star);
      }
      var sub = document.createElement("span");
      sub.className = "rm-file-sub";
      sub.textContent = pageLabel(e);
      meta.appendChild(name);
      meta.appendChild(sub);
      a.appendChild(frame);
      a.appendChild(meta);
      li.appendChild(a);
      ul.appendChild(li);
    });
  }

  function splitEntries(entries) {
    var folders = [];
    var files = [];
    (entries || []).forEach(function (e) {
      if (e.isFolder) folders.push(e);
      else files.push(e);
    });
    return { folders: folders, files: files };
  }

  function applySort() {
    var sel = document.getElementById("rm-sort");
    var mode = sel ? sel.value : "modified";
    function cmp(a, b) {
      if (mode === "name") {
        return (a.getAttribute("data-name") || "").localeCompare(b.getAttribute("data-name") || "", undefined, {
          sensitivity: "base",
        });
      }
      return String(b.getAttribute("data-modified") || "").localeCompare(String(a.getAttribute("data-modified") || ""));
    }
    [".rm-folder-grid", ".rm-file-grid"].forEach(function (selGrid) {
      var ul = document.querySelector(selGrid);
      if (!ul) return;
      var items = Array.prototype.slice.call(ul.children);
      items.sort(cmp);
      items.forEach(function (it) {
        ul.appendChild(it);
      });
    });
  }

  function applySearch() {
    var q = (document.getElementById("rm-search") && document.getElementById("rm-search").value) || "";
    q = q.trim().toLowerCase();
    document.querySelectorAll(".rm-folder-item, .rm-file-item").forEach(function (el) {
      var name = (el.getAttribute("data-name") || "").toLowerCase();
      el.classList.toggle("is-hidden", q && name.indexOf(q) < 0);
    });
  }

  function loadPdfJs() {
    if (window.pdfjsLib) return Promise.resolve(window.pdfjsLib);
    if (pdfjsReady) return pdfjsReady;
    pdfjsReady = new Promise(function (resolve, reject) {
      var s = document.createElement("script");
      s.src = PDFJS_CDN;
      s.async = true;
      s.onload = function () {
        if (!window.pdfjsLib) {
          reject(new Error("pdfjsLib missing"));
          return;
        }
        window.pdfjsLib.GlobalWorkerOptions.workerSrc = WORKER_CDN;
        resolve(window.pdfjsLib);
      };
      s.onerror = function () {
        reject(new Error("pdf.js failed to load"));
      };
      document.head.appendChild(s);
    });
    return pdfjsReady;
  }

  async function renderPdfThumb(canvas) {
    var url = canvas.getAttribute("data-pdf-url");
    if (!url || canvas.getAttribute("data-rendered") === "1") return;
    canvas.setAttribute("data-rendered", "1");
    try {
      var pdfjsLib = await loadPdfJs();
      var pdf = await pdfjsLib.getDocument({ url: url, withCredentials: true }).promise;
      var pageNum = parseInt(canvas.getAttribute("data-pdf-page") || "1", 10);
      if (!pageNum || pageNum < 1) pageNum = 1;
      if (pageNum > pdf.numPages) pageNum = pdf.numPages;
      var page = await pdf.getPage(pageNum);
      var viewport = page.getViewport({ scale: 1 });
      var scale = canvas.width / viewport.width;
      var vp = page.getViewport({ scale: scale });
      canvas.height = vp.height;
      var ctx = canvas.getContext("2d");
      await page.render({ canvasContext: ctx, viewport: vp }).promise;
      var frame = canvas.closest(".rm-page-frame");
      if (frame) frame.classList.add("has-preview");
    } catch (err) {
      console.warn("[documents] pdf thumb", err);
    }
  }

  function observeThumbs() {
    document.querySelectorAll(".rm-page-frame .rm-thumb-img").forEach(function (img) {
      function mark() {
        if (img.naturalWidth) {
          var frame = img.closest(".rm-page-frame");
          if (frame) frame.classList.add("has-preview");
        }
      }
      if (img.complete) mark();
      else img.addEventListener("load", mark, { once: true });
    });
    if (thumbObserver) thumbObserver.disconnect();
    var canvases = document.querySelectorAll(".rm-thumb-canvas[data-pdf-url]");
    if (!canvases.length) return;
    thumbObserver = new IntersectionObserver(
      function (entries) {
        entries.forEach(function (en) {
          if (!en.isIntersecting) return;
          thumbObserver.unobserve(en.target);
          renderPdfThumb(en.target);
        });
      },
      { rootMargin: "120px" }
    );
    canvases.forEach(function (c) {
      thumbObserver.observe(c);
    });
  }

  function wireChrome() {
    var searchBtn = document.getElementById("rm-search-toggle");
    var searchBar = document.getElementById("rm-search-bar");
    var searchInput = document.getElementById("rm-search");
    if (searchBtn && searchBar) {
      searchBtn.addEventListener("click", function () {
        var on = searchBar.hasAttribute("hidden");
        if (on) searchBar.removeAttribute("hidden");
        else searchBar.setAttribute("hidden", "hidden");
        searchBtn.setAttribute("aria-expanded", on ? "true" : "false");
        if (on && searchInput) searchInput.focus();
      });
    }
    if (searchInput) searchInput.addEventListener("input", applySearch);

    var uploadBtn = document.getElementById("rm-upload-toggle");
    var fileInput = document.getElementById("doc-upload");
    var uploadForm = document.getElementById("rm-upload-form");
    if (uploadBtn && fileInput) {
      uploadBtn.addEventListener("click", function () {
        fileInput.click();
      });
      fileInput.addEventListener("change", function () {
        if (fileInput.files && fileInput.files.length && uploadForm) uploadForm.submit();
      });
    }

    var folderBtn = document.getElementById("rm-folder-toggle");
    var dialog = document.getElementById("rm-folder-dialog");
    var cancel = document.getElementById("rm-folder-cancel");
    if (folderBtn && dialog && dialog.showModal) {
      folderBtn.addEventListener("click", function () {
        dialog.showModal();
        var name = document.getElementById("folder-name");
        if (name) name.focus();
      });
    }
    if (cancel && dialog) {
      cancel.addEventListener("click", function () {
        dialog.close();
      });
    }

    var sort = document.getElementById("rm-sort");
    if (sort) sort.addEventListener("change", applySort);
  }

  function shouldSkip() {
    var ae = document.activeElement;
    if (!ae) return false;
    if (ae.closest && ae.closest("#rm-upload-form, #rm-folder-dialog, .rm-search-bar")) return true;
    if (ae.tagName === "INPUT" && ae.type === "file") return true;
    return false;
  }

  async function poll() {
    if (document.hidden || shouldSkip()) return;
    if (!window.apiService || typeof window.apiService.listDocument !== "function") return;
    try {
      var tree = await window.apiService.listDocument();
      var snap = snapshot(tree);
      var folderUl = document.querySelector(".rm-folder-grid");
      var fileUl = document.querySelector(".rm-file-grid");
      var hasItems = !!(folderUl && folderUl.children.length) || !!(fileUl && fileUl.children.length);
      if (lastSnap === "") {
        lastSnap = snap;
        if (hasItems) {
          observeThumbs();
          return;
        }
      }
      if (snap === lastSnap) return;
      lastSnap = snap;
      var folderId = activeFolderId();
      var rootEntries = tree.Entries || tree.entries || [];
      var entries = folderId ? findFolder(rootEntries, folderId) : rootEntries;
      if (entries === null) entries = [];
      var parts = splitEntries(entries);
      if (folderUl) renderFolders(folderUl, parts.folders);
      if (fileUl) renderFiles(fileUl, parts.files);
      applySort();
      applySearch();
      observeThumbs();
    } catch (e) {
      console.warn("[documents]", e);
    }
  }

  function init() {
    if (!document.body.classList.contains("page-documents") && !document.querySelector(".rm-files")) {
      return;
    }
    wireChrome();
    applySort();
    observeThumbs();
    poll();
    setInterval(poll, INTERVAL_MS);
    document.addEventListener("visibilitychange", function () {
      if (!document.hidden) poll();
    });
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", init);
  } else {
    init();
  }
})();
