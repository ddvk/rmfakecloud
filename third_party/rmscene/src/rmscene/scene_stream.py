"""Read structure of reMarkable tablet lines format v6

With help from ddvk's v6 reader, and enum values from remt.

"""

from __future__ import annotations

import logging
import math
import typing as tp
from abc import ABC, abstractmethod
from collections.abc import Iterable, Iterator
from dataclasses import KW_ONLY, dataclass, replace
from uuid import UUID, uuid4

from packaging.version import Version

from . import scene_items as si
from .crdt_sequence import CrdtSequence, CrdtSequenceItem
from .scene_tree import SceneTree
from .tagged_block_common import CrdtId, LwwValue, TagType, UnexpectedBlockError
from .tagged_block_reader import MainBlockInfo, TaggedBlockReader
from .tagged_block_writer import TaggedBlockWriter

_logger = logging.getLogger(__name__)


############################################################
# Top-level block types
############################################################


@dataclass
class Block(ABC):
    BLOCK_TYPE: tp.ClassVar

    # Store any unrecognised data we can't understand
    _: KW_ONLY
    extra_data: bytes = b""

    def version_info(self, writer: TaggedBlockWriter) -> tuple[int, int]:
        """Return (min_version, current_version) to use when writing."""
        return (1, 1)

    def get_block_type(self) -> int:
        """Return block type for this block.

        By default, returns the block's BLOCK_TYPE attribute, but this method
        can be overriden if a single block subclass can handle multiple block
        types.

        """
        return self.BLOCK_TYPE

    @classmethod
    def lookup(cls, block_type: int) -> tp.Optional[tp.Type[Block]]:
        if getattr(cls, "BLOCK_TYPE", None) == block_type:
            return cls
        for subclass in cls.__subclasses__():
            if match := subclass.lookup(block_type):
                return match
        return None

    @classmethod
    def read(self, reader: TaggedBlockReader) -> Optional[Block]:
        """
        Maybe parse a block from the reader stream.
        """
        with reader.read_block() as block_info:
            if block_info is None:
                return

            block_type = Block.lookup(block_info.block_type)
            if block_type:
                try:
                    block = block_type.from_stream(reader)
                except Exception as e:
                    _logger.warning("Error reading block: %s", e)
                    reader.data.data.seek(block_info.offset)
                    data = reader.data.read_bytes(block_info.size)
                    block = UnreadableBlock(str(e), data, block_info)
            else:
                msg = (
                    f"Unknown block type {block_info.block_type}. "
                    f"Skipping {block_info.size} bytes."
                )
                _logger.warning(msg)
                data = reader.data.read_bytes(block_info.size)
                block = UnreadableBlock(msg, data, block_info)

        # Keep any unparsed extra data
        block.extra_data = block_info.extra_data
        return block

    def write(self, writer: TaggedBlockWriter):
        """Write the block header and content to the stream."""
        min_version, current_version = self.version_info(writer)
        with writer.write_block(self.get_block_type(), min_version, current_version):
            self.to_stream(writer)
            # Write any leftover extra data that wasn't parsed
            writer.data.write_bytes(self.extra_data)

    @classmethod
    @abstractmethod
    def from_stream(cls, reader: TaggedBlockReader) -> Block:
        """Read content of block from stream."""
        raise NotImplementedError()

    @abstractmethod
    def to_stream(self, writer: TaggedBlockWriter):
        """Write content of block to stream."""
        raise NotImplementedError()


@dataclass
class UnreadableBlock(Block):
    """Represent a block which could not be read for some reason."""

    error: str
    data: bytes
    info: MainBlockInfo

    def get_block_type(self) -> int:
        return self.info.block_type

    @classmethod
    def from_stream(cls, reader: TaggedBlockReader) -> Block:
        raise NotImplementedError()

    def to_stream(self, writer: TaggedBlockWriter):
        writer.data.write_bytes(self.data)


