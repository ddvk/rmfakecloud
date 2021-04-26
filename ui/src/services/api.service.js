import constants from "../common/constants"

class ApiServices {
    header() {
        var token = localStorage.getItem("token")

        return {
            "Content-Type": "application/json",
            Authorization: `Bearer ${token}`,
        };
    }

    upload(files) {
        const formData = new FormData();
        files.forEach(f => {
            formData.append("file", f)
        });

        let { Authorization } = this.header()
        return fetch(`${constants.ROOT_URL}/documents/upload`, {
            method: "POST",
            headers: {
                Authorization
            },
            body: formData
        })
            .then(handleError)

    }

    resetPassword(resetPasswordForm) {

        return fetch(`${constants.ROOT_URL}/resetPassword`, {
            method: "POST",
            headers: this.header(),
            body: JSON.stringify({
                ...resetPasswordForm,
            }),
        })
    }


    listDocument() {
        return fetch(`${constants.ROOT_URL}/documents`, {
            method: "GET",
            headers: this.header()
        })
            .then(r => {
                handleError(r)
                return r.json()
            })
    }
    getCode() {
        return fetch(`${constants.ROOT_URL}/newcode`, {
            method: "GET",
            headers: this.header()
        })
            .then(r => {
                handleError(r)
                return r.json()
            })
    }
    download(id) {
        return fetch(`${constants.ROOT_URL}/documents/${id}`, {
            method: "GET",
            headers: this.header()
        })
            .then(r => {
                handleError(r)

                return r.blob()
            })
    }
}

function handleError(r) {
    if (r.status === 401) {
        localStorage.removeItem("token");
        localStorage.removeItem("currentUser");
        window.location.replace("/login")
        throw new Error("not authorized")
    }
    if (!r.ok) {
        throw new Error(r.status)
    }
}

export default new ApiServices()
