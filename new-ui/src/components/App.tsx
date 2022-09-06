import { Helmet } from 'react-helmet-async'
import { useTranslation } from 'react-i18next'

import { fullSiteTitle } from '../utils/site'
import { listDocuments } from '../api'

import Uploader from './Uploader'

function App() {
  const { t } = useTranslation()

  listDocuments()
    .then((response) => console.log(response))
    .catch((err) => console.error(err))

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
          </div>
        </div>
      </div>
    </>
  )
}

export default App
