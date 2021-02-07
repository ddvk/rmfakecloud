import React from "react";

import Row from "react-bootstrap/Row";
import useFetch from "../../hooks/useFetch";

import Table from "react-bootstrap/Table";

const userListUrl = "users";

export default function UserList() {
  const { data: userList, error, loading } = useFetch(`${userListUrl}`);

  if (loading) {
    return <div>Loading...</div>;
  }
  if (error) {
    return <div>{error.message}</div>;
  }

  if (!userList.length) {
    return <div>No users</div>;
  }

  debugger;

  return (
    <Row>
      <Table>
        <thead>
          <tr>
            <th>#</th>
            <th>Email</th>
            <th>Name</th>
            <th>Created</th>
          </tr>
        </thead>
        <tbody>
          {userList.map((x, index) => (
            <tr key={x.userid}>
              <td>{index}</td>
              <td>{x.email}</td>
              <td>{x.Name}</td>
              {/* TODO: format datetime */}
              <td>{x.CreatedAt}</td>
            </tr>
          ))}
        </tbody>
      </Table>
    </Row>
  );
}
