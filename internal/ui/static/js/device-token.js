/**
 * Device token result page: copy token to clipboard.
 */
(function () {
  "use strict";

  function init() {
    var btn = document.getElementById("device-token-copy");
    var ta = document.getElementById("device-token-value");
    if (!btn || !ta) return;
    btn.addEventListener("click", function () {
      var text = ta.value || "";
      if (!text) return;
      if (navigator.clipboard && navigator.clipboard.writeText) {
        navigator.clipboard.writeText(text).then(
          function () {
            btn.textContent = "Copied";
          },
          function () {
            ta.select();
            document.execCommand("copy");
            btn.textContent = "Copied";
          }
        );
      } else {
        ta.select();
        document.execCommand("copy");
        btn.textContent = "Copied";
      }
    });
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", init);
  } else {
    init();
  }
})();
