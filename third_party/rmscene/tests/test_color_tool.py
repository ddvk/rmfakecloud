import logging
import os
from pathlib import Path

import pytest

from rmscene import SceneGlyphItemBlock, SceneLineItemBlock, read_blocks
from rmscene.scene_items import Pen, PenColor

logger = logging.getLogger(__name__)


DATA_PATH = Path(__file__).parent / "data"


def _hex_lines(b, n=32):
    return [b[i * n : (i + 1) * n].hex() for i in range(len(b) // n + 1)]


FILE_NAME = os.path.join(DATA_PATH, "Color_and_tool_v3.14.4.rm")


@pytest.mark.parametrize(
    "block_type,colors,tools",
    [
        (SceneGlyphItemBlock, {PenColor.HIGHLIGHT}, None),
        (SceneLineItemBlock, {PenColor.HIGHLIGHT}, {Pen.SHADER}),
        (SceneLineItemBlock, {PenColor.GREEN_2, PenColor.CYAN, PenColor.MAGENTA}, {Pen.BALLPOINT_2}),
    ],
)
def test_color_tool_parsing(block_type, colors, tools):
    
    with open(FILE_NAME, "rb") as f:
        result = read_blocks(f)
        for el in result:
            if not getattr(el, "item", None):
                continue
            if not getattr(el.item, "value", None):
                continue
            if isinstance(el, block_type) and el.item.value.color in colors:
                if tools is None:
                    continue
                assert el.item.value.tool in tools, "Tool and colors don't match"