import React, { useMemo, useState } from 'react';
import { useDropzone } from 'react-dropzone';

import apiservice from "../../services/api.service";

const baseStyle = {
  flex: 1,
  display: 'flex',
  flexDirection: 'column',
  alignItems: 'center',
  padding: '20px',
  borderWidth: 2,
  borderRadius: 5,
  borderColor: '#fff',
  borderStyle: 'dashed',
  backgroundColor: '#000',
  color: '#fff',
  outline: 'none',
  transition: 'border .24s ease-in-out',
  margin: '50px 0'
};

const activeStyle = {
  borderColor: '#2196f3'
};

const acceptStyle = {
  borderColor: '#00e676'
};

const rejectStyle = {
  borderColor: '#ff1744'
};

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

  const style = useMemo(() => ({
    ...baseStyle,
    ...(isDragActive ? activeStyle : {}),
    ...(isDragAccept ? acceptStyle : {}),
    ...(isDragReject ? rejectStyle : {})
  }), [
    isDragActive,
    isDragReject,
    isDragAccept
  ]);

  if (!uploading) {
    return (
        <div {...getRootProps({ style })}>
          <input {...getInputProps()} />
          <p>Drag 'n' drop some files here, or click to select files to upload</p>
          <p>{lasterror}</p>
        </div>
    )
  } else {
    return <div> Uploading...</div>
  }
}
