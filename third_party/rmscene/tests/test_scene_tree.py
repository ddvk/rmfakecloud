import pytest
from io import BytesIO
from pathlib import Path
from uuid import UUID
from rmscene import (
    read_blocks,
    write_blocks,
    LwwValue,
    TaggedBlockWriter,
    TaggedBlockReader,
)
from rmscene.scene_tree import *
from rmscene.scene_stream import *
from rmscene import scene_items as si
from rmscene.crdt_sequence import CrdtSequenceItem

import logging

logger = logging.getLogger(__name__)


DATA_PATH = Path(__file__).parent / "data"


# @pytest.mark.parametrize(
#     "test_file",
#     [
#         "Normal_AB.rm",
#         "Normal_A_stroke_2_layers.rm",
#         "Bold_Heading_Bullet_Normal.rm",
#         "Lines_v2.rm"
#     ]
# )
# def test_full_roundtrip(test_file):
#     with open(DATA_PATH / test_file, "rb") as f:
#         data = f.read()

#     input_buf = BytesIO(data)
#     output_buf = BytesIO()
#     options = {
#         "line_version": (2 if test_file == "Lines_v2.rm" else 1)
#     }

#     write_blocks(output_buf, read_blocks(input_buf), options)

#     assert _hex_lines(input_buf.getvalue()) == _hex_lines(output_buf.getvalue())


def tree_structure(item):
    if isinstance(item, si.Group):
        return (
            item.node_id,
            item.label.value,
            [tree_structure(child) for child in item.children.values()],
        )
    else:
        return item


def test_basic_tree_structure():
    # this is the bare minimum structure
    blocks = [
        SceneTreeBlock(
            tree_id=CrdtId(0, 11),
            node_id=CrdtId(0, 0),
            is_update=True,
            parent_id=CrdtId(0, 1),
        ),
        TreeNodeBlock(
            si.Group(CrdtId(0, 1)),
        ),
        TreeNodeBlock(
            si.Group(
                node_id=CrdtId(0, 11),
                label=LwwValue(CrdtId(0, 12), "Layer 1"),
            )
        ),
        SceneGroupItemBlock(
            parent_id=CrdtId(0, 1),
            item=CrdtSequenceItem(
                item_id=CrdtId(0, 13),
                left_id=CrdtId(0, 0),
                right_id=CrdtId(0, 0),
                deleted_length=0,
                value=CrdtId(0, 11),
            ),
        ),
    ]

    tree = SceneTree()
    build_tree(tree, blocks)

    assert tree_structure(tree.root) == (
        CrdtId(0, 1),
        "",
        [(CrdtId(0, 11), "Layer 1", [])],
    )

    assert list(tree.root.children.values()) == [
        si.Group(
            node_id=CrdtId(0, 11), children=[], label=LwwValue(CrdtId(0, 12), "Layer 1")
        )
    ]


