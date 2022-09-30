import { Helmet } from 'react-helmet-async'
import { useTranslation } from 'react-i18next'
import { useState } from 'react'

import { fullSiteTitle } from '../utils/site'

import Uploader from './Uploader'
import FileTreeView from './FileTreeView'

function App() {
  const { t } = useTranslation()
  const [docsReloadCnt, setDocsReloadCnt] = useState(0)

  return (
    <>
      <Helmet>
        <title>{fullSiteTitle(t('nav.documents'))}</title>
      </Helmet>
      <div className="min-h-[calc(100vh-63px)] bg-slate-900 text-neutral-400 md:min-h-[calc(100vh-57px)]">
        <div className="mx-auto max-w-4xl">
          <div className="relative overflow-hidden py-8">
            <h1 className="mx-4 mb-8 text-2xl font-semibold text-neutral-200">
              {t('nav.documents')}
            </h1>

            <Uploader
              onFilesUploaded={() => {
                setDocsReloadCnt(docsReloadCnt + 1)
              }}
            />

            <FileTreeView reloadCnt={docsReloadCnt} />
          </div>
        </div>
      </div>
    </>
  )
}

export default App
