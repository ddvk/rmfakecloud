/**
 * Admin server log viewer: refresh from /ui/api/logs.
 */
(function () {
  "use strict";

  var INTERVAL_MS = 8000;

  function render(lines) {
    var pre = document.getElementById("admin-logs-view");
    if (!pre) return;
    if (!lines || !lines.length) {
      pre.textContent = "No log lines captured yet.";
      return;
    }
    var text = lines
      .map(function (l) {
        return l.text || "";
      })
      .join("\n");
    var atBottom = pre.scrollHeight - pre.scrollTop - pre.clientHeight < 48;
    pre.textContent = text;
    if (atBottom) {
      pre.scrollTop = pre.scrollHeight;
    }
  }

  function load() {
    return fetch("/ui/api/logs?limit=200", { credentials: "same-origin" })
      .then(function (r) {
        if (!r.ok) throw new Error(String(r.status));
        return r.json();
      })
      .then(function (data) {
        render((data && data.lines) || []);
      })
      .catch(function () {});
  }

  function wireDialogDismiss(dialog, cancelBtn) {
    if (!dialog) return;
    if (cancelBtn) {
      cancelBtn.addEventListener("click", function () {
        dialog.close();
      });
    }
    dialog.addEventListener("click", function (e) {
      if (e.target === dialog) dialog.close();
    });
  }

  function initNewUserDialog() {
    var dialog = document.getElementById("admin-new-user-dialog");
    var openBtn = document.getElementById("admin-new-user-open");
    var cancelBtn = document.getElementById("admin-new-user-cancel");
    if (!dialog || !openBtn || !dialog.showModal) return;
    openBtn.addEventListener("click", function () {
      dialog.showModal();
      var first = dialog.querySelector("input");
      if (first) first.focus();
    });
    wireDialogDismiss(dialog, cancelBtn);
  }

  function initEditUserDialog() {
    var dialog = document.getElementById("admin-edit-user-dialog");
    var form = document.getElementById("admin-edit-user-form");
    var cancelBtn = document.getElementById("admin-edit-user-cancel");
    var useridInput = document.getElementById("edit-userid");
    var emailInput = document.getElementById("edit-email");
    var nameInput = document.getElementById("edit-name");
    if (!dialog || !form || !dialog.showModal) return;

    document.querySelectorAll(".admin-edit-user-open").forEach(function (btn) {
      btn.addEventListener("click", function () {
        var id = btn.getAttribute("data-userid") || "";
        form.action = "/admin/users/" + encodeURIComponent(id) + "/update";
        if (useridInput) useridInput.value = id;
        if (emailInput) emailInput.value = btn.getAttribute("data-email") || "";
        if (nameInput) nameInput.value = btn.getAttribute("data-name") || "";
        dialog.showModal();
        if (useridInput) useridInput.focus();
      });
    });
    wireDialogDismiss(dialog, cancelBtn);
  }

  function init() {
    initNewUserDialog();
    initEditUserDialog();
    var pre = document.getElementById("admin-logs-view");
    if (!pre) return;
    pre.scrollTop = pre.scrollHeight;
    var btn = document.getElementById("admin-logs-refresh");
    if (btn) {
      btn.addEventListener("click", function () {
        load();
      });
    }
    setInterval(load, INTERVAL_MS);
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", init);
  } else {
    init();
  }
})();