# Example file layers.stroke.rm
#
# SceneTreeBlocks -- all have parent (0, 1)
#
# (0, 11) update=True
# (1, 27) update=True
# (1, 30) update=True
# (1, 40) update=False --> "node_id" (1, 27), parent (0, 0)
# (1, 41) update=True
#
# In ddvk's reader this is described as a "node move" info?
#
# TreeNodeBlock
#
# (0, 1)
# (0, 11) Layer 1
# (1, 27) Layer 2
# (1, 30) Layer 3
# (1, 41) Layer 4
#
# These don't have any structure (parents etc), just information -- they are
# giving data about a node in the tree? ddvk's reader says it's an error if the
# node here has not already been seen, and if it's not marked as a layer (i.e.
# its parent is the root node). It doesn't add anything new to the tree, just
# attaches the info.
#
# All of the following are "items", so are expected to be added with a parent
# which is an already defined node. Items have left/right IDs, so are ordered,
# and leave tombstones.
#
# SceneGroupItemBlock
#
# Parent: (0, 1)
#         (0, 13)       -> (0, 11)
# (0, 13) (1, 29)       -> (1, 27)
# (0, 29) (1, 32)       -> (1, 30)
# (0, 32) (1, 43)       -> (1, 41)
#
# The values refer to the "tree blocks" above, with new IDs -- like the Groups
# have their own identity and order (amongst what?), but are referring to nodes
# (layers).
#
# Is this because this defines the order of the layers? Otherwise it's not yet
# specified. So a "GroupItem" defines the order of further nested nodes.
#
# Line, parent: (0, 11)
#
#         (1, 14)       -> [line data]
# (1, 14) (1, 15)       -> DELETED
# ...
# (1, 19) (1, 20)       -> DELETED
# (1, 20) (1, 21)       -> [line data]
# ...
# (1, 25) (1, 26)       -> [line data]      at this point, the next layer must have been added
# (1, 26) (1, 46)       -> [line data]
# ...
# (1, 51) (1, 52)       -> [line data]
#
# Line, parent: (1, 30)
#
#         (1, 33)       -> [line data]
# (1, 33) (1, 34)       -> [line data]
# ...
# (1, 38) (1, 39)       -> [line data]
#
#
# So the actual line data has a parent which is a layer. (0, 11) or (1, 30)


