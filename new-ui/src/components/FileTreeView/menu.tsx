import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Transition } from '@headlessui/react'
import { toast } from 'react-toastify'
import { Link } from 'react-router-dom'
import { SaveIcon, EyeIcon, PencilAltIcon, TrashIcon } from '@heroicons/react/outline'

import { deleteDocument } from '../../api/document'
import { HashDoc } from '../../utils/models'
import { ConfirmationDialog } from '../Dialog'

export interface FileMenuProps {
  doc?: HashDoc | null
  onDocDeleted?: (doc: HashDoc) => void
  onDocEditing?: (doc: HashDoc) => void
}

export default function FileMenu(params: FileMenuProps) {
  const { doc, onDocDeleted } = params
  const { t } = useTranslation()
  const [isOpenDialog, setIsOpenDialog] = useState(false)
  const [dialogIsLoading, setDialogIsLoading] = useState(false)

  return (
    <>
      <ConfirmationDialog
        content={t('site.dialog.doc_delete_content', { name: doc?.name })}
        isLoading={dialogIsLoading}
        isOpen={isOpenDialog}
        title={t('site.dialog.delete_title')}
        onClose={() => {
          setIsOpenDialog(false)
          setDialogIsLoading(false)
        }}
        onConfirm={() => {
          if (!doc) {
            return
          }

          setDialogIsLoading(true)

          deleteDocument(doc.id)
            .then(() => {
              toast.success(t('notifications.document_deleted'))
              onDocDeleted && onDocDeleted(doc)

              return 'ok'
            })
            .catch((err) => {
              throw err
            })
            .finally(() => {
              setDialogIsLoading(false)
              setIsOpenDialog(false)
            })
        }}
      />
      <Transition
        as="div"
        className="sticky bottom-0 z-10 w-full"
        enter="transition-translate-y duration-300"
        enterFrom="translate-y-full"
        enterTo="translate-y-0"
        leave="transition-translate-y duration-300"
        leaveFrom="translate-y-0"
        leaveTo="translate-y-full"
        show={new Boolean(doc).valueOf()}
      >
        <div className="-mx-4 max-w-4xl border-t border-slate-100/20 md:mx-auto md:border-none">
          <div className="rounded bg-slate-900 md:border-x md:border-t md:border-slate-100/20">
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

                  setIsOpenDialog(true)
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
