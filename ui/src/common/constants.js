const constants = {
    ROOT_URL :  "/ui/api",
    // Optional: set VITE_ADOBE_PDF_CLIENT_ID for "Original PDF" (no drawings) viewer
    ADOBE_CLIENT_ID: typeof import.meta !== "undefined" && import.meta.env?.VITE_ADOBE_PDF_CLIENT_ID
        ? import.meta.env.VITE_ADOBE_PDF_CLIENT_ID
        : "",
}

export default constants