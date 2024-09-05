import React, { useMemo, useState } from 'react';
import { useDropzone } from 'react-dropzone';

import apiservice from "../../services/api.service";

import styles from "./Home.module.scss";

export default function StyledDropzone(props) {
  const [uploading, setUploading] = useState(false);
  const [lasterror, setLastError] = useState();
  const uploadFolder = props.uploadFolder;
  var onDrop = async (acceptedFiles) => {
    try {
      setUploading(true);
      await apiservice.upload(uploadFolder, acceptedFiles)
      setLastError(null)
      props.filesUploaded()
    } catch (e) {
      setLastError(e)
      console.error(e)
    }
    finally{
      setUploading(false);
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
    return `${styles.base} 
            ${isDragActive ? styles.active : ''} 
            ${isDragAccept ? styles.accept : ''} 
            ${isDragReject ? styles.reject : ''}`
  }, [isDragActive, isDragReject, isDragAccept])

  const hint = "Drag 'n' drop some files here, or click to select files to upload"
  const wasError = lasterror !== undefined && lasterror !== ""

  if (!uploading) {
    return (
        <div {...getRootProps({ className })}>
          <input {...getInputProps()} />
          <p>{wasError ? lasterror : hint}</p>
        </div>
    )
  } else {
    return <div><p>Uploading...</p></div>
  }
}
