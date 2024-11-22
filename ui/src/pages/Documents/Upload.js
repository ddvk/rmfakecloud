import { useMemo, useState } from "react";
import { useDropzone } from "react-dropzone";
import Spinner from "react-bootstrap/Spinner";
import { toast } from "react-toastify";
import apiservice from "../../services/api.service";

import styles from "./Documents.module.scss";

async function uploadFilesInBatches(files, uploadFolder, batchSize = 100) {
  const totalFiles = files.length;
  let uploadedCount = 0;

  while (uploadedCount < totalFiles) {
    const batch = files.slice(uploadedCount, uploadedCount + batchSize);
    try {
      await apiservice.upload(uploadFolder, batch);
      uploadedCount += batchSize;
    } catch (e) {
      console.error("Batch upload error:", e);
      throw e;
    }
  }
}

export default function StyledDropzone(props) {
  const [uploading, setUploading] = useState(false);
  const uploadFolder = props.uploadFolder;

  var onDrop = async (acceptedFiles) => {
    try {
      setUploading(true);
      // TODO: add loading and error handling
      await uploadFilesInBatches(acceptedFiles, uploadFolder);
      // await delay(100)
      props.filesUploaded();
    } catch (e) {
      toast.error("upload error" + e.toString());
    } finally {
      setUploading(false);
    }
  };

  const {
    getRootProps,
    getInputProps,
    isDragActive,
    isDragAccept,
    isDragReject,
  } = useDropzone({
    accept: "application/pdf, application/zip, application/epub+zip",
    onDropAccepted: onDrop,
    maxSize: 1 * 1024 * 1024 * 1024, // 1GB
  });

  const className = useMemo(() => {
    return `${styles.upload}
            ${isDragActive ? styles.uploadActive : ""}
            ${isDragAccept ? styles.uploadAccept : ""}
            ${isDragReject ? styles.uploadReject : ""}`;
  }, [isDragActive, isDragReject, isDragAccept]);

  const hint =
    "Drag 'n' drop some files here, or click to select files to upload";
  if (!uploading) {
    return (
      <div {...getRootProps({ className })}>
        <input {...getInputProps()} />
        <p>{hint}</p>
      </div>
    );
  } else {
    return (
      <div>
        <Spinner animation="grow" />
      </div>
    );
  }
}
