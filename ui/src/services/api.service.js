import constants from "../common/constants";
import { jwtDecode } from "jwt-decode";

class ApiServices {
  header() {
    return {
      "Content-Type": "application/json",
    };
  }
  checkLogin() {
    if (localStorage.getItem("currentUser")) {
      return fetch(`${constants.ROOT_URL}/`, {
        method: "HEAD",
      }).then(handleError);
    }
  }
  login(loginData) {
    return fetch(`${constants.ROOT_URL}/login`, {
      method: "POST",
      headers: this.header(),
      credentials: "same-origin",
      body: JSON.stringify(loginData),
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
        if (!text || typeof text !== "string" || !/^[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+/.test(text)) {
          throw new Error("Invalid response from server. Check that the app URL and API are correct.");
        }
        let user = jwtDecode(text);
        localStorage.setItem("currentUser", JSON.stringify(user));
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

    return fetch(`${constants.ROOT_URL}/documents/upload`, {
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

  resetPassword(resetPasswordForm) {
    return fetch(`${constants.ROOT_URL}/profile`, {
      method: "POST",
      headers: this.header(),
      body: JSON.stringify({
        ...resetPasswordForm,
      }),
    });
  }

  listDocument() {
    return fetch(`${constants.ROOT_URL}/documents`, {
      method: "GET",
      headers: this.header(),
    }).then((r) => {
      handleError(r);
      return r.json();
    });
  }

  _isDocumentId(id) {
    return typeof id === "string" && /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i.test(id);
  }

  getTemplateUrl(id) {
    if (this._isDocumentId(id)) return `${constants.ROOT_URL}/documents/${id}/template`;
    return `${constants.ROOT_URL}/templates/${id}`;
  }

  getTemplate(id) {
    return fetch(this.getTemplateUrl(id), {
      method: "GET",
      credentials: "same-origin",
    }).then((r) => {
      handleError(r);
      return r.text();
    });
  }

  getMethodUrl(id) {
    if (this._isDocumentId(id)) return `${constants.ROOT_URL}/documents/${id}/template`;
    return `${constants.ROOT_URL}/methods/${id}`;
  }

  getMethod(id) {
    return fetch(this.getMethodUrl(id), {
      method: "GET",
      credentials: "same-origin",
    }).then((r) => {
      handleError(r);
      return r.text();
    });
  }

  getCode() {
    return fetch(`${constants.ROOT_URL}/newcode`, {
      method: "GET",
      headers: this.header(),
      credentials: "same-origin",
    }).then((r) => {
      handleError(r);
      return r.json();
    });
  }

  getCodeStatus() {
    return fetch(`${constants.ROOT_URL}/newcode/status`, {
      method: "GET",
      credentials: "same-origin",
    }).then((r) => {
      handleError(r);
      return r.json();
    });
  }

  listRegisteredDevices() {
    return fetch(`${constants.ROOT_URL}/devices`, {
      method: "GET",
      credentials: "same-origin",
    }).then((r) => {
      handleError(r);
      return r.json();
    });
  }

  deleteDocument(id) {
    return fetch(`${constants.ROOT_URL}/documents/${id}`, {
      method: "DELETE",
      headers: this.header(),
    }).then((r) => handleError(r));
  }
  download(id, exportType) {
    let url = `${constants.ROOT_URL}/documents/${id}`;
    if (exportType) url += `?type=${exportType}`;
    return fetch(url, {
      method: "GET",
      credentials: "same-origin",
    }).then((r) => {
      handleError(r);
      return r.blob();
    });
  }

  getDocumentMetadata(id) {
    return fetch(`${constants.ROOT_URL}/documents/${id}/metadata`, {
      method: "GET",
      credentials: "same-origin",
    }).then((r) => {
      handleError(r);
      return r.json();
    });
  }

  getDocumentPageBackgroundUrl(id, pageNum) {
    return `${constants.ROOT_URL}/documents/${id}/page/${pageNum}/background`;
  }

  getDocumentPagePngUrl(id, pageNum) {
    return `${constants.ROOT_URL}/documents/${id}/page/${pageNum}`;
  }

  /** Cover image URL for EPUB thumbnails (cover.htm/html/xhtml, or *0000.xhtml first img). */
  getEpubCoverThumbUrl(id) {
    return `${constants.ROOT_URL}/documents/${id}/epub/cover-thumb`;
  }

  getDocumentPageOverlayUrl(id, pageNum) {
    return `${constants.ROOT_URL}/documents/${id}/page/${pageNum}/overlay.svg`;
  }

  getDocumentPageOverlaySvg(id, pageNum) {
    return fetch(this.getDocumentPageOverlayUrl(id, pageNum), {
      method: "GET",
      credentials: "same-origin",
    }).then((r) => {
      handleError(r);
      return r.text();
    });
  }

  createFolder(data) {
    return fetch(`${constants.ROOT_URL}/folders`, {
      method: "POST",
      headers: this.header(),
      body: JSON.stringify(data),
    }).then((r) => {
      handleError(r);
      return r.json();
    });
  }
  updateuser(usr) {
    return fetch(`${constants.ROOT_URL}/users`, {
      method: "PUT",
      headers: this.header(),
      body: JSON.stringify(usr),
    }).then((r) => handleError(r));
  }
  createuser(usr) {
    return fetch(`${constants.ROOT_URL}/users`, {
      method: "POST",
      headers: this.header(),
      body: JSON.stringify(usr),
    }).then((r) => handleError(r));
  }
  deleteuser(userid) {
    return fetch(`${constants.ROOT_URL}/users/${userid}`, {
      method: "DELETE",
      headers: this.header(),
    }).then((r) => handleError(r));
  }

  listintegration() {
    return fetch(`${constants.ROOT_URL}/integrations`, {
      method: "GET",
      headers: this.header(),
    }).then((r) => {
      handleError(r);
      return r.json();
    });
  }
  updateintegration(integration) {
    return fetch(`${constants.ROOT_URL}/integrations/${integration.id}`, {
      method: "PUT",
      headers: this.header(),
      body: JSON.stringify(integration),
    }).then((r) => handleError(r));
  }
  createintegration(integration) {
    return fetch(`${constants.ROOT_URL}/integrations`, {
      method: "POST",
      headers: this.header(),
      body: JSON.stringify(integration),
    }).then((r) => handleError(r));
  }
  deleteintegration(integrationid) {
    return fetch(`${constants.ROOT_URL}/integrations/${integrationid}`, {
      method: "DELETE",
      headers: this.header(),
    }).then((r) => handleError(r));
  }
}

function removeUser(){
  localStorage.removeItem("currentUser");
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

const apiServices = new ApiServices()
export default apiServices
