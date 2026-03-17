import DocumentTree from "./Tree";
import apiservice from "../../services/api.service"
import { useEffect, useRef, useState } from "react";
import { useParams, useHistory } from "react-router-dom";
import { Container, Row, Col } from "react-bootstrap";
import File from "./File";
import Folder from "./Folder";
import TemplateViewer from "./TemplateViewer";
import MethodViewer from "./MethodViewer";
import EpubViewer from "./EpubViewer";
import Navbar from 'react-bootstrap/Navbar';
import { BsSearch } from "react-icons/bs";
import Form from 'react-bootstrap/Form';
import Button from 'react-bootstrap/Button';
import InputGroup from 'react-bootstrap/InputGroup';
import { toast } from "react-toastify";
import { useAuthState } from "../../common/useAuthContext";

import styles from "./Documents.module.scss";

function isEpub(data) {
  if (!data) return false;
  if (data.type === "epub") return true;
  const name = (data.name || "").toLowerCase();
  return name.endsWith(".epub");
}

export default function DocumentList() {
  const [selected, setSelected] = useState(null);
  const [term, setTerm] = useState("");
  const [showSearch, setShowSearch] = useState(false);
  const [counter, setCounter] = useState(0);
  const [entries, setEntries] = useState([])
  const [initialSelectionSet, setInitialSelectionSet] = useState(false);
  const [treeHeight, setTreeHeight] = useState(700);

  const { itemId } = useParams();
  const history = useHistory();
  const { state: { user } } = useAuthState();

  const treeRef = useRef(null);
  const treeContainerRef = useRef(null);
  const lastSelectedId = useRef(null);

  useEffect(() => {
    lastSelectedId.current = selected?.id || null;
  }, [selected]);

  useEffect(() => {
    if (lastSelectedId.current && treeRef.current && typeof treeRef.current.get === 'function') {
      const node = treeRef.current.get(lastSelectedId.current);
      if (node && node !== selected) {
        setSelected(node);
      }
    }
  }, [entries]);

  const toggleNode = (node) => {
    if (node == null) {
      return
    }
    if (typeof node.toggle !== 'function') {
      return
    }

    node.toggle()
  }

  // select from tree. node must extend NodeApi from react-arborist
  const onSelect = (node) => {
    setSelected(node);
    toggleNode(node);

    // Update URL with selected item ID
    if (node && node.id) {
      // Don't add root, templates, methods, or trash to URL, keep as /documents
      if (node.id === 'root' || node.id === 'trash' || node.id === 'templates' || node.id === 'methods') {
        history.push('/documents');
      } else {
        history.push(`/documents/${node.id}`);
      }
    }
  };

  const onUpdate = () => {
    setCounter(prev => prev+1);
  };

  useEffect(() => {
    // Only auto-select first item if there's no itemId in URL
    if (
      !initialSelectionSet &&
      !itemId &&
      selected === null &&
      treeRef.current &&
      treeRef.current.root &&
      treeRef.current.root.children[0]
    ) {
      setSelected(treeRef.current.root.children[0]);
      setInitialSelectionSet(true);
    }
  }, [entries, selected, initialSelectionSet, itemId]);

  useEffect(() => {
    const resizeObserver = new ResizeObserver((event) => {
      setTreeHeight(event[0].contentBoxSize[0].blockSize);
    });

    if (treeContainerRef.current) {
      resizeObserver.observe(treeContainerRef.current);
    }

    return () => {
      resizeObserver.disconnect();
    };
  }, []);

	useEffect(() => {
		const loadDocs = async () => {
			const { Trash, Entries, Templates, Methods } = await apiservice.listDocument()

			const root = {
				id: "root",
				name: "My Files",
				isFolder: true,
				icon: "device",
				children: Entries || [],
			}
			const templates = {
				id: "templates",
				name: "Templates",
				isFolder: true,
				icon: "templates",
				children: Templates?.[0]?.children ?? [],
			}
			const methods = {
				id: "methods",
				name: "rm Methods",
				isFolder: true,
				icon: "methods",
				children: Methods?.[0]?.children ?? [],
			}
			const trash = {
				id: "trash",
				name: "Trash",
				isFolder: true,
				icon: "trash",
				children: Trash || [],
			}
			setEntries([root, templates, methods, trash]);
		}

		loadDocs().catch(e => toast.error(e));
	},[counter])

  // Helper function to recursively search for an item by ID in the tree
  // Returns both the item and its parent chain
  const findItemInEntries = (entries, targetId, parent = null) => {
    for (const entry of entries) {
      if (entry.id === targetId) {
        return { item: entry, parent };
      }
      if (entry.children && entry.children.length > 0) {
        const found = findItemInEntries(entry.children, targetId, entry);
        if (found) return found;
      }
    }
    return null;
  };

  // Helper to build parent chain for breadcrumb
  const buildParentChain = (parentItem) => {
    if (!parentItem) return null;

    const parentNode = {
      id: parentItem.id,
      data: parentItem,
      isLeaf: !parentItem.isFolder,
      isRoot: parentItem.id === 'root' || parentItem.id === 'trash' || parentItem.id === 'templates' || parentItem.id === 'methods',
      // Add a dummy toggle function for compatibility
      toggle: () => {},
    };

    // If this parent is not root/trash, try to find its parent
    if (parentItem.id !== 'root' && parentItem.id !== 'trash' && parentItem.id !== 'templates' && parentItem.id !== 'methods') {
      const grandparentResult = findItemInEntries(entries, parentItem.id);
      if (grandparentResult && grandparentResult.parent) {
        parentNode.parent = buildParentChain(grandparentResult.parent);
      }
    } else {
      // This is root, templates, methods, or trash - add the internal react-arborist root above it
      parentNode.parent = {
        id: '__REACT_ARBORIST_INTERNAL_ROOT__',
        data: { id: '__REACT_ARBORIST_INTERNAL_ROOT__', name: '' },
        isLeaf: false,
        parent: null,
      };
    }

    return parentNode;
  };

  // Handle URL navigation: restore selection from URL parameter
  useEffect(() => {
    // Only proceed if we have entries and an itemId
    if (!entries.length || !itemId || initialSelectionSet) {
      return;
    }

    // Find the item in our data
    const result = findItemInEntries(entries, itemId);

    if (!result) {
      // Item doesn't exist in our data
      toast.warning(`Item not found, returning to root`);
      history.push('/documents');
      return;
    }

    const { item: foundItem, parent: parentItem } = result;

    // Create a pseudo-node object that matches what onSelect expects
    // React-arborist wraps the data, so the node has both top-level properties
    // and a 'data' property containing the actual item
    const pseudoNode = {
      id: foundItem.id,
      data: foundItem,
      isLeaf: !foundItem.isFolder,
      children: (foundItem.children || []).map(child => ({
        id: child.id,
        data: child,
        isLeaf: !child.isFolder,
      })),
      parent: parentItem ? buildParentChain(parentItem) : null,
      isRoot: foundItem.id === 'root' || foundItem.id === 'trash' || foundItem.id === 'templates' || foundItem.id === 'methods',
    };

    // Set the selection directly
    setSelected(pseudoNode);
    setInitialSelectionSet(true);

    // Try to open parent folders in the tree if possible
    if (treeRef.current && typeof treeRef.current.openParents === 'function') {
      // Give tree a moment to render, then open parents
      setTimeout(() => {
        if (treeRef.current && typeof treeRef.current.openParents === 'function') {
          treeRef.current.openParents(itemId);
        }
      }, 100);
    }
  }, [entries, itemId, initialSelectionSet]);

  return (
    <Container fluid style={{height: "100%", display: "flex", flexDirection: "column", overflow: "hidden", padding: "25px 0 20px 0"}}>
        <Row style={{flex: "1 1 auto", minHeight: 0}}>
          <Col md={4} style={{display: "flex", flexDirection: "column", height: "100%"}}>
            <Navbar style={{flexShrink: 0}}>
              <div className={`${styles.stretch} ${styles.userid}`}>{user.UserID}</div>
              <Button variant="outline" onClick={() => { setShowSearch(!showSearch); setTerm("") }}><BsSearch/></Button>
            </Navbar>

            {showSearch && <div style={{flexShrink: 0}}>
              <InputGroup className="mb-3">
                <InputGroup.Text>
                  <BsSearch />
                </InputGroup.Text>

                <Form.Control autoFocus size="sm" type="text" value={term} onChange={(e) => { setTerm(e.currentTarget.value); }} />
              </InputGroup>
            </div>}

            <div ref={treeContainerRef} className={styles.treeContainer} style={{flex: "1 1 auto", minHeight: 0, overflow: "auto"}}>
              <DocumentTree selection={selected} onSelect={onSelect} treeRef={treeRef} term={term} entries={entries} height={treeHeight} />
            </div>
          </Col>
          <Col md={8} style={{display: "flex", flexDirection: "column", height: "100%"}}>
            <div style={{flex: "1 1 auto", minHeight: 0, overflow: "auto"}}>
              {selected && selected.isLeaf && selected.data?.type === "template" && <TemplateViewer file={selected} onSelect={onSelect} />}
              {selected && selected.isLeaf && selected.data?.type === "method" && <MethodViewer file={selected} onSelect={onSelect} />}
              {selected && selected.isLeaf && isEpub(selected.data) && <EpubViewer file={selected} onSelect={onSelect} />}
              {selected && selected.isLeaf && !isEpub(selected.data) && selected.data?.type !== "template" && selected.data?.type !== "method" && <File file={selected} onSelect={onSelect} />}
              {selected && !selected.isLeaf && <Folder selection={selected} onSelect={onSelect} onUpdate={onUpdate} counter={counter} />}
            </div>
          </Col>
        </Row>
    </Container>
  );
}
