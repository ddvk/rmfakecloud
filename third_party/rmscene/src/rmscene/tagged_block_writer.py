"""Read structure of remarkable .rm files version 6.

Based on my investigation of the format with lots of help from ddvk's v6 reader
code.

"""

from __future__ import annotations

from collections.abc import Iterator
from contextlib import contextmanager
from io import BytesIO
import logging
import typing as tp

from .tagged_block_common import (
    TagType,
    DataStream,
    CrdtId,
    LwwValue,
    UnexpectedBlockError,
)


_logger = logging.getLogger(__name__)


class TaggedBlockWriter:
    """Write blocks and values to a remarkable v6 file stream."""

    def __init__(self, data: tp.BinaryIO, options: tp.Optional[dict] = None):
        if options is None:
            options = {}
        self.options = options
        rm_data = DataStream(data)
        self.data = rm_data
        self._in_block: bool = False

    def write_header(self) -> None:
        """Write the file header.

        This should be the first call when starting to write a new file.

        """
        self.data.write_header()

    ## Write simple values

    def write_id(self, index: int, value: CrdtId):
        """Write a tagged CRDT ID."""
        self.data.write_tag(index, TagType.ID)
        self.data.write_crdt_id(value)

    def write_bool(self, index: int, value: bool):
        """Write a tagged bool."""
        self.data.write_tag(index, TagType.Byte1)
        self.data.write_bool(value)

    def write_byte(self, index: int, value: int):
        """Write a tagged byte as an unsigned integer."""
        self.data.write_tag(index, TagType.Byte1)
        self.data.write_uint8(value)

    def write_int(self, index: int, value: int):
        """Write a tagged 4-byte unsigned integer."""
        self.data.write_tag(index, TagType.Byte4)
        # TODO: is this supposed to be signed or unsigned?
        self.data.write_uint32(value)

    def write_float(self, index: int, value: float):
        """Write a tagged 4-byte float."""
        self.data.write_tag(index, TagType.Byte4)
        self.data.write_float32(value)

    def write_double(self, index: int, value: float):
        """Write a tagged 8-byte double."""
        self.data.write_tag(index, TagType.Byte8)
        self.data.write_float64(value)

    def write_color(self, index: int, value: tuple[int, int, int, int]):
        """Write a tagged RGBA color packed as a little-endian uint32 (stored BGRA)."""
        r, g, b, a = value
        self.write_int(index, b | (g << 8) | (r << 16) | (a << 24))

    ## Blocks

    @contextmanager
    def write_block(
        self, block_type: int, min_version: int, current_version: int
    ) -> Iterator[None]:
        """Write a top-level block header.

        Within this block, other writes are accumulated, so that the
        whole block can be written out with its length at the end.

        """
        if self._in_block:
            raise UnexpectedBlockError("Already in a block")

        previous_data = self.data
        block_buf = BytesIO()
        block_data = DataStream(block_buf)
        try:
            self.data = block_data
            self._in_block = True
            yield
        finally:
            self.data = previous_data

        assert self._in_block
        self._in_block = False

        self.data.write_uint32(len(block_buf.getbuffer()))
        self.data.write_uint8(0)
        self.data.write_uint8(min_version)
        self.data.write_uint8(current_version)
        self.data.write_uint8(block_type)
        self.data.write_bytes(block_buf.getbuffer())

    @contextmanager
    def write_subblock(self, index: int) -> Iterator[None]:
        """Write a subblock tag and length once the with-block has exited.

        Within this block, other writes are accumulated, so that the
        whole block can be written out with its length at the end.
        """
        previous_data = self.data
        subblock_buf = BytesIO()
        subblock_data = DataStream(subblock_buf)
        try:
            self.data = subblock_data
            yield
        finally:
            self.data = previous_data

        self.data.write_tag(index, TagType.Length4)
        self.data.write_uint32(len(subblock_buf.getbuffer()))
        self.data.write_bytes(subblock_buf.getbuffer())
        _logger.debug("Wrote subblock %d: %s", index, subblock_buf.getvalue().hex())

    ## Higher level constructs

    def write_lww_bool(self, index: int, value: LwwValue[bool]):
        "Write a LWW bool."
        with self.write_subblock(index):
            self.write_id(1, value.timestamp)
            self.write_bool(2, value.value)

    def write_lww_byte(self, index: int, value: LwwValue[int]):
        "Write a LWW byte."
        with self.write_subblock(index):
            self.write_id(1, value.timestamp)
            self.write_byte(2, value.value)

    def write_lww_float(self, index: int, value: LwwValue[float]):
        "Write a LWW float."
        with self.write_subblock(index):
            self.write_id(1, value.timestamp)
            self.write_float(2, value.value)

    def write_lww_id(self, index: int, value: LwwValue[CrdtId]):
        "Write a LWW ID."
        with self.write_subblock(index):
            # XXX ddvk has these the other way round?
            self.write_id(1, value.timestamp)
            self.write_id(2, value.value)

    def write_lww_bytes(self, index: int, value: LwwValue[bytes]):
        "Write a LWW bytes value."
        with self.write_subblock(index):
            self.write_id(1, value.timestamp)
            self.write_bytes(2, value.value)

    def write_lww_string(self, index: int, value: LwwValue[str]):
        "Write a LWW string."
        with self.write_subblock(index):
            self.write_id(1, value.timestamp)
            self.write_string(2, value.value)

    def write_bytes(self, index: int, value: bytes):
        """Write a standard bytes block."""
        with self.write_subblock(index):
            self.data.write_bytes(value)

    def write_string(self, index: int, value: str):
        """Write a standard string block."""
        with self.write_subblock(index):
            b = value.encode()
            bytes_length = len(b)
            is_ascii = True  # XXX not sure if this is right meaning?
            self.data.write_varuint(bytes_length)
            self.data.write_bool(is_ascii)
            self.data.write_bytes(b)

    def write_string_with_format(self, index: int, text: str, fmt: int):
        """Write a string block with formatting."""
        with self.write_subblock(index):
            b = text.encode()
            bytes_length = len(b)
            is_ascii = True  # XXX not sure if this is right meaning?
            self.data.write_varuint(bytes_length)
            self.data.write_bool(is_ascii)
            self.data.write_bytes(b)
            self.write_int(2, fmt)

    def write_int_pair(self, index: int, value: tuple[int, int]):
        """Read a sub block containing two uint32"""
        with self.write_subblock(index):
            self.data.write_uint32(value[0])
            self.data.write_uint32(value[1])
