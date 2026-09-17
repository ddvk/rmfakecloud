(function (global) {
  "use strict";
  function ensureHost() {
    var el = document.getElementById("rm-toast-host");
    if (el) return el;
    el = document.createElement("div");
    el.id = "rm-toast-host";
    el.setAttribute("aria-live", "polite");
    el.style.cssText =
      "position:fixed;z-index:99999;right:1rem;bottom:1rem;display:flex;flex-direction:column;gap:0.5rem;max-width:20rem;";
    document.body.appendChild(el);
    return el;
  }
  function show(type, msg) {
    var host = ensureHost();
    var item = document.createElement("div");
    item.className = "rm-toast rm-toast-" + type;
    item.textContent = String(msg);
    item.style.cssText =
      "padding:0.65rem 0.85rem;border-radius:0.35rem;background:#211e1c;color:#f8f7f6;border:1px solid rgba(255,255,255,0.15);box-shadow:0 8px 24px rgba(0,0,0,0.35);font:inherit;";
    if (type === "error") item.style.borderColor = "#ff1744";
    if (type === "success") item.style.borderColor = "#00e676";
    host.appendChild(item);
    setTimeout(function () {
      if (item.parentNode) item.parentNode.removeChild(item);
    }, 4000);
  }
  global.toast = {
    success: function (m) {
      show("success", m);
    },
    error: function (m) {
      show("error", m);
    },
    info: function (m) {
      show("info", m);
    },
  };
})(typeof window !== "undefined" ? window : globalThis);
