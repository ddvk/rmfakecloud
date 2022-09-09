import { Transition } from '@headlessui/react'
import { useState } from 'react'
import { Helmet } from 'react-helmet-async'
import { useTranslation } from 'react-i18next'
import { PulseLoader } from 'react-spinners'

import { newCode } from '../api/auth'
import { fullSiteTitle } from '../utils/site'

function CodeDisplay({ code }: { code: string }) {
  const innerDom = []

  for (let i = 0; i < code.length; i++) {
    innerDom.push(
      <Transition.Child
        key={`char-${i}`}
        as="div"
        className="h-full relative mt-[50%] md:mt-[35%] translate-y-[-50%]"
        enter={`transition-opacity ease-in-out duration-[2s]`}
        enterFrom="opacity-0"
        enterTo="opacity-100"
        leave="transition-opacity duration-[2s]"
        leaveFrom="opacity-100"
        leaveTo="opacity-0"
      >
        {code[i]}
      </Transition.Child>
    )
  }

  return (
    <Transition
      as="div"
      show={true}
      appear={true}
      className="relative min-h-[50vh] flex justify-around text-5xl text-neutral-200 font-bold font-[CascadiaCodePL]"
    >
      {innerDom}
    </Transition>
  )
}

export default function Devices() {
  const { t } = useTranslation()
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [code, setCode] = useState('')
  const [showCode, setShowCode] = useState(false)

  const onClickHandler = () => {
    setIsSubmitting(true)
    setShowCode(false)
    newCode()
      .then((response) => {
        setCode(response.data)
        setShowCode(true)
        return 'ok'
      })
      .catch((error) => {
        throw error
      })
      .finally(() => {
        setIsSubmitting(false)
      })
  }

  return (
    <>
      <Helmet>
        <title>{fullSiteTitle(t('nav.devices'))}</title>
      </Helmet>
      <div className="min-h-[calc(100vh-63px)] bg-slate-900 text-neutral-400 md:min-h-[calc(100vh-57px)]">
        <div className="mx-auto max-w-4xl">
          <div className="mx-4 py-8">
            <h1 className="mb-8 text-2xl font-semibold text-neutral-200">{t('nav.devices')}</h1>
            {showCode ? <CodeDisplay code={code} /> : <></>}
            <div className="fixed bottom-8 w-[calc(100%-32px)] md:mx-auto md:max-w-[calc(896px-32px)]">
              <button
                className="w-full rounded bg-blue-700 py-3 font-bold text-neutral-200 hover:bg-blue-600 focus:outline-none disabled:bg-blue-500"
                type="button"
                disabled={isSubmitting}
                onClick={onClickHandler}
              >
                {isSubmitting ? (
                  <PulseLoader
                    color="#e5e5e5"
                    cssOverride={{ lineHeight: 0, padding: '6px 0' }}
                    size={8}
                    speedMultiplier={0.8}
                  />
                ) : (
                  t('devices.code-btn')
                )}
              </button>
            </div>
          </div>
        </div>
      </div>
    </>
  )
}