@dataclass
class SceneInfo(Block):
    BLOCK_TYPE: tp.ClassVar = 0x0D

    def version_info(self, _) -> tuple[int, int]:
        """Return (min_version, current_version) to use when writing."""
        return (0, 1)

    current_layer: LwwValue[CrdtId]
    background_visible: tp.Optional[LwwValue[bool]]
    root_document_visible: tp.Optional[LwwValue[bool]]
    paper_size: tp.Optional[tuple[int, int]]

    @classmethod
    def from_stream(cls, stream: TaggedBlockReader) -> SceneInfo:
        current_layer = stream.read_lww_id(1)
        background_visible = stream.read_lww_bool(2) if stream.bytes_remaining_in_block() > 0 else None
        root_document_visible = stream.read_lww_bool(3) if stream.bytes_remaining_in_block() > 0 else None
        paper_size = stream.read_int_pair(5) if stream.bytes_remaining_in_block() > 0 else None

        return SceneInfo(current_layer=current_layer,
                         background_visible=background_visible,
                         root_document_visible=root_document_visible,
                         paper_size=paper_size)

    def to_stream(self, writer: TaggedBlockWriter):
        writer.write_lww_id(1, self.current_layer)
        if self.background_visible:
            writer.write_lww_bool(2, self.background_visible)
        if self.root_document_visible:
            writer.write_lww_bool(3, self.root_document_visible)
        if self.paper_size:
            writer.write_int_pair(5, self.paper_size)


@dataclass
class SceneImageInfoBlock(Block):
    BLOCK_TYPE: tp.ClassVar = 0x0E
    images: dict[UUID, si.ImageInfo]

    def version_info(self, _) -> tuple[int, int]:
        return (3, 3)

    @classmethod
    def from_stream(cls, stream: TaggedBlockReader) -> SceneImageInfoBlock:
        with stream.read_subblock(1):
            images = {}
            num_images = stream.data.read_varuint()
            for _ in range(num_images):
                with stream.read_subblock(0):
                    uuid = UUID(bytes_le=stream.data.read_bytes(16))
                    filename = stream.read_lww_string(1)
                    flags = stream.read_lww_bytes(2)
                    images[uuid] = si.ImageInfo(filename=filename, flags=flags)
        return SceneImageInfoBlock(images=images)

    def to_stream(self, writer: TaggedBlockWriter):
        with writer.write_subblock(1):
            writer.data.write_varuint(len(self.images))
            for uuid, image in self.images.items():
                with writer.write_subblock(0):
                    writer.data.write_bytes(uuid.bytes_le)
                    writer.write_lww_string(1, image.filename)
                    writer.write_lww_bytes(2, image.flags)


@dataclass
class AuthorIdsBlock(Block):
    BLOCK_TYPE: tp.ClassVar = 0x09

    author_uuids: dict[int, UUID]

    @classmethod
    def from_stream(cls, stream: TaggedBlockReader) -> AuthorIdsBlock:
        _logger.debug("Reading %s", cls.__name__)
        num_subblocks = stream.data.read_varuint()
        author_ids = {}
        for _ in range(num_subblocks):
            with stream.read_subblock(0):
                uuid_length = stream.data.read_varuint()
                if uuid_length != 16:
                    raise ValueError("Expected UUID length to be 16 bytes")
                uuid = UUID(bytes_le=stream.data.read_bytes(uuid_length))
                author_id = stream.data.read_uint16()
                author_ids[author_id] = uuid
        return AuthorIdsBlock(author_ids)

    def to_stream(self, writer: TaggedBlockWriter):
        _logger.debug("Writing %s", type(self).__name__)
        num_subblocks = len(self.author_uuids)
        writer.data.write_varuint(num_subblocks)
        for author_id, uuid in self.author_uuids.items():
            with writer.write_subblock(0):
                writer.data.write_varuint(len(uuid.bytes_le))
                writer.data.write_bytes(uuid.bytes_le)
                writer.data.write_uint16(author_id)


@dataclass
class MigrationInfoBlock(Block):
    BLOCK_TYPE: tp.ClassVar = 0x00

    migration_id: CrdtId
    is_device: bool
    _unknown: bool = False

    @classmethod
    def from_stream(cls, stream: TaggedBlockReader) -> MigrationInfoBlock:
        "Parse migration info"
        _logger.debug("Reading %s", cls.__name__)
        migration_id = stream.read_id(1)
        is_device = stream.read_bool(2)
        if stream.bytes_remaining_in_block():
            unknown = stream.read_bool(3)
        else:
            unknown = False
        return MigrationInfoBlock(migration_id, is_device, unknown)

    def to_stream(self, writer: TaggedBlockWriter):
        _logger.debug("Writing %s", type(self).__name__)
        version = writer.options.get("version", Version("9.9.9"))
        writer.write_id(1, self.migration_id)
        writer.write_bool(2, self.is_device)
        if version >= Version("3.2.2"):
            writer.write_bool(3, self._unknown)


