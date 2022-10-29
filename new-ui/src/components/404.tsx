import { Helmet } from 'react-helmet-async'
import { useTranslation } from 'react-i18next'
import { Navigate, useMatch } from 'react-router-dom'

import { fullSiteTitle } from '../utils/site'

export default function NotFound() {
  const { t } = useTranslation()

  if (!useMatch('/404')) {
    return (
      <Navigate
        replace={true}
        to="/404"
      />
    )
  }

  return (
    <>
      <Helmet>
        <title>{fullSiteTitle(t('site.titles.404'))}</title>
      </Helmet>
      <div className="flex items-center justify-center text-neutral-400">
        <div className="flex text-6xl">
          <svg
            className="mt-0.5 mr-5 h-14 w-14"
            fill="none"
            stroke="currentColor"
            strokeWidth={1.5}
            viewBox="0 0 24 24"
            xmlns="http://www.w3.org/2000/svg"
          >
            <path
              d="M3 3l8.735 8.735m0 0a.374.374 0 11.53.53m-.53-.53l.53.53m0 0L21 21M14.652 9.348a3.75 3.75 0 010 5.304m2.121-7.425a6.75 6.75 0 010 9.546m2.121-11.667c3.808 3.807 3.808 9.98 0 13.788m-9.546-4.242a3.733 3.733 0 01-1.06-2.122m-1.061 4.243a6.75 6.75 0 01-1.625-6.929m-.496 9.05c-3.068-3.067-3.664-7.67-1.79-11.334M12 12h.008v.008H12V12z"
              strokeLinecap="round"
              strokeLinejoin="round"
            />
          </svg>
          <h1>404</h1>
        </div>
      </div>
    </>
  )
}
