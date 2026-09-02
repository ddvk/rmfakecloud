#!/usr/bin/env python3
"""Render a reMarkable v6 notebook to PDF using rmc + pypdf."""
import argparse
import os
import subprocess
import sys
import tempfile

from pypdf import PdfWriter, PdfReader


def default_rmc():
    rmc = os.environ.get("RMC_BIN")
    if rmc:
        return rmc
    # rmc lives next to the venv python that runs this script
    return os.path.join(os.path.dirname(os.path.abspath(sys.executable)), "rmc")


def convert_page(rm_path, pdf_path, rmc_bin):
    cmd = [rmc_bin, rm_path, "-o", pdf_path]
    res = subprocess.run(cmd, capture_output=True, text=True)
    if res.returncode != 0:
        raise RuntimeError(f"rmc failed for {rm_path}: {res.stderr}")


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
            out_pdf = os.path.join(tmpdir, f"{pid}.pdf")
            convert_page(rm_path, out_pdf, args.rmc)
            writer.append(PdfReader(out_pdf))

    with open(args.output, "wb") as f:
        writer.write(f)


if __name__ == "__main__":
    main()