@dataclass
class TreeNodeBlock(Block):
    BLOCK_TYPE: tp.ClassVar = 0x02

    def version_info(self, writer: TaggedBlockWriter) -> tuple[int, int]:
        """Return (min_version, current_version) to use when writing."""
        version = writer.options.get("version", Version("9999"))
        # XXX this is a guess about which version this changed in
        return (1, 2) if (version >= Version("3.4")) else (1, 1)

    group: si.Group

    @classmethod
    def from_stream(cls, stream: TaggedBlockReader) -> TreeNodeBlock:
        "Parse tree node block."
        _logger.debug("Reading %s", cls.__name__)

        group = si.Group(
            node_id=stream.read_id(1),
            label=stream.read_lww_string(2),
            visible=stream.read_lww_bool(3),
        )

        # XXX this may need to be generalised for other examples
        if stream.bytes_remaining_in_block() > 0:
            group.anchor_id = stream.read_lww_id(7)
            group.anchor_type = stream.read_lww_byte(8)
            group.anchor_threshold = stream.read_lww_float(9)
            group.anchor_origin_x = stream.read_lww_float(10)

        return cls(group)

    def to_stream(self, writer: TaggedBlockWriter):
        _logger.debug("Writing %s", type(self).__name__)
        group = self.group
        writer.write_id(1, group.node_id)
        writer.write_lww_string(2, group.label)
        writer.write_lww_bool(3, group.visible)
        if group.anchor_id is not None:
            # FIXME group together in an anchor type?
            assert (
                group.anchor_type is not None
                and group.anchor_threshold is not None
                and group.anchor_origin_x is not None
            )
            writer.write_lww_id(7, group.anchor_id)
            writer.write_lww_byte(8, group.anchor_type)
            writer.write_lww_float(9, group.anchor_threshold)
            writer.write_lww_float(10, group.anchor_origin_x)


@dataclass
class PageInfoBlock(Block):
    BLOCK_TYPE: tp.ClassVar = 0x0A

    def version_info(self, _) -> tuple[int, int]:
        """Return (min_version, current_version) to use when writing."""
        return (0, 1)

    loads_count: int
    merges_count: int
    text_chars_count: int
    text_lines_count: int
    type_folio_use_count: int = 0

    @classmethod
    def from_stream(cls, stream: TaggedBlockReader) -> PageInfoBlock:
        "Parse page info block"
        _logger.debug("Reading %s", cls.__name__)
        info = PageInfoBlock(
            loads_count=stream.read_int(1),
            merges_count=stream.read_int(2),
            text_chars_count=stream.read_int(3),
            text_lines_count=stream.read_int(4),
        )
        if stream.bytes_remaining_in_block():
            info.type_folio_use_count = stream.read_int(5)
        return info

    def to_stream(self, writer: TaggedBlockWriter):
        _logger.debug("Writing %s", type(self).__name__)
        writer.write_int(1, self.loads_count)
        writer.write_int(2, self.merges_count)
        writer.write_int(3, self.text_chars_count)
        writer.write_int(4, self.text_lines_count)
        version = writer.options.get("version", Version("9999"))
        if version >= Version("3.2.2"):
            writer.write_int(5, self.type_folio_use_count)


@dataclass
class SceneTreeBlock(Block):
    BLOCK_TYPE: tp.ClassVar = 0x01

    # XXX not sure what the difference is
    tree_id: CrdtId
    node_id: CrdtId
    is_update: bool
    parent_id: CrdtId

    @classmethod
    def from_stream(cls, stream: TaggedBlockReader) -> SceneTreeBlock:
        "Parse scene tree block"
        _logger.debug("Reading %s", cls.__name__)

        # XXX not sure what the difference is. This "tree_id" is used as the
        # plain "Id" in the SceneTree.NodeMap in ddvk's reader. If the parent_id
        # is equal to the root_id (1, 1), this node represents a layer.
        tree_id = stream.read_id(1)
        node_id = stream.read_id(2)
        is_update = stream.read_bool(3)
        with stream.read_subblock(4):
            parent_id = stream.read_id(1)
            # XXX can there sometimes be something else here?

        return SceneTreeBlock(tree_id, node_id, is_update, parent_id)

    def to_stream(self, writer: TaggedBlockWriter):
        _logger.debug("Writing %s", type(self).__name__)
        writer.write_id(1, self.tree_id)
        writer.write_id(2, self.node_id)
        writer.write_bool(3, self.is_update)
        with writer.write_subblock(4):
            writer.write_id(1, self.parent_id)