def test_text_and_strokes():
    line1 = si.Line(
        color=si.PenColor.RED,
        tool=si.Pen.PENCIL_2,
        points=[],
        thickness_scale=2.0,
        starting_length=0.0,
    )
    line2 = si.Line(
        color=si.PenColor.BLACK,
        tool=si.Pen.FINELINER_2,
        points=[],
        thickness_scale=2.0,
        starting_length=0.0,
    )
    blocks = [
        SceneTreeBlock(
            tree_id=CrdtId(0, 13),
            node_id=CrdtId(0, 0),
            is_update=True,
            parent_id=CrdtId(0, 1),
        ),
        SceneTreeBlock(
            tree_id=CrdtId(1, 17),
            node_id=CrdtId(0, 0),
            is_update=True,
            parent_id=CrdtId(0, 1),
        ),
        SceneTreeBlock(
            tree_id=CrdtId(1, 20),
            node_id=CrdtId(0, 0),
            is_update=True,
            parent_id=CrdtId(0, 13),
        ),
        SceneTreeBlock(
            tree_id=CrdtId(1, 26),
            node_id=CrdtId(0, 0),
            is_update=True,
            parent_id=CrdtId(1, 17),
        ),
        RootTextBlock(
            block_id=CrdtId(0, 0),
            value=si.Text(
                items=CrdtSequence([
                    CrdtSequenceItem(
                        item_id=CrdtId(1, 14),
                        left_id=CrdtId(0, 0),
                        right_id=CrdtId(0, 0),
                        deleted_length=0,
                        value="A",
                    )
                ]),
                styles={},
                pos_x=-468.0,
                pos_y=234.0,
                width=936.0,
            )
        ),
        TreeNodeBlock(
            si.Group(node_id=CrdtId(0, 1)),
        ),
        TreeNodeBlock(
            si.Group(
                node_id=CrdtId(0, 13),
                label=LwwValue(timestamp=CrdtId(0, 15), value="Layer 1"),
                visible=LwwValue(timestamp=CrdtId(0, 16), value=True),
            )
        ),
        TreeNodeBlock(
            si.Group(
                node_id=CrdtId(1, 17),
                label=LwwValue(timestamp=CrdtId(1, 18), value="Layer 2"),
                visible=LwwValue(timestamp=CrdtId(0, 0), value=True),
            )
        ),
        TreeNodeBlock(
            si.Group(
                node_id=CrdtId(1, 20),
                label=LwwValue(timestamp=CrdtId(0, 0), value=""),
                visible=LwwValue(timestamp=CrdtId(0, 0), value=True),
                anchor_id=LwwValue(timestamp=CrdtId(1, 22), value=CrdtId(1, 14)),
                anchor_type=LwwValue(timestamp=CrdtId(1, 23), value=2),
                anchor_threshold=LwwValue(timestamp=CrdtId(1, 24), value=67.02755737304688),
                anchor_origin_x=LwwValue(timestamp=CrdtId(1, 20), value=-464.0),
            )
        ),
        TreeNodeBlock(
            si.Group(
                node_id=CrdtId(1, 26),
                label=LwwValue(timestamp=CrdtId(0, 0), value=""),
                visible=LwwValue(timestamp=CrdtId(0, 0), value=True),
                anchor_id=LwwValue(timestamp=CrdtId(1, 28), value=CrdtId(1, 14)),
                anchor_type=LwwValue(timestamp=CrdtId(1, 29), value=2),
                anchor_threshold=LwwValue(timestamp=CrdtId(1, 30), value=67.02755737304688),
                anchor_origin_x=LwwValue(timestamp=CrdtId(1, 26), value=-464.0),
            )
        ),
        SceneGroupItemBlock(
            parent_id=CrdtId(0, 1),
            item=CrdtSequenceItem(
                item_id=CrdtId(0, 14),
                left_id=CrdtId(0, 0),
                right_id=CrdtId(0, 0),
                deleted_length=0,
                value=CrdtId(0, 13),
            ),
        ),
        SceneGroupItemBlock(
            parent_id=CrdtId(0, 1),
            item=CrdtSequenceItem(
                item_id=CrdtId(1, 19),
                left_id=CrdtId(0, 14),
                right_id=CrdtId(0, 0),
                deleted_length=0,
                value=CrdtId(1, 17),
            ),
        ),
        SceneGroupItemBlock(
            parent_id=CrdtId(0, 13),
            item=CrdtSequenceItem(
                item_id=CrdtId(1, 21),
                left_id=CrdtId(0, 0),
                right_id=CrdtId(0, 0),
                deleted_length=0,
                value=CrdtId(1, 20),
            ),
        ),
        SceneGroupItemBlock(
            parent_id=CrdtId(1, 17),
            item=CrdtSequenceItem(
                item_id=CrdtId(1, 27),
                left_id=CrdtId(0, 0),
                right_id=CrdtId(0, 0),
                deleted_length=0,
                value=CrdtId(1, 26),
            ),
        ),
        SceneLineItemBlock(
            parent_id=CrdtId(1, 20),
            item=CrdtSequenceItem(
                item_id=CrdtId(1, 25),
                left_id=CrdtId(0, 0),
                right_id=CrdtId(0, 0),
                deleted_length=0,
                value=line1,
            ),
        ),
        SceneLineItemBlock(
            parent_id=CrdtId(1, 26),
            item=CrdtSequenceItem(
                item_id=CrdtId(1, 31),
                left_id=CrdtId(0, 0),
                right_id=CrdtId(0, 0),
                deleted_length=0,
                value=line2,
            ),
        ),
    ]

    tree = SceneTree()
    build_tree(tree, blocks)

    assert tree_structure(tree.root) == (
        CrdtId(0, 1),
        "",
        [
            (
                CrdtId(0, 13),
                "Layer 1",
                [
                    (CrdtId(1, 20), "", [line1]),
                ],
            ),
            (
                CrdtId(1, 17),
                "Layer 2",
                [
                    (CrdtId(1, 26), "", [line2]),
                ],
            ),
        ],
    )

    # TODO should also include the text items, and anchor reference
    #
    # Does this mean the tree should also be able to look up items by ID? And
    # maybe actually children list should just contain item ids?
    #
    # Otherwise how do you make use of the "anchor id" value? Could set the
    # explicit reference.
    #
    # When writing back to blocks, need to be able to flatten the structure
    # again (i.e. find out the item id for the anchor, if referenced directly).
    #
    # Currently the item id is held in the CrdtSequenceItem, not in its value.


def test_scene_info_paper_size():
    with open(DATA_PATH / "Color_and_tool_v3.14.4.rm", "rb") as f:
        tree = read_tree(f)
    assert tree.scene_info is not None
    assert tree.scene_info.paper_size == (1620, 2160)
