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
from rmscene.scene_stream import *
from rmscene.tagged_block_common import HEADER_V6
from rmscene.tagged_block_reader import MainBlockInfo
from rmscene.crdt_sequence import CrdtSequenceItem
from rmscene.scene_items import ParagraphStyle

import logging

logger = logging.getLogger(__name__)


DATA_PATH = Path(__file__).parent / "data"


def _hex_lines(b, n=32):
    return [b[i * n : (i + 1) * n].hex() for i in range(len(b) // n + 1)]


LINES_V2_FILES = [
    "Lines_v2.rm",
    "Wikipedia_highlighted_p2.rm",
]


TEST_FILES_AND_VERSIONS = [
    ("Normal_AB.rm", "3.0"),
    ("Normal_A_stroke_2_layers.rm", "3.0"),
    ("Normal_A_stroke_2_layers_v3.2.2.rm", "3.2.2"),
    ("Normal_A_stroke_2_layers_v3.3.2.rm", "3.3.2"),
    ("Bold_Heading_Bullet_Normal.rm", "3.0"),
    ("Lines_v2.rm", "3.1"),
    ("Lines_v2_updated.rm", "3.2"),  # extra 7fXXXX part of Line data was added
    ("Wikipedia_highlighted_p1.rm", "3.1"),
    ("Wikipedia_highlighted_p2.rm", "3.1"),
    ("With_SceneInfo_Block.rm", "3.4"),  # XXX version?
    ("Color_and_tool_v3.14.4.rm", "3.14"),
    ("More_color_highlight_shader_v3.15.4.2.rm", "3.15"),
]


@pytest.mark.parametrize("test_file,version", TEST_FILES_AND_VERSIONS)
def test_full_roundtrip(test_file, version):
    with open(DATA_PATH / test_file, "rb") as f:
        data = f.read()

    # XXX not sure why this is a problem -- reMarkable seems to be inconsistent
    # about the min version written to line block headers?
    if version in ("3.2.2", "3.3.2"):
        # This is not a very good way of doing it...
        data = data.replace(bytes.fromhex("010205"), bytes.fromhex("020205"))

    input_buf = BytesIO(data)
    output_buf = BytesIO()
    options = {"version": version}

    write_blocks(output_buf, read_blocks(input_buf), options)

    assert _hex_lines(input_buf.getvalue()) == _hex_lines(output_buf.getvalue())


# Temporarily add test files here that add new data fields before updating the
# parsing code properly.
FULL_PARSING_XFAILS = []

TEST_FILES_FOR_FULL_PARSING = [
    pytest.param(
        filename,
        marks=pytest.mark.xfail if filename in FULL_PARSING_XFAILS else [],
    )
    for filename, _ in TEST_FILES_AND_VERSIONS
]


@pytest.mark.parametrize("test_file", TEST_FILES_FOR_FULL_PARSING)
def test_files_fully_parsed(test_file):
    with open(DATA_PATH / test_file, "rb") as f:
        result = list(read_blocks(f))

    # Check none of the blocks were unreadable and do not have extra data
    for block in result:
        assert not isinstance(block, UnreadableBlock)
        assert not block.extra_data
        if isinstance(block, SceneItemBlock):
            assert not block.extra_value_data


def test_normal_ab():
    with open(DATA_PATH / "Normal_AB.rm", "rb") as f:
        result = list(read_blocks(f))

    assert result == [
        AuthorIdsBlock(author_uuids={1: UUID("495ba59f-c943-2b5c-b455-3682f6948906")}),
        MigrationInfoBlock(migration_id=CrdtId(1, 1), is_device=True),
        PageInfoBlock(
            loads_count=1, merges_count=0, text_chars_count=3, text_lines_count=1
        ),
        SceneTreeBlock(
            tree_id=CrdtId(0, 11),
            node_id=CrdtId(0, 0),
            is_update=True,
            parent_id=CrdtId(0, 1),
        ),
        RootTextBlock(
            block_id=CrdtId(0, 0),
            value=si.Text(
                items=CrdtSequence(
                    [
                        CrdtSequenceItem(
                            item_id=CrdtId(1, 16),
                            left_id=CrdtId(0, 0),
                            right_id=CrdtId(0, 0),
                            deleted_length=0,
                            value="AB",
                        )
                    ]
                ),
                styles={
                    CrdtId(0, 0): LwwValue(
                        timestamp=CrdtId(1, 15), value=si.ParagraphStyle.PLAIN
                    ),
                },
                pos_x=-468.0,
                pos_y=234.0,
                width=936.0,
            ),
        ),
        TreeNodeBlock(
            group=si.Group(node_id=CrdtId(0, 1)),
        ),
        TreeNodeBlock(
            group=si.Group(
                node_id=CrdtId(0, 11),
                label=LwwValue(CrdtId(0, 12), "Layer 1"),
            ),
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


def test_read_glyph_range():
    with open(DATA_PATH / "Wikipedia_highlighted_p1.rm", "rb") as f:
        result = [
            block for block in read_blocks(f) if isinstance(block, SceneGlyphItemBlock)
        ]

    assert result[0].item.value.text == "The reMarkable uses electronic paper"


@pytest.mark.parametrize(
    "block",
    [
        AuthorIdsBlock(author_uuids={1: UUID("495ba59f-c943-2b5c-b455-3682f6948906")}),
        AuthorIdsBlock(
            author_uuids={
                1: UUID("495ba59f-c943-2b5c-b455-3682f6948906"),
                2: UUID("cd83324a-917f-11ed-bb7b-3c0754484e34"),
            }
        ),
        MigrationInfoBlock(migration_id=CrdtId(1, 1), is_device=True),
        PageInfoBlock(
            loads_count=3, merges_count=2, text_chars_count=3, text_lines_count=1
        ),
        SceneTreeBlock(
            tree_id=CrdtId(0, 11),
            node_id=CrdtId(0, 0),
            is_update=True,
            parent_id=CrdtId(0, 1),
        ),
        RootTextBlock(
            block_id=CrdtId(0, 0),
            value=si.Text(
                items=CrdtSequence(
                    [
                        CrdtSequenceItem(
                            item_id=CrdtId(1, 16),
                            left_id=CrdtId(0, 0),
                            right_id=CrdtId(0, 0),
                            deleted_length=0,
                            value="AB",
                        )
                    ]
                ),
                styles={
                    CrdtId(0, 0): LwwValue(
                        timestamp=CrdtId(1, 15), value=si.ParagraphStyle.PLAIN
                    ),
                },
                pos_x=-468.0,
                pos_y=234.0,
                width=936.0,
            ),
        ),
        TreeNodeBlock(
            group=si.Group(
                node_id=CrdtId(0, 11),
                label=LwwValue(CrdtId(0, 12), "Layer 1"),
            ),
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
        SceneGlyphItemBlock(
            parent_id=CrdtId(0, 11),
            item=CrdtSequenceItem(
                item_id=CrdtId(1, 17),
                left_id=CrdtId(1, 16),
                right_id=CrdtId(0, 0),
                deleted_length=0,
                value=si.GlyphRange(
                    start=1536,
                    length=23,
                    text="display technology.[13]",
                    color=si.PenColor.YELLOW,
                    rectangles=[
                        si.Rectangle(
                            x=-809.061564750815,
                            y=1724.1146737357485,
                            w=333.5427440226558,
                            h=56.30432956921868,
                        ),
                        si.Rectangle(
                            x=-485.51105154941456,
                            y=1730.4364894378523,
                            w=58.22011730763188,
                            h=33.42225280328421,
                        ),
                    ],
                ),
            ),
        ),
    ],
)
def test_blocks_roundtrip(block):
    buf = BytesIO()
    writer = TaggedBlockWriter(buf)
    reader = TaggedBlockReader(buf)

    block.write(writer)
    buf.seek(0)
    logger.info("After writing block %s", type(block))
    logger.info("Buffer: %s", buf.getvalue().hex())

    block2 = Block.read(reader)

    assert block2 == block


def test_write_blocks():
    blocks = [
        MigrationInfoBlock(migration_id=CrdtId(1, 1), is_device=True),
    ]

    buf = BytesIO()
    write_blocks(buf, blocks, options={"version": "3.1"})

    assert buf.getvalue()[:43] == b"reMarkable .lines file, version=6          "
    assert buf.getvalue()[43:].hex() == "05000000000101001f01012101"


def test_blocks_keep_unknown_data_in_main_block():
    # The "E1 FF" represents new, unknown data -- note that this might need
    # to be changed in future if the next id starts to actually be used in a
    # future update!
    data_hex = """
    2E000000 0000010D
    1C 06000000
       1F 0000
       2F 0000
    2C 05000000
       1F 0000 21 01
    3C 05000000
       1F 0000 21 01
    5C 08000000
       7C050000 50070000
    E1 FF
    """
    buf = BytesIO(HEADER_V6 + bytes.fromhex(data_hex))
    block = next(read_blocks(buf))
    assert isinstance(block, SceneInfo)
    assert block.extra_data == bytes.fromhex("E1 FF")


def test_blocks_keep_unknown_data_in_value_subblock():
    # The "8f 010f" is represents new, unknown data -- note that this might need
    # to be changed in future if the next id starts to actually be used in a
    # future update!
    data_hex = """
    59000000 00020205
    1f 0219
    2f 021e
    3f 0000
    4f 0000
    54 0000 0000
    6c 4300 0000
       03
       14 0f000000
       24 00000000
       38 00000000 0000f03f
       44 00000000
       5c 1c000000
          f8fe82c2 f42a30c3 03000800 0000b869
          83c2622d 30c30000 08000000
       6f 0001
       7f 010f
       8f 0101
    """
    buf = BytesIO(HEADER_V6 + bytes.fromhex(data_hex))
    block = next(read_blocks(buf))
    assert isinstance(block, SceneLineItemBlock)
    assert block.extra_value_data == bytes.fromhex("8f 0101")


def test_error_in_block_contained():
    # First block will cause a parsing error at `0xff`. Second block should
    # still be parsed.
    data_hex = """
    06000000 00010103
    1f 0219
    aa bbcc
    05000000 00010100
    1f 0219
    21 01
    """
    buf = BytesIO(HEADER_V6 + bytes.fromhex(data_hex))
    blocks = list(read_blocks(buf))
    assert blocks == [
        UnreadableBlock(
            error="Bad tag type 0xA at position 58",
            data=bytes.fromhex("1f0219aabbcc"),
            info=MainBlockInfo(
                offset=51,
                size=6,
                extra_data=b"",
                block_type=3,
                min_version=1,
                current_version=1,
            ),
        ),
        MigrationInfoBlock(migration_id=CrdtId(0x02, 0x19), is_device=True),
    ]

    # Check all data is preserved
    buf2 = BytesIO()
    # Version 3 to be consistent with the input data used here
    write_blocks(buf2, blocks, options={"version": "3.0"})

    assert buf2.getvalue() == buf.getvalue()


from hypothesis import given, strategies as st


crdt_id_strategy = st.builds(
    CrdtId,
    st.integers(min_value=0, max_value=2**8 - 1),
    st.integers(min_value=0, max_value=2**64 - 1),
)
st.register_type_strategy(CrdtId, crdt_id_strategy)

author_ids_block_strategy = st.builds(
    AuthorIdsBlock,
    st.dictionaries(st.integers(min_value=0, max_value=65535), st.uuids()),
    extra_data=st.binary(),
)

block_strategy = st.one_of(
    [
        author_ids_block_strategy,
        st.builds(MigrationInfoBlock),
    ]
)


@given(block_strategy)
def test_blocks_roundtrip_2(block):
    buf = BytesIO()
    writer = TaggedBlockWriter(buf)
    reader = TaggedBlockReader(buf)

    block.write(writer)
    buf.seek(0)
    logger.info("After writing block %s", type(block))
    logger.info("Buffer: %s", buf.getvalue().hex())
    block2 = Block.read(reader)
    assert block2 == block


@given(...)
def test_write_id(crdt_id: CrdtId):
    buf = BytesIO()
    s = TaggedBlockWriter(buf)
    s.write_id(3, crdt_id)
