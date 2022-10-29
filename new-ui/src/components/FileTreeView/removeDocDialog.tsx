import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'react-toastify'

import { deleteDocument } from '../../api/document'
import { HashDoc } from '../../utils/models'
import { ConfirmationDialog } from '../Dialog'

interface RemoveDocDialogProps {
  doc: HashDoc | null
  onDocDeleted?: (doc: HashDoc) => void
  onDismissDialog?: () => void
}

export default function RemoveDocDialog({
  doc,
  onDocDeleted,
  onDismissDialog
}: RemoveDocDialogProps) {
  const [isOpen, setIsOpen] = useState(false)
  const [isLoading, setIsLoading] = useState(false)

  const { t } = useTranslation()

  useEffect(() => {
    setIsOpen(Boolean(doc))
  }, [doc])

  const type = doc?.type || 'CollectionType'

  return (
    <ConfirmationDialog
      content={t(`site.dialog.${type}.doc_delete_content`, { name: doc?.name })}
      isLoading={isLoading}
      isOpen={isOpen}
      title={t(`site.dialog.${type}.delete_title`)}
      onClose={() => {
        setIsLoading(false)
        setIsOpen(false)
        setTimeout(() => {
          onDismissDialog && onDismissDialog()
        }, 200)
      }}
      onConfirm={() => {
        if (!doc) {
          return
        }

        if (type === 'CollectionType' && doc.children !== undefined && doc.children.length > 0) {
          setIsOpen(false)
          setTimeout(() => {
            onDismissDialog && onDismissDialog()
          }, 200)
          toast.error(t('notifications.CollectionType.nonempty'))

          return
        }

        setIsLoading(true)

        deleteDocument(doc.id)
          .then(() => {
            toast.success(t(`notifications.${type}.document_deleted`))
            onDocDeleted && onDocDeleted(doc)

            return 'ok'
          })
          .catch((err) => {
            throw err
          })
          .finally(() => {
            setIsLoading(false)
            setIsOpen(false)
          })
      }}
    />
  )
}
