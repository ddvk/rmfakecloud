import pytest
from io import BytesIO
from rmscene import (
    TaggedBlockReader,
    TaggedBlockWriter,
    CrdtId,
    LwwValue,
    UnexpectedBlockError,
)


def test_write_id_zero():
    buf = BytesIO()
    s = TaggedBlockWriter(buf)
    s.write_id(3, CrdtId(0, 0))
    assert buf.getvalue().hex() == "3f0000"


def test_write_int():
    buf = BytesIO()
    s = TaggedBlockWriter(buf)
    s.write_int(3, 0xCDAB)
    assert buf.getvalue().hex() == "34abcd0000"


def test_write_block():
    buf = BytesIO()
    s = TaggedBlockWriter(buf)
    with s.write_block(5, 1, 2):
        s.write_int(3, 0x1234)
    assert buf.getvalue().hex() == "05000000000102053434120000"


def test_write_block_error_if_nested():
    buf = BytesIO()
    s = TaggedBlockWriter(buf)
    with pytest.raises(UnexpectedBlockError):
        with s.write_block(5, 1, 2):
            with s.write_block(4, 1, 1):
                s.write_int(3, 0x1234)


def test_write_block_error_recovery():
    buf = BytesIO()
    s = TaggedBlockWriter(buf)
    try:
        with s.write_block(5, 1, 2):
            s.write_int(3, 0x1234)
            raise Exception
    except:
        pass

    # The subblock redirection should have recovered
    s.write_bool(7, True)
    assert buf.getvalue().hex() == "7101"


def test_write_subblock():
    buf = BytesIO()
    s = TaggedBlockWriter(buf)
    with s.write_subblock(2):
        s.write_int(3, 0x1234)
    assert buf.getvalue().hex() == "2c050000003434120000"


def test_write_subblock_nested():
    buf = BytesIO()
    s = TaggedBlockWriter(buf)
    with s.write_subblock(1):
        with s.write_subblock(2):
            s.write_int(3, 0x1234)
    assert buf.getvalue().hex() == "1c0a0000002c050000003434120000"


def test_write_subblock_error_recovery():
    buf = BytesIO()
    s = TaggedBlockWriter(buf)
    try:
        with s.write_subblock(2):
            s.write_int(3, 0x1234)
            raise Exception
    except:
        pass

    # The subblock redirection should have recovered
    s.write_bool(7, True)
    assert buf.getvalue().hex() == "7101"


@pytest.mark.parametrize(
    "data_type,value",
    [
        ("id", CrdtId(1, 5)),
        ("id", CrdtId(0, 0)),
        ("bool", True),
        ("bool", False),
        ("byte", 7),
        ("int", 45),
        ("float", 8.0),
        ("double", 5.4),
        ("lww_bool", LwwValue(CrdtId(1, 4), True)),
        ("lww_byte", LwwValue(CrdtId(1, 4), 7)),
        ("lww_float", LwwValue(CrdtId(1, 4), 8.0)),
        ("lww_id", LwwValue(CrdtId(1, 4), CrdtId(1, 5))),
        ("lww_string", LwwValue(CrdtId(1, 4), "hello")),
        ("string", "abc"),
        ("string", "a×"),
    ],
)
@pytest.mark.parametrize("index", [0, 3, 20])
def test_values_roundtrip(data_type, value, index):
    buf = BytesIO()
    writer = TaggedBlockWriter(buf)
    reader = TaggedBlockReader(buf)
    write = getattr(writer, f"write_{data_type}")
    read = getattr(reader, f"read_{data_type}")
    write(index, value)
    buf.seek(0)
    assert read(index) == value
