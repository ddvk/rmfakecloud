import pytest
from io import BytesIO
from rmscene.tagged_block_common import (
    DataStream
)


@pytest.mark.parametrize(
    "value,hexstr",
    [
        (0, "00"),
        (3, "03"),
    ],
)
def test_write_uint8(value, hexstr):
    buf = BytesIO()
    s = DataStream(buf)
    s.write_uint8(value)
    assert buf.getvalue().hex() == hexstr


@pytest.mark.parametrize(
    "value,hexstr",
    [
        (0x00, "00"),
        (0x03, "03"),
        (0x7f, "7f"),
        (0x8c, "8c01"),
        (0x9c, "9c01"),
        (0x3fff, "ff7f"),
    ],
)
def test_write_varuint(value, hexstr):
    buf = BytesIO()
    s = DataStream(buf)
    s.write_varuint(value)
    assert buf.getvalue().hex() == hexstr
    buf.seek(0)
    assert s.read_varuint() == value
