(function () {
  "use strict";

  var API = "/ui/api/webauthn";

  function b64urlToBuf(val) {
    if (!val) return new ArrayBuffer(0);
    if (val instanceof ArrayBuffer) return val;
    if (ArrayBuffer.isView(val)) return val.buffer.slice(val.byteOffset, val.byteOffset + val.byteLength);
    var s = String(val).replace(/-/g, "+").replace(/_/g, "/");
    var pad = (4 - (s.length % 4)) % 4;
    if (pad) s += "====".slice(0, pad);
    var bin = atob(s);
    var out = new Uint8Array(bin.length);
    for (var i = 0; i < bin.length; i++) out[i] = bin.charCodeAt(i);
    return out.buffer;
  }

  function bufToB64url(buf) {
    var bytes = buf instanceof ArrayBuffer ? new Uint8Array(buf) : new Uint8Array(buf.buffer || buf);
    var s = "";
    for (var i = 0; i < bytes.length; i++) s += String.fromCharCode(bytes[i]);
    return btoa(s).replace(/\+/g, "-").replace(/\//g, "_").replace(/=+$/g, "");
  }

  function copyCreds(list) {
    if (!list || !list.length) return list;
    return list.map(function (c) {
      var o = {};
      Object.keys(c).forEach(function (k) {
        o[k] = c[k];
      });
      if (c.id) o.id = b64urlToBuf(c.id);
      return o;
    });
  }

  function nativeLib() {
    return {
      browserSupportsWebAuthn: function () {
        return typeof window.PublicKeyCredential === "function";
      },
      startRegistration: async function (opts) {
        var json = opts && opts.optionsJSON ? opts.optionsJSON : opts;
        if (!json || !json.challenge || !json.user) {
          throw new Error("Invalid registration options");
        }
        var publicKey = {};
        Object.keys(json).forEach(function (k) {
          publicKey[k] = json[k];
        });
        publicKey.challenge = b64urlToBuf(json.challenge);
        publicKey.user = {};
        Object.keys(json.user).forEach(function (k) {
          publicKey.user[k] = json.user[k];
        });
        publicKey.user.id = b64urlToBuf(json.user.id);
        publicKey.excludeCredentials = copyCreds(json.excludeCredentials);
        var cred = await navigator.credentials.create({ publicKey: publicKey });
        if (!cred) throw new Error("Registration was not completed");
        var resp = cred.response;
        var out = {
          id: cred.id,
          rawId: bufToB64url(cred.rawId),
          type: cred.type,
          clientExtensionResults: cred.getClientExtensionResults ? cred.getClientExtensionResults() : {},
          response: {
            clientDataJSON: bufToB64url(resp.clientDataJSON),
            attestationObject: bufToB64url(resp.attestationObject),
          },
        };
        if (cred.authenticatorAttachment) out.authenticatorAttachment = cred.authenticatorAttachment;
        if (resp.getTransports) out.response.transports = resp.getTransports();
        return out;
      },
      startAuthentication: async function (opts) {
        var json = opts && opts.optionsJSON ? opts.optionsJSON : opts;
        if (!json || !json.challenge) {
          throw new Error("Invalid authentication options");
        }
        var publicKey = {};
        Object.keys(json).forEach(function (k) {
          publicKey[k] = json[k];
        });
        publicKey.challenge = b64urlToBuf(json.challenge);
        publicKey.allowCredentials = copyCreds(json.allowCredentials);
        var cred = await navigator.credentials.get({ publicKey: publicKey });
        if (!cred) throw new Error("Authentication was not completed");
        var resp = cred.response;
        var out = {
          id: cred.id,
          rawId: bufToB64url(cred.rawId),
          type: cred.type,
          clientExtensionResults: cred.getClientExtensionResults ? cred.getClientExtensionResults() : {},
          response: {
            authenticatorData: bufToB64url(resp.authenticatorData),
            clientDataJSON: bufToB64url(resp.clientDataJSON),
            signature: bufToB64url(resp.signature),
            userHandle: resp.userHandle ? bufToB64url(resp.userHandle) : undefined,
          },
        };
        if (cred.authenticatorAttachment) out.authenticatorAttachment = cred.authenticatorAttachment;
        return out;
      },
    };
  }

  function swa() {
    var lib = window.SimpleWebAuthnBrowser;
    if (lib && typeof lib.startAuthentication === "function") return lib;
    return nativeLib();
  }

  function jsonHeaders() {
    return { "Content-Type": "application/json" };
  }

  async function readError(r) {
    var text = await r.text();
    var msg = r.statusText || "Request failed";
    try {
      if (text && text.charAt(0) === "{") {
        var j = JSON.parse(text);
        if (j.error) msg = j.error;
      } else if (text) {
        msg = text;
      }
    } catch (_) {}
    throw new Error(msg);
  }

  async function postJSON(url, body) {
    var r = await fetch(url, {
      method: "POST",
      headers: jsonHeaders(),
      credentials: "same-origin",
      body: body ? JSON.stringify(body) : undefined,
    });
    if (!r.ok) await readError(r);
    var ct = r.headers.get("content-type") || "";
    if (ct.indexOf("application/json") !== -1) return r.json();
    return r.text();
  }

  async function del(url) {
    var r = await fetch(url, {
      method: "DELETE",
      headers: jsonHeaders(),
      credentials: "same-origin",
    });
    if (!r.ok) await readError(r);
  }

  function setBusy(btn, busy) {
    if (!btn) return;
    btn.disabled = !!busy;
    btn.setAttribute("aria-busy", busy ? "true" : "false");
  }

  function showError(msg) {
    var el = document.getElementById("passkey-error") || document.getElementById("flash") || document.querySelector("[data-passkey-error]");
    if (el) {
      el.textContent = msg;
      el.hidden = false;
    } else {
      console.error("[webauthn]", msg);
      window.alert(msg);
    }
  }

  async function handleLogin(e) {
    if (e) e.preventDefault();
    var lib = swa();
    if (!lib || typeof lib.startAuthentication !== "function") {
      showError("WebAuthn library not loaded");
      return;
    }
    if (typeof lib.browserSupportsWebAuthn === "function" && !lib.browserSupportsWebAuthn()) {
      showError("Passkeys are not supported in this browser");
      return;
    }

    var btn = document.getElementById("passkey-login");
    setBusy(btn, true);
    try {
      var begin = await postJSON(API + "/login/begin");
      var credential = await lib.startAuthentication({ optionsJSON: begin.publicKey });
      await postJSON(API + "/login/finish", {
        sessionId: begin.sessionId,
        credential: credential,
      });
      var next = (btn && btn.getAttribute("data-redirect")) || "/documents";
      window.location.href = next;
    } catch (err) {
      showError("Passkey login failed: " + (err.message || String(err)));
    } finally {
      setBusy(btn, false);
    }
  }

  async function handleRegister(e) {
    if (e) e.preventDefault();
    var lib = swa();
    if (!lib || typeof lib.startRegistration !== "function") {
      showError("WebAuthn library not loaded");
      return;
    }

    var btn = document.getElementById("passkey-register");
    var nameInput = document.getElementById("passkey-name");
    var name = nameInput && nameInput.value ? nameInput.value.trim() : "";
    setBusy(btn, true);
    try {
      var begin = await postJSON(API + "/register/begin");
      var credential = await lib.startRegistration({ optionsJSON: begin.publicKey });
      await postJSON(API + "/register/finish", {
        sessionId: begin.sessionId,
        credential: credential,
        name: name || "Passkey",
      });
      if (nameInput) nameInput.value = "";
      window.location.reload();
    } catch (err) {
      showError(err.message || String(err));
    } finally {
      setBusy(btn, false);
    }
  }

  async function handleDelete(btn) {
    var id = btn.getAttribute("data-cred-delete");
    if (!id) return;
    if (!window.confirm("Remove this passkey?")) return;
    setBusy(btn, true);
    try {
      await del(API + "/credentials/" + encodeURIComponent(id));
      var row = btn.closest("tr") || btn.closest("li") || btn.parentElement;
      if (row) row.remove();
    } catch (err) {
      showError(err.message || String(err));
      setBusy(btn, false);
    }
  }

  function init() {
    var loginBtn = document.getElementById("passkey-login");
    if (loginBtn) {
      loginBtn.addEventListener("click", handleLogin);
    }

    var registerBtn = document.getElementById("passkey-register");
    if (registerBtn) {
      registerBtn.addEventListener("click", handleRegister);
      var form = registerBtn.closest("form");
      if (form) {
        form.addEventListener("submit", function (e) {
          e.preventDefault();
          handleRegister(e);
        });
      }
    }

    document.querySelectorAll("[data-cred-delete]").forEach(function (btn) {
      btn.addEventListener("click", function (e) {
        e.preventDefault();
        handleDelete(btn);
      });
    });
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", init);
  } else {
    init();
  }
})();
