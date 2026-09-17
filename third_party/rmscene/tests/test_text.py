import pytest
from rmscene.text import (
    expand_text_item,
    expand_text_items,
    TextDocument,
    CrdtStr,
    Paragraph,
)
from rmscene import scene_items as si
from rmscene import CrdtId, CrdtSequenceItem, CrdtSequence


def cid(k):
    "Shorthand for making end-markers or IDs with author-id = 1."
    return CrdtId(0 if k == 0 else 1, k)


def make_item(item_id, left_id, right_id, *args):
    "Shorthand for creating ids."
    return CrdtSequenceItem(cid(item_id), cid(left_id), cid(right_id), *args)


def test_expand_text_1():
    result = expand_text_item(make_item(17, 0, 0, 0, "AAAA"))
    assert list(result) == [
        make_item(17, 0, 18, 0, "A"),
        make_item(18, 17, 19, 0, "A"),
        make_item(19, 18, 20, 0, "A"),
        make_item(20, 19, 0, 0, "A"),
    ]


def test_expand_text_2():
    result = expand_text_item(make_item(34, 20, 21, 0, "x"))
    assert list(result) == [
        make_item(34, 20, 21, 0, "x"),
    ]


def test_expand_text_3():
    result = expand_text_item(make_item(21, 20, 0, 0, "A\nB"))
    assert list(result) == [
        make_item(21, 20, 22, 0, "A"),
        make_item(22, 21, 23, 0, "\n"),
        make_item(23, 22, 0, 0, "B"),
    ]


def test_expand_text_empty():
    result = expand_text_item(make_item(21, 20, 0, 2, ""))
    assert list(result) == [
        make_item(21, 20, 22, 1, ""),
        make_item(22, 21, 0, 1, ""),
    ]


START_BOLD = 1
END_BOLD = 2
START_ITALIC = 3
END_ITALIC = 4


def doc_from_items(items):
    root_text = si.Text(
        items=CrdtSequence(items),
        styles={},
        pos_x=-468.0,
        pos_y=234.0,
        width=936.0,
    )
    doc = TextDocument.from_scene_item(root_text)
    return doc


def test_inline_formatting_italic_over_paragraphs():
    doc = doc_from_items(
        [
            make_item(20, 0, 0, 0, "A"),
            make_item(21, 20, 0, 0, "B\nC"),
            make_item(24, 23, 0, 0, "D"),
            # Start italic between A and B
            make_item(30, 20, 21, 0, START_ITALIC),
            # End italic between C and D
            make_item(31, 23, 24, 0, END_ITALIC),
        ]
    )

    assert doc.contents == [
        Paragraph(
            contents=[
                CrdtStr(
                    "A",
                    [CrdtId(1, 20)],
                    {"font-weight": "normal", "font-style": "normal"},
                ),
                CrdtStr(
                    "B",
                    [CrdtId(1, 21)],
                    {"font-weight": "normal", "font-style": "italic"},
                ),
            ],
            start_id=CrdtId(0, 0),
        ),
        Paragraph(
            contents=[
                CrdtStr(
                    "C",
                    [CrdtId(1, 23)],
                    {"font-weight": "normal", "font-style": "italic"},
                ),
                CrdtStr(
                    "D",
                    [CrdtId(1, 24)],
                    {"font-weight": "normal", "font-style": "normal"},
                ),
            ],
            start_id=CrdtId(1, 22),
        ),
    ]


def test_inline_formatting_italic_over_paragraphs():
    doc = doc_from_items(
        [
            make_item(20, 0, 0, 0, "A"),
            make_item(21, 20, 0, 0, "B\nC"),
            make_item(24, 23, 0, 0, "D"),
            # Start italic between A and B
            make_item(30, 20, 21, 0, START_ITALIC),
            # End italic between C and D
            make_item(31, 23, 24, 0, END_ITALIC),
        ]
    )

    assert doc.contents == [
        Paragraph(
            contents=[
                CrdtStr(
                    "A",
                    [CrdtId(1, 20)],
                    {"font-weight": "normal", "font-style": "normal"},
                ),
                CrdtStr(
                    "B",
                    [CrdtId(1, 21)],
                    {"font-weight": "normal", "font-style": "italic"},
                ),
            ],
            start_id=CrdtId(0, 0),
        ),
        Paragraph(
            contents=[
                CrdtStr(
                    "C",
                    [CrdtId(1, 23)],
                    {"font-weight": "normal", "font-style": "italic"},
                ),
                CrdtStr(
                    "D",
                    [CrdtId(1, 24)],
                    {"font-weight": "normal", "font-style": "normal"},
                ),
            ],
            start_id=CrdtId(1, 22),
        ),
    ]


def test_inline_formatting_bold_italic_interleaved_over_paragraphs():
    doc = doc_from_items(
        [
            make_item(20, 0, 0, 0, "ABC\nDEF"),
            # Start italic between A and B
            make_item(30, 20, 21, 0, START_ITALIC),
            # Start bold between B and C
            make_item(31, 21, 22, 0, START_BOLD),
            # End italic between D and E
            make_item(32, 24, 25, 0, END_ITALIC),
            # End bold between E and F
            make_item(33, 25, 26, 0, END_BOLD),
        ]
    )

    assert doc.contents == [
        Paragraph(
            contents=[
                CrdtStr(
                    "A",
                    [CrdtId(1, 20)],
                    {"font-weight": "normal", "font-style": "normal"},
                ),
                CrdtStr(
                    "B",
                    [CrdtId(1, 21)],
                    {"font-weight": "normal", "font-style": "italic"},
                ),
                CrdtStr(
                    "C",
                    [CrdtId(1, 22)],
                    {"font-weight": "bold", "font-style": "italic"},
                ),
            ],
            start_id=CrdtId(0, 0),
        ),
        Paragraph(
            contents=[
                CrdtStr(
                    "D",
                    [CrdtId(1, 24)],
                    {"font-weight": "bold", "font-style": "italic"},
                ),
                CrdtStr(
                    "E",
                    [CrdtId(1, 25)],
                    {"font-weight": "bold", "font-style": "normal"},
                ),
                CrdtStr(
                    "F",
                    [CrdtId(1, 26)],
                    {"font-weight": "normal", "font-style": "normal"},
                ),
            ],
            start_id=CrdtId(1, 23),
        ),
    ]
