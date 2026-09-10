import pytest
from io import BytesIO
from rmscene import TaggedBlockReader, UnexpectedBlockError, BlockOverflowError, CrdtId


def stream(hex_string: str) -> TaggedBlockReader:
    data = bytes.fromhex(hex_string)
    return TaggedBlockReader(BytesIO(data))


class TestBlock:
    TEST_DATA = (
        "04000000"  # length
        "00010205"  # header
        "ff000000"  # data
        "00000000"  # more data to test overflow
    )

    def test_read_block(self):
        s = stream(self.TEST_DATA)
        with s.read_block() as block_info:
            assert block_info is not None
            assert block_info.size == 4
            assert block_info.block_type == 5
            assert block_info.min_version == 1
            assert block_info.current_version == 2
            assert block_info.offset == 8

            value = s.data.read_uint32()
            assert value == 0xFF

    def test_bytes_remaining(self):
        s = stream(self.TEST_DATA)
        with s.read_block():
            assert s.bytes_remaining_in_block() == 4
            s.data.read_uint32()
            assert s.bytes_remaining_in_block() == 0

    def test_error_on_overflow(self):
        s = stream(self.TEST_DATA)
        with pytest.raises(BlockOverflowError):
            with s.read_block():
                s.data.read_uint32()
                s.data.read_uint32()

    def test_warns_if_incomplete(self, caplog):
        s = stream(self.TEST_DATA)
        with s.read_block():
            pass  # not reading anything
        assert "not been read" in caplog.records[0].message

    def test_skips_to_end_of_block_if_not_all_read(self):
        s = stream(self.TEST_DATA)
        with s.read_block():
            assert s.data.tell() == 8
            # not reading anything
        assert s.data.tell() == 12

    def test_error_if_already_in_block(self):
        s = stream(self.TEST_DATA)
        with s.read_block():
            with pytest.raises(UnexpectedBlockError):
                with s.read_block():
                    pass


class TestSubblock:
    TEST_DATA = (
        "5c" "04000000" "ff000000" "00000000"  # tag  # length  # data
    )  # more data to test overflow

    def test_read_subblock(self):
        s = stream(self.TEST_DATA)
        with s.read_subblock(5) as block_info:
            assert block_info.size == 4
            value = s.data.read_uint32()
            assert value == 0xFF

    def test_has_subblock(self):
        s = stream(self.TEST_DATA)
        # Rewinds so can be called repeatedly
        assert s.has_subblock(5) == True
        assert s.has_subblock(5) == True
        assert s.has_subblock(4) == False
        assert s.has_subblock(5) == True

    def test_error_on_wrong_index(self):
        s = stream(self.TEST_DATA)
        with pytest.raises(UnexpectedBlockError):
            with s.read_subblock(3):
                pass

    def test_error_on_wrong_tag(self):
        s = stream(self.TEST_DATA)
        with pytest.raises(UnexpectedBlockError):
            s.read_int(5)

    def test_error_on_wrong_index_does_not_consume_tag(self):
        s = stream(self.TEST_DATA)
        with pytest.raises(UnexpectedBlockError):
            with s.read_subblock(3):
                pass

        # We can still read the next block after catching the error.
        with s.read_subblock(5) as block_info:
            assert block_info.size == 4

    def test_error_on_overflow(self):
        s = stream(self.TEST_DATA)
        with pytest.raises(BlockOverflowError):
            with s.read_subblock(5):
                s.data.read_uint32()
                s.data.read_uint32()

    def test_warns_if_incomplete(self, caplog):
        s = stream(self.TEST_DATA)
        with s.read_subblock(5):
            pass  # not reading anything
        assert "not been read" in caplog.records[0].message


def test_has_subblock_returns_False_with_bad_data():
    s = stream("1d000000")
    assert s.has_subblock(1) == False
    assert s.has_subblock(6) == False
    assert s.data.tell() == 0


def test_has_subblock_returns_False_at_end_of_file():
    s = stream("")
    assert s.has_subblock(2) == False


def test_has_subblock_checks_for_end_of_block():
    # See https://github.com/ricklupton/rmscene/issues/17#issuecomment-1701071477
    #
    # Construct some potentially confusing data -- the 0x2c is the start of the
    # next block, but if we don't take care, `has_subblock(2)` could see it as a
    # subblock instead.
    data_hex = """
    03000000 00010103
    1f 0219
    2c000000 00010100
    """

    s = stream(data_hex)
    with s.read_block():
        assert s.read_id(1)
        assert s.has_subblock(2) == False


def test_read_int():
    s = stream("34abcd0000")
    assert s.read_int(3) == 0xCDAB


def test_read_int_wrong_index():
    s = stream("34abcd0000")
    with pytest.raises(UnexpectedBlockError):
        s.read_int(2)


def test_read_int_default_value():
    s = stream("34abcd0000")
    assert s.read_int_optional(2, -1) == -1
    assert s.read_int_optional(3) == 0xCDAB
    assert s.read_int_optional(4) == None


def test_read_lww_string():
    s = stream("1c0d000000" "1f0101" "2c05000000" "0301616263")
    lww = s.read_lww_string(1)
    assert lww.timestamp == CrdtId(1, 1)
    assert lww.value == "abc"


def test_read_string_ascii():
    s = stream("1c05000000" "0301616263")
    result = s.read_string(1)
    assert result == "abc"


def test_read_string_utf():
    s = stream("1c05000000" "030161c397")
    result = s.read_string(1)
    assert result == "a×"
