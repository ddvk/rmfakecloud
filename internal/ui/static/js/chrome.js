(function (global) {
  "use strict";
  var RM = global.RM;
  var h = RM.h;

  function readNav() {
    var el = document.getElementById("rm-nav");
    if (!el) return [];
    try {
      return JSON.parse(el.textContent || "[]");
    } catch (_) {
      return [];
    }
  }

  var NAV_LABELS = {
    documents: "Documents",
    integrations: "Integrations",
    connect: "Connect",
    screenshare: "Screen share",
    admin: "Admin",
    help: "Help",
    profile: "Settings",
  };

  function navFromTheme(layout) {
    var server = readNav();
    var byId = {};
    server.forEach(function (it) {
      byId[it.id] = it;
    });
    var items = (layout && layout.nav && layout.nav.items) || [];
    if (!items.length) return server;

    var seen = {};
    var out = items
      .filter(function (it) {
        return it && it.id && it.visible !== false && String(it.visible) !== "false";
      })
      .map(function (it) {
        var src = byId[it.id] || {};
        seen[it.id] = true;
        return {
          id: it.id,
          href: src.href || it.href || "/" + it.id,
          label: it.label || src.label || NAV_LABELS[it.id] || it.id,
          icon: it.icon || src.icon || it.id,
          adminOnly: it.adminOnly != null ? !!it.adminOnly : !!src.adminOnly,
        };
      });

    server.forEach(function (it) {
      if (seen[it.id]) return;
      out.push(it);
    });
    return out.length ? out : server;
  }

  function resolveChromeStyle(explicit, layout) {
    var normalize = global.RMShells && global.RMShells.normalizeChromeStyle;
    var fromLayout = layout && layout.chrome && layout.chrome.style;
    var style = explicit || fromLayout || "remarkable";
    return normalize ? normalize(style) : style;
  }

  function ShellChrome(props) {
    var theme = global.useShellTheme ? global.useShellTheme() : { layout: {}, formFactor: "desktop" };
    var formFactor =
      props.forceMobile || theme.formFactor === "mobile" || (global.detectFormFactor && global.detectFormFactor() === "mobile")
        ? "mobile"
        : "desktop";
    var chromeStyle = resolveChromeStyle(props.chromeStyle, theme.layout);
    var nav = navFromTheme(theme.layout);
    if (!nav.length) nav = readNav();

    var Shell = (global.RMShells && global.RMShells[chromeStyle]) || global.RMShells.remarkable;
    return h(Shell, {
      nav: nav,
      mainSlot: props.mainSlot,
      layout: theme.layout,
      formFactor: formFactor,
    });
  }

  function mount() {
    if (document.body.classList.contains("page-login") || document.body.classList.contains("page-error")) {
      return;
    }
    var main = document.getElementById("main");
    if (!main) return;

    var root = document.getElementById("rm-shell-root");
    if (!root) {
      root = document.createElement("div");
      root.id = "rm-shell-root";
      document.body.insertBefore(root, document.body.firstChild);
    }

    var flashes = Array.prototype.slice.call(root.querySelectorAll(".flash"));
    flashes.forEach(function (f) {
      document.body.insertBefore(f, root);
    });

    var mainHost = document.createElement("div");
    mainHost.id = "rm-main-host";
    mainHost.appendChild(main);

    var chromeMatch = (document.body.className || "").match(/chrome-([a-z]+)/);
    var chromeStyle =
      global.RMShells && global.RMShells.normalizeChromeStyle
        ? global.RMShells.normalizeChromeStyle(chromeMatch ? chromeMatch[1] : "remarkable")
        : chromeMatch
          ? chromeMatch[1]
          : "remarkable";

    function Root() {
      var theme = global.useShellTheme
        ? global.useShellTheme()
        : { formFactor: global.detectFormFactor && global.detectFormFactor() };
      RM.useEffect(function () {
        if (global.reloadShellTheme) global.reloadShellTheme();
        if (global.enhanceTablePagination) {
          global.enhanceTablePagination(".profile-panel .data-table");
          global.enhanceTablePagination(".admin-panel .data-table");
          global.enhanceTablePagination(".integrations-panel .data-table");
          global.enhanceTablePagination(".themes-panel .data-table");
        }
      }, []);
      return h(ShellChrome, {
        chromeStyle: chromeStyle,
        forceMobile: theme.formFactor === "mobile",
        mainSlot: mainHost,
      });
    }

    try {
      RM.render(Root, root);
      document.body.classList.add("has-js-shell");
    } catch (err) {
      console.error("[chrome] shell mount failed", err);
    }

    if (global.apiService && global.apiService.listPasscodeResets) {
      mountPasscodeResets();
    }
  }

  function mountPasscodeResets() {
    var host = document.createElement("div");
    host.id = "rm-passcode-resets";
    document.body.appendChild(host);
    async function tick() {
      try {
        var list = await global.apiService.listPasscodeResets();
        if (!list || !list.length) {
          host.innerHTML = "";
          return;
        }
        host.innerHTML =
          '<div class="flash flash-info rm-passcode-banner" role="status">Passcode reset request(s) pending. <button type="button" class="btn btn-sm btn-primary" id="rm-approve-passcode">Approve</button></div>';
        var btn = document.getElementById("rm-approve-passcode");
        if (btn && list[0].uuid) {
          btn.onclick = async function () {
            await global.apiService.approvePasscodeReset(list[0].uuid);
            tick();
          };
        }
      } catch (_) {}
    }
    tick();
    setInterval(tick, 10000);
  }

  global.RMChrome = { mount: mount, readNav: readNav, navFromTheme: navFromTheme };

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", mount);
  } else {
    mount();
  }
})(typeof window !== "undefined" ? window : globalThis);
