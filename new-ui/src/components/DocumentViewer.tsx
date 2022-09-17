import { useEffect, useState } from 'react'
import { Navigate, useParams } from 'react-router-dom'
import { AxiosError } from 'axios'
import { StatusCodes } from 'http-status-codes'
import { Document, Page, pdfjs } from 'react-pdf'
import { PulseLoader } from 'react-spinners'
import { Helmet } from 'react-helmet-async'
import { useTranslation } from 'react-i18next'
import 'react-pdf/dist/esm/Page/AnnotationLayer.css'
import 'react-pdf/dist/esm/Page/TextLayer.css'

import { fullSiteTitle } from '../utils/site'
import useWindowDimensions from '../utils/hooks'
import { HashDocMetadata } from '../utils/models'
import { getMetadata } from '../api/document'

export default function DocumentViewer() {
  const { docId } = useParams()
  const { t } = useTranslation()
  const [metadata, setMetadata] = useState<HashDocMetadata | null>(null)
  const [isLoadingDocument, setIsLoadingDocument] = useState(true)
  const [isDocNotFound, setIsDocNotFound] = useState(false)
  const [numPages, setNumPages] = useState<number | null>(null)
  const [loadingProgress, setLoadingProgress] = useState(0)
  const [pageScale, setPageScale] = useState(1)
  const [pageWidth, setPageWidth] = useState<number | null>(null)
  const { width: windowWidth } = useWindowDimensions()

  useEffect(() => {
    if (!pageWidth) {
      return
    }
    if (windowWidth > pageWidth) {
      setPageScale(1)

      return
    }

    setPageScale((windowWidth - 32) / pageWidth)
  }, [windowWidth, pageWidth])

  pdfjs.GlobalWorkerOptions.workerSrc = '/lib/pdf.worker.min.js'

  useEffect(() => {
    if (!docId) {
      setIsDocNotFound(true)

      return
    }

    getMetadata(docId)
      .then((response) => {
        const md = response.data as HashDocMetadata

        if (!md) {
          setIsDocNotFound(true)

          return false
        }

        setMetadata(md)

        return true
      })
      .catch((err: AxiosError) => {
        if (err.response && err.response.status === StatusCodes.NOT_FOUND) {
          setIsDocNotFound(true)
        }
      })
  }, [docId])

  return isDocNotFound ? (
    <Navigate
      replace={true}
      to="/404"
    />
  ) : (
    <>
      <Helmet>
        <title>{fullSiteTitle(metadata?.VissibleName)}</title>
      </Helmet>
      <Document
        className={`mx-auto px-4 ${isLoadingDocument ? 'flex items-center' : ''}`}
        error={<p className="font-bold text-red-900">{t('document_viewer.pdf.load_error')}</p>}
        externalLinkTarget="_blank"
        file={`/ui/api/documents/${docId}`}
        loading={<DocumentViewerLoader progress={loadingProgress} />}
        options={{ cMapUrl: '/lib/cmaps/', cMapPacked: true }}
        onLoadError={(err) => {
          // eslint-disable-next-line no-console
          console.error(err)
        }}
        onLoadProgress={({ loaded, total }) => {
          setLoadingProgress(Math.floor((loaded / total) * 100))
        }}
        onLoadSuccess={(pdf) => {
          setIsLoadingDocument(false)
          setNumPages(pdf.numPages)
        }}
      >
        {Array.from(new Array(numPages), (_el, index) => (
          <Page
            key={`page_${index + 1}`}
            className="my-4 block"
            pageNumber={index + 1}
            scale={pageScale}
            onLoadSuccess={(page) => {
              setPageWidth(page.width)
            }}
          />
        ))}
      </Document>
    </>
  )
}

function DocumentViewerLoader({ progress }: { progress?: number }) {
  return (
    <>
      <PulseLoader
        color="#e5e5e5"
        cssOverride={{ padding: '6px 0', display: 'block' }}
        size={8}
        speedMultiplier={0.8}
      />
      {progress !== undefined ? (
        <p className="text-center font-bold text-neutral-400">{progress}%</p>
      ) : (
        <></>
      )}
    </>
  )
}
