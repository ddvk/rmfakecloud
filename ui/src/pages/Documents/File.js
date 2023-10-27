import { useState } from "react";
import { Button, ButtonGroup } from "react-bootstrap";
import { Document, Page } from "react-pdf";
import Navbar from 'react-bootstrap/Navbar';
import { FaChevronRight, FaChevronLeft, } from "react-icons/fa6";
import { AiOutlineDownload } from "react-icons/ai";

import apiservice from "../../services/api.service"

export default function FileViewer({ file }) {
  const [page, setPage] = useState(1);
  const [pages, setPages] = useState(1);

  const onLoadSuccess = (pdf) => {
    setPage(1);
    setPages(pdf.numPages);
  };
  const onPrev = () => {
    setPage((p) => {
      return Math.max(p - 1, 1);
    });
  };
  const onNext = () => {
    setPage((p) => {
      return Math.min(p + 1, pages);
    });
  };

  const onDownloadClick = () => {
    //setDownloadError(null)
    //const {id, name} = dwn
    apiservice.download(file.id)
      .then(blob => {
        var url = window.URL.createObjectURL(blob)
        var a = document.createElement('a')
        a.href = url
        a.download = file.id + '.pdf'
        document.body.appendChild(a)
        a.click()
        a.remove()
      })
      .catch(e => {
        //setDownloadError('cant download ' + e)
      })

  }

  return (
    <>
      <Navbar>
        { file && (<span>{file.name}</span>) }
        {pages > 1 && (
        <div>
          <ButtonGroup aria-label="Basic example">
            <Button size="sm" variant="secondary" onClick={onPrev}><FaChevronLeft /></Button>
            <Button size="sm" variant="secondary" onClick={onNext}><FaChevronRight /></Button>
          </ButtonGroup>
          <span style={{ margin: '0 10px' }}>
            Page: {page} of {pages}
          </span>
        </div>
        )}
        <div style={{ flex: 1 }}></div>

        <Button size="sm" onClick={onDownloadClick}>
          <AiOutlineDownload />
        </Button>

      </Navbar>

      {file && (
        <div style={{ height: '100vh' }}>
        <Document file={file.downloadUrl} onLoadSuccess={onLoadSuccess}>
          <Page pageNumber={page} />
        </Document>
        </div>
      )}
    </>
  );
}
