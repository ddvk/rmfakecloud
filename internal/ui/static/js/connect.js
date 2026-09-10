/**
 * Connect pairing code + live draining gauge (5 minute TTL).
 * Polls whether the displayed code is still active; after a successful
 * tablet pair the code is consumed and a new one is issued automatically.
 */
(function () {
  "use strict";

  var CODE_TTL_MS = 5 * 60 * 1000;
  var POLL_MS = 2000;

  function $(id) {
    return document.getElementById(id);
  }

  function formatLabel(secs) {
    var s = Math.max(0, Math.ceil(secs));
    var m = Math.floor(s / 60);
    var r = s % 60;
    return m + ":" + String(r).padStart(2, "0");
  }

  function buildGauge(el, type) {
    el.innerHTML = "";
    el.classList.add("connect-gauge", "connect-gauge-" + type);
    el.setAttribute("role", "progressbar");
    el.setAttribute("aria-valuemin", "0");
    el.setAttribute("aria-valuemax", "100");
    el.setAttribute("data-ttl", "300");

    if (type === "bar") {
      var track = document.createElement("div");
      track.className = "connect-gauge-bar-track";
      var fill = document.createElement("div");
      fill.className = "connect-gauge-bar-fill";
      fill.style.width = "100%";
      track.appendChild(fill);
      el.appendChild(track);
      var label = document.createElement("div");
      label.className = "connect-gauge-label";
      label.id = "connect-gauge-label";
      el.appendChild(label);
      return;
    }

    var size = 160;
    var stroke = 10;
    var r = (size - stroke) / 2;
    var c = 2 * Math.PI * r;
    var ns = "http://www.w3.org/2000/svg";
    var svg = document.createElementNS(ns, "svg");
    svg.setAttribute("width", String(size));
    svg.setAttribute("height", String(size));
    svg.setAttribute("viewBox", "0 0 " + size + " " + size);
    svg.setAttribute("aria-hidden", "true");
    svg.classList.add("connect-gauge-inner");

    var g = document.createElementNS(ns, "g");
    g.setAttribute("transform", "rotate(-90 " + size / 2 + " " + size / 2 + ")");

    var trackCircle = document.createElementNS(ns, "circle");
    trackCircle.setAttribute("cx", String(size / 2));
    trackCircle.setAttribute("cy", String(size / 2));
    trackCircle.setAttribute("r", String(r));
    trackCircle.setAttribute("fill", "none");
    trackCircle.setAttribute("stroke-width", String(stroke));
    trackCircle.setAttribute("class", "connect-gauge-track");
    g.appendChild(trackCircle);

    var fillCircle = document.createElementNS(ns, "circle");
    fillCircle.setAttribute("cx", String(size / 2));
    fillCircle.setAttribute("cy", String(size / 2));
    fillCircle.setAttribute("r", String(r));
    fillCircle.setAttribute("fill", "none");
    fillCircle.setAttribute("stroke-width", String(stroke));
    fillCircle.setAttribute("stroke-dasharray", String(c));
    fillCircle.setAttribute("stroke-dashoffset", "0");
    fillCircle.setAttribute("stroke-linecap", "round");
    fillCircle.setAttribute("class", "connect-gauge-circle-fill connect-gauge-fill");
    fillCircle.dataset.circumference = String(c);
    g.appendChild(fillCircle);
    svg.appendChild(g);

    var text = document.createElementNS(ns, "text");
    text.setAttribute("x", "50%");
    text.setAttribute("y", "50%");
    text.setAttribute("dominant-baseline", "middle");
    text.setAttribute("text-anchor", "middle");
    text.setAttribute("class", "connect-gauge-text connect-gauge-label");
    text.id = "connect-gauge-label";
    svg.appendChild(text);
    el.appendChild(svg);
  }

  function setProgress(el, type, progress, label) {
    var p = Math.max(0, Math.min(1, progress));
    el.setAttribute("aria-valuenow", String(Math.round(p * 100)));
    if (type === "bar") {
      var fill = el.querySelector(".connect-gauge-bar-fill");
      if (fill) fill.style.width = p * 100 + "%";
      var lab = el.querySelector(".connect-gauge-label");
      if (lab) lab.textContent = label;
      return;
    }
    var circle = el.querySelector(".connect-gauge-circle-fill");
    if (circle) {
      var circ = parseFloat(circle.dataset.circumference || "0");
      circle.style.strokeDashoffset = String(circ * (1 - p));
      circle.setAttribute("stroke-dashoffset", String(circ * (1 - p)));
    }
    var text = el.querySelector(".connect-gauge-label");
    if (text) text.textContent = label;
  }

  async function fetchCode() {
    if (window.apiService && typeof window.apiService.getCode === "function") {
      return window.apiService.getCode();
    }
    var r = await fetch("/ui/api/newcode", {
      method: "GET",
      credentials: "same-origin",
      headers: { "Content-Type": "application/json" },
    });
    if (!r.ok) throw new Error(r.statusText || "Failed to fetch code");
    return r.json();
  }

  async function fetchCodeStatus() {
    if (window.apiService && typeof window.apiService.getCodeStatus === "function") {
      return window.apiService.getCodeStatus();
    }
    var r = await fetch("/ui/api/code", {
      method: "GET",
      credentials: "same-origin",
      headers: { Accept: "application/json" },
    });
    if (!r.ok) throw new Error(r.statusText || "Failed to check code");
    return r.json();
  }

  function init() {
    var codeEl = $("connect-code");
    var gaugeEl = $("connect-gauge");
    if (!codeEl || !gaugeEl) return;

    var statusEl = $("connect-status");
    var gaugeType = (gaugeEl.getAttribute("data-gauge-type") || "circle").toLowerCase();
    if (gaugeType !== "bar") gaugeType = "circle";
    buildGauge(gaugeEl, gaugeType);

    var issuedAt = 0;
    var activeCode = "";
    var busy = false;
    var raf = 0;
    var pollTimer = 0;
    var statusClearTimer = 0;

    function setStatus(msg, paired) {
      if (!statusEl) return;
      statusEl.textContent = msg || "";
      statusEl.hidden = !msg;
      statusEl.classList.toggle("is-paired", !!paired);
      if (statusClearTimer) clearTimeout(statusClearTimer);
      if (msg && paired) {
        statusClearTimer = setTimeout(function () {
          statusEl.textContent = "";
          statusEl.hidden = true;
          statusEl.classList.remove("is-paired");
        }, 8000);
      }
    }

    async function refresh(opts) {
      if (busy) return;
      busy = true;
      var paired = opts && opts.paired;
      try {
        var code = await fetchCode();
        activeCode = typeof code === "string" ? code : String(code);
        codeEl.textContent = activeCode;
        issuedAt = Date.now();
        if (paired) {
          setStatus("Device paired. New code ready for another tablet.", true);
        }
      } catch (e) {
        activeCode = "";
        codeEl.textContent = "········";
        console.error("[connect]", e);
      } finally {
        busy = false;
      }
    }

    async function pollActive() {
      if (busy || !activeCode || document.hidden) return;
      try {
        var st = await fetchCodeStatus();
        var serverCode = st && st.code ? String(st.code) : "";
        var active = !!(st && st.active && serverCode);
        if (!active || serverCode !== activeCode) {
          var remaining = issuedAt ? CODE_TTL_MS - (Date.now() - issuedAt) : 0;
          // Still within TTL → code was consumed by a successful pair (or replaced elsewhere).
          var paired = remaining > 2000;
          await refresh({ paired: paired });
        }
      } catch (e) {
        console.error("[connect] status", e);
      }
    }

    function frame() {
      if (!issuedAt) {
        setProgress(gaugeEl, gaugeType, 1, "");
      } else {
        var remaining = Math.max(0, CODE_TTL_MS - (Date.now() - issuedAt));
        var progress = remaining / CODE_TTL_MS;
        setProgress(gaugeEl, gaugeType, progress, formatLabel(remaining / 1000));
        if (remaining <= 0 && !busy) refresh();
      }
      raf = requestAnimationFrame(frame);
    }

    refresh();
    raf = requestAnimationFrame(frame);
    pollTimer = setInterval(pollActive, POLL_MS);

    var refreshBtn = document.getElementById("connect-refresh");
    if (refreshBtn) {
      refreshBtn.addEventListener("click", function () {
        setStatus("");
        refresh();
      });
    }

    document.addEventListener("visibilitychange", function () {
      if (!document.hidden) pollActive();
    });

    window.addEventListener(
      "pagehide",
      function () {
        cancelAnimationFrame(raf);
        if (pollTimer) clearInterval(pollTimer);
        if (statusClearTimer) clearTimeout(statusClearTimer);
      },
      { once: true }
    );
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", init);
  } else {
    init();
  }
})();
