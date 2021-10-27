import React from "react";
import useFetch from "../../hooks/useFetch";
import Spinner from "../Spinner";
import Table from "react-bootstrap/Table";
import { Link } from "react-router-dom";

const userListUrl = "users";

export default function UserList() {
  const { data: userList, error, loading } = useFetch(`${userListUrl}`);

  if (loading) {
    return <Spinner />;
  }
  if (error) {
    return <div>{error.message}</div>;
  }

  if (!userList.length) {
    return <div>No users</div>;
  }

  return (
      <Table className="table-dark">
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
              <td>
                <Link to={`/userList/${x.userid}`}>{x.email}</Link>
              </td>
              <td>{x.Name}</td>
              {/* TODO: format datetime */}
              <td>{x.CreatedAt}</td>
            </tr>
          ))}
        </tbody>
      </Table>
  );
}
