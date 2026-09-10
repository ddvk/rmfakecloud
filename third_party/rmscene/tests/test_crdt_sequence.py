from rmscene.crdt_sequence import CrdtSequenceItem, CrdtSequence
from rmscene import CrdtId


def cid(k):
    "Shorthand for making end-markers or IDs with author-id = 1."
    return CrdtId(0 if k == 0 else 1, k)


def make_item(item_id, left_id, right_id, *args):
    "Shorthand for creating ids."
    return CrdtSequenceItem(cid(item_id), cid(left_id), cid(right_id), *args)


def test_empty():
    assert list(CrdtSequence([])) == []


def test_just_one():
    items = [
        make_item(1, 0, 0, 0, "A"),
    ]
    result = list(CrdtSequence(items))
    assert result == [cid(1)]


def test_two():
    items = [
        make_item(1, 0, 0, 0, "A"),
        make_item(2, 1, 0, 0, "B"),
    ]
    result1 = list(CrdtSequence(items))
    result2 = list(CrdtSequence(reversed(items)))
    assert result1 == result2


def test_overlapping():
    # Test concurrent-insert: two different authors insert at the same position
    # (identical left_id and right_id) without knowledge of each other. The
    # higher author ID should come first, matching observed reMarkable device
    # behaviour (verified with test-crdt-ordering.rm: "A12_Z" not "A_12Z").
    A = CrdtSequenceItem(CrdtId(1, 1), CrdtId(0, 0), CrdtId(0, 0), 0, "A")
    Z = CrdtSequenceItem(CrdtId(1, 2), CrdtId(1, 1), CrdtId(0, 0), 0, "Z")
    author2 = CrdtSequenceItem(CrdtId(2, 1), CrdtId(1, 1), CrdtId(1, 2), 0, "12")
    author1 = CrdtSequenceItem(CrdtId(1, 3), CrdtId(1, 1), CrdtId(1, 2), 0, "_")
    items = [A, Z, author2, author1]
    seq = CrdtSequence(items)
    result1 = list(seq)
    result2 = list(CrdtSequence(reversed(items)))
    assert result1 == result2
    assert result1 == [CrdtId(1, 1), CrdtId(2, 1), CrdtId(1, 3), CrdtId(1, 2)]
    assert "".join(seq.values()) == "A12_Z"


def test_overlapping_sorted_by_id():
    items = [
        make_item(8, 0, 0, 0, "A"),
        make_item(9, 8, 0, 0, "B"),
        make_item(3, 0, 0, 0, "C"),
    ]
    result1 = list(CrdtSequence(items))
    result2 = list(CrdtSequence(reversed(items)))
    assert result1 == result2
    assert result1 == [cid(3), cid(8), cid(9)]


def test_unknown_id():
    items = [
        make_item(28, 0, 15, 0, "A"),
        make_item(31, 30, 15, 2, ""),
        make_item(33, 32, 15, 0, "B"),
        make_item(15, 0, 0, 0, "C"),
    ]
    result1 = list(CrdtSequence(items))
    result2 = list(CrdtSequence(reversed(items)))
    assert result1 == result2
    assert result1 == [cid(28), cid(31), cid(33), cid(15)]


def test_unknown_id_at_right():
    items = [
        make_item(14, 0, 0, 0, "A"),
        make_item(19, 14, 15, 0, "V"),
        # item 15 is not defined -- this shouldn't happen, as there should be a
        # tombstone, but don't want to raise errors in this case
    ]
    result = list(CrdtSequence(items))
    assert result == [cid(14), cid(19)]


def test_iterates_in_order():
    # Order should be "AB"
    items = [
        make_item(1, 0, 0, 0, "A"),
        make_item(2, 1, 0, 0, "B"),
    ]

    for test_items in (items, reversed(items)):
        seq = CrdtSequence(test_items)
        assert list(seq.keys()) == [cid(1), cid(2)]
        assert list(seq.values()) == ["A", "B"]
        assert list(seq.items()) == list(zip(seq.keys(), seq.values()))


from hypothesis import given, note, strategies as st
from hypothesis.stateful import Bundle, RuleBasedStateMachine, rule, precondition, invariant


class CrdtSequenceComparison(RuleBasedStateMachine):
    """State machine that adds and deletes text in the sequence"""
    def __init__(self):
        super().__init__()
        # All items, even deleted ones
        self.items: dict[CrdtId, CrdtSequenceItem] = {}
        # List of items corresponding to string
        self.string_items: list[CrdtSequenceItem] = []
        self.string = ""
        self.last_id = 1

    # keys = Bundle("keys")
    # values = Bundle("values")

    # @rule(target=keys, k=st.binary())
    # def add_key(self, k):
    #     return k

    @rule(data=st.data(), c=st.characters())
    def add_char(self, data, c):
        i = data.draw(st.integers(min_value=0, max_value=len(self.string)))
        new_item = CrdtSequenceItem(
            item_id=cid(self.last_id),
            left_id=self.string_items[i-1].item_id if i >= 1 else cid(0),
            right_id=self.string_items[i].item_id if i < len(self.string_items) else cid(0),
            deleted_length=0,
            value=c
        )
        note(f"new_item: {new_item}")
        self.last_id += 1
        self.string = self.string[:i] + c + self.string[i:]
        self.string_items = self.string_items[:i] + [new_item] + self.string_items[i:]
        self.items[new_item.item_id] = new_item

    @rule(data=st.data())
    def add_empty_item(self, data):
        i = data.draw(st.integers(min_value=0, max_value=len(self.string)))
        new_item = CrdtSequenceItem(
            item_id=cid(self.last_id),
            left_id=self.string_items[i-1].item_id if i >= 1 else cid(0),
            right_id=self.string_items[i].item_id if i < len(self.string_items) else cid(0),
            deleted_length=0,
            value=""
        )
        note(f"new_item: {new_item}")
        self.last_id += 1
        self.items[new_item.item_id] = new_item

    @precondition(lambda self: len(self.string) > 0)
    @rule(data=st.data())
    def delete_char(self, data):
        i = data.draw(st.integers(min_value=0, max_value=len(self.string) - 1))
        item_id = self.string_items[i].item_id
        note(f"deleting_item: {item_id}")
        self.string = self.string[:i] + self.string[i+1:]
        self.string_items = self.string_items[:i] + self.string_items[i+1:]
        self.items[item_id].value = ""
        self.items[item_id].deleted_length = 1

    @invariant()
    def values_agree(self):
        seq = CrdtSequence(self.items.values())
        assert "".join(seq.values()) == self.string


TestCrdtSequenceComparison = CrdtSequenceComparison.TestCase
