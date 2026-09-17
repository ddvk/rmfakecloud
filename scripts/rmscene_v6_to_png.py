#!/usr/bin/env python3
"""
Render a single v6 .rm page to PNG.

reMarkable v6 scene coordinates are centered on X (0 = page center) with Y=0 at
the top of the page. For notebooks that page is the device viewport (1404×1872).
For PDF annotations the page is the PDF page in screen pixels:

  doc_w = page_width_pt  * (226/72)
  doc_h = page_height_pt * (226/72)
  px = (x + doc_w/2) * (fit_width / doc_w)
  py = y * (fit_width / doc_w)

When PDF page size in points is passed, output is width-fitted to 1404 (natural
aspect) so it multiplies cleanly onto a pdftoppm raster of the same size.

Usage:
  python3 scripts/rmscene_v6_to_png.py page.rm > page.png
  python3 scripts/rmscene_v6_to_png.py - < page.rm > page.png
  python3 scripts/rmscene_v6_to_png.py --pdf-pts 612 792 page.rm > page.png
  python3 scripts/rmscene_v6_to_png.py --pdf-pts 612 792 - < page.rm > page.png
"""

from __future__ import annotations

import math
import pathlib
import sys
from io import BytesIO

_REPO_ROOT = pathlib.Path(__file__).resolve().parents[1]
sys.path.insert(0, str(_REPO_ROOT / "third_party" / "rmscene" / "src"))

from rmscene import read_tree  # noqa: E402
from rmscene.scene_items import Line  # noqa: E402
from rmscene.scene_items import Pen  # noqa: E402

DEVICE_W, DEVICE_H = 1404, 1872
SCREEN_DPI = 226.0
FIT_WIDTH = 1404


def parse_args(argv: list[str]) -> tuple[str, float | None, float | None]:
    pdf_w = pdf_h = None
    args = list(argv[1:])
    if args and args[0] == "--pdf-pts":
        if len(args) < 4:
            raise SystemExit("usage: --pdf-pts WIDTH_PT HEIGHT_PT <file.rm|->")
        pdf_w = float(args[1])
        pdf_h = float(args[2])
        path = args[3]
    elif len(args) >= 1:
        path = args[0]
    else:
        raise SystemExit(
            "usage: rmscene_v6_to_png.py [--pdf-pts W H] <file.rm|->"
        )
    return path, pdf_w, pdf_h


def main() -> int:
    try:
        from PIL import Image, ImageDraw, PngImagePlugin
    except ImportError as e:
        print("Pillow required: pip install Pillow", file=sys.stderr)
        print(str(e), file=sys.stderr)
        return 2

    try:
        path, pdf_w_pt, pdf_h_pt = parse_args(sys.argv)
    except SystemExit as e:
        print(str(e), file=sys.stderr)
        return 1

    if path == "-":
        data = sys.stdin.buffer.read()
    else:
        with open(path, "rb") as f:
            data = f.read()

    if pdf_w_pt and pdf_h_pt and pdf_w_pt > 0 and pdf_h_pt > 0:
        doc_w = pdf_w_pt * SCREEN_DPI / 72.0
        doc_h = pdf_h_pt * SCREEN_DPI / 72.0
        scale = FIT_WIDTH / doc_w
        out_w = FIT_WIDTH
        out_h = max(int(round(doc_h * scale)), 1)
        x0 = doc_w / 2.0
        map_pt = lambda x, y: ((x + x0) * scale, y * scale)
        view_w, view_h = out_w, out_h
        origin_x = origin_y = 0
    else:
        out_w, out_h = DEVICE_W, DEVICE_H
        scale = 1.0
        x0 = DEVICE_W / 2.0
        map_pt = lambda x, y: (x + x0, y)
        view_w, view_h = DEVICE_W, DEVICE_H
        origin_x = origin_y = 0
        doc_w, doc_h = float(DEVICE_W), float(DEVICE_H)

    tree = read_tree(BytesIO(data))

    strokes: list[tuple[list[tuple[float, float]], int]] = []
    min_x, min_y = 0.0, 0.0
    max_x, max_y = float(out_w), float(out_h)

    for item in tree.walk():
        if not isinstance(item, Line):
            continue
        if item.tool in (Pen.ERASER, Pen.ERASER_AREA):
            continue
        if len(item.points) < 2:
            continue
        sw = max(int(round((item.points[0].width * scale) / 4)), 1)
        pad = sw / 2.0 + 1.0
        pts = [map_pt(float(p.x), float(p.y)) for p in item.points]
        strokes.append((pts, sw))
        for x, y in pts:
            if x - pad < min_x:
                min_x = x - pad
            if y - pad < min_y:
                min_y = y - pad
            if x + pad > max_x:
                max_x = x + pad
            if y + pad > max_y:
                max_y = y + pad

    min_x = math.floor(min_x)
    min_y = math.floor(min_y)
    max_x = math.ceil(max_x)
    max_y = math.ceil(max_y)
    width = max(int(max_x - min_x), 1)
    height = max(int(max_y - min_y), 1)
    origin_x = int(-min_x)
    origin_y = int(-min_y)

    img = Image.new("RGB", (width, height), "white")
    draw = ImageDraw.Draw(img)
    for pts, sw in strokes:
        shifted = [(x - min_x, y - min_y) for x, y in pts]
        for i in range(len(shifted) - 1):
            draw.line([shifted[i], shifted[i + 1]], fill="black", width=sw)

    meta = PngImagePlugin.PngInfo()
    # Device/page viewport inside this (possibly expanded) PNG.
    meta.add_text("rmDeviceOrigin", f"{origin_x},{origin_y}")
    meta.add_text("rmDeviceSize", f"{view_w},{view_h}")

    buf = BytesIO()
    img.save(buf, format="PNG", pnginfo=meta)
    sys.stdout.buffer.write(buf.getvalue())
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
