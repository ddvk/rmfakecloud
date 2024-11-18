import constants from "../common/constants";
import jwt_decode from "jwt-decode";

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
      body: JSON.stringify(loginData),
    })
      .then((r) => {
        if (!r.ok) {
          throw new Error(r.statusText);
        }
        return r.text();
      })
      .then((text) => {
        let user = jwt_decode(text);
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
    }).then(r => r.json()); //.then(handleError);
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
  getCode() {
    return fetch(`${constants.ROOT_URL}/newcode`, {
      method: "GET",
      headers: this.header(),
    }).then((r) => {
      handleError(r);
      return r.json();
    });
  }
  download(id) {
    return fetch(`${constants.ROOT_URL}/documents/${id}`, {
      method: "GET",
      // headers: this.header()
    }).then((r) => {
      handleError(r);
      return r.blob();
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
