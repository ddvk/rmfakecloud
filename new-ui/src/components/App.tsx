import { Helmet } from 'react-helmet-async'
import { useTranslation } from 'react-i18next'

import { fullSiteTitle } from '../utils/site'

import Uploader from './Uploader'
import FileTreeView from './FileTreeView'

function App() {
  const { t } = useTranslation()

  return (
    <>
      <Helmet>
        <title>{fullSiteTitle(t('nav.documents'))}</title>
      </Helmet>
      <div className="min-h-[calc(100vh-63px)] bg-slate-900 text-neutral-400">
        <div className="mx-auto max-w-4xl">
          <div className="mx-4 py-8">
            <h1 className="mb-8 text-2xl font-semibold text-neutral-200">{t('nav.documents')}</h1>

            <Uploader />

            <FileTreeView />
          </div>
        </div>
      </div>
    </>
  )
}

export default App