def point_from_stream(stream: TaggedBlockReader, version: int = 2) -> si.Point:
    if version not in (1, 2):
        raise ValueError("Unknown version %s" % version)
    d = stream.data
    x = d.read_float32()
    y = d.read_float32()
    if version == 1:
        # calculation based on ddvk's reader
        # XXX removed rounding so that can round-trip correctly?
        speed = d.read_float32() * 4
        # speed = int(round(d.read_float32() * 4))
        direction = 255 * d.read_float32() / (math.pi * 2)
        # direction = int(round(255 * d.read_float32() / (math.pi * 2)))
        width = int(round(d.read_float32() * 4))
        pressure = d.read_float32() * 255
        # pressure = int(round(d.read_float32() * 255))
    else:
        speed = d.read_uint16()
        width = d.read_uint16()
        direction = d.read_uint8()
        pressure = d.read_uint8()
    return si.Point(x, y, speed, direction, width, pressure)


def point_serialized_size(version: int = 2) -> int:
    if version == 1:
        return 0x18
    elif version == 2:
        return 0x0E
    else:
        raise ValueError("Unknown version %s" % version)


def point_to_stream(point: si.Point, writer: TaggedBlockWriter, version: int = 2):
    if version not in (1, 2):
        raise ValueError("Unknown version %s" % version)
    d = writer.data
    d.write_float32(point.x)
    d.write_float32(point.y)
    _logger.debug("Writing Point v%d: %s", version, point)
    if version == 1:
        # calculation based on ddvk's reader
        d.write_float32(point.speed / 4)
        d.write_float32(point.direction * (2 * math.pi) / 255)
        d.write_float32(point.width / 4)
        d.write_float32(point.pressure / 255)
    else:
        d.write_uint16(point.speed)
        d.write_uint16(point.width)
        d.write_uint8(point.direction)
        d.write_uint8(point.pressure)


def line_from_stream(stream: TaggedBlockReader, version: int = 2) -> si.Line:
    _logger.debug("Reading Line version %d", version)
    tool_id = stream.read_int(1)
    tool = si.Pen(tool_id)
    color_id = stream.read_int(2)
    color = si.PenColor(color_id)
    thickness_scale = stream.read_double(3)
    starting_length = stream.read_float(4)
    with stream.read_subblock(5) as block_info:
        data_length = block_info.size
        point_size = point_serialized_size(version)
        if data_length % point_size != 0:
            raise ValueError(
                "Point data size mismatch: %d is not multiple of point_size"
                % data_length
            )
        num_points = data_length // point_size
        points = [point_from_stream(stream, version=version) for _ in range(num_points)]

    # XXX unused
    timestamp = stream.read_id(6)

    if stream.bytes_remaining_in_block() >= 3:
        try:
            move_id = stream.read_id(7)
        except UnexpectedBlockError as _:
            move_id = None
    else:
        move_id = None

    color_rgba = stream.read_color_optional(8)

    return si.Line(color, tool, points, thickness_scale, starting_length, move_id, color_rgba)


def line_to_stream(line: si.Line, writer: TaggedBlockWriter, version: int = 2):
    _logger.debug("Writing Line version %d", version)
    writer.write_int(1, line.tool)
    writer.write_int(2, line.color)
    writer.write_double(3, line.thickness_scale)
    writer.write_float(4, line.starting_length)
    with writer.write_subblock(5):
        for point in line.points:
            point_to_stream(point, writer, version)

    # XXX didn't save
    timestamp = CrdtId(0, 1)
    writer.write_id(6, timestamp)
    if line.move_id is not None:
        writer.write_id(7, line.move_id)
    if line.color_rgba is not None:
        writer.write_color(8, line.color_rgba)


