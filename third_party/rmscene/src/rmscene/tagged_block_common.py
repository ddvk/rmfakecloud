"""Helpers for reading/writing tagged block files.

"""

from __future__ import annotations

from collections.abc import Iterator
from contextlib import contextmanager
from dataclasses import dataclass
from io import BytesIO
import enum
import logging
import struct
import typing as tp


_logger = logging.getLogger(__name__)


HEADER_V6 = b"reMarkable .lines file, version=6          "


class TagType(enum.IntEnum):
    "Tag type representing the type of following data."
    ID = 0xF
    Length4 = 0xC
    Byte8 = 0x8
    Byte4 = 0x4
    Byte1 = 0x1


class UnexpectedBlockError(Exception):
    """Unexpected tag or index in block stream."""


@dataclass(eq=True, order=True, frozen=True)
class CrdtId:
    "An identifier or timestamp."
    part1: int
    part2: int

    def __repr__(self) -> str:
        return f"CrdtId({self.part1}, {self.part2})"


class DataStream:
    """Read basic values from a remarkable v6 file stream."""

    def __init__(self, data: tp.BinaryIO):
        self.data = data

    def tell(self) -> int:
        return self.data.tell()

    def read_header(self) -> None:
        """Read the file header.

        This should be the first call when starting to read a new file.

        """
        header = self.read_bytes(len(HEADER_V6))
        if header != HEADER_V6:
            raise ValueError("Wrong header: %r" % header)

    def write_header(self) -> None:
        """Write the file header.

        This should be the first call when starting to read a new file.

        """
        self.write_bytes(HEADER_V6)

    def check_tag(self, expected_index: int, expected_type: TagType) -> bool:
        """Check that INDEX and TAG_TYPE are next.

        Returns True if the expected index and tag type are found. Does not
        advance the stream.

        """
        pos = self.data.tell()
        try:
            index, tag_type = self._read_tag_values()
            return (index == expected_index) and (tag_type == expected_type)
        except (ValueError, EOFError):
            return False
        finally:
            self.data.seek(pos)  # Go back

    def read_tag(
        self, expected_index: int, expected_type: TagType
    ) -> tuple[int, TagType]:
        """Read a tag from the stream.

        Raise an error if the expected index and tag type are not found, and
        rewind the stream.

        """
        pos = self.data.tell()
        index, tag_type = self._read_tag_values()

        if index != expected_index:
            self.data.seek(pos)  # Go back
            raise UnexpectedBlockError(
                "Expected index %d, got %d, at position %d"
                % (expected_index, index, self.data.tell())
            )

        if tag_type != expected_type:
            self.data.seek(pos)  # Go back
            raise UnexpectedBlockError(
                "Expected tag type %s (0x%X), got 0x%X at position %d"
                % (
                    expected_type.name,
                    expected_type.value,
                    tag_type,
                    self.data.tell(),
                )
            )

        return index, tag_type

    def _read_tag_values(self) -> tuple[int, TagType]:
        """Read tag values from the stream."""

        x = self.read_varuint()

        # First part is an index number that identifies if this is the right
        # data we're expecting
        index = x >> 4

        # Second part is a tag type that identifies what kind of data it is
        tag_type = x & 0xF
        try:
            tag_type = TagType(tag_type)
        except ValueError as e:
            raise ValueError(
                "Bad tag type 0x%X at position %d" % (tag_type, self.data.tell())
            )

        return index, tag_type

    def write_tag(self, index: int, tag_type: TagType):
        """Write a tag to the stream."""
        x = index << 4 | int(tag_type)
        self.write_varuint(x)

    def read_bytes(self, n: int) -> bytes:
        "Read `n` bytes, raising `EOFError` if there are not enough."
        result = self.data.read(n)
        if len(result) != n:
            raise EOFError()
        return result

    def write_bytes(self, b: bytes):
        "Write bytes to underlying stream."
        self.data.write(b)

    def _read_struct(self, pattern: str):
        pattern = "<" + pattern
        n = struct.calcsize(pattern)
        return struct.unpack(pattern, self.read_bytes(n))[0]

    def _write_struct(self, pattern: str, value):
        pattern = "<" + pattern
        self.data.write(struct.pack(pattern, value))

    def read_bool(self) -> bool:
        """Read a bool from the data stream."""
        return self._read_struct("?")

    def read_uint8(self) -> int:
        """Read a uint8 from the data stream."""
        return self._read_struct("B")

    def read_uint16(self) -> int:
        """Read a uint16 from the data stream."""
        return self._read_struct("H")

    def read_uint32(self) -> int:
        """Read a uint32 from the data stream."""
        return self._read_struct("I")

    def read_float32(self) -> float:
        """Read a float32 from the data stream."""
        return self._read_struct("f")

    def read_float64(self) -> float:
        """Read a float64 (double) from the data stream."""
        return self._read_struct("d")

    def read_varuint(self) -> int:
        """Read a varuint from the data stream."""
        shift = 0
        result = 0
        while True:
            i = ord(self.read_bytes(1))
            result |= (i & 0x7F) << shift
            shift += 7
            if not (i & 0x80):
                break
        return result

    def read_crdt_id(self) -> CrdtId:
        # Based on ddvk's reader.go
        # TODO: should be var unit?
        part1 = self.read_uint8()
        part2 = self.read_varuint()
        # result = (part1 << 48) | part2
        return CrdtId(part1, part2)

    def write_bool(self, value: bool):
        """Write a bool to the data stream."""
        self._write_struct("?", value)

    def write_uint8(self, value: int):
        """Write a uint8 to the data stream."""
        return self._write_struct("B", value)

    def write_uint16(self, value: int):
        """Write a uint16 to the data stream."""
        return self._write_struct("H", value)

    def write_uint32(self, value: int):
        """Write a uint32 to the data stream."""
        return self._write_struct("I", value)

    def write_float32(self, value: float):
        """Write a float32 to the data stream."""
        return self._write_struct("f", value)

    def write_float64(self, value: float):
        """Write a float64 (double) to the data stream."""
        return self._write_struct("d", value)

    def write_varuint(self, value: int):
        """Write a varuint to the data stream."""
        if value < 0:
            raise ValueError("value is negative")
        b = bytearray()
        while True:
            to_write = value & 0x7F
            value >>= 7
            if value:
                b.append(to_write | 0x80)
            else:
                b.append(to_write)
                break
        self.data.write(b)

    def write_crdt_id(self, value: CrdtId):
        """Write a `CrdtId` to the data stream."""
        # Based on ddvk's reader.go
        # TODO: should be var unit?
        if value.part1 >= 2**8 or value.part2 >= 2**64:
            raise ValueError("CrdtId too large: %s" % value)
        self.write_uint8(value.part1)
        self.write_varuint(value.part2)
        # result = (part1 << 48) | part2


_T = tp.TypeVar("_T")

# This makes sense to be frozen, since the value should not be changed without
# updating the timestamp.
@dataclass(eq=True, frozen=True)
class LwwValue(tp.Generic[_T]):
    "Container for a last-write-wins value."
    timestamp: CrdtId
    value: _T
