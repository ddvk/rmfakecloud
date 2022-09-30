import { Transition } from '@headlessui/react'
import { EyeIcon, PencilAltIcon, SaveIcon, TrashIcon } from '@heroicons/react/outline'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Link } from 'react-router-dom'

import { HashDoc } from '../../utils/models'

import RemoveDocDialog from './removeDocDialog'

export interface FileMenuProps {
  doc?: HashDoc | null
  onDocDeleted?: (doc: HashDoc) => void
  onDocEditing?: (doc: HashDoc) => void
}

export default function FileMenu(params: FileMenuProps) {
  const { doc, onDocDeleted } = params
  const { t } = useTranslation()

  const [removingDoc, setRemovingDoc] = useState<HashDoc | null>(null)

  return (
    <>
      <RemoveDocDialog
        doc={removingDoc}
        onDismissDialog={() => {
          setRemovingDoc(null)
        }}
        onDocDeleted={onDocDeleted}
      />
      <Transition
        as="div"
        className={`fixed inset-x-0 bottom-0 z-10 w-full ${removingDoc ? 'md:pr-4' : ''}`}
        enter="transition-translate-y duration-300"
        enterFrom="translate-y-full"
        enterTo="translate-y-0"
        leave="transition-translate-y duration-300"
        leaveFrom="translate-y-0"
        leaveTo="translate-y-full"
        show={new Boolean(doc).valueOf()}
      >
        <div className="-mx-4 max-w-4xl border-t border-slate-100/20 md:mx-auto md:border-none">
          <div className="rounded bg-slate-900 md:border-x md:border-t md:border-slate-100/20 md:mx-4">
            <div className="flex justify-between md:justify-around">
              <Link
                className="basis-1/4 p-4 "
                target="_blank"
                to={`/ui/api/documents/${doc?.id}/`}
              >
                <div className="hover:text-neutral-200">
                  <SaveIcon className="mx-auto mb-1 h-6 w-6" />
                  <p className="text-center text-xs">
                    {t('documents.file_tree_view.menu.download')}
                  </p>
                </div>
              </Link>
              <Link
                className="basis-1/4 p-4 "
                target="_blank"
                to={`/documents/${doc?.id}/viewer`}
              >
                <div className="hover:text-neutral-200">
                  <EyeIcon className="mx-auto mb-1 h-6 w-6" />
                  <p className="text-center text-xs">{t('documents.file_tree_view.menu.view')}</p>
                </div>
              </Link>
              <div
                className="basis-1/4 cursor-pointer p-4 hover:text-neutral-200"
                onClick={() => {
                  const { onDocEditing } = params

                  onDocEditing && doc && onDocEditing(doc)
                }}
              >
                <PencilAltIcon className="mx-auto mb-1 h-6 w-6" />
                <p className="text-center text-xs">{t('documents.file_tree_view.menu.rename')}</p>
              </div>
              <div
                className="basis-1/4 cursor-pointer p-4 hover:text-neutral-200"
                onClick={() => {
                  if (!doc) {
                    return
                  }

                  setRemovingDoc(doc)
                }}
              >
                <TrashIcon className="mx-auto mb-1 h-6 w-6" />
                <p className="text-center text-xs">{t('documents.file_tree_view.menu.remove')}</p>
              </div>
            </div>
          </div>
        </div>
      </Transition>
    </>
  )
}