@dataclass
class SceneItemBlock(Block):
    parent_id: CrdtId
    item: CrdtSequenceItem
    extra_value_data: bytes = b""

    ITEM_TYPE: tp.ClassVar[int] = 0

    @classmethod
    def from_stream(cls, stream: TaggedBlockReader) -> SceneItemBlock:
        "Group item block?"
        _logger.debug("Reading %s", cls.__name__)

        assert stream.current_block
        block_type = stream.current_block.block_type
        if block_type == SceneGlyphItemBlock.BLOCK_TYPE:
            subclass = SceneGlyphItemBlock
        elif block_type == SceneGroupItemBlock.BLOCK_TYPE:
            subclass = SceneGroupItemBlock
        elif block_type == SceneLineItemBlock.BLOCK_TYPE:
            subclass = SceneLineItemBlock
        elif block_type == SceneTextItemBlock.BLOCK_TYPE:
            subclass = SceneTextItemBlock
        elif block_type == SceneTombstoneItemBlock.BLOCK_TYPE:
            subclass = SceneTombstoneItemBlock
        elif block_type == SceneImageItemBlock.BLOCK_TYPE:
            subclass = SceneImageItemBlock
        else:
            raise ValueError(
                "unknown scene type %d in %s" % (block_type, stream.current_block)
            )

        parent_id = stream.read_id(1)
        item_id = stream.read_id(2)
        left_id = stream.read_id(3)
        right_id = stream.read_id(4)
        deleted_length = stream.read_int(5)

        if stream.has_subblock(6):
            with stream.read_subblock(6) as block_info:
                item_type = stream.data.read_uint8()
                assert item_type == subclass.ITEM_TYPE
                value = subclass.value_from_stream(stream)
            # Keep known extra data from within the value subblock
            extra_value_data = block_info.extra_data
        else:
            value = None
            extra_value_data = b""

        return subclass(
            parent_id,
            CrdtSequenceItem(item_id, left_id, right_id, deleted_length, value),
            extra_value_data=extra_value_data,
        )

    def to_stream(self, writer: TaggedBlockWriter):
        _logger.debug("Writing %s", type(self).__name__)
        writer.write_id(1, self.parent_id)
        writer.write_id(2, self.item.item_id)
        writer.write_id(3, self.item.left_id)
        writer.write_id(4, self.item.right_id)
        writer.write_int(5, self.item.deleted_length)

        if self.item.value is not None:
            with writer.write_subblock(6):
                writer.data.write_uint8(self.ITEM_TYPE)
                self.value_to_stream(writer, self.item.value)

                writer.data.write_bytes(self.extra_value_data)

    @classmethod
    @abstractmethod
    def value_from_stream(cls, reader: TaggedBlockReader) -> tp.Any:
        """Read the specific content of this block"""
        raise NotImplementedError()

    @abstractmethod
    def value_to_stream(self, writer: TaggedBlockWriter, value: tp.Any):
        """Write the specific content of this block"""
        raise NotImplementedError()


# These share the same structure so can share the same implementation?


def glyph_range_from_stream(stream: TaggedBlockReader) -> si.GlyphRange:
    # Since reMarkable version 3.6, the start and length are optional
    start = stream.read_int_optional(2)
    length = stream.read_int_optional(3)

    color_id = stream.read_int(4)
    color = si.PenColor(color_id)
    text = stream.read_string(5)

    if length is None:
        length = len(text)

    # Note: the decoded text length is not always the same as the length in the
    # glyph range...
    if len(text) != length:
        _logger.debug(
            "GlyphRange text length %d != length value %d: %r",
            len(text),
            length,
            text,
        )

    with stream.read_subblock(6):
        num_rects = stream.data.read_varuint()
        rectangles = [
            si.Rectangle(*[stream.data.read_float64() for _ in range(4)])
            for _ in range(num_rects)
        ]

    color_rgba = stream.read_color_optional(10)

    return si.GlyphRange(start, length, text, color, rectangles, color_rgba)


def glyph_range_to_stream(stream: TaggedBlockWriter, item: si.GlyphRange):
    if item.start is not None:
        stream.write_int(2, item.start)
        stream.write_int(3, item.length)
    stream.write_int(4, item.color)
    stream.write_string(5, item.text)
    with stream.write_subblock(6):
        stream.data.write_varuint(len(item.rectangles))
        for rect in item.rectangles:
            stream.data.write_float64(rect.x)
            stream.data.write_float64(rect.y)
            stream.data.write_float64(rect.w)
            stream.data.write_float64(rect.h)
    if item.color_rgba is not None:
        stream.write_color(10, item.color_rgba)


class SceneTombstoneItemBlock(SceneItemBlock):
    BLOCK_TYPE: tp.ClassVar = 0x08

    @classmethod
    def value_from_stream(cls, reader: TaggedBlockReader):
        pass

    def value_to_stream(self, writer: TaggedBlockWriter, value):
        pass


class SceneGlyphItemBlock(SceneItemBlock):
    BLOCK_TYPE: tp.ClassVar = 0x03
    ITEM_TYPE: tp.ClassVar = 0x01

    @classmethod
    def value_from_stream(cls, reader: TaggedBlockReader) -> si.GlyphRange:
        value = glyph_range_from_stream(reader)
        return value

    def value_to_stream(self, writer: TaggedBlockWriter, value):
        glyph_range_to_stream(writer, value)


class SceneGroupItemBlock(SceneItemBlock):
    BLOCK_TYPE: tp.ClassVar = 0x04
    ITEM_TYPE: tp.ClassVar = 0x02

    @classmethod
    def value_from_stream(cls, reader: TaggedBlockReader) -> CrdtId:
        # XXX don't know what this means
        value = reader.read_id(2)
        return value

    def value_to_stream(self, writer: TaggedBlockWriter, value: CrdtId):
        writer.write_id(2, value)


