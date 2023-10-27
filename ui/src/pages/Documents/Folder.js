import Navbar from 'react-bootstrap/Navbar';
//import apiservice from "../../services/api.service"

export default function Folder({ folder }) {
  return (
    <>
      <Navbar>
        { folder && (<h6>{folder.name}</h6>) }
      </Navbar>

      <div>
        {JSON.stringify(folder)}
      </div>
    </>
  );
}
