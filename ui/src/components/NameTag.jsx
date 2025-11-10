import { BsChevronRight } from "react-icons/bs";
import { Button } from "react-bootstrap";

export default function NameTag({ node, onSelect }) {
    if (node.parent) {
        return (<>
            <NameTag node={node.parent} onSelect={onSelect} />
            {!node.parent.isRoot && <BsChevronRight />}
            <Button variant="outline" onClick={() => onSelect(node)}>{node.data.name}</Button>
        </>)
    }
    // No parent means this is the root - render it
    return <Button variant="outline" onClick={() => onSelect(node)}>{node.data.name}</Button>
}