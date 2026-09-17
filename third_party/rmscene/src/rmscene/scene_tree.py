"""Build scene tree structure from block data.

"""

from __future__ import annotations

import logging
import typing as tp

from .tagged_block_common import CrdtId
from .crdt_sequence import CrdtSequenceItem
from . import scene_items as si

if tp.TYPE_CHECKING:
    from .scene_stream import SceneInfo, SceneImageInfoBlock

_logger = logging.getLogger(__name__)


ROOT_ID = CrdtId(0, 1)


class SceneTree:
    def __init__(self):
        self.root = si.Group(ROOT_ID)
        self._node_ids = {self.root.node_id: self.root}
        self.scene_info: tp.Optional["SceneInfo"] = None
        self.root_text: tp.Optional[si.Text] = None
        self.image_info: tp.Optional["SceneImageInfoBlock"] = None

    def __contains__(self, node_id: CrdtId):
        return node_id in self._node_ids

    def __getitem__(self, node_id: CrdtId):
        return self._node_ids[node_id]

    def add_node(self, node_id: CrdtId, parent_id: CrdtId):
        if node_id in self._node_ids:
            raise ValueError("Node %s already in tree" % node_id)
        node = si.Group(node_id)
        self._node_ids[node_id] = node
        # parent = self._node_ids[parent_id]
        # parent.children.add(item)

    def add_item(self, item: CrdtSequenceItem[si.SceneItem], parent_id: CrdtId):
        if parent_id not in self._node_ids:
            raise ValueError("Parent id not known: %s" % parent_id)
        parent = self._node_ids[parent_id]
        parent.children.add(item)

    def walk(self) -> tp.Iterator[si.SceneItem]:
        """Iterate through all leaf items (not groups)."""
        yield from _walk_items(self.root)


def _walk_items(item):
    if isinstance(item, si.Group):
        for child in item.children.values():
            yield from _walk_items(child)
    else:
        yield item
