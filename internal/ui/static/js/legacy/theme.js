(function (global) {
  "use strict";
  var RM = global.RM;
  var X = global.RMXslt;
  if (!RM || !X) throw new Error("rm.js and xslt.js required for theme.js");

  function detectFormFactor() {
    try {
      if (window.matchMedia && window.matchMedia("(max-width: 768px)").matches) return "mobile";
    } catch (_) {}
    return "desktop";
  }

  var store = {
    layout: X.defaultLayout(),
    themeId: "default",
    formFactor: detectFormFactor(),
    colorOverrides: {},
    loading: true,
    listeners: [],
  };

  function emit() {
    store.listeners.slice().forEach(function (fn) {
      try {
        fn();
      } catch (e) {
        console.error(e);
      }
    });
  }

  function patch(partial) {
    Object.assign(store, partial);
    emit();
  }

  var xslCache = null;

  async function loadXsl() {
    if (xslCache) return xslCache;
    var pair = await Promise.all([
      X.fetchXML("/ui/api/themes/assets/theme-to-css.xsl"),
      X.fetchXML("/ui/api/themes/assets/theme-to-layout.xsl"),
    ]);
    xslCache = { cssXsl: pair[0], layoutXsl: pair[1] };
    return xslCache;
  }

  async function applyThemeXML(xmlString, overrides, factor) {
    overrides = overrides || {};
    factor = factor || detectFormFactor();
    try {
      var cache = await loadXsl();
      var themeDoc = X.parseXMLString(xmlString);
      themeDoc = X.applyColorOverrides(themeDoc, overrides);
      themeDoc = X.applyFormFactor(themeDoc, factor);
      var css = X.transformToText(themeDoc, cache.cssXsl);
      X.injectThemeCSS(css);
      var layoutDoc = X.transformWithXSLT(themeDoc, cache.layoutXsl);
      var parsed = X.parseLayoutDocument(layoutDoc);
      parsed.formFactor = factor;
      patch({ layout: parsed, formFactor: factor });
    } catch (e) {
      console.warn("theme apply failed", e);
      patch({ layout: X.defaultLayout() });
    }
  }

  async function reload() {
    patch({ loading: true });
    try {
      var tid = "default";
      var overrides = {};
      var factor = detectFormFactor();
      var auth = global.useAuthState ? global.useAuthState() : null;
      var user = auth && auth.state && auth.state.user;
      if (user) {
        try {
          var pref = await global.apiService.getProfileTheme();
          tid = pref.themeId || "default";
          overrides = pref.themeColorOverrides || {};
          if (pref.formFactor) factor = pref.formFactor;
        } catch (_) {}
      }
      var theme = await global.apiService.getTheme(tid);
      if (detectFormFactor() === "mobile") factor = "mobile";
      patch({ themeId: tid, colorOverrides: overrides });
      await applyThemeXML(theme.xml, overrides, factor);
    } catch (e) {
      console.warn(e);
      patch({ themeId: "default", layout: X.defaultLayout() });
    } finally {
      patch({ loading: false });
    }
  }

  async function applyThemeId(id, overrides) {
    var theme = await global.apiService.getTheme(id);
    patch({ themeId: id, colorOverrides: overrides || store.colorOverrides });
    await applyThemeXML(theme.xml, overrides || store.colorOverrides, store.formFactor);
  }

  function useShellTheme() {
    var tick = RM.useState(0);
    var setTick = tick[1];
    RM.useEffect(function () {
      var sub = function () {
        setTick(function (n) {
          return n + 1;
        });
      };
      store.listeners.push(sub);
      return function () {
        store.listeners = store.listeners.filter(function (x) {
          return x !== sub;
        });
      };
    }, []);
    return {
      layout: store.layout,
      themeId: store.themeId,
      formFactor: store.formFactor,
      colorOverrides: store.colorOverrides,
      loading: store.loading,
      reload: reload,
      applyThemeId: applyThemeId,
    };
  }

  function ThemeProvider(props) {
    RM.useEffect(function () {
      reload();
      var onResize = function () {
        var f = detectFormFactor();
        if (f !== store.formFactor) {
          patch({ formFactor: f });
          reload();
        }
      };
      window.addEventListener("resize", onResize);
      return function () {
        window.removeEventListener("resize", onResize);
      };
    }, []);
    return RM.h("div", { className: "rm-theme-provider", style: { display: "contents" } }, props.children);
  }

  global.useShellTheme = useShellTheme;
  global.ThemeProvider = ThemeProvider;
  global.detectFormFactor = detectFormFactor;
  global.reloadShellTheme = reload;
})(typeof window !== "undefined" ? window : globalThis);
