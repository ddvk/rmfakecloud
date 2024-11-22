import {useState} from "react";
import { toast } from "react-toastify";
import { Alert, Button, Modal, Table } from "react-bootstrap";

import useFetch from "../../hooks/useFetch";
import Spinner from "../../components/Spinner";
import UserProfileModal from "./UserProfileModal";
import NewUserModal from "./NewUserModal";
import apiService from "../../services/api.service";
import { formatDate } from "../../common/date";

const userListUrl = "users";

const NewUser = 1;
const UpdateUser = 2;

export default function UserList() {
  const [index, setIndex] = useState(false);
  const { data: userList, error, loading } = useFetch(`${userListUrl}`, index);
  const [ state, setState ] = useState({showModal: 0, modalUser: null});
  const refresh = () =>{
    setIndex(previous => previous+1)
  }

  function openModal(index) {
    let user = userList[index];
    setState({
      showModal: UpdateUser,
      modalUser: user,
    });
  }

  function closeModal() {
    setState({
      showModal: 0,
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
  const newUser = e => {
    setState({
      showModal: NewUser,
    });
  }

  const onSave  = () => {
    closeModal();
    refresh();
  }

  const remove = async (e, id) => {
    e.preventDefault()
    e.stopPropagation()
    if (!window.confirm(`Are you sure you want to delete user: ${id}?`))
      return false

    try{
      await apiService.deleteuser(id)
      refresh()
    } catch(e){
      toast.error('Error:'+ e)
    }
  }
  // const handleSave = async e => {
  //   e.preventDefault()
  //   try {
  //     await apiService.createuser({
  //       userid: profileForm.userid,
  //       email: profileForm.email,
  //       newPassword: profileForm.password,
  //     });
  //     hide()
  //     refresh()
  //
  //   } catch (e) {
  //     setFormErrors({ error: e.toString()});
  //   }
  // }
  //
  // const handleChange = ({ target }) => {
  //   setProfileForm({ ...profileForm, [target.name]: target.value });
  // }

  return (
    <>
      <h3>Users</h3>
      <Table striped bordered hover>
        <thead>
          <tr>
            <th>#</th>
            <th>UserId</th>
            <th>Email</th>
            <th>Name</th>
            <th>Role</th>
            <th>Created At</th>
            <th><Button onClick={newUser}>New User</Button></th>
          </tr>
        </thead>
        <tbody>
          {userList.map((x, index) => (
            <tr key={x.userid} onClick={() => openModal(index)} style={{ cursor: "pointer" }}>
              <td>{index}</td>
              <td>{x.userid}</td>
              <td>{x.email}</td>
              <td>{x.Name}</td>
              <td>{x.isAdmin && "admin"}</td>
              <td>{formatDate(x.CreatedAt)}</td>
              <td><Button variant="danger" onClick={(e) => remove(e,x.userid)}>Delete</Button></td>
            </tr>
          ))}
        </tbody>
      </Table>

      <Modal show={state.showModal === UpdateUser} onHide={closeModal} className="transparent-modal">
        <UserProfileModal user={state.modalUser} onSave={onSave} onClose={closeModal} headerText={`Change User Email/Password: ${state.modalUser?.userid}`} />
      </Modal>
      <Modal show={state.showModal === NewUser} onHide={closeModal} className="transparent-modal">
        <NewUserModal onSave={onSave} onClose={closeModal} />
      </Modal>
    </>
  );
}
