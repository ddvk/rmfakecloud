import Upload from "./Upload";
import Tree from "./Tree";
import { useState, useRef } from "react";
import { Container, Row, Col, Button } from "react-bootstrap";
import { Document, Page } from "react-pdf";
import constants from "../../common/constants";

export default function DocumentList() {
  const [counter, setCounter] = useState(0);
  const [folder, setFolder] = useState("");
  const [file, setFile] = useState(null);
  const [page, setPage] = useState(1);
  const [pages, setPages] = useState(1);
  const containerRef = useRef(null);

  const callback = () => {
    //TODO: really hacky, just to force the tree to update
    setCounter(counter + 1);
  };
  const setParent = (f) => {
    console.log("setting folder" + f);
    setFolder(f);
  };

  const onFileSelected = (f) => {
    setFile(`${constants.ROOT_URL}/documents/${f}`);
  };
  const onLoadSuccess = (pdf) => {
    setPage(1);
    setPages(pdf.numPages);
  };
  const onPrev = () => {
    console.log(containerRef.current.offsetWidth);
    setPage((p) => {
      return p - 1;
    });
  };
  const onNext = () => {
    setPage((p) => {
      return p + 1;
    });
  };

  return (
    <Container fluid>
      <Row>
        <Col md={6}>
          <Upload filesUploaded={callback} uploadFolder={folder} />
          <Tree
            counter={counter}
            onFileSelected={onFileSelected}
            onFolderChanged={setParent}
          />
        </Col>
        <Col md={6} ref={containerRef}>
          <Button onClick={onPrev}>Prev</Button>
          <Button onClick={onNext}>Next</Button>
          <label>
            Page: {page} of {pages}
          </label>
          {file && (
            <Document file={file} onLoadSuccess={onLoadSuccess}>
              <Page
                pageNumber={page}
                width={containerRef.current.offsetWidth - 30}
              />
            </Document>
          )}
        </Col>
      </Row>
    </Container>
  );
}
