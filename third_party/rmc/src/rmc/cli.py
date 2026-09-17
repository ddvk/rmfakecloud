"""CLI for converting rm files."""

import os
import sys
import io
from pathlib import Path
from contextlib import contextmanager
import click
from rmscene import read_tree, read_blocks, write_blocks, simple_text_document
from .exporters.svg import tree_to_svg
from .exporters.pdf import svg_to_pdf
from .exporters.markdown import print_text

import logging


@click.command
@click.version_option()
@click.option('-v', '--verbose', count=True)
@click.option("-f", "--from", "from_", metavar="FORMAT", help="Format to convert from (default: guess from filename)")
@click.option("-t", "--to", metavar="FORMAT", help="Format to convert to (default: guess from filename)")
@click.option("-o", "--output", type=click.Path(), help="Output filename (default: write to standard out)")
@click.argument("input", nargs=-1, type=click.Path(exists=True))
def cli(verbose, from_, to, output, input):
    """Convert to/from reMarkable v6 files.

    Available FORMATs are: `rm` (reMarkable file), `markdown`, `svg`, `pdf`,
    `blocks`, `blocks-data`.

    Formats `blocks` and `blocks-data` dump the internal structure of the `rm`
    file, with and without detailed data values respectively.

    """

    if verbose >= 2:
        logging.basicConfig(level=logging.DEBUG)
    elif verbose >= 1:
        logging.basicConfig(level=logging.INFO)
    else:
        logging.basicConfig(level=logging.WARNING)

    input = [Path(p) for p in input]
    if output is not None:
        output = Path(output)

    if from_ is None:
        if not input:
            raise click.UsageError("Must specify input filename or --from")
        from_ = guess_format(input[0])
    if to is None:
        if output is None:
            raise click.UsageError("Must specify --output or --to")
        to = guess_format(output)

    if from_ == "rm":
        with open_output(to, output) as fout:
            for fn in input:
                convert_rm(Path(fn), to, fout)
    elif from_ == "markdown":
        text = "".join(
            Path(fn).read_text() for fn in input
        )
        with open_output(to, output) as fout:
            convert_text(text, fout)
    else:
        raise click.UsageError("source format %s not implemented yet" % from_)


@contextmanager
def open_output(to, output):
    to_binary = to in ("pdf", "rm")
    if output is None:
        # Write to stdout
        if to_binary:
            with os.fdopen(sys.stdout.fileno(), "wb", closefd=False) as f:
                yield f
        else:
            yield sys.stdout
    else:
        with open(output, "w" + ("b" if to_binary else "t")) as f:
            yield f


def guess_format(p: Path):
    # XXX could be neater
    if p.suffix == ".rm":
        return "rm"
    if p.suffix == ".svg":
        return "svg"
    elif p.suffix == ".pdf":
        return "pdf"
    elif p.suffix == ".md" or p.suffix == ".markdown":
        return "markdown"
    else:
        return "blocks"


from rmscene import scene_items as si
def tree_structure(item):
    if isinstance(item, si.Group):
        return (
            item.node_id,
            (
                item.label.value,
                item.visible.value,
                (
                    item.anchor_id.value if item.anchor_id else None,
                    item.anchor_type.value if item.anchor_type else None,
                    item.anchor_threshold.value if item.anchor_threshold else None,
                    item.anchor_origin_x.value if item.anchor_origin_x else None,
                )
            ),
            [tree_structure(child) for child in item.children.values() if child],
        )
    else:
        return item


def convert_rm(filename: Path, to, fout):
    image_path = filename.with_suffix("")
    with open(filename, "rb") as f:
        if to == "blocks":
            pprint_blocks(f, fout)
        elif to == "blocks-data":
            pprint_blocks(f, fout, data=False)
        elif to == "tree":
            # Experimental dumping of tree structure
            pprint_tree(f, fout, data=True)
        elif to == "tree-data":
            # Experimental dumping of tree structure
            pprint_tree(f, fout, data=False)
        elif to == "markdown":
            print_text(f, fout)
        elif to == "svg":
            tree = read_tree(f)
            tree_to_svg(tree, fout, image_path=image_path)
        elif to == "pdf":
            buf = io.StringIO()
            tree = read_tree(f)
            tree_to_svg(tree, buf, image_path=image_path)
            buf.seek(0)
            svg_to_pdf(buf, fout)
        else:
            raise click.UsageError("Unknown format %s" % to)


def pprint_blocks(f, fout, data=True) -> None:
    import pprint
    depth = None if data else 1
    result = read_blocks(f)
    for el in result:
        print(file=fout)
        pprint.pprint(el, depth=depth, stream=fout)


def pprint_tree(f, fout, data=True) -> None:
    tree = read_tree(f)

    import pprint
    import re

    def pprint_Line(self, object, stream, indent, allowance, context, level):
        min_x = min(p.x for p in object.points)
        min_y = min(p.y for p in object.points)
        max_x = max(p.x for p in object.points)
        max_y = max(p.y for p in object.points)
        rep = re.sub(r"points=\[.*\]",
                     f"points=[({min_x: 4.0f},{min_y: 4.0f})-({max_x: 4.0f},{max_y: 4.0f})]",
                     repr(object))
        stream.write(rep)

    pprint.PrettyPrinter._dispatch[si.Line.__repr__] = pprint_Line

    depth = None if data else 1
    pprint.pprint(tree_structure(tree.root), stream=fout)
    pprint.pprint(tree_structure(tree.root_text), depth=depth, stream=fout)



def convert_text(text, fout):
    write_blocks(fout, simple_text_document(text))


if __name__ == "__main__":
    cli()
