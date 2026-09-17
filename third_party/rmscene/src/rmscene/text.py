"""Process text from remarkable scene files.

"""

from __future__ import annotations

from collections.abc import Iterable
from collections import defaultdict
from dataclasses import dataclass, field
import logging
import typing as tp

from . import scene_items as si
from .tagged_block_common import CrdtId, LwwValue
from .crdt_sequence import CrdtSequence, CrdtSequenceItem

_logger = logging.getLogger(__name__)


def expand_text_item(
    item: CrdtSequenceItem[str | int],
) -> tp.Iterator[CrdtSequenceItem[str | int]]:
    """Expand TextItem into single-character TextItems.

    Text is stored as strings in TextItems, each with an associated ID for the
    block. This ID identifies the character at the start of the block. The
    subsequent characters' IDs are implicit.

    This function expands a TextItem into multiple single-character TextItems,
    so that each character has an explicit ID.

    """

    if item.deleted_length > 0:
        assert item.value == ""
        chars = [""] * item.deleted_length
        deleted_length = 1
    elif isinstance(item.value, int):
        yield item
        return
    else:
        # Actually the value can be empty
        # assert len(item.value) > 0
        chars = item.value
        deleted_length = 0

    if not chars:
        _logger.warning("Unexpected empty text item: %s", item)
        return

    item_id = item.item_id
    left_id = item.left_id
    for c in chars[:-1]:
        right_id = CrdtId(item_id.part1, item_id.part2 + 1)
        yield CrdtSequenceItem(item_id, left_id, right_id, deleted_length, c)
        left_id = item_id
        item_id = right_id
    yield CrdtSequenceItem(item_id, left_id, item.right_id, deleted_length, chars[-1])


def expand_text_items(
    items: Iterable[CrdtSequenceItem[str | int]],
) -> tp.Iterator[CrdtSequenceItem[str | int]]:
    """Expand a sequence of TextItems into single-character TextItems."""
    for item in items:
        yield from expand_text_item(item)


@dataclass
class CrdtStr:
    """String with CrdtIds for chars and optional properties.

    The properties apply to the whole `CrdtStr`. Use a list of
    `CrdtStr`s to represent a sequence of spans of text with different
    properties.

    """

    s: str = ""
    i: list[CrdtId] = field(default_factory=list)
    properties: dict = field(default_factory=dict)

    def __str__(self):
        return self.s


@dataclass
class Paragraph:
    """Paragraph of text."""

    contents: list[CrdtStr]
    start_id: CrdtId
    style: LwwValue[si.ParagraphStyle] = field(
        default_factory=lambda: LwwValue(CrdtId(0, 0), si.ParagraphStyle.PLAIN)
    )

    def __str__(self):
        return "".join(str(s) for s in self.contents)


@dataclass
class TextDocument:
    contents: list[Paragraph]

    @classmethod
    def from_scene_item(cls, text: si.Text):
        """Extract spans of text with associated formatting and char ids.

        This uses the inline formatting introduced in v3.3.2.
        """

        char_formats = {k: lww.value for k, lww in text.styles.items()}
        if si.END_MARKER not in char_formats:
            char_formats[si.END_MARKER] = si.ParagraphStyle.PLAIN

        # Expand from strings to characters
        char_items = CrdtSequence(expand_text_items(text.items.sequence_items()))
        keys = list(char_items)
        properties = {"font-weight": "normal", "font-style": "normal"}

        def handle_formatting_code(code):
            if code == 1:
                properties["font-weight"] = "bold"
            elif code == 2:
                properties["font-weight"] = "normal"
            if code == 3:
                properties["font-style"] = "italic"
            elif code == 4:
                properties["font-style"] = "normal"
            else:
                _logger.warning("Unknown formatting code in text: %d", code)
            return properties

        def parse_paragraph_contents():
            if keys and char_items[keys[0]] == "\n":
                start_id = keys.pop(0)
            else:
                start_id = si.END_MARKER
            contents = []
            while keys:
                char = char_items[keys[0]]
                if isinstance(char, int):
                    handle_formatting_code(char)
                elif char == "\n":
                    # End of paragraph
                    break
                else:
                    assert len(char) <= 1
                    # Start a new string if text properties have changed
                    if not contents or contents[-1].properties != properties:
                        contents += [CrdtStr(properties=properties.copy())]
                    contents[-1].s += char
                    contents[-1].i += [keys[0]]
                keys.pop(0)

            return start_id, contents

        paragraphs = []
        while keys:
            start_id, contents = parse_paragraph_contents()
            if start_id in text.styles:
                p = Paragraph(contents, start_id, text.styles[start_id])
            else:
                p = Paragraph(contents, start_id)
            paragraphs += [p]

        doc = cls(paragraphs)
        return doc
