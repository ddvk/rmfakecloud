#!/usr/bin/env python3
"""
Render a single v6 .rm page to PNG (1404×1872) using rmscene + Pillow.

Reads from a file path, or from stdin when the argument is '-'.

  pip install Pillow   # if not already installed

Usage:
  python3 scripts/rmscene_v6_to_png.py page.rm > page.png
  python3 scripts/rmscene_v6_to_png.py - < page.rm > page.png
"""

from __future__ import annotations

import pathlib
import sys
from io import BytesIO

_REPO_ROOT = pathlib.Path(__file__).resolve().parents[1]
sys.path.insert(0, str(_REPO_ROOT / "third_party" / "rmscene" / "src"))

from rmscene import read_tree  # noqa: E402
from rmscene.scene_items import Line  # noqa: E402
from rmscene.scene_items import Pen  # noqa: E402

W, H = 1404, 1872


def main() -> int:
    try:
        from PIL import Image, ImageDraw
    except ImportError as e:
        print("Pillow required: pip install Pillow", file=sys.stderr)
        print(str(e), file=sys.stderr)
        return 2

    if len(sys.argv) < 2:
        print("usage: rmscene_v6_to_png.py <file.rm|- >", file=sys.stderr)
        return 1
    path = sys.argv[1]
    if path == "-":
        data = sys.stdin.buffer.read()
    else:
        with open(path, "rb") as f:
            data = f.read()

    tree = read_tree(BytesIO(data))
    img = Image.new("RGB", (W, H), "white")
    draw = ImageDraw.Draw(img)

    for item in tree.walk():
        if not isinstance(item, Line):
            continue
        if item.tool in (Pen.ERASER, Pen.ERASER_AREA):
            continue
        if len(item.points) < 2:
            continue
        sw = max(int(item.points[0].width) // 4, 1)
        pts = [(float(p.x), float(p.y)) for p in item.points]
        for i in range(len(pts) - 1):
            draw.line([pts[i], pts[i + 1]], fill="black", width=sw)

    buf = BytesIO()
    img.save(buf, format="PNG")
    sys.stdout.buffer.write(buf.getvalue())
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
