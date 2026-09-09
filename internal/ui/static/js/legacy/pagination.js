(function (global) {
  "use strict";
  var RM = global.RM;
  var STORAGE_KEY = "rm-settings-page-size";

  function usePagination(items, opts) {
    opts = opts || {};
    var defaultSize = opts.pageSize || 10;
    try {
      var stored = parseInt(localStorage.getItem(STORAGE_KEY) || "", 10);
      if (stored === 5 || stored === 10 || stored === 25) defaultSize = stored;
    } catch (_) {}
    var pageState = RM.useState(0);
    var page = pageState[0];
    var setPage = pageState[1];
    var sizeState = RM.useState(defaultSize);
    var pageSize = sizeState[0];
    var setPageSizeRaw = sizeState[1];
    var list = items || [];
    var pageCount = Math.max(1, Math.ceil(list.length / pageSize) || 1);
    if (page > pageCount - 1) {
      // clamp on next render
    }
    var safePage = Math.min(page, pageCount - 1);
    var start = safePage * pageSize;
    var pageItems = list.slice(start, start + pageSize);

    function setPageSize(n) {
      setPageSizeRaw(n);
      setPage(0);
      try {
        localStorage.setItem(STORAGE_KEY, String(n));
      } catch (_) {}
    }

    function Pager() {
      if (list.length <= pageSize && pageSize === defaultSize && list.length <= 10) {
        // still show pager if more than page size
      }
      if (list.length === 0) return null;
      return RM.h(
        "div",
        { className: "rm-pager", role: "navigation", "aria-label": "Pagination" },
        RM.h(
          "button",
          {
            type: "button",
            className: "btn btn-secondary btn-sm",
            disabled: safePage <= 0,
            onClick: function () {
              setPage(Math.max(0, safePage - 1));
            },
          },
          "Previous"
        ),
        RM.h(
          "span",
          { className: "rm-pager-status" },
          "Page " + (safePage + 1) + " of " + pageCount
        ),
        RM.h(
          "button",
          {
            type: "button",
            className: "btn btn-secondary btn-sm",
            disabled: safePage >= pageCount - 1,
            onClick: function () {
              setPage(Math.min(pageCount - 1, safePage + 1));
            },
          },
          "Next"
        ),
        RM.h(
          "label",
          { className: "rm-pager-size" },
          "Per page ",
          RM.h(
            "select",
            {
              value: String(pageSize),
              "aria-label": "Page size",
              onChange: function (e) {
                setPageSize(parseInt(e.target.value, 10));
              },
            },
            RM.h("option", { value: "5" }, "5"),
            RM.h("option", { value: "10" }, "10"),
            RM.h("option", { value: "25" }, "25")
          )
        )
      );
    }

    return {
      page: safePage,
      pageSize: pageSize,
      pageCount: pageCount,
      pageItems: pageItems,
      setPage: setPage,
      setPageSize: setPageSize,
      Pager: Pager,
    };
  }

  /** Enhance a static HTML table: hide rows not on current page. */
  function enhanceTable(tableSelector, opts) {
    opts = opts || {};
    var table = document.querySelector(tableSelector);
    if (!table || !table.tBodies || !table.tBodies[0]) return;
    var tbody = table.tBodies[0];
    var rows = Array.prototype.slice.call(tbody.rows);
    if (!rows.length) return;
    var pageSize = opts.pageSize || 10;
    try {
      var stored = parseInt(localStorage.getItem(STORAGE_KEY) || "", 10);
      if (stored === 5 || stored === 10 || stored === 25) pageSize = stored;
    } catch (_) {}
    var page = 0;
    var host = document.createElement("div");
    host.className = "rm-pager";
    table.parentNode.insertBefore(host, table.nextSibling);

    function render() {
      var pageCount = Math.max(1, Math.ceil(rows.length / pageSize));
      if (page >= pageCount) page = pageCount - 1;
      rows.forEach(function (row, i) {
        var show = i >= page * pageSize && i < (page + 1) * pageSize;
        row.hidden = !show;
      });
      host.innerHTML = "";
      var prev = document.createElement("button");
      prev.type = "button";
      prev.className = "btn btn-secondary btn-sm";
      prev.textContent = "Previous";
      prev.disabled = page <= 0;
      prev.onclick = function () {
        page--;
        render();
      };
      var status = document.createElement("span");
      status.className = "rm-pager-status";
      status.textContent = "Page " + (page + 1) + " of " + pageCount;
      var next = document.createElement("button");
      next.type = "button";
      next.className = "btn btn-secondary btn-sm";
      next.textContent = "Next";
      next.disabled = page >= pageCount - 1;
      next.onclick = function () {
        page++;
        render();
      };
      var label = document.createElement("label");
      label.className = "rm-pager-size";
      label.appendChild(document.createTextNode("Per page "));
      var sel = document.createElement("select");
      sel.setAttribute("aria-label", "Page size");
      [5, 10, 25].forEach(function (n) {
        var o = document.createElement("option");
        o.value = String(n);
        o.textContent = String(n);
        if (n === pageSize) o.selected = true;
        sel.appendChild(o);
      });
      sel.onchange = function () {
        pageSize = parseInt(sel.value, 10);
        page = 0;
        try {
          localStorage.setItem(STORAGE_KEY, String(pageSize));
        } catch (_) {}
        render();
      };
      label.appendChild(sel);
      host.appendChild(prev);
      host.appendChild(status);
      host.appendChild(next);
      host.appendChild(label);
    }
    render();
  }

  global.usePagination = usePagination;
  global.enhanceTablePagination = enhanceTable;
})(typeof window !== "undefined" ? window : globalThis);
