"""Data structures for the contents of a scene."""

import enum
import logging
import typing as tp
from dataclasses import dataclass, field

from .crdt_sequence import CrdtSequence
from .tagged_block_common import CrdtId, LwwValue
from .text import expand_text_items

_logger = logging.getLogger(__name__)


## Base class


@dataclass
class SceneItem:
    """Base class for items stored in scene tree."""


## Image


@dataclass
class ImageInfo(SceneItem):
    filename: LwwValue[str]
    flags: LwwValue[bytes]


@dataclass
class Image(SceneItem):
    uuid: LwwValue[bytes]
    vertices: list[float]
    move_id: tp.Optional[CrdtId] = None
    filename: tp.Optional[str] = None


## Group


@dataclass
class Group(SceneItem):
    """A Group represents a group of nested items.

    Groups are used to represent layers.

    node_id is the id that this sub-tree is stored as a "SceneTreeBlock".

    children is a sequence of other SceneItems.

    `anchor_id` refers to a text character which provides the anchor y-position
    for this group. There are two values that seem to be special:
    - `0xfffffffffffe` seems to be used for lines right at the top of the page?
    - `0xffffffffffff` seems to be used for lines right at the bottom of the page?

    """

    node_id: CrdtId
    children: CrdtSequence[SceneItem] = field(default_factory=CrdtSequence)
    label: LwwValue[str] = LwwValue(CrdtId(0, 0), "")
    visible: LwwValue[bool] = LwwValue(CrdtId(0, 0), True)

    anchor_id: tp.Optional[LwwValue[CrdtId]] = None
    anchor_type: tp.Optional[LwwValue[int]] = None
    anchor_threshold: tp.Optional[LwwValue[float]] = None
    anchor_origin_x: tp.Optional[LwwValue[float]] = None


## Strokes


@enum.unique
class PenColor(enum.IntEnum):
    """
    Color index value.
    """

    # XXX list from remt pre-v6

    BLACK = 0
    GRAY = 1
    WHITE = 2

    YELLOW = 3
    GREEN = 4
    PINK = 5

    BLUE = 6
    RED = 7

    GRAY_OVERLAP = 8

    # All highlight colors share the same value.
    # The actual color is stored in the optional `color_rgba` field of Line.
    HIGHLIGHT = 9

    GREEN_2 = 10
    CYAN = 11
    MAGENTA = 12
    
    YELLOW_2 = 13


@enum.unique
class Pen(enum.IntEnum):
    """
    Stroke pen id representing reMarkable tablet tools.

    Tool examples: ballpoint, fineliner, highlighter or eraser.
    """

    # XXX this list is from remt pre-v6

    BALLPOINT_1 = 2
    BALLPOINT_2 = 15
    CALIGRAPHY = 21
    ERASER = 6
    ERASER_AREA = 8
    FINELINER_1 = 4
    FINELINER_2 = 17
    HIGHLIGHTER_1 = 5
    HIGHLIGHTER_2 = 18
    MARKER_1 = 3
    MARKER_2 = 16
    MECHANICAL_PENCIL_1 = 7
    MECHANICAL_PENCIL_2 = 13
    PAINTBRUSH_1 = 0
    PAINTBRUSH_2 = 12
    PENCIL_1 = 1
    PENCIL_2 = 14
    SHADER = 23

    @classmethod
    def is_highlighter(cls, value: int) -> bool:
        return value in (cls.HIGHLIGHTER_1, cls.HIGHLIGHTER_2)


@dataclass
class Point:
    x: float
    y: float
    speed: int
    direction: int
    width: int
    pressure: int


@dataclass
class Line(SceneItem):
    color: PenColor
    tool: Pen
    points: list[Point]
    thickness_scale: float
    starting_length: float
    move_id: tp.Optional[CrdtId] = None
    color_rgba: tp.Optional[tuple[int, int, int, int]] = None


## Text


@enum.unique
class ParagraphStyle(enum.IntEnum):
    """
    Text paragraph style.
    """

    BASIC = 0
    PLAIN = 1
    HEADING = 2
    BOLD = 3
    BULLET = 4
    BULLET2 = 5
    CHECKBOX = 6
    CHECKBOX_CHECKED = 7


END_MARKER = CrdtId(0, 0)


@dataclass
class Text(SceneItem):
    """Block of text.

    `items` are a CRDT sequence of strings. The `item_id` for each string refers
    to its first character; subsequent characters implicitly have sequential
    ids.

    When formatting is present, some of `items` have a value of an integer
    formatting code instead of a string.

    `styles` are LWW values representing a mapping of character IDs to
    `ParagraphStyle` values. These formats apply to each line of text (until the
    next newline).

    `pos_x`, `pos_y` and `width` are dimensions for the text block.

    """

    items: CrdtSequence[str | int]
    styles: dict[CrdtId, LwwValue[ParagraphStyle]]
    pos_x: float
    pos_y: float
    width: float


## Glyph range


@dataclass
class Rectangle:
    x: float
    y: float
    w: float
    h: float


@dataclass
class GlyphRange(SceneItem):
    """Highlighted text

    `start` is only available in SceneGlyphItemBlock version=0, prior to ReMarkable v3.6

    `length` is the length of the text

    `text` is the highlighted text itself

    `color` represents the highlight color

    `rectangles` represent the locations of the highlight.
    """
    start: tp.Optional[int]
    length: int
    text: str
    color: PenColor
    rectangles: list[Rectangle]
    color_rgba: tp.Optional[tuple[int, int, int, int]] = None
