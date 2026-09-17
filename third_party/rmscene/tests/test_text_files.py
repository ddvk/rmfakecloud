from uuid import UUID
from io import BytesIO
from pathlib import Path
from rmscene.scene_stream import *
from rmscene.scene_items import Text, ParagraphStyle
from rmscene.text import TextDocument, CrdtStr

DATA_PATH = Path(__file__).parent / "data"


def _hex_lines(b, n=32):
    return [b[i * n : (i + 1) * n].hex() for i in range(len(b) // n + 1)]


def extract_doc(filename):
    with open(filename, "rb") as f:
        tree = read_tree(f)
        assert tree.root_text
        doc = TextDocument.from_scene_item(tree.root_text)
        return doc


def show_str_formatting(x):
    # Basic logic -- could be smarter about removing adjacent
    # unnecessary opening/closing tags
    s = x.s
    if x.properties.get("font-weight") == "bold":
        s = f"<b>{s}</b>"
    if x.properties.get("font-style") == "italic":
        s = f"<i>{s}</i>"
    return s


def formatted_lines(doc):
    return [
        (p.style.value, "".join(show_str_formatting(s) for s in p.contents))
        for p in doc.contents
    ]


def extract_paragraphs(filename):
    with open(filename, "rb") as f:
        tree = read_tree(f)
        assert tree.root_text


def test_normal_ab():
    lines = formatted_lines(extract_doc(DATA_PATH / "Normal_AB.rm"))
    assert lines == [(ParagraphStyle.PLAIN, "AB")]


def test_list():
    lines = formatted_lines(extract_doc(DATA_PATH / "Bold_Heading_Bullet_Normal.rm"))
    assert lines == [
        (ParagraphStyle.BOLD, "A"),
        (ParagraphStyle.HEADING, "new line"),
        (ParagraphStyle.BULLET, "B is a letter of the alphabet"),
        (ParagraphStyle.PLAIN, "C"),
    ]


def test_inline_formats():
    doc = extract_doc(DATA_PATH / "Normal_A_stroke_2_layers_v3.3.2.rm")
    lines = formatted_lines(doc)
    assert lines == [
        (ParagraphStyle.PLAIN, "A"),
        (ParagraphStyle.PLAIN, "v3.2.2"),
        (
            ParagraphStyle.PLAIN,
            "Normal <b>bold</b> <i>italic</i>",
        ),
        (
            ParagraphStyle.PLAIN,
            "<b>Bold</b> <i>italic</i> normal",
        ),
        (ParagraphStyle.BOLD, "Bold line"),
        (ParagraphStyle.PLAIN, "Normal line"),
        (ParagraphStyle.HEADING, "Heading line"),
    ]


def test_crdt_concurrent_author_ordering():
    # Two authors concurrently inserted between 'A' and 'Z'; the higher author
    # ID (tablet, author 2) should appear before the lower (desktop, author 1),
    # giving "A12_Z" not "A_12Z".
    lines = formatted_lines(extract_doc(DATA_PATH / "test-crdt-ordering.rm"))
    assert lines == [(ParagraphStyle.HEADING, "A12_Z")]


def test_simple_text_document():
    test_file = "Normal_AB.rm"
    with open(DATA_PATH / test_file, "rb") as f:
        expected = f.read()

    output_buf = BytesIO()
    author_id = UUID("495ba59f-c943-2b5c-b455-3682f6948906")
    write_blocks(
        output_buf, simple_text_document("AB", author_id), options={"version": "3.0"}
    )

    assert _hex_lines(output_buf.getvalue()) == _hex_lines(expected)
