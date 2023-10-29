import Tree from "./Tree";
import { useState } from "react";
import { Container, Row, Col } from "react-bootstrap";
import File from "./File";
import Folder from "./Folder";
import './Documents.scss';
import Navbar from 'react-bootstrap/Navbar';
import { BsSearch } from "react-icons/bs";
import Form from 'react-bootstrap/Form';
import Button from 'react-bootstrap/Button';
import InputGroup from 'react-bootstrap/InputGroup';

export default function DocumentList() {
  const [selectedId, setSelectedId] = useState("root");
  const [selected, setSelected] = useState(null);
  const [term, setTerm] = useState("");
  const [showSearch, setShowSearch] = useState(false);

  const isFolder = selected && selected.data && selected.data.isFolder;

  const onSelect = (node) => {
    setSelected(node);
  };

  const onSelectById = (id) => {
    setSelectedId(id);
  };

  const onCloseFile = () => {
    setSelectedId(selected.parent.id);
  }

  return (
    <Container>
      <Row className="mt-2">
        <Col md={4}>
          <Navbar>
            <h6 style={{ flex: 1}}>My Documents</h6>
            <h6>
              <Button variant="outline" onClick={() => { setShowSearch(!showSearch); setTerm("") }}><BsSearch/></Button>
            </h6>
          </Navbar>

          {showSearch && <div>
            <InputGroup className="mb-3">
              <InputGroup.Text>
                <BsSearch />
              </InputGroup.Text>

              <Form.Control autoFocus size="sm" type="text" value={term} onChange={(e) => { setTerm(e.currentTarget.value); }} />
            </InputGroup>
          </div>}

          <Tree selectedId={selectedId} onSelect={onSelect} term={term} />
        </Col>
        <Col md={8}>
          {selected && !isFolder && <File file={selected} onClose={onCloseFile} />}
          {selected && isFolder && <Folder folder={selected} onSelect={onSelectById} />}
        </Col>
      </Row>
    </Container>
  );
}
