import React, {useState} from "react";
import useFetch from "../hooks/useFetch";
import Spinner from "./Spinner";
import { formatDate } from "../common/date";
import {Alert, Button, Card, Modal, Table} from "react-bootstrap";
import UserProfileModal from "./UserProfileModal";

const userListUrl = "users";

export default function UserList() {
  const { data: userList, error, loading } = useFetch(`${userListUrl}`);
  const [ state, setState ] = useState({showModal: false, modalUser: []});

  function openModal(userid) {
    setState({
      showModal: true,
      modalUser: userid,
    });
  }

  function closeModal() {
    setState({
      showModal: false,
      modalUser: null,
    });
  }

  if (loading) {
    return <Spinner />
  }

  if (error) {
    return (
        <Alert variant="danger">
            <Alert.Heading>An Error Occurred</Alert.Heading>
            {`Error ${error.status}: ${error.statusText}`}
        </Alert>
    );
  }

  if (!userList.length) {
    return <div>No users</div>;
  }

  return (
    <Card
      bg="dark"
      text="white"
    >
      <Card.Header>User List</Card.Header>
      <Table striped bordered hover className="table-dark">
        <thead>
        <tr>
          <th>#</th>
          <th>Email</th>
          <th>Name</th>
          <th>Created At</th>
        </tr>
        </thead>
        <tbody>
          {userList.map((x, index) => (
            <tr key={x.userid} onClick={() => openModal(x.userid)} style={{ cursor: "pointer" }}>
              <td>{index}</td>
              <td>{x.email}</td>
              <td>{x.Name}</td>
              <td>{formatDate(x.CreatedAt)}</td>
            </tr>
          ))}
        </tbody>
      </Table>
      <Modal show={state.showModal} onHide={closeModal} className="transparent-modal">
        <Card
          bg="dark"
          text="white"
        >
          <Card.Header>
            <span>User Management: '{state.modalUser}'</span>
          </Card.Header>
          <Card.Body>
            {state.showModal && <UserProfileModal userid={state.modalUser} />}
          </Card.Body>
          <Card.Footer style={{ display: "flex", justifyContent: "end" }}>
            <Button onClick={closeModal}>Close</Button>
          </Card.Footer>
        </Card>
      </Modal>
    </Card>
  );
}
