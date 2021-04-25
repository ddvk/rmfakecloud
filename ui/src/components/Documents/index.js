import Upload from './Upload'
import Tree from './Tree'
import {useState} from 'react';

export default function DocumentList() {

    const [counter, setCounter] = useState(0);

    const callback = () => {
        //TODO: really hacky, just to force the tree to update
        setCounter(counter+1)
    }

    return (
        <>
            <Upload filesUploaded={callback} />
            <Tree counter={counter} />
        </>
    )
}
