import { useMemo, useState } from 'react';
import { useDropzone } from 'react-dropzone';
import Spinner from 'react-bootstrap/Spinner';
import { toast } from "react-toastify";
import apiservice from "../../services/api.service";

import styles from "./Documents.module.scss";

export default function StyledDropzone(props) {
  const [uploading, setUploading] = useState(false);
  const uploadFolder = props.uploadFolder;

  var onDrop = async (acceptedFiles) => {
    try {
      setUploading(true);
      // TODO: add loading and error handling
      await apiservice.upload(uploadFolder, acceptedFiles)
      // await delay(100)
      props.filesUploaded();
    } catch (e) {
	  toast.error('upload error' + e.toString())
    }
    finally{
      setUploading(false);
    }
  }

  const {
    getRootProps,
    getInputProps,
    isDragActive,
    isDragAccept,
    isDragReject
  } = useDropzone({ accept: 'application/pdf, application/zip, application/epub+zip', onDropAccepted: onDrop });

  const className = useMemo(() => {
    return `${styles.upload} 
            ${isDragActive ? styles.uploadActive : ''} 
            ${isDragAccept ? styles.uploadAccept : ''} 
            ${isDragReject ? styles.uploadReject : ''}`
  }, [isDragActive, isDragReject, isDragAccept])

  const hint = "Drag 'n' drop some files here, or click to select files to upload"
  if (!uploading) {
    return (
        <div {...getRootProps({ className })}>
          <input {...getInputProps()} />
          <p>{ hint}</p>
        </div>
    )
  } else {
    return <div><Spinner animation="grow" /></div>
  }
}