class SceneLineItemBlock(SceneItemBlock):
    BLOCK_TYPE: tp.ClassVar = 0x05
    ITEM_TYPE: tp.ClassVar = 0x03

    def version_info(self, writer: TaggedBlockWriter) -> tuple[int, int]:
        """Return (min_version, current_version) to use when writing."""
        version = writer.options.get("version", Version("9999"))
        return (2, 2) if (version > Version("3.0")) else (1, 1)

    @classmethod
    def value_from_stream(cls, reader: TaggedBlockReader) -> si.Line:
        assert reader.current_block is not None
        version = reader.current_block.current_version
        value = line_from_stream(reader, version)
        return value

    def value_to_stream(self, writer: TaggedBlockWriter, value: si.Line):
        # XXX make sure this version ends up in block header
        version = writer.options.get("version", Version("9999"))
        line_version = 2 if (version > Version("3.0")) else 1
        line_to_stream(value, writer, version=line_version)


# XXX missing "PathItemBlock"? with ITEM_TYPE 0x04


class SceneImageItemBlock(SceneItemBlock):
    BLOCK_TYPE: tp.ClassVar = 0x0F
    ITEM_TYPE: tp.ClassVar = 0x07

    _ints = [0, 1, 2, 2, 3, 0]

    def version_info(self, writer: TaggedBlockWriter) -> tuple[int, int]:
        return (2, 2)

    @classmethod
    def value_from_stream(cls, reader: TaggedBlockReader) -> si.Image:
        uuid = reader.read_lww_bytes(1)
        timestamp = reader.read_id(2)

        with reader.read_subblock(3):
            length = reader.data.read_varuint()
            if length != 16:
                raise ValueError("unexpected image vertex length %s" % length)
            vertices = [reader.data.read_float32() for _ in range(16)]

        with reader.read_subblock(4):
            length = reader.data.read_varuint()
            if length != len(cls._ints):
                raise ValueError("unexpected image ints length %s" % length)
            ints = [reader.data.read_uint32() for _ in range(len(cls._ints))]
            if ints != cls._ints:
                _logger.debug("image magic ints differ: %s", ints)

        return si.Image(uuid=uuid, vertices=vertices, move_id=timestamp)

    def value_to_stream(self, writer: TaggedBlockWriter, value: si.Image):
        writer.write_lww_bytes(1, value.uuid)
        writer.write_id(2, value.move_id)

        with writer.write_subblock(3):
            writer.data.write_varuint(len(value.vertices))
            for v in value.vertices:
                writer.data.write_float32(v)

        with writer.write_subblock(4):
            writer.data.write_varuint(len(self._ints))
            for v in self._ints:
                writer.data.write_uint32(v)


class SceneTextItemBlock(SceneItemBlock):
    BLOCK_TYPE: tp.ClassVar = 0x06
    ITEM_TYPE: tp.ClassVar = 0x05

    @classmethod
    def value_from_stream(cls, reader: TaggedBlockReader) -> tp.Any:
        return None

    def value_to_stream(self, writer: TaggedBlockWriter, value):
        pass


def text_item_from_stream(stream: TaggedBlockReader) -> CrdtSequenceItem[str | int]:
    with stream.read_subblock(0):
        item_id = stream.read_id(2)
        left_id = stream.read_id(3)
        right_id = stream.read_id(4)
        deleted_length = stream.read_int(5)

        if stream.has_subblock(6):
            text, fmt = stream.read_string_with_format(6)
            # It seems that formats are stored on empty strings, so it's one or the other
            if fmt is not None:
                if text:
                    _logger.error(
                        "Unhandled combined text and format: %s, %s", text, fmt
                    )
                value = fmt
            else:
                value = text
        else:
            value = ""

    return CrdtSequenceItem(item_id, left_id, right_id, deleted_length, value)


def text_item_to_stream(item: CrdtSequenceItem[str | int], writer: TaggedBlockWriter):
    with writer.write_subblock(0):
        writer.write_id(2, item.item_id)
        writer.write_id(3, item.left_id)
        writer.write_id(4, item.right_id)
        writer.write_int(5, item.deleted_length)

        if item.value:
            if isinstance(item.value, str):
                writer.write_string(6, item.value)
            elif isinstance(item.value, int):
                writer.write_string_with_format(6, "", item.value)


