import React, {useState} from "react";
import useFetch from "../hooks/useFetch";
import Spinner from "./Spinner";
import {Table, Button, Modal, Form} from "react-bootstrap";
import { Link } from "react-router-dom";
import apiService from "../services/api.service";
const userListUrl = "users";

export default function UserList() {
  const [show, setShow] = useState(false);
  const [index, setIndex] = useState(false);
  const { data: userList, error, loading } = useFetch(`${userListUrl}`, index);
  const [formErrors, setFormErrors] = useState({});

  const refresh = () =>{
    setIndex(previous => previous+1)
  }

  const [profileForm, setProfileForm] = useState({
    userid: "",
    email: "",
    password: ""
  });

  if (loading) {
    return <Spinner />;
  }
  if (error) {
    return <div>{error.message}</div>;
  }

  if (!userList.length) {
    return <div>No users</div>;
  }
  const newUser = e => {
    setShow(true)
  }
  const handleClose = e => {
    setShow(false)
  }
  const remove = async id => {
    if (!window.confirm(`Are you sure you want to delete user: ${id}?`))
      return

    try{
      await apiService.deleteuser(id)
      refresh()
    } catch(e){
      //TODO:
    }
  }
  const handleSave = async e => {
    e.preventDefault()
    try {
      await apiService.createuser({
        userid: profileForm.userid,
        email: profileForm.email,
        newPassword: profileForm.newPassword,
      });
      setShow(false)
      refresh()
      
    } catch (e) {
      setFormErrors({ error: e.toString()});
    }
  }
  
  const handleChange = ({ target }) => {
    setProfileForm({ ...profileForm, [target.name]: target.value });
  }

  return (
    <>
      <Table className="table-dark">
        <thead>
          <tr>
            <th>#</th>
            <th>UserId</th>
            <th>Email</th>
            <th>Name</th>
            <th>Created</th>
            <th><Button onClick={newUser}>New User</Button></th>
          </tr>
        </thead>
        <tbody>
          {userList.map((x, index) => (
            <tr key={x.userid}>
              <td>{index}</td>
              <td>
                <Link to={`/users/${x.userid}`}>{x.userid}</Link>
              </td>
              <td>{x.email}</td>
              <td>{x.Name}</td>
              {/* TODO: format datetime */}
              <td>{x.CreatedAt}</td>
              <td><Button variant="danger" onClick={() => remove(x.userid)}>Delete</Button></td>
            </tr>
          ))}
        </tbody>
      </Table>
      <Modal show={show} onHide={handleClose}>
        <Modal.Header closeButton>
          <Modal.Title>Modal heading</Modal.Title>
        </Modal.Header>
        <Modal.Body>
        <Form autocomplete="chrome-off" onSubmit={handleSave}>
          <Form.Label>UserId</Form.Label>
          <Form.Control
            className="font-weight-bold"
            placeholder="userid"
            name="userid"
            value={profileForm.userid}
            onChange={handleChange}
          />
        <Form.Group controlId="formEmail">
          <Form.Label>Email address</Form.Label>
          <Form.Control
            type="email"
            className="font-weight-bold"
            placeholder="Enter email"
            name="email"
            value={profileForm.email}
            onChange={handleChange}
          />
        </Form.Group>
        <Form.Group controlId="formPasswordRepeat">
          <Form.Label>New Password</Form.Label>
          <Form.Control
            type="password"
            placeholder="new password"
            value={profileForm.newPassword}
            name="newPassword"
            onChange={handleChange}
          />
        </Form.Group>
        {formErrors.error && (
          <div className="alert alert-danger">{formErrors.error}</div>
        )}
        <input type="submit" hidden/>
      </Form>
</Modal.Body>
        <Modal.Footer>
          <Button variant="secondary" onClick={handleClose}>
            Close
          </Button>
          <Button variant="primary"  onClick={handleSave}>
            Save Changes
          </Button>
        </Modal.Footer>
      </Modal>
      </>
  );
}
