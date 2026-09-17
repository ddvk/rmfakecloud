/**
 * reMarkable-style documents browser: folder cluster, framed thumbs, dock.
 */
(function () {
  "use strict";

  var INTERVAL_MS = 15000;
  var lastSnap = "";
  var lastTree = null;
  var PDFJS_CDN = "https://cdnjs.cloudflare.com/ajax/libs/pdf.js/3.11.174/pdf.min.js";
  var WORKER_CDN = "https://cdnjs.cloudflare.com/ajax/libs/pdf.js/3.11.174/pdf.worker.min.js";
  var pdfjsReady = null;
  var thumbObserver = null;
  var selectMode = false;
  var selected = Object.create(null);
  var docsChannel = null;
  try {
    if (typeof BroadcastChannel !== "undefined") {
      docsChannel = new BroadcastChannel("rmfakecloud-docs");
    }
  } catch (_) {}

  function isTemplatesAdmin() {
    return document.body.classList.contains("page-templates");
  }

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
      acc.push((e.isFolder ? "F:" : "D:") + e.id + ":" + (e.name || "") + ":" + (e.size || 0) + ":" + (e.formatLabel || e.FormatLabel || ""));
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
    var n = String(e.name || "").toLowerCase();
    if (n.endsWith(".pdf")) return "pdf";
    if (n.endsWith(".epub")) return "epub";
    return "notebook";
  }

  function notebookThumbURL(id, page) {
    var n = page || 1;
    if (n < 1) n = 1;
    return "/documents/" + encodeURIComponent(id) + "/page/" + n + "/thumb.png";
  }

  function epubThumbURL(id) {
    return "/documents/" + encodeURIComponent(id) + "/epub-thumb.png";
  }

  function thumbPage(e) {
    var pages = e.pageCount || e.PageCount || 0;
    var page = e.currentPage || e.CurrentPage || 0;
    var n = page + 1;
    if (n < 1) n = 1;
    if (pages > 0 && n > pages) n = pages;
    return n;
  }

  function formatBadge(e) {
    var lab = String(e.formatLabel || e.FormatLabel || "").trim();
    if (lab) return lab;
    var t = docType(e);
    if (t === "pdf") return "PDF";
    if (t === "epub") return "EPUB";
    return "RM";
  }

  function pageLabel(e) {
    var pages = e.pageCount || e.PageCount || 0;
    var n = thumbPage(e);
    if (pages > 0) return "Page " + n + " of " + pages;
    return formatBadge(e);
  }

  function folderHasChildren(e) {
    var kids = e && (e.children || e.Entries || e.entries);
    return !!(kids && kids.length);
  }

  function folderIconSvg(filled) {
    if (filled) {
      return (
        '<svg class="rm-folder-icon is-full" viewBox="0 0 24 20" width="18" height="15" aria-hidden="true" focusable="false">' +
        '<path fill="currentColor" d="M2 4h8l2 2h10v12H2z"/>' +
        "</svg>"
      );
    }
    return (
      '<svg class="rm-folder-icon is-empty" viewBox="0 0 24 20" width="18" height="15" aria-hidden="true" focusable="false">' +
      '<path fill="currentColor" fill-rule="evenodd" d="M2 4h8l2 2h10v12H2V4zm1.5 1.5v11h17v-9H11.2l-2-2H3.5z"/>' +
      "</svg>"
    );
  }

  function renderFolders(ul, folders) {
    ul.innerHTML = "";
    folders.forEach(function (e) {
      var filled = folderHasChildren(e);
      var li = document.createElement("li");
      li.className = "rm-folder-item";
      li.setAttribute("data-id", e.id || "");
      li.setAttribute("data-name", e.name || "");
      li.setAttribute("data-modified", e.lastModified || "");
      li.setAttribute("data-empty", filled ? "false" : "true");
      li.setAttribute("data-pinned", e.pinned || e.Pinned ? "true" : "false");
      var check = document.createElement("label");
      check.className = "rm-item-check";
      var box = document.createElement("input");
      box.type = "checkbox";
      box.className = "rm-select-box";
      box.value = e.id || "";
      box.setAttribute("data-kind", "folder");
      box.setAttribute("data-name", e.name || "");
      box.setAttribute("data-pinned", e.pinned || e.Pinned ? "true" : "false");
      box.setAttribute("aria-label", "Select " + (e.name || e.id || "folder"));
      if (selected[e.id]) box.checked = true;
      check.appendChild(box);
      var a = document.createElement("a");
      a.href = "/documents?folder=" + encodeURIComponent(e.id);
      a.innerHTML = folderIconSvg(filled) + '<span class="rm-folder-name"></span>';
      a.querySelector(".rm-folder-name").textContent = e.name || e.id;
      if (e.pinned || e.Pinned) {
        var star = document.createElement("span");
        star.className = "rm-star";
        star.setAttribute("aria-label", "Favorite");
        star.textContent = "★";
        a.appendChild(star);
      }
      li.appendChild(check);
      li.appendChild(a);
      if (selected[e.id]) li.classList.add("is-selected");
      ul.appendChild(li);
    });
  }

  function renderFiles(ul, files) {
    ul.innerHTML = "";
    files.forEach(function (e) {
      var t = docType(e);
      var li = document.createElement("li");
      li.className = "rm-file-item";
      li.setAttribute("data-id", e.id || "");
      li.setAttribute("data-name", e.name || "");
      li.setAttribute("data-modified", e.lastModified || "");
      li.setAttribute("data-type", t);
      var check = document.createElement("label");
      check.className = "rm-item-check";
      var box = document.createElement("input");
      box.type = "checkbox";
      box.className = "rm-select-box";
      box.value = e.id || "";
      box.setAttribute("data-kind", "file");
      box.setAttribute("data-name", e.name || "");
      box.setAttribute("data-pinned", e.pinned || e.Pinned ? "true" : "false");
      box.setAttribute("aria-label", "Select " + (e.name || e.id || "file"));
      if (selected[e.id]) box.checked = true;
      check.appendChild(box);
      var a = document.createElement("a");
      a.href = "/documents/" + encodeURIComponent(e.id);
      a.className = "rm-file-link";
      var frame = document.createElement("span");
      frame.className = "rm-page-frame is-" + t;
      frame.setAttribute("data-label", formatBadge(e));
      li.setAttribute("data-pinned", e.pinned || e.Pinned ? "true" : "false");
      var page = thumbPage(e);
      var writings = !!(e.hasWritings || e.HasWritings);
      if ((t === "pdf" || t === "epub") && writings) {
        frame.classList.add("has-preview");
        var ann = document.createElement("img");
        ann.className = "rm-thumb-img";
        ann.width = 180;
        ann.height = 240;
        ann.alt = "";
        ann.decoding = "async";
        ann.src = "/documents/" + encodeURIComponent(e.id) + "/page/" + page + "/thumb.png";
        ann.addEventListener("load", function () {
          frame.classList.add("has-preview");
        });
        frame.appendChild(ann);
      } else if (t === "pdf") {
        var canvas = document.createElement("canvas");
        canvas.className = "rm-thumb-canvas";
        canvas.width = 180;
        canvas.height = 240;
        canvas.setAttribute("data-pdf-url", "/ui/api/documents/" + encodeURIComponent(e.id) + "?type=pdf");
        canvas.setAttribute("data-pdf-page", String(page));
        canvas.setAttribute("aria-hidden", "true");
        frame.appendChild(canvas);
      } else if (t === "notebook") {
        frame.classList.add("has-preview");
        var img = document.createElement("img");
        img.className = "rm-thumb-img";
        img.width = 180;
        img.height = 240;
        img.alt = "";
        img.decoding = "async";
        img.src = notebookThumbURL(e.id, page);
        img.addEventListener("load", function () {
          frame.classList.add("has-preview");
        });
        frame.appendChild(img);
      } else if (t === "epub") {
        frame.classList.add("has-preview");
        var epubImg = document.createElement("img");
        epubImg.className = "rm-thumb-img";
        epubImg.width = 180;
        epubImg.height = 240;
        epubImg.alt = "";
        epubImg.decoding = "async";
        epubImg.src = epubThumbURL(e.id);
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
      li.appendChild(check);
      li.appendChild(a);
      if (selected[e.id]) li.classList.add("is-selected");
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
      img.addEventListener("error", function () {
        var url = img.getAttribute("src");
        if (!url || img.getAttribute("data-thumb-retried") === "1") return;
        img.setAttribute("data-thumb-retried", "1");
        fetch(url, { credentials: "same-origin" })
          .then(function (r) {
            if (!r.ok) throw new Error(String(r.status));
            return r.blob();
          })
          .then(function (blob) {
            var obj = URL.createObjectURL(blob);
            img.onload = function () {
              mark();
              URL.revokeObjectURL(obj);
            };
            img.src = obj;
          })
          .catch(function () {});
      });
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

  function selectedList() {
    return Object.keys(selected).map(function (id) {
      return selected[id];
    });
  }

  function setSelectMode(on) {
    selectMode = !!on;
    var panel = document.querySelector(".rm-files");
    var bar = document.getElementById("rm-select-bar");
    var btn = document.getElementById("rm-select-toggle");
    if (panel) panel.classList.toggle("is-select-mode", selectMode);
    if (bar) {
      if (selectMode) bar.removeAttribute("hidden");
      else bar.setAttribute("hidden", "hidden");
    }
    if (btn) {
      btn.setAttribute("aria-pressed", selectMode ? "true" : "false");
      btn.setAttribute("aria-label", selectMode ? "Done selecting" : "Select items");
      btn.classList.toggle("is-active", selectMode);
    }
    if (!selectMode) {
      selected = Object.create(null);
      document.querySelectorAll(".rm-folder-item, .rm-file-item").forEach(function (el) {
        el.classList.remove("is-selected");
        var box = el.querySelector(".rm-select-box");
        if (box) box.checked = false;
      });
    }
    updateSelectBar();
  }

  function updateSelectBar() {
    var list = selectedList();
    var count = document.getElementById("rm-select-count");
    var renameBtn = document.getElementById("rm-rename-toggle");
    var moveBtn = document.getElementById("rm-move-toggle");
    var favBtn = document.getElementById("rm-favorite-toggle");
    var delBtn = document.getElementById("rm-delete-selected");
    if (count) count.textContent = list.length + " selected";
    if (renameBtn) renameBtn.disabled = list.length !== 1;
    if (moveBtn) moveBtn.disabled = list.length < 1;
    if (delBtn) delBtn.disabled = list.length < 1;
    if (favBtn) {
      favBtn.disabled = list.length < 1 || isTemplatesAdmin();
      var allPinned = list.length > 0 && list.every(function (it) {
        return !!it.pinned;
      });
      favBtn.textContent = allPinned ? "☆ Unfavorite" : "★ Favorite";
      favBtn.setAttribute("aria-label", allPinned ? "Remove from favorites" : "Add to favorites");
    }
  }

  function selectableItems() {
    return document.querySelectorAll(
      ".rm-folder-item:not(.is-hidden), .rm-file-item:not(.is-hidden):not([data-builtin='true'])"
    );
  }

  function itemPinned(id) {
    var li = document.querySelector(
      '.rm-folder-item[data-id="' +
        String(id).replace(/\\/g, "\\\\").replace(/"/g, '\\"') +
        '"], .rm-file-item[data-id="' +
        String(id).replace(/\\/g, "\\\\").replace(/"/g, '\\"') +
        '"]'
    );
    if (!li) return false;
    return li.getAttribute("data-pinned") === "true";
  }

  function toggleItem(id, kind, name, force, pinned) {
    if (!id) return;
    var on = force === undefined ? !selected[id] : !!force;
    if (on) {
      selected[id] = {
        id: id,
        kind: kind || "file",
        name: name || "",
        pinned: pinned === undefined ? itemPinned(id) : !!pinned,
      };
    } else delete selected[id];
    var li = document.querySelector(
      '.rm-folder-item[data-id="' +
        String(id).replace(/\\/g, "\\\\").replace(/"/g, '\\"') +
        '"], .rm-file-item[data-id="' +
        String(id).replace(/\\/g, "\\\\").replace(/"/g, '\\"') +
        '"]'
    );
    if (li) {
      li.classList.toggle("is-selected", on);
      var box = li.querySelector(".rm-select-box");
      if (box) box.checked = on;
    }
    updateSelectBar();
  }

  function findEntry(entries, id, parentId) {
    entries = entries || [];
    for (var i = 0; i < entries.length; i++) {
      var e = entries[i];
      if (e.id === id) return { entry: e, parentId: parentId || "" };
      if (e.isFolder && e.children) {
        var found = findEntry(e.children, id, e.id);
        if (found) return found;
      }
    }
    return null;
  }

  function collectFolders(entries, path, out, skipIds) {
    (entries || []).forEach(function (e) {
      if (!e.isFolder) return;
      if (skipIds && skipIds[e.id]) return;
      var label = path ? path + " / " + (e.name || e.id) : e.name || e.id;
      out.push({ id: e.id, label: label });
      if (e.children) collectFolders(e.children, label, out, skipIds);
    });
  }

  function descendantIds(entry, acc) {
    if (!entry || !entry.isFolder) return;
    (entry.children || []).forEach(function (c) {
      acc[c.id] = true;
      if (c.isFolder) descendantIds(c, acc);
    });
  }

  function refreshNow() {
    lastSnap = "";
    return poll();
  }

  /** Refresh this tab and ask other open My Files tabs for the same account to refresh. */
  function refreshAccountWide() {
    var p = refreshNow();
    try {
      if (docsChannel) {
        docsChannel.postMessage({ type: "docs-refresh", t: Date.now() });
      }
      localStorage.setItem("rmfakecloud-docs-refresh", String(Date.now()));
    } catch (_) {}
    return p;
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

    var addBtn = document.getElementById("rm-add-toggle");
    var addMenu = document.getElementById("rm-add-menu");
    function setAddMenuOpen(open) {
      if (!addMenu || !addBtn) return;
      if (open) addMenu.removeAttribute("hidden");
      else addMenu.setAttribute("hidden", "hidden");
      addBtn.setAttribute("aria-expanded", open ? "true" : "false");
      addBtn.classList.toggle("is-active", !!open);
    }
    function closeAddMenu() {
      setAddMenuOpen(false);
    }
    if (addBtn && addMenu) {
      addBtn.addEventListener("click", function (ev) {
        ev.stopPropagation();
        setAddMenuOpen(addMenu.hasAttribute("hidden"));
      });
      document.addEventListener("click", function (ev) {
        if (!addMenu.hasAttribute("hidden") && !addMenu.contains(ev.target) && ev.target !== addBtn && !addBtn.contains(ev.target)) {
          closeAddMenu();
        }
      });
      document.addEventListener("keydown", function (ev) {
        if (ev.key === "Escape") closeAddMenu();
      });
    }

    var uploadBtn = document.getElementById("rm-upload-toggle");
    var fileInput = document.getElementById("doc-upload");
    var uploadForm = document.getElementById("rm-upload-form");
    if (uploadBtn && fileInput) {
      uploadBtn.addEventListener("click", function () {
        closeAddMenu();
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
        closeAddMenu();
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

    var selectBtn = document.getElementById("rm-select-toggle");
    if (selectBtn) {
      selectBtn.addEventListener("click", function () {
        setSelectMode(!selectMode);
      });
    }

    var panel = document.querySelector(".rm-files");
    if (panel) {
      panel.addEventListener("click", function (ev) {
        if (!selectMode) return;
        var box = ev.target.closest && ev.target.closest(".rm-select-box");
        if (box) {
          ev.stopPropagation();
          toggleItem(
            box.value,
            box.getAttribute("data-kind"),
            box.getAttribute("data-name"),
            box.checked,
            box.getAttribute("data-pinned") === "true"
          );
          return;
        }
        var link = ev.target.closest && ev.target.closest(".rm-folder-item a, .rm-file-item a");
        if (link) {
          ev.preventDefault();
          var li = link.closest(".rm-folder-item, .rm-file-item");
          if (!li) return;
          var id = li.getAttribute("data-id");
          var kind = li.classList.contains("rm-folder-item") ? "folder" : "file";
          toggleItem(id, kind, li.getAttribute("data-name"), undefined, li.getAttribute("data-pinned") === "true");
        }
      });
      panel.addEventListener("change", function (ev) {
        if (!selectMode) return;
        var box = ev.target;
        if (!box || !box.classList || !box.classList.contains("rm-select-box")) return;
        toggleItem(
          box.value,
          box.getAttribute("data-kind"),
          box.getAttribute("data-name"),
          box.checked,
          box.getAttribute("data-pinned") === "true"
        );
      });
    }

    var selectAll = document.getElementById("rm-select-all");
    if (selectAll) {
      selectAll.addEventListener("click", function () {
        selectableItems().forEach(function (li) {
          var id = li.getAttribute("data-id");
          if (!id) return;
          var kind = li.classList.contains("rm-folder-item") ? "folder" : "file";
          toggleItem(id, kind, li.getAttribute("data-name"), true, li.getAttribute("data-pinned") === "true");
        });
      });
    }

    var favBtn = document.getElementById("rm-favorite-toggle");
    if (favBtn) {
      favBtn.addEventListener("click", function () {
        var list = selectedList();
        if (!list.length || isTemplatesAdmin()) return;
        var allPinned = list.every(function (it) {
          return !!it.pinned;
        });
        var wantPinned = !allPinned;
        var redirect = activeFolderId();
        Promise.all(
          list.map(function (it) {
            var body = new URLSearchParams();
            body.set("pinned", wantPinned ? "true" : "false");
            body.set("redirect", redirect);
            return fetch("/documents/" + encodeURIComponent(it.id) + "/pin", {
              method: "POST",
              credentials: "same-origin",
              headers: { "Content-Type": "application/x-www-form-urlencoded" },
              body: body.toString(),
            }).then(function (r) {
              if (!r.ok) throw new Error("Favorite update failed");
              return r;
            });
          })
        )
          .then(function () {
            selected = Object.create(null);
            updateSelectBar();
            return refreshAccountWide();
          })
          .catch(function (err) {
            window.alert(err && err.message ? err.message : "Favorite update failed");
          });
      });
    }

    var renameBtn = document.getElementById("rm-rename-toggle");
    var renameDialog = document.getElementById("rm-rename-dialog");
    var renameForm = document.getElementById("rm-rename-form");
    var renameCancel = document.getElementById("rm-rename-cancel");
    if (renameBtn && renameDialog && renameDialog.showModal) {
      renameBtn.addEventListener("click", function () {
        var list = selectedList();
        if (list.length !== 1) return;
        var item = list[0];
        var hit = lastTree ? findEntry(lastTree.Entries || lastTree.entries || [], item.id) : null;
        var parentId = hit ? hit.parentId : activeFolderId();
        var nameInput = document.getElementById("rm-rename-name");
        var parentInput = document.getElementById("rm-rename-parent");
        var redirectInput = document.getElementById("rm-rename-redirect");
        if (nameInput) nameInput.value = item.name || (hit && hit.entry && hit.entry.name) || "";
        if (parentInput) parentInput.value = parentId;
        if (redirectInput) redirectInput.value = activeFolderId();
        if (renameForm) {
          if (isTemplatesAdmin()) {
            renameForm.action = "/admin/templates/" + encodeURIComponent(item.id) + "/update";
          } else {
            renameForm.action = "/documents/" + encodeURIComponent(item.id) + "/update";
          }
        }
        renameDialog.showModal();
        if (nameInput) {
          nameInput.focus();
          nameInput.select();
        }
      });
    }
    if (renameCancel && renameDialog) {
      renameCancel.addEventListener("click", function () {
        renameDialog.close();
      });
    }
    if (renameForm) {
      renameForm.addEventListener("submit", function (ev) {
        if (isTemplatesAdmin()) {
          // Native POST to /admin/templates/:id/update
          return;
        }
        if (!window.apiService || typeof window.apiService.updateDocument !== "function") return;
        ev.preventDefault();
        var list = selectedList();
        if (list.length !== 1) return;
        var item = list[0];
        var nameInput = document.getElementById("rm-rename-name");
        var parentInput = document.getElementById("rm-rename-parent");
        var name = nameInput ? nameInput.value.trim() : "";
        if (!name) return;
        window.apiService
          .updateDocument({
            documentId: item.id,
            name: name,
            parentId: parentInput ? parentInput.value : "",
          })
          .then(function () {
            if (renameDialog) renameDialog.close();
            selected[item.id] = { id: item.id, kind: item.kind, name: name };
            return refreshAccountWide();
          })
          .catch(function (err) {
            window.alert(err && err.message ? err.message : "Rename failed");
          });
      });
    }

    var moveBtn = document.getElementById("rm-move-toggle");
    var moveDialog = document.getElementById("rm-move-dialog");
    var moveForm = document.getElementById("rm-move-form");
    var moveCancel = document.getElementById("rm-move-cancel");
    var moveTarget = document.getElementById("rm-move-target");
    if (moveBtn && moveDialog && moveDialog.showModal) {
      moveBtn.addEventListener("click", function () {
        var list = selectedList();
        if (!list.length || !moveTarget) return;
        var skip = Object.create(null);
        list.forEach(function (it) {
          if (it.kind === "folder") {
            skip[it.id] = true;
            var hit = lastTree ? findEntry(lastTree.Entries || lastTree.entries || [], it.id) : null;
            if (hit && hit.entry) descendantIds(hit.entry, skip);
          }
        });
        var folders = [];
        if (lastTree) collectFolders(lastTree.Entries || lastTree.entries || [], "", folders, skip);
        moveTarget.innerHTML = "";
        var rootOpt = document.createElement("option");
        rootOpt.value = "";
        rootOpt.textContent = "My Files";
        moveTarget.appendChild(rootOpt);
        folders.forEach(function (f) {
          var opt = document.createElement("option");
          opt.value = f.id;
          opt.textContent = f.label;
          moveTarget.appendChild(opt);
        });
        moveDialog.showModal();
      });
    }
    if (moveCancel && moveDialog) {
      moveCancel.addEventListener("click", function () {
        moveDialog.close();
      });
    }
    if (moveForm) {
      moveForm.addEventListener("submit", function (ev) {
        ev.preventDefault();
        if (!window.apiService || typeof window.apiService.updateDocument !== "function") return;
        var list = selectedList();
        var parentId = moveTarget ? moveTarget.value : "";
        var cur = activeFolderId();
        if (parentId === cur) {
          if (moveDialog) moveDialog.close();
          return;
        }
        Promise.all(
          list.map(function (it) {
            var hit = lastTree ? findEntry(lastTree.Entries || lastTree.entries || [], it.id) : null;
            var name = (hit && hit.entry && hit.entry.name) || it.name || "";
            return window.apiService.updateDocument({
              documentId: it.id,
              name: name,
              parentId: parentId,
            });
          })
        )
          .then(function () {
            if (moveDialog) moveDialog.close();
            selected = Object.create(null);
            updateSelectBar();
            return refreshAccountWide();
          })
          .catch(function (err) {
            window.alert(err && err.message ? err.message : "Move failed");
          });
      });
    }

    var delBtn = document.getElementById("rm-delete-selected");
    if (delBtn) {
      delBtn.addEventListener("click", function () {
        var list = selectedList();
        if (!list.length) return;
        var label = list.length === 1 ? '"' + (list[0].name || list[0].id) + '"' : list.length + " items";
        if (!window.confirm("Delete " + label + "? This cannot be undone from the web UI.")) return;
        if (isTemplatesAdmin()) {
          Promise.all(
            list.map(function (it) {
              return fetch("/admin/templates/" + encodeURIComponent(it.id) + "/delete", {
                method: "POST",
                credentials: "same-origin",
              }).then(function (r) {
                if (!r.ok && r.status >= 400) throw new Error("Delete failed");
                return r;
              });
            })
          )
            .then(function () {
              window.location.reload();
            })
            .catch(function (err) {
              window.alert(err && err.message ? err.message : "Delete failed");
            });
          return;
        }
        if (!window.apiService || typeof window.apiService.deleteDocument !== "function") return;
        Promise.all(
          list.map(function (it) {
            return window.apiService.deleteDocument(it.id);
          })
        )
          .then(function () {
            selected = Object.create(null);
            updateSelectBar();
            return refreshAccountWide();
          })
          .catch(function (err) {
            window.alert(err && err.message ? err.message : "Delete failed");
          });
      });
    }
  }

  function shouldSkip() {
    var ae = document.activeElement;
    if (!ae) return false;
    if (ae.closest && ae.closest("#rm-upload-form, #rm-folder-dialog, #rm-rename-dialog, #rm-move-dialog, .rm-search-bar, .rm-select-bar")) return true;
    if (ae.tagName === "INPUT" && ae.type === "file") return true;
    return false;
  }

  async function poll() {
    if (isTemplatesAdmin()) return;
    if (document.hidden || shouldSkip()) return;
    if (!window.apiService || typeof window.apiService.listDocument !== "function") return;
    try {
      var tree = await window.apiService.listDocument();
      lastTree = tree;
      var snap = snapshot(tree);
      var folderUl = document.querySelector(".rm-folder-grid");
      var fileUl = document.querySelector(".rm-file-grid");
      var hasItems = !!(folderUl && folderUl.children.length) || !!(fileUl && fileUl.children.length);
      if (lastSnap === "") {
        lastSnap = snap;
        if (hasItems) {
          observeThumbs();
          updateSelectBar();
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
      updateSelectBar();
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
    updateSelectBar();
    if (docsChannel) {
      docsChannel.onmessage = function (ev) {
        if (ev && ev.data && ev.data.type === "docs-refresh") refreshNow();
      };
    }
    window.addEventListener("storage", function (ev) {
      if (ev.key === "rmfakecloud-docs-refresh") refreshNow();
    });
    // Form actions redirect back with a flash — nudge other tabs for this account.
    if (document.querySelector(".flash-success, .flash-info, .flash-error")) {
      try {
        if (docsChannel) docsChannel.postMessage({ type: "docs-refresh", t: Date.now() });
        localStorage.setItem("rmfakecloud-docs-refresh", String(Date.now()));
      } catch (_) {}
    }
    if (isTemplatesAdmin()) {
      return;
    }
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
