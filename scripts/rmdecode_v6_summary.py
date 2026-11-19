#!/usr/bin/env python3
"""
Summarize a version-6 reMarkable .rm file using rmscene.

  pip install 'rmscene>=0.6'

Usage:
  python3 scripts/rmdecode_v6_summary.py /path/to/page.rm
"""

from __future__ import annotations

import json
import sys


def main() -> int:
    if len(sys.argv) < 2:
        print("usage: rmdecode_v6_summary.py <file.rm>", file=sys.stderr)
        return 1
    path = sys.argv[1]
    try:
        from rmscene import read_tree
        from rmscene.scene_items import GlyphRange, Line, Text
    except ImportError as e:
        json.dump(
            {
                "error": "rmscene not installed",
                "hint": "pip install 'rmscene>=0.6'",
                "detail": str(e),
            },
            sys.stdout,
            indent=2,
        )
        print()
        return 2

    with open(path, "rb") as f:
        tree = read_tree(f)

    n_lines = 0
    n_points = 0
    n_text = 0
    n_glyph = 0
    tools: dict[str, int] = {}
    colors: dict[str, int] = {}

    for item in tree.walk():
        if isinstance(item, Line):
            n_lines += 1
            n_points += len(item.points)
            tname = item.tool.name if hasattr(item.tool, "name") else str(item.tool)
            cname = item.color.name if hasattr(item.color, "name") else str(item.color)
            tools[tname] = tools.get(tname, 0) + 1
            colors[cname] = colors.get(cname, 0) + 1
        elif isinstance(item, Text):
            n_text += 1
        elif isinstance(item, GlyphRange):
            n_glyph += 1

    out = {
        "version": 6,
        "lines": n_lines,
        "points": n_points,
        "text_blocks": n_text,
        "glyph_ranges": n_glyph,
        "tools": tools,
        "colors": colors,
    }
    json.dump(out, sys.stdout, indent=2)
    print()
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