def text_format_from_stream(
    stream: TaggedBlockReader,
) -> tuple[CrdtId, LwwValue[si.ParagraphStyle]]:
    # These are character ids, but not with an initial tag like other ids have.
    char_id = stream.data.read_crdt_id()

    # This seems to be the item ID for this format data? It doesn't appear
    # elsewhere in the file. Sometimes coincides with a character id but I don't
    # think it is referring to it.
    timestamp = stream.read_id(1)

    with stream.read_subblock(2):
        # XXX not sure what this is format?
        c = stream.data.read_uint8()
        assert c == 17
        format_code = stream.data.read_uint8()
        try:
            format_type = si.ParagraphStyle(format_code)
        except ValueError:
            _logger.warning("Unrecognised text format code %d.", format_code)
            _logger.debug(
                "Unrecognised text format code %d at position %d.",
                format_code,
                stream.data.tell(),
            )
            format_type = si.ParagraphStyle.PLAIN  # fallback

    return (char_id, LwwValue(timestamp, format_type))


def text_format_to_stream(
    char_id: CrdtId, value: LwwValue[si.ParagraphStyle], writer: TaggedBlockWriter
):
    format_type = value.value

    writer.data.write_crdt_id(char_id)
    writer.write_id(1, value.timestamp)
    with writer.write_subblock(2):
        # XXX not sure what this is format?
        c = 17
        writer.data.write_uint8(c)
        writer.data.write_uint8(format_type)


@dataclass
class RootTextBlock(Block):
    BLOCK_TYPE: tp.ClassVar = 0x07

    block_id: CrdtId
    value: si.Text

    @classmethod
    def from_stream(cls, stream: TaggedBlockReader) -> RootTextBlock:
        "Parse root text block."
        _logger.debug("Reading %s", cls.__name__)

        block_id = stream.read_id(1)
        assert block_id == CrdtId(0, 0)

        with stream.read_subblock(2):

            # Text items
            with stream.read_subblock(1):
                with stream.read_subblock(1):
                    num_subblocks = stream.data.read_varuint()
                    text_items = [
                        text_item_from_stream(stream) for _ in range(num_subblocks)
                    ]

            # Formatting
            with stream.read_subblock(2):
                with stream.read_subblock(1):
                    num_subblocks = stream.data.read_varuint()
                    text_formats = dict(
                        text_format_from_stream(stream) for _ in range(num_subblocks)
                    )

        # Last section
        with stream.read_subblock(3):
            # "pos_x" and "pos_y" from ddvk? Gives negative number -- possibly could
            # be bounding box?
            pos_x = stream.data.read_float64()
            pos_y = stream.data.read_float64()

        # "width" from ddvk
        width = stream.read_float(4)

        value = si.Text(
            items=CrdtSequence(text_items),
            styles=text_formats,
            pos_x=pos_x,
            pos_y=pos_y,
            width=width,
        )
        return RootTextBlock(block_id, value)

    def to_stream(self, writer: TaggedBlockWriter):
        _logger.debug("Writing %s", type(self).__name__)
        writer.write_id(1, self.block_id)

        with writer.write_subblock(2):

            # Text items
            text_items = self.value.items.sequence_items()
            with writer.write_subblock(1):
                with writer.write_subblock(1):
                    writer.data.write_varuint(len(text_items))
                    for item in text_items:
                        text_item_to_stream(item, writer)

            # Formatting
            text_formats = self.value.styles
            with writer.write_subblock(2):
                with writer.write_subblock(1):
                    writer.data.write_varuint(len(text_formats))
                    for key, item in text_formats.items():
                        text_format_to_stream(key, item, writer)

        # Last section
        with writer.write_subblock(3):
            writer.data.write_float64(self.value.pos_x)
            writer.data.write_float64(self.value.pos_y)

        # "width" from ddvk
        writer.write_float(4, self.value.width)


## Functions to read and write streams of blocks


def _read_blocks(stream: TaggedBlockReader) -> Iterator[Block]:
    """
    Parse blocks from reMarkable v6 file.
    """
    while True:
        maybe_block = Block.read(stream)
        if maybe_block:
            yield maybe_block
        else:
            # no more blocks
            return


def read_blocks(data: tp.BinaryIO) -> Iterator[Block]:
    """
    Parse reMarkable file and return iterator of document items.

    :param data: reMarkable file data.
    """
    stream = TaggedBlockReader(data)
    stream.read_header()
    yield from _read_blocks(stream)


def write_blocks(
    data: tp.BinaryIO, blocks: Iterable[Block], options: tp.Optional[dict] = None
):
    """
    Write blocks to file.
    """
    if options is not None and "version" in options:
        options["version"] = Version(options["version"])
    stream = TaggedBlockWriter(data, options=options)
    stream.write_header()
    for block in blocks:
        block.write(stream)


