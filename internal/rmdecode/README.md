# .rm (reMarkable lines) decoder

This repo includes tooling to inspect raw page files (`.rm` / “lines” format):

- **v3 / v5** — decoded in **Go** via [`github.com/juruen/rmapi/encoding/rm`](https://github.com/juruen/rmapi) (`internal/rmdecode`, `cmd/rmdecode`).
- **v6** (tablet OS 3.x) — summaries via **Python + [rmscene](https://github.com/ricklupton/rmscene)** (`scripts/rmdecode_v6_summary.py`). Upstream **rmscene** is vendored as a git submodule at **`third_party/rmscene`** (see below).

Production-grade v6 tooling in the wider ecosystem is **[rmscene](https://github.com/ricklupton/rmscene) + [rmc](https://github.com/ricklupton/rmc)** (SVG/PDF export and other conversions). This repository wires **Go for v3/v5** and **Python + rmscene for v6 summaries**; use **rmc** separately when you need full SVG/PDF export from `.rm` files.

## Versions

| Header prefix | Decoder |
|---------------|---------|
| `version=3` / `version=5` | Go: `DecodeLegacy` → `github.com/juruen/rmapi/encoding/rm` |
| `version=6` (tablet OS 3.x) | Tagged blocks / scene tree: [rmscene](https://github.com/ricklupton/rmscene) via `scripts/rmdecode_v6_summary.py`; for SVG/PDF use [rmc](https://github.com/ricklupton/rmc), which bundles rmscene. |

## CLI

From the repo root:

```bash
# Text dump (v3/v5) or JSON summary (v6; needs rmscene)
go run ./cmd/rmdecode /path/to/page.rm

# All writings → SVG (v3/v5 native; v6 needs `rmc` on PATH)
go run ./cmd/rmdecode -format svg -o out.svg /path/to/page.rm

# All writings → PDF
go run ./cmd/rmdecode -format pdf -o out.pdf /path/to/page.rm
# or infer format from extension:
go run ./cmd/rmdecode -o out.svg /path/to/page.rm
```

- **v3 / v5**: SVG and PDF are generated in Go (`RenderWritingsSVG`, `RenderWritingsPDF`).
- **v6**: SVG/PDF are produced by invoking **`rmc`** (`pipx install rmc` or `pip install rmc`). Text summary still uses Python + **rmscene** (`scripts/rmdecode_v6_summary.py`).

### Submodule `third_party/rmscene`

After cloning this repo, initialize submodules:

```bash
git submodule update --init --recursive
```

Use the vendored source without PyPI (set `PYTHONPATH` to the package root):

```bash
export PYTHONPATH="$(pwd)/third_party/rmscene/src"
python3 scripts/rmdecode_v6_summary.py /path/to/page.rm
```

Alternatively install from PyPI or editable from the submodule (venv recommended):

```bash
pip install 'rmscene>=0.6'
# or: pip install -e ./third_party/rmscene
```

Full SVG/PDF export for v6 (not included in this repo’s CLI):

```bash
pipx install rmc   # or: pip install rmc
rmc -t svg -o out.svg page.rm
```

### `.rmdoc` → PNG for every `.rm` page (paging)

From the repo root, after `git submodule update --init --recursive` (for **v6** via vendored **rmscene**):

```bash
pip install Pillow   # for v6 pages
go run ./cmd/rmdoc2png -o ./out_pngs path/to/document.rmdoc
```

- Walks the **zip** and processes every `*.rm` member in archive order.
- Output files: **`001_<sanitized_path>.png`**, **`002_…`**, … (1-based page index + original stem).
- **v3/v5**: rendered in Go (`RenderWritingsPNG`, [gg](https://github.com/fogleman/gg)).
- **v6**: `scripts/rmscene_v6_to_png.py` + `third_party/rmscene` + **Pillow**.
- Default output directory is `<rmdoc_basename>_png` if `-o` is omitted.

## References

- [Older binary layout (v3)](https://plasma.ninja/blog/devices/remarkable/binary/format/2017/12/26/reMarkable-lines-file-format.html)
- [Notes on v6 vs v5](https://github.com/chemag/maxio/blob/master/version6.md)
