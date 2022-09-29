import { Fragment } from 'react'
import { Transition, Dialog } from '@headlessui/react'
import { useTranslation } from 'react-i18next'
import { PulseLoader } from 'react-spinners'

export interface ConfirmationDialogProps {
  isOpen: boolean
  isLoading?: boolean
  onClose: () => void
  onConfirm: () => void
  title: string
  content?: string
}

export function ConfirmationDialog({ isOpen, onClose, ...params }: ConfirmationDialogProps) {
  const { title, content, onConfirm, isLoading } = params
  const { t } = useTranslation()

  return (
    <Transition
      appear
      as={Fragment}
      show={isOpen}
    >
      <Dialog
        as="div"
        className="relative z-50"
        onClose={() => onClose()}
      >
        <Transition.Child
          as={Fragment}
          enter="ease-out duration-300"
          enterFrom="opacity-0"
          enterTo="opacity-100"
          leave="ease-in duration-200"
          leaveFrom="opacity-100"
          leaveTo="opacity-0"
        >
          <div className="fixed inset-0 bg-black/50 backdrop-blur" />
        </Transition.Child>

        <div className="fixed inset-0 overflow-y-auto">
          <div className="flex min-h-full items-center justify-center p-4 text-center">
            <Transition.Child
              as={Fragment}
              enter="ease-out duration-300"
              enterFrom="opacity-0 scale-95"
              enterTo="opacity-100 scale-100"
              leave="ease-in duration-200"
              leaveFrom="opacity-100 scale-100"
              leaveTo="opacity-0 scale-95"
            >
              <Dialog.Panel
                as="div"
                className="w-full max-w-md overflow-hidden rounded-lg bg-slate-800 px-4 py-6 text-left align-middle text-neutral-400"
              >
                <Dialog.Title
                  as="h3"
                  className="text-lg font-bold"
                >
                  {title}
                </Dialog.Title>

                {content ? (
                  <div className="mt-2">
                    <p className="text-sm text-neutral-400">{content}</p>
                  </div>
                ) : (
                  <></>
                )}

                <div className="mt-6 flex">
                  <button
                    className="mr-4 basis-1/2 rounded border border-slate-600 py-2 hover:border-slate-800 hover:bg-slate-700 focus-visible:outline-none"
                    type="button"
                    onClick={() => {
                      onClose()
                    }}
                  >
                    {t('site.dialog.cancel-btn')}
                  </button>
                  <button
                    className="basis-1/2 rounded border border-red-900 bg-red-800 py-2 text-neutral-400 hover:border-red-800 hover:bg-red-700 focus-visible:outline-none disabled:border-red-900/30 disabled:bg-red-800/30"
                    disabled={isLoading}
                    type="button"
                    onClick={() => {
                      onConfirm()
                    }}
                  >
                    {isLoading ? (
                      <PulseLoader
                        color="#e5e5e5"
                        size={8}
                        speedMultiplier={0.8}
                      />
                    ) : (
                      t('site.dialog.confirm-btn')
                    )}
                  </button>
                </div>
              </Dialog.Panel>
            </Transition.Child>
          </div>
        </div>
      </Dialog>
    </Transition>
  )
}