def build_tree(tree: SceneTree, blocks: Iterable[Block]):
    """Read `blocks` and add contents to `tree`."""
    for b in blocks:
        if isinstance(b, SceneTreeBlock):
            # XXX check node_id and is_update
            # pending_tree_nodes[b.tree_id] = b
            tree.add_node(b.tree_id, parent_id=b.parent_id)
        elif isinstance(b, TreeNodeBlock):
            # Expect this node to already exist; adding information
            # if b.node_id not in pending_tree_nodes:
            if b.group.node_id not in tree:
                raise ValueError(
                    "Node does not exist for TreeNodeBlock: %s" % b.group.node_id
                )
            node = tree[b.group.node_id]
            node.label = b.group.label
            node.visible = b.group.visible
            node.anchor_id = b.group.anchor_id
            node.anchor_type = b.group.anchor_type
            node.anchor_threshold = b.group.anchor_threshold
            node.anchor_origin_x = b.group.anchor_origin_x
        elif isinstance(b, SceneGroupItemBlock):
            # Add this entry to children of parent_id
            node_id = b.item.value
            if node_id is None:
                continue
            if node_id not in tree:
                raise ValueError(
                    "Node does not exist for SceneGroupItemBlock: %s" % node_id
                )
            item = replace(b.item, value=tree[node_id])
            tree.add_item(item, b.parent_id)
        elif isinstance(b, (SceneLineItemBlock, SceneGlyphItemBlock)):
            # Add this entry to children of parent_id
            tree.add_item(b.item, b.parent_id)
        elif isinstance(b, SceneImageItemBlock):
            node = b.item.value
            if node is None:
                continue
            if tree.image_info is not None:
                try:
                    info = tree.image_info.images[UUID(bytes_le=node.uuid.value)]
                    node.filename = info.filename.value
                except Exception as e:
                    _logger.warning("Could not resolve image filename: %s", e)
            tree.add_item(b.item, b.parent_id)
        elif isinstance(b, SceneInfo):
            tree.scene_info = b
        elif isinstance(b, SceneImageInfoBlock):
            tree.image_info = b
        elif isinstance(b, RootTextBlock):
            if tree.root_text is not None:
                _logger.error(
                    "Overwriting root text\n  Old: %s\n  New: %s",
                    tree.root_text,
                    b.value,
                )
            tree.root_text = b.value

    pass


def read_tree(data: tp.BinaryIO) -> SceneTree:
    """
    Parse reMarkable file and return `SceneTree`.

    :param data: reMarkable file data.
    """
    tree = SceneTree()
    build_tree(tree, read_blocks(data))
    return tree


def simple_text_document(text: str, author_uuid=None) -> Iterator[Block]:
    """Return the basic blocks to represent `text` as plain text.

    TODO: replace this with a way to generate the tree with given text, and a
    function to write a tree to blocks.

    """

    if author_uuid is None:
        author_uuid = uuid4()

    yield AuthorIdsBlock(author_uuids={1: author_uuid})

    yield MigrationInfoBlock(migration_id=CrdtId(1, 1), is_device=True)

    yield PageInfoBlock(
        loads_count=1,
        merges_count=0,
        text_chars_count=len(text) + 1,
        text_lines_count=text.count("\n") + 1,
    )

    yield SceneTreeBlock(
        tree_id=CrdtId(0, 11),
        node_id=CrdtId(0, 0),
        is_update=True,
        parent_id=CrdtId(0, 1),
    )

    yield RootTextBlock(
        block_id=CrdtId(0, 0),
        value=si.Text(
            items=CrdtSequence(
                [
                    CrdtSequenceItem(
                        item_id=CrdtId(1, 16),
                        left_id=CrdtId(0, 0),
                        right_id=CrdtId(0, 0),
                        deleted_length=0,
                        value=text,
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
    )

    yield TreeNodeBlock(
        si.Group(
            node_id=CrdtId(0, 1),
        )
    )

    yield TreeNodeBlock(
        si.Group(
            node_id=CrdtId(0, 11),
            label=LwwValue(timestamp=CrdtId(0, 12), value="Layer 1"),
        )
    )

    yield SceneGroupItemBlock(
        parent_id=CrdtId(0, 1),
        item=CrdtSequenceItem(
            item_id=CrdtId(0, 13),
            left_id=CrdtId(0, 0),
            right_id=CrdtId(0, 0),
            deleted_length=0,
            value=CrdtId(0, 11),
        ),
    )
