import React, {useState} from "react";
import useFetch from "../../hooks/useFetch";
import Spinner from "../../components/Spinner";
import {Alert, Button, Card, Container, Modal, Table} from "react-bootstrap";
import IntegrationModal from "./IntegrationModal";
import NewIntegrationModal from "./NewIntegrationModal";
import apiService from "../../services/api.service";
import { toast } from "react-toastify";
const integrationListUrl = "integrations";

const NewIntegration = 1;
const UpdateIntegration = 2;
const Integrations = () => {
  const [index, setIndex] = useState(0);
  const { data: integrationList, error, loading } = useFetch(`${integrationListUrl}`, index);
  const [ state, setState ] = useState({showModal: 0, modalIntegration: null});
  const refresh = () =>{
    setIndex(previous => previous+1)
  }

  function openModal(index: number) {
    if (!integrationList) return;
    let integration = integrationList[index];
    setState({
      showModal: UpdateIntegration,
      modalIntegration: integration,
    });
  }
  function closeModal() {
    setState({
      showModal: 0,
      modalIntegration: null,
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

  const newIntegration = e => {
    setState({
      showModal: NewIntegration,
    });
  }

  const onSave  = () => {
    closeModal();
    refresh();
  }

  const remove = async (e, id, name) => {
    e.preventDefault()
    e.stopPropagation()
    if (!window.confirm(`Are you sure you want to delete integration: ${name}?`))
      return false

    try{
      await apiService.deleteintegration(id)
      refresh()
    } catch(e){
        toast.error('Error:'+ e)
    }
  }

  return (
    <Container>
      <h3>Integrations</h3>
      <Card>
        <Table striped bordered hover className="mb-0">
          <thead>
            <tr>
              <th>#</th>
              <th>IntegrationId</th>
              <th>Name</th>
              <th>Provider</th>
              <th><Button onClick={newIntegration}>New Integration</Button></th>
            </tr>
          </thead>
          <tbody>
            {!integrationList.length && (
              <tr>
                <td colspan="5" class="text-center">No integration</td>
              </tr>
            )}
            {integrationList.map((i, index) => (
              <tr key={i.ID} onClick={() => openModal(index)} style={{ cursor: "pointer" }}>
                <td>{index}</td>
                <td>{i.ID}</td>
                <td>{i.Name}</td>
                <td>{i.Provider}</td>
                <td><Button variant="danger" onClick={(e) => remove(e,i.ID,i.Name)}>Delete</Button></td>
              </tr>
            ))}
          </tbody>
        </Table>
        <Modal show={state.showModal === UpdateIntegration} onHide={closeModal} className="transparent-modal">
          <IntegrationModal integration={state.modalIntegration} onSave={onSave} onClose={closeModal} headerText={`Change Integration: ${state.modalIntegration?.Name}`} />
        </Modal>
        <Modal show={state.showModal === NewIntegration} onHide={closeModal} className="transparent-modal">
          <NewIntegrationModal onSave={onSave} onClose={closeModal} />
        </Modal>
      </Card>
    </Container>
  );
};

export default Integrations;
