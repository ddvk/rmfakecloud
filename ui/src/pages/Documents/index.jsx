import DocumentTree from "./Tree";
import apiservice from "../../services/api.service"
import { useEffect, useRef, useState } from "react";
import { Container, Row, Col } from "react-bootstrap";
import File from "./File";
import Folder from "./Folder";
import Navbar from 'react-bootstrap/Navbar';
import { BsSearch } from "react-icons/bs";
import Form from 'react-bootstrap/Form';
import Button from 'react-bootstrap/Button';
import InputGroup from 'react-bootstrap/InputGroup';
import { toast } from "react-toastify";
import { useAuthState } from "../../common/useAuthContext";

import styles from "./Documents.module.scss";

export default function DocumentList() {
  const [selected, setSelected] = useState(null);
  const [term, setTerm] = useState("");
  const [showSearch, setShowSearch] = useState(false);
  const [counter, setCounter] = useState(0);
  const [entries, setEntries] = useState([])
  const [initialSelectionSet, setInitialSelectionSet] = useState(false);

  const { state: { user } } = useAuthState();

  const treeRef = useRef(null);
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

  // const findInTree = (id) => {
  //   treeRef.current.openParents(id)
  //   return treeRef.current.get(id)
  // }

  // select from tree. node must extend NodeApi from react-arborist
  const onSelect = (node) => {
    setSelected(node);
    toggleNode(node)
  };

  const onUpdate = () => {
    setCounter(prev => prev+1);
  };

  useEffect(() => {
    if (
      !initialSelectionSet &&
      selected === null &&
      treeRef.current &&
      treeRef.current.root &&
      treeRef.current.root.children[0]
    ) {
      setSelected(treeRef.current.root.children[0]);
      setInitialSelectionSet(true);
    }
  }, [entries, selected, initialSelectionSet]);

	useEffect(() => {
		const loadDocs = async () => {
			const { Trash, Entries } = await apiservice.listDocument()

			const root = {
				id: "root",
				name: "My Files",
				isFolder: true,
				icon: "device",
				children: Entries,
			}
			const trash = {
				id: "trash",
				name: "Trash",
				isFolder: true,
				icon: "trash",
				children: Trash,
			}
			setEntries([root, trash]);
		}

		loadDocs().catch(e => toast.error(e));
	},[counter])

  return (
    <Container fluid>
        <Row className="mt-2">
          <Col md={4}>
            <Navbar>
              <div className={`${styles.stretch} ${styles.userid}`}>{user.UserID}</div>
              <Button variant="outline" onClick={() => { setShowSearch(!showSearch); setTerm("") }}><BsSearch/></Button>
            </Navbar>

            {showSearch && <div>
              <InputGroup className="mb-3">
                <InputGroup.Text>
                  <BsSearch />
                </InputGroup.Text>

                <Form.Control autoFocus size="sm" type="text" value={term} onChange={(e) => { setTerm(e.currentTarget.value); }} />
              </InputGroup>
            </div>}

            <DocumentTree selection={selected} onSelect={onSelect} treeRef={treeRef} term={term} entries={entries} />
          </Col>
          <Col md={8}>
            {selected && selected.isLeaf && <File file={selected} onSelect={onSelect} />}
            {selected && !selected.isLeaf && <Folder selection={selected} onSelect={onSelect} onUpdate={onUpdate} counter={counter} />}
          </Col>
        </Row>
    </Container>
  );
}
