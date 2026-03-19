# Unreleased

## Features

- Document type from `.content` file when present and size > 4 bytes (pdf, epub, notebook).
- Template thumbnails use `iconData` from `.template` file when available (base64 SVG).
- EPUB table of contents (TOC) uses black font color.
- PDF object fallback: "Alternative text: Download PDF" when plugin cannot render.

## Changed

- PDFs rendered solely as raw `application/pdf` via browser/Adobe plugin; removed react-pdf and PNG+overlay viewer.
- Error boundary ("Something went wrong") adds meta refresh to `/` and link to home.

---

# 0.0.25

## Features

- Software compatibility with 3.20 (cdb45df0b8314e637b5cdb722b10f0b262d74f56)
- Handle messaging integrations (a88aee6ea5ad846cd8aaab2bcbe2f82d2898e5f4)
- [Webhook messaging integration](https://ddvk.github.io/rmfakecloud/usage/integrations/#messaging-webhook) (479887ee4b335cd99f8a4cb4afeb7577681a217b)
- New option: `RMAPI_HWR_LANG_OVERRIDE` to override the language specified in myScript requests (#352)

## Internal change

- Refactor hash function (#365)
