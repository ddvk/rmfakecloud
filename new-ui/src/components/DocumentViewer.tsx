import { useEffect, useState } from 'react'
import { Navigate, useParams } from 'react-router-dom'
import { AxiosError } from 'axios'
import { StatusCodes } from 'http-status-codes'
import { Document, Page, pdfjs } from 'react-pdf'
import { PulseLoader } from 'react-spinners'
import { Helmet } from 'react-helmet-async'
import { useTranslation } from 'react-i18next'
import { List } from 'react-virtualized'
import 'react-pdf/dist/esm/Page/AnnotationLayer.css'
import 'react-pdf/dist/esm/Page/TextLayer.css'

import { fullSiteTitle } from '../utils/site'
import useWindowDimensions from '../utils/hooks'
import { HashDocMetadata } from '../utils/models'
import { getMetadata } from '../api/document'

interface PageInfo {
  totalPageNum?: number
  originalWidth?: number
  originalHeight?: number
  scale: number
  isLoadingDocument?: boolean
}

export default function DocumentViewer() {
  const { docId } = useParams()
  const { t } = useTranslation()
  const [metadata, setMetadata] = useState<HashDocMetadata | null>(null)
  const [isDocNotFound, setIsDocNotFound] = useState(false)
  const [loadingProgress, setLoadingProgress] = useState(0)
  const { width: windowWidth } = useWindowDimensions()
  const [pageInfo, setPageInfo] = useState<PageInfo>({ scale: 1, isLoadingDocument: true })
  const {
    totalPageNum,
    scale: pageScale,
    originalWidth: pageWidth,
    originalHeight: pageHeight,
    isLoadingDocument
  } = pageInfo

  useEffect(() => {
    if (!pageWidth) {
      return
    }
    if (windowWidth > pageWidth) {
      setPageInfo((prev) => {
        return {
          ...prev,
          scale: 1,
          isLoadingDocument: false
        }
      })

      return
    }

    setPageInfo((prev) => {
      return {
        ...prev,
        scale: (windowWidth - 32) / pageWidth,
        isLoadingDocument: false
      }
    })
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
          pdf
            .getPage(1)
            .then((page) => {
              const [x1, y1, x2, y2] = page.view

              setPageInfo({
                totalPageNum: pdf.numPages,
                originalHeight: y2 - y1,
                originalWidth: x2 - x1,
                scale: (() => {
                  const pageWidth = x2 - x1

                  if (windowWidth > pageWidth) {
                    return 1
                  }

                  return (windowWidth - 32) / pageWidth
                })()
              })

              return true
            })
            .catch((err) => {
              // eslint-disable-next-line no-console
              console.error(err)
            })
        }}
      >
        {!isLoadingDocument && pageHeight && totalPageNum && pageWidth && (
          <List
            height={pageHeight * pageScale * totalPageNum + 16 * (totalPageNum + 1)}
            rowCount={totalPageNum}
            rowHeight={pageHeight * pageScale + 16}
            rowRenderer={({ key, index, style }) => (
              <div
                key={key}
                className="my-4"
                style={style}
              >
                <Page
                  pageNumber={index + 1}
                  scale={pageScale}
                />
              </div>
            )}
            width={pageWidth * pageScale}
          />
        )}
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
