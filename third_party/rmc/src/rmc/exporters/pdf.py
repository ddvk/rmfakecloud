"""Convert blocks to pdf file.

Code originally from https://github.com/lschwetlick/maxio through
https://github.com/chemag/maxio .
"""

import logging
from tempfile import NamedTemporaryFile
from subprocess import check_call

from .svg import rm_to_svg

_logger = logging.getLogger(__name__)


def rm_to_pdf(rm_path, pdf_path, debug=0):
    """Convert `rm_path` to PDF at `pdf_path`."""
    with NamedTemporaryFile(suffix=".svg") as f_temp:
        rm_to_svg(rm_path, f_temp.name)

        # use inkscape to convert svg to pdf
        check_call(["inkscape", f_temp.name, "--export-filename", pdf_path])


def svg_to_pdf(svg_file, pdf_file):
    """Read svg data from `svg_file` and write PDF data to `pdf_file`."""

    with NamedTemporaryFile("w", suffix=".svg") as fsvg, NamedTemporaryFile("rb", suffix=".pdf") as fpdf:
        fsvg.write(svg_file.read())
        fsvg.flush() # Make sure content is writen to the file
        
        # use inkscape to convert svg to pdf
        try:
            print("Convert SVG to PDF using Inkscape")
            check_call(["inkscape", fsvg.name, "--export-filename", fpdf.name])
        except FileNotFoundError:
            print("Inkscape not found in path")

            try:
                print("Convert SVG to PDF using Inkscape (default MacOS path)")
                check_call(["/Applications/Inkscape.app/Contents/MacOS/inkscape", fsvg.name, "--export-filename", fpdf.name])
            except FileNotFoundError:
                pass

        pdf_file.write(fpdf.read())
        pdf_file.flush()
