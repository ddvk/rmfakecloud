import Upload from './Upload'
import Tree from './Tree'
import {useState} from 'react';

export default function DocumentList() {

    const [counter, setCounter] = useState(0);
    const [folder, setFolder] = useState("");

    const callback = () => {
        //TODO: really hacky, just to force the tree to update
        setCounter(counter+1)
    }
    const setParent = f => {
        console.log("setting folder"+ f)
        setFolder(f)
    }

    return (
        <>
            <Upload filesUploaded={callback} uploadFolder={folder} />
            <Tree counter={counter} onFolderChanged={setParent}/>
             
        </>
    )
}
