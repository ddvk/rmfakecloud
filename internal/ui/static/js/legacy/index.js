(function (global) {
  "use strict";
  // Load order documented in chrome.js / shell.xsl:
  // constants → api.service → xslt → auth → theme → useFetch → toast → link → pagination
  global.RMLegacy = {
    ready: true,
  };
})(typeof window !== "undefined" ? window : globalThis);
