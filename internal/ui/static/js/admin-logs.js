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

  function init() {
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
