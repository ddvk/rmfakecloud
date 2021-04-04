import React from "react";

import Row from "react-bootstrap/Row";
import Card from "react-bootstrap/Card";
import useFetch from "../../hooks/useFetch";

//const listUrl = '/document-storage/json/2/docs';
const documentListUrl = "list";

export default function FileListFunctional() {
  //const skuRef = useRef();
  //const { id } = useParams();
  //const navigate = useNavigate();

  const { data: documentList, error, loading } = useFetch(`${documentListUrl}`);

  if (loading) {
    return <div>Loading...</div>;
  }
  if (error) {
    return <div>{error.message}</div>;
  }

  if (!documentList.length) {
    return <div>No documents</div>;
  }

  return (
    <Row>
      {documentList.map((x) => (
        <Card key={x.ID} style={{ width: "20rem" }} className="m-3">
          <Card.Img variant="top" src={x.ImageUrl}></Card.Img>
          <Card.Body>
            <Card.Title>{x.Name}</Card.Title>
            <Card.Text>{x.ParentId}</Card.Text>
          </Card.Body>
        </Card>
      ))}
    </Row>
  );
}
