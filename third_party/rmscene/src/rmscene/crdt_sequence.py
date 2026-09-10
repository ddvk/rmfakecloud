"""Data structure representing CRDT sequence.

"""

import logging
import typing as tp
from typing import Iterable
from collections import defaultdict
from dataclasses import dataclass

from .tagged_block_common import CrdtId

_logger = logging.getLogger(__name__)


# If the type constraint is for a CrdtSequenceItem[Superclass], then a
# CrdtSequenceItem[Subclass] would do, so it is covariant.

_T = tp.TypeVar("_T", covariant=True)


@dataclass
class CrdtSequenceItem(tp.Generic[_T]):
    item_id: CrdtId
    left_id: CrdtId
    right_id: CrdtId
    deleted_length: int
    value: _T


# As a mutable container, CrdtSequence is invariant.
_Ti = tp.TypeVar("_Ti", covariant=False)


class CrdtSequence(tp.Generic[_Ti]):
    """Ordered CRDT Sequence container.

    The Sequence contains `CrdtSequenceItem`s, each of which has an ID and
    left/right IDs establishing a partial order.

    Iterating through the `CrdtSequence` yields IDs following this order.

    """

    def __init__(self, items=None):
        if items is None:
            items = []
        self._items = {item.item_id: item for item in items}

    def __eq__(self, other):
        if isinstance(other, CrdtSequence):
            return self._items == other._items
        if isinstance(other, (list, tuple)):
            return self == CrdtSequence(other)
        raise NotImplemented

    def __repr__(self):
        return "CrdtSequence(%s)" % (", ".join(str(i) for i in self._items.values()))

    ## Access values, in order

    def __iter__(self) -> tp.Iterator[CrdtId]:
        """Return ids in order"""
        yield from toposort_items(self._items.values())

    def keys(self) -> list[CrdtId]:
        """Return CrdtIds in order."""
        return list(self)

    def values(self) -> list[_Ti]:
        """Return list of sorted values."""
        return [self[item_id] for item_id in self]

    def items(self) -> Iterable[tuple[CrdtId, _Ti]]:
        """Return list of sorted key, value pairs."""
        return [(item_id, self[item_id]) for item_id in self]

    def __getitem__(self, key: CrdtId) -> _Ti:
        """Return item with key"""
        return self._items[key].value

    ## Access SequenceItems

    def sequence_items(self) -> list[CrdtSequenceItem[_Ti]]:
        """Iterate through CrdtSequenceItems."""
        return list(self._items.values())

    ## Modify sequence

    def add(self, item: CrdtSequenceItem[_Ti]):
        if item.item_id in self._items:
            raise ValueError("Already have item %s" % item.item_id)
        self._items[item.item_id] = item


END_MARKER = CrdtId(0, 0)


def toposort_items(items: Iterable[CrdtSequenceItem]) -> Iterable[CrdtId]:
    """Sort SequenceItems based on left and right ids.

    Returns `CrdtId`s in the sorted order.

    This uses Kahn's algorithm for topological sorting. Standard Kahn's
    algorithm is O(V + E), i.e. linear. We use a heap to maintain
    deterministic ordering when multiple items become ready simultaneously
    (matching the previous implementation's use of sorted()). This adds
    a log factor, making the overall complexity O(V log V + E), which
    simplifies to O(V log V) since each item has exactly 2 edges.
    """
    import heapq

    item_dict = {}
    for item in items:
        item_dict[item.item_id] = item
    if not item_dict:
        return  # nothing to do

    def _side_id(item, side):
        side_id = getattr(item, f"{side}_id")
        if side_id == END_MARKER or side_id not in item_dict:
            if side_id != END_MARKER:
                _logger.debug("Ignoring unknown %s_id %s of %s", side, side_id, item)
            return "__start" if side == "left" else "__end"
        else:
            return side_id

    # Build dependency graph with O(1) lookups
    # in_degree[item] = number of unprocessed dependencies
    # dependents[item] = list of items that depend on this item
    in_degree = defaultdict(int)
    dependents = defaultdict(list)

    # Track all nodes we need to process
    all_nodes = set()
    all_nodes.add("__start")
    all_nodes.add("__end")

    for item in item_dict.values():
        item_id = item.item_id
        left_id = _side_id(item, "left")
        right_id = _side_id(item, "right")

        all_nodes.add(item_id)
        all_nodes.add(left_id)
        all_nodes.add(right_id)

        # item depends on left_id (item comes after left_id)
        in_degree[item_id] += 1
        dependents[left_id].append(item_id)

        # right_id depends on item (right_id comes after item)
        in_degree[right_id] += 1
        dependents[item_id].append(right_id)

    # Initialize in_degree for nodes with no incoming edges
    for node in all_nodes:
        if node not in in_degree:
            in_degree[node] = 0

    # Use a heap to yield items in deterministic order when multiple items
    # become ready simultaneously (i.e. concurrent inserts at the same
    # position). For concurrent items with the same left_id/right_id, the
    # reMarkable places higher author IDs first.
    def sort_key(node):
        if node == "__start":
            return (0, 0, 0)
        elif node == "__end":
            return (2, 0, 0)
        else:
            return (1, -node.part1, node.part2)

    # Start with nodes that have no dependencies
    ready = []
    for node in all_nodes:
        if in_degree[node] == 0:
            heapq.heappush(ready, (sort_key(node), node))

    while ready:
        _, node = heapq.heappop(ready)

        # Yield if it's an actual item (not __start or __end)
        if node in item_dict:
            yield node

        # Stop if we've reached __end
        if node == "__end":
            break

        # Process dependents
        for dependent in dependents[node]:
            in_degree[dependent] -= 1
            if in_degree[dependent] == 0:
                heapq.heappush(ready, (sort_key(dependent), dependent))

    # Check for cycles (items with remaining dependencies)
    remaining = [n for n in all_nodes if in_degree[n] > 0 and n != "__end"]
    if remaining:
        raise ValueError("cyclic dependency")
