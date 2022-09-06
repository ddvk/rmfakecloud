import { UploadedFile } from 'api/document'
import { useState } from 'react'
import Dropzone from 'react-dropzone'
import { useTranslation } from 'react-i18next'
import { toast } from 'react-toastify'
import Transfer from './transfer'

interface UploadState {
  uploadingOrUploaded: number
  total: number
  current?: UploadedFile
  mode: 'progress_bar' | 'file_selector' | 'show_success'
}

export default function Uploader() {
  const { t } = useTranslation()
  const transfer = new Transfer(1, true)
  const [uploadState, setUploadState] = useState<UploadState>({
    uploadingOrUploaded: 0,
    total: 0,
    mode: 'file_selector'
  })

  transfer.onChange = (file, files) => {
    let uploadingOrUploaded = 0
    let uploadSuccess = true

    files.forEach((file) => {
      if (file.status !== 'uploaded') {
        uploadSuccess = false
      }
      if (file.status !== 'pending') {
        uploadingOrUploaded += 1
      }
    })

    setUploadState({
      uploadingOrUploaded,
      total: files.length,
      current: file,
      mode: uploadSuccess ? 'show_success' : files.length > 0 ? 'progress_bar' : 'file_selector'
    })
  }

  transfer.onDuplicatedFiles = (files) => {
    if (files.length > 0) {
      const fileNames = files
        .map((file) => {
          return file.name
        })
        .join(',')

      toast.warn(`${fileNames} ${t('documents.uploader.duplicated_document')}`, {
        position: 'top-center',
        theme: 'dark'
      })
    }
  }

  return (
    <>
      <Dropzone
        accept={{ 'application/pdf': ['.pdf'], 'application/epub+zip': ['.epub'] }}
        maxSize={25 * 1024 * 1024} // 25MB
        onDrop={(acceptedFiles) => {
          transfer.uploadFiles(acceptedFiles)
        }}
      >
        {({ getRootProps, getInputProps }) => (
          <section>
            <div
              {...getRootProps()}
              className="flex flex-col h-44 md:h-80 rounded border-4 border-dashed border-slate-100/10 items-center justify-center text-slate-100/20 p-4 text-center whitespace-nowrap"
            >
              <input {...getInputProps()} />
              {uploadState.mode === 'file_selector' ? (
                <>
                  <p className="md:text-lg md:mb-1 font-semibold text-ellipsis relative max-w-full overflow-hidden">
                    {t('documents.uploader.title')}
                  </p>
                  <p className="text-xs md:text-base text-ellipsis relative max-w-full overflow-hidden">
                    {t('documents.uploader.subtitle')}
                  </p>
                </>
              ) : uploadState.mode === 'progress_bar' ? (
                <>
                  <p className="md:text-lg md:mb-1 font-semibold text-ellipsis relative max-w-full overflow-hidden">
                    [{uploadState.uploadingOrUploaded}/{uploadState.total}]{' '}
                    {t('documents.uploader.progress_bar.title')}
                  </p>
                  <p className="text-xs md:text-base text-ellipsis relative max-w-full overflow-hidden">
                    {t('documents.uploader.progress_bar.current_file')}(
                    {uploadProgress(uploadState.current)}%){uploadState.current?.file.name}
                  </p>
                </>
              ) : (
                <>
                  <p className="font-semibold text-ellipsis relative max-w-full overflow-hidden md:mb-1">
                    <svg
                      xmlns="http://www.w3.org/2000/svg"
                      fill="none"
                      viewBox="0 0 24 24"
                      strokeWidth={1.5}
                      stroke="currentColor"
                      className="w-12 h-12"
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        d="M9 12.75L11.25 15 15 9.75M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                      />
                    </svg>
                  </p>
                  <p className="text-xs md:text-base text-ellipsis relative max-w-full overflow-hidden">
                    {t('documents.uploader.upload_success', { count: uploadState.total })}
                  </p>
                </>
              )}
            </div>
          </section>
        )}
      </Dropzone>
    </>
  )
}

function uploadProgress(file?: UploadedFile): number {
  if (!file) {
    return 0
  }

  return Math.round((file.uploadedSize / file.file.size) * 100)
}
