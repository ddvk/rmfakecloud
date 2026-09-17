"""Export text content of rm files as Markdown."""

from rmscene import read_tree
from rmscene import scene_items as si

from rmscene.text import TextDocument


def print_text(f, fout):
    tree = read_tree(f)

    # Find out what anchor characters are used
    anchor_ids = set(collect_anchor_ids(tree.root))

    if tree.root_text:
        print_root_text(tree.root_text, fout, anchor_ids)

    JOIN_TOLERANCE = 2
    print("\n\n# Highlights", file=fout)
    last_pos = 0
    for item in tree.walk():
        if isinstance(item, si.GlyphRange):
            if item.start > last_pos + JOIN_TOLERANCE:
                print(file=fout)
            print(">", item.text, file=fout)
            last_pos = item.start + len(item.text)
    print(file=fout)


def print_root_text(root_text: si.Text, fout, anchor_ids):
    doc = TextDocument.from_scene_item(root_text)
    for p in doc.contents:
        annotated_line = annotate_anchor_ids(anchor_ids,
                                             str(p),
                                             [char_id for s in p.contents for char_id in s.i])
        if p.style.value == si.ParagraphStyle.BULLET:
            fout.write("- " + annotated_line + "\n")
        elif p.style.value == si.ParagraphStyle.BULLET2:
            fout.write("  + " + annotated_line + "\n")
        elif p.style.value == si.ParagraphStyle.BOLD:
            fout.write("> " + annotated_line + "\n")
        elif p.style.value == si.ParagraphStyle.HEADING:
            fout.write("# " + annotated_line + "\n")
        elif p.style.value == si.ParagraphStyle.PLAIN:
            fout.write(annotated_line + "\n")
        else:
            fout.write(("[unknown format %s] " % p.style.value) + annotated_line + "\n")


def annotate_anchor_ids(anchor_ids, line, ids):
    """Annotate appearances of `anchor_ids` in `line`."""
    result = ""
    for char, char_id in zip(line, ids):
        if char_id in anchor_ids:
            result += f"<<{char_id.part1},{char_id.part2}>>"
        result += char
    return result


def collect_anchor_ids(item):
    if isinstance(item, si.Group):
        if item.anchor_id is not None:
            yield item.anchor_id.value
        for child in item.children.values():
            yield from collect_anchor_ids(child)
