import FileList from "./FileList";
import Navbar from 'react-bootstrap/Navbar';
import { Button } from "react-bootstrap";
import { BsChevronRight } from "react-icons/bs";
//import apiservice from "../../services/api.service"

export default function Folder({ folder, onSelect }) {

  const NameTag = ({ node }) => {
    if (node.parent) {
      return (<>
        <NameTag node={node.parent} />
        { !node.parent.isRoot && <BsChevronRight /> }
        <Button variant="outline" onClick={() => onSelect(node.id)}>{node.data.name}</Button>
      </>)
    }
    return <></>
  }

  const { data } = folder;

  return (
    <>
      <Navbar >
        { folder && (<div><NameTag node={folder} /></div>) }
      </Navbar>
      <FileList files={data.children} onSelect={onSelect} />
    </>
  );
}
