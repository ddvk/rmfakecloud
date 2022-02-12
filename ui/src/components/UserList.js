import React from "react";
import useFetch from "../hooks/useFetch";
import Spinner from "./Spinner";
import { Link } from "react-router-dom";
import { formatDate } from "../common/date";
import { Alert, Card, Table } from "react-bootstrap";

const userListUrl = "users";

export default function UserList() {
  const { data: userList, error, loading } = useFetch(`${userListUrl}`);

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
      <Table className="table-dark">
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
            <tr key={x.userid}>
              <td>{index}</td>
              <td>
                <Link to={`/userList/${x.userid}`}>{x.email}</Link>
              </td>
              <td>{x.Name}</td>
              <td>{formatDate(x.CreatedAt)}</td>
            </tr>
          ))}
        </tbody>
      </Table>
    </Card>
  );
}
