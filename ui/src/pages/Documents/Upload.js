import React, { useMemo, useState } from 'react';
import { useDropzone } from 'react-dropzone';

import apiservice from "../../services/api.service";

import styles from "./Documents.module.scss";

export default function StyledDropzone(props) {
  const [uploading, setUploading] = useState(false);
  const [lasterror, setLastError] = useState();
  const uploadFolder = props.uploadFolder;

  var onDrop = async (acceptedFiles) => {
    try {
      setUploading(true);
      await apiservice.upload(uploadFolder, acceptedFiles)
      setLastError(null)
    } catch (e) {
      setLastError(e)
      console.error(e)
    }
    finally{
      setUploading(false);
      props.filesUploaded()
    }
    console.log('done')
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
  const wasError = lasterror !== undefined && lasterror !== null && lasterror !== ""

  if (!uploading) {
    return (
        <div {...getRootProps({ className })}>
          <input {...getInputProps()} />
          <p>{wasError ? lasterror : hint}</p>
        </div>
    )
  } else {
    return <div><Spinner animation="grow" /></div>
  }
}
