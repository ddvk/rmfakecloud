(function (global) {
  "use strict";
  var constants = global.constants || { ROOT_URL: "/ui/api" };
  function jwtDecode(token) {
    var parts = String(token).split(".");
    if (parts.length < 2) throw new Error("invalid jwt");
    var json = parts[1].replace(/-/g, "+").replace(/_/g, "/");
    while (json.length % 4) json += "=";
    return JSON.parse(atob(json));
  }


class ApiServices {
  header() {
    return {
      "Content-Type": "application/json",
    };
  }
  checkLogin() {
    if (localStorage.getItem("currentUser")) {
      return fetch(`${constants.ROOT_URL}/`, { credentials: "same-origin",
        method: "HEAD",
      }).then(handleError);
    }
  }
  login(loginData) {
    return fetch(`${constants.ROOT_URL}/login`, { credentials: "same-origin",
      method: "POST",
      headers: this.header(),
      body: JSON.stringify(loginData),
    })
      .then((r) => {
        if (!r.ok) {
          throw new Error(r.statusText);
        }
        return r.text();
      })
      .then((text) => {
        let user = jwtDecode(text);
        localStorage.setItem("currentUser", JSON.stringify(user));
        localStorage.setItem("authToken", text);
        return user;
      });
  }
  webAuthnStatus() {
    return fetch(`${constants.ROOT_URL}/webauthn/status`, {
      method: "GET",
      headers: this.header(),
      credentials: "same-origin",
    }).then(async (r) => {
      if (!r.ok) return { enabled: false };
      return r.json();
    }).catch(() => ({ enabled: false }));
  }
  webAuthnRegisterBegin() {
    return fetch(`${constants.ROOT_URL}/webauthn/register/begin`, {
      method: "POST",
      headers: this.header(),
      credentials: "same-origin",
    }).then(async (r) => {
      handleError(r);
      return r.json();
    });
  }
  webAuthnRegisterFinish(sessionId, credential, name) {
    return fetch(`${constants.ROOT_URL}/webauthn/register/finish`, {
      method: "POST",
      headers: this.header(),
      credentials: "same-origin",
      body: JSON.stringify({ sessionId, credential, name }),
    }).then(async (r) => {
      handleError(r);
      return r.json();
    });
  }
  listWebAuthnCredentials() {
    return fetch(`${constants.ROOT_URL}/webauthn/credentials`, {
      method: "GET",
      headers: this.header(),
      credentials: "same-origin",
    }).then(async (r) => {
      handleError(r);
      return r.json();
    });
  }
  deleteWebAuthnCredential(id) {
    return fetch(`${constants.ROOT_URL}/webauthn/credentials/${encodeURIComponent(id)}`, {
      method: "DELETE",
      headers: this.header(),
      credentials: "same-origin",
    }).then((r) => handleError(r));
  }
  webAuthnLoginBegin() {
    return fetch(`${constants.ROOT_URL}/webauthn/login/begin`, {
      method: "POST",
      headers: this.header(),
      credentials: "same-origin",
    }).then(async (r) => {
      handleError(r);
      return r.json();
    });
  }
  webAuthnLoginFinish(sessionId, credential) {
    return fetch(`${constants.ROOT_URL}/webauthn/login/finish`, {
      method: "POST",
      headers: this.header(),
      credentials: "same-origin",
      body: JSON.stringify({ sessionId, credential }),
    })
      .then(async (r) => {
        const text = await r.text();
        if (!r.ok) {
          let msg = r.statusText;
          try {
            if (text && text.startsWith("{")) {
              const j = JSON.parse(text);
              if (j.error) msg = j.error;
            }
          } catch (_) {}
          throw new Error(msg);
        }
        return text;
      })
      .then((text) => {
        let user = jwtDecode(text);
        localStorage.setItem("currentUser", JSON.stringify(user));
        localStorage.setItem("authToken", text);
        return user;
      });
  }
  logout() {
    removeUser();
    fetch(`${constants.ROOT_URL}/logout`);
  }

  upload(parent, files) {
    const formData = new FormData();
    formData.append("parent", parent);
    files.forEach((f) => {
      // file extensions which are not lowecase break the upload

      // set the file extension to be lower case
      let sub = f.name.split(".")
      const ext = sub[sub.length - 1].toLowerCase()
      sub.pop()
      sub.push(ext)

      // copy file data to new (lowecase) file
      const newFile = new File([f], sub.join("."), {type: f.type});
      formData.append("file", newFile);
    });

    return fetch(`${constants.ROOT_URL}/documents/upload`, { credentials: "same-origin",
      method: "POST",
      body: formData,
    }).then(async (r) => {
      if (r.status === 409) {
        const body = await r.json();
        const err = new Error(body.error);
        err.status = 409;
        err.docId = body.docId;
        throw err;
      }
      if (!r.ok) {
        const body = await r.json();
        throw new Error(body.error || r.statusText);
      }
      return r.json();
    });
  }

  listPasscodeResets() {
    return fetch(`${constants.ROOT_URL}/passcode/resets`, { credentials: "same-origin",
      method: "GET",
      headers: this.header(),
    }).then((r) => {
      handleError(r);
      return r.json();
    });
  }
  approvePasscodeReset(uuid) {
    return fetch(`${constants.ROOT_URL}/passcode/resets/${uuid}/approve`, { credentials: "same-origin",
      method: "POST",
      headers: this.header(),
    }).then((r) => handleError(r));
  }
  dismissPasscodeReset(uuid) {
    return fetch(`${constants.ROOT_URL}/passcode/resets/${uuid}`, { credentials: "same-origin",
      method: "DELETE",
      headers: this.header(),
    }).then((r) => handleError(r));
  }

  resetPassword(resetPasswordForm) {
    return fetch(`${constants.ROOT_URL}/profile`, { credentials: "same-origin",
      method: "POST",
      headers: this.header(),
      body: JSON.stringify({
        ...resetPasswordForm,
      }),
    });
  }

  listDocument() {
    return fetch(`${constants.ROOT_URL}/documents`, { credentials: "same-origin",
      method: "GET",
      headers: this.header(),
    }).then((r) => {
      handleError(r);
      return r.json();
    });
  }
  getCode() {
    return fetch(`${constants.ROOT_URL}/newcode`, { credentials: "same-origin",
      method: "GET",
      headers: this.header(),
    }).then((r) => {
      handleError(r);
      return r.json();
    });
  }

  deleteDocument(id) {
    return fetch(`${constants.ROOT_URL}/documents/${id}`, { credentials: "same-origin",
      method: "DELETE",
      headers: this.header(),
    }).then((r) => handleError(r));
  }
  download(id, exportType) {
    let url = `${constants.ROOT_URL}/documents/${id}`;
    if (exportType) url += `?type=${exportType}`;
    return fetch(url, {
      method: "GET",
    }).then((r) => {
      handleError(r);
      return r.blob();
    });
  }

  createFolder(data) {
    return fetch(`${constants.ROOT_URL}/folders`, { credentials: "same-origin",
      method: "POST",
      headers: this.header(),
      body: JSON.stringify(data),
    }).then((r) => {
      handleError(r);
      return r.json();
    });
  }
  updateuser(usr) {
    return fetch(`${constants.ROOT_URL}/users`, { credentials: "same-origin",
      method: "PUT",
      headers: this.header(),
      body: JSON.stringify(usr),
    }).then((r) => handleError(r));
  }
  createuser(usr) {
    return fetch(`${constants.ROOT_URL}/users`, { credentials: "same-origin",
      method: "POST",
      headers: this.header(),
      body: JSON.stringify(usr),
    }).then((r) => handleError(r));
  }
  deleteuser(userid) {
    return fetch(`${constants.ROOT_URL}/users/${userid}`, { credentials: "same-origin",
      method: "DELETE",
      headers: this.header(),
    }).then((r) => handleError(r));
  }

  listintegration() {
    return fetch(`${constants.ROOT_URL}/integrations`, { credentials: "same-origin",
      method: "GET",
      headers: this.header(),
    }).then((r) => {
      handleError(r);
      return r.json();
    });
  }
  updateintegration(integration) {
    return fetch(`${constants.ROOT_URL}/integrations/${integration.id}`, { credentials: "same-origin",
      method: "PUT",
      headers: this.header(),
      body: JSON.stringify(integration),
    }).then((r) => handleError(r));
  }
  createintegration(integration) {
    return fetch(`${constants.ROOT_URL}/integrations`, { credentials: "same-origin",
      method: "POST",
      headers: this.header(),
      body: JSON.stringify(integration),
    }).then((r) => handleError(r));
  }
  deleteintegration(integrationid) {
    return fetch(`${constants.ROOT_URL}/integrations/${integrationid}`, { credentials: "same-origin",
      method: "DELETE",
      headers: this.header(),
    }).then((r) => handleError(r));
  }

  listThemes() {
    return fetch(`${constants.ROOT_URL}/themes`, {
      method: "GET",
      headers: this.header(),
      credentials: "same-origin",
    }).then(async (r) => {
      handleError(r);
      return r.json();
    });
  }
  getTheme(id) {
    return fetch(`${constants.ROOT_URL}/themes/${encodeURIComponent(id)}`, {
      method: "GET",
      headers: this.header(),
      credentials: "same-origin",
    }).then(async (r) => {
      handleError(r);
      return r.json();
    });
  }
  getProfileTheme() {
    return fetch(`${constants.ROOT_URL}/profile/theme`, {
      method: "GET",
      headers: this.header(),
      credentials: "same-origin",
    }).then(async (r) => {
      handleError(r);
      return r.json();
    });
  }
  putProfileTheme(body) {
    return fetch(`${constants.ROOT_URL}/profile/theme`, {
      method: "PUT",
      headers: this.header(),
      credentials: "same-origin",
      body: JSON.stringify(body),
    }).then(async (r) => {
      handleError(r);
      return r.json();
    });
  }
  saveTheme(body) {
    return fetch(`${constants.ROOT_URL}/themes`, {
      method: "POST",
      headers: this.header(),
      credentials: "same-origin",
      body: JSON.stringify(body),
    }).then(async (r) => {
      handleError(r);
      return r.json();
    });
  }
  updateTheme(id, body) {
    return fetch(`${constants.ROOT_URL}/themes/${encodeURIComponent(id)}`, {
      method: "PUT",
      headers: this.header(),
      credentials: "same-origin",
      body: JSON.stringify(body),
    }).then(async (r) => {
      handleError(r);
      return r.json();
    });
  }
  publishTheme(id, published = true) {
    return fetch(`${constants.ROOT_URL}/themes/${encodeURIComponent(id)}/publish`, {
      method: "POST",
      headers: this.header(),
      credentials: "same-origin",
      body: JSON.stringify({ published }),
    }).then(async (r) => {
      handleError(r);
      return r.json();
    });
  }
  deleteTheme(id) {
    return fetch(`${constants.ROOT_URL}/themes/${encodeURIComponent(id)}`, {
      method: "DELETE",
      headers: this.header(),
      credentials: "same-origin",
    }).then((r) => handleError(r));
  }
}

function removeUser(){
  localStorage.removeItem("currentUser");
  localStorage.removeItem("authToken");
}
function handleError(r) {
  if (!r.ok) {
    if (r.status === 401) {
      removeUser();
      window.location.reload(true);
      return
    }
    if (r.headers.get("Content-Type").startsWith("application/json")) {
      return r.json().then(d => {throw new Error(d.error)});
    }
    if (r.status === 400) {
      return r.text().then(text => {throw new Error(text)})
    }
    return Promise.reject(r.status)
  }
}

var apiServices = new ApiServices();

  global.apiService = apiServices;
  global.ApiServices = ApiServices;
})(typeof window !== "undefined" ? window : globalThis);
