#!/usr/bin/env python3
"""Render reMarkable v6 notebook pages to a single PDF.

Pipeline per page: rmc (.rm -> SVG), then svglib+reportlab (SVG -> PDF).
Pages are merged in the given order with pypdf. No Inkscape required.
"""
import argparse
import os
import subprocess
import sys
import tempfile

from pypdf import PdfReader, PdfWriter
from svglib.svglib import svg2rlg
from reportlab.graphics import renderPDF


def default_rmc():
    rmc = os.environ.get("RMC_BIN")
    if rmc:
        return rmc
    # rmc lives next to the venv python that runs this script
    return os.path.join(os.path.dirname(os.path.abspath(sys.executable)), "rmc")


def rm_to_svg(rm_path, svg_path, rmc_bin):
    res = subprocess.run(
        [rmc_bin, rm_path, "-t", "svg", "-o", svg_path],
        capture_output=True, text=True,
    )
    if res.returncode != 0 or not os.path.exists(svg_path):
        raise RuntimeError(f"rmc failed for {rm_path}: {res.stderr}")


def svg_to_pdf(svg_path, pdf_path):
    drawing = svg2rlg(svg_path)
    if drawing is None:
        raise RuntimeError(f"svglib could not parse {svg_path}")
    renderPDF.drawToFile(drawing, pdf_path)


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--rmc", default=default_rmc(), help="path to rmc binary")
    parser.add_argument("--output", required=True)
    parser.add_argument("--page", action="append", required=True, help="page_id:rm_file")
    args = parser.parse_args()

    pages = []
    for p in args.page:
        pid, path = p.split(":", 1)
        pages.append((pid, path))

    writer = PdfWriter()
    with tempfile.TemporaryDirectory() as tmpdir:
        for pid, rm_path in pages:
            svg_path = os.path.join(tmpdir, f"{pid}.svg")
            pdf_path = os.path.join(tmpdir, f"{pid}.pdf")
            rm_to_svg(rm_path, svg_path, args.rmc)
            svg_to_pdf(svg_path, pdf_path)
            writer.append(PdfReader(pdf_path))

    with open(args.output, "wb") as f:
        writer.write(f)


if __name__ == "__main__":
    main()
