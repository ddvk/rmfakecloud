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

interface PageInfo {
  current?: number
  totalPageNum?: number
}

export default function DocumentViewer() {
  const { docId } = useParams()
  const { t } = useTranslation()
  const [metadata, setMetadata] = useState<HashDocMetadata | null>(null)
  const [isLoadingDocument, setIsLoadingDocument] = useState(true)
  const [isDocNotFound, setIsDocNotFound] = useState(false)
  const [loadingProgress, setLoadingProgress] = useState(0)
  const [pageScale, setPageScale] = useState(1)
  const [pageWidth, setPageWidth] = useState<number | null>(null)
  const { width: windowWidth } = useWindowDimensions()
  const [pageInfo, setPageInfo] = useState<PageInfo>({})
  const { current, totalPageNum } = pageInfo

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

  useEffect(() => {
    function handleScroll() {
      const docHeight = document.body.offsetHeight
      const posHeight = window.innerHeight + window.scrollY
      const { current, totalPageNum } = pageInfo

      // scroll to bottom
      if (posHeight >= docHeight - 50 && current && totalPageNum) {
        if (current < totalPageNum) {
          setPageInfo({
            current: current + 1,
            totalPageNum
          })

          return
        }
      }
    }

    window.addEventListener('scroll', handleScroll)

    return () => window.removeEventListener('scroll', handleScroll)
  }, [pageInfo])

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
          setPageInfo({
            totalPageNum: pdf.numPages,
            current: pdf.numPages > 0 ? 1 : 0
          })
        }}
      >
        {Array.from(new Array(totalPageNum), (_el, index) => {
          const page = index + 1

          if (current && page - current <= 2)
            return (
              <Page
                key={`page_${index + 1}`}
                className="my-2 block"
                pageNumber={index + 1}
                scale={pageScale}
                onLoadSuccess={(page) => {
                  setPageWidth(page.width / pageScale)
                }}
              />
            )
        })}
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
