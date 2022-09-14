import { v4 as uuidv4 } from 'uuid'
import {
  DocumentTextIcon,
  SaveIcon,
  EyeIcon,
  PencilAltIcon,
  TrashIcon,
  FolderIcon
} from '@heroicons/react/outline'
import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Transition } from '@headlessui/react'
import { toast } from 'react-toastify'

import { EpubIcon, PDFIcon } from '../../utils/icons'
import { listDocuments } from '../../api'
import { deleteDocument } from '../../api/document'
import { ConfirmationDialog } from '../Dialog'
import { HashDoc } from '../../utils/models'

type HashDocElementProp = {
  doc: HashDoc
  onClickDoc?: (doc: HashDoc) => void
  onClickDir?: (dir: HashDoc) => void
} & React.HTMLAttributes<HTMLDivElement>

function FileElement(params: HashDocElementProp) {
  const { doc, onClickDoc, className, ...remainParams } = params

  return (
    <div
      className={`flex cursor-pointer py-6 ${className || ''}`}
      {...remainParams}
      onClick={() => {
        onClickDoc && onClickDoc(doc)
      }}
    >
      {doc.extension === '.epub' ? (
        <EpubIcon className="top-[-1px] mr-2 h-6 w-6 shrink-0" />
      ) : doc.extension === '.pdf' ? (
        <PDFIcon className="top-[-1px] mr-2 h-6 w-6 shrink-0" />
      ) : (
        <DocumentTextIcon className="top-[-1px] mr-2 h-6 w-6 shrink-0" />
      )}
      <p className="max-w-[calc(100%-28px)] overflow-hidden text-ellipsis whitespace-nowrap leading-6">
        {doc.name}
      </p>
    </div>
  )
}

function DirElement(params: HashDocElementProp) {
  const { doc, onClickDoc, className, ...remainParams } = params

  return (
    <div
      className={`flex cursor-pointer py-6 ${className || ''}`}
      {...remainParams}
      onClick={() => {
        onClickDoc && onClickDoc(doc)
      }}
    >
      <FolderIcon className="top-[-1px] mr-2 h-6 w-6 shrink-0" />
      <p className="max-w-[calc(100%-28px)] overflow-hidden text-ellipsis whitespace-nowrap leading-6">
        {doc.name}
      </p>
    </div>
  )
}

function TreeElement(params: HashDocElementProp) {
  const { doc, ...remainParams } = params

  if (doc.type === 'DocumentType') {
    return (
      <FileElement
        doc={doc}
        {...remainParams}
      />
    )
  }

  if (undefined !== doc.children) {
    return (
      <DirElement
        doc={doc}
        {...remainParams}
      />
    )
  }

  return <></>
}

interface FileMenuProps {
  doc?: HashDoc | null
  onDocDeleted?: (doc: HashDoc) => void
  onDocEditing?: (doc: HashDoc) => void
}

function FileMenu(params: FileMenuProps) {
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
        className="fixed inset-x-0 bottom-0 z-10"
        enter="transition-translate-y duration-300"
        enterFrom="translate-y-full"
        enterTo="translate-y-0"
        leave="transition-translate-y duration-300"
        leaveFrom="translate-y-0"
        leaveTo="translate-y-full"
        show={new Boolean(doc).valueOf()}
      >
        <div className="mx-auto max-w-4xl border-t border-slate-100/20 md:border-none">
          <div className="rounded bg-slate-900 md:mx-4 md:border-x md:border-t md:border-slate-100/20">
            <div className="flex justify-between md:justify-around">
              <div className="basis-1/4 cursor-pointer p-4 hover:text-neutral-200">
                <SaveIcon className="mx-auto mb-1 h-6 w-6" />
                <p className="text-center text-xs">{t('documents.file_tree_view.menu.download')}</p>
              </div>
              <div className="basis-1/4 cursor-pointer p-4 hover:text-neutral-200">
                <EyeIcon className="mx-auto mb-1 h-6 w-6" />
                <p className="text-center text-xs">{t('documents.file_tree_view.menu.view')}</p>
              </div>
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

interface BreakcrumbItem {
  title: string
  docs: HashDoc[]
}

function Breadcrumbs(params: {
  items: BreakcrumbItem[]
  className?: string
  onClickBreadcrumb?: (item: BreakcrumbItem, index: number) => void
}) {
  const { items, className, onClickBreadcrumb } = params

  const innerDom = items.map((item, i) => {
    return (
      <li
        key={`breakcrumb-item-${i}`}
        className="cursor-pointer after:mx-2 after:text-neutral-400 after:content-['>'] last:after:hidden"
        onClick={(e) => {
          e.preventDefault()
          if (onClickBreadcrumb) {
            onClickBreadcrumb(item, i)
          }
        }}
      >
        {item.title}
      </li>
    )
  })

  return (
    <div className={className}>
      <ul className="flex text-sm font-semibold text-sky-600">{innerDom}</ul>
    </div>
  )
}

export default function FileTreeView() {
  const { t } = useTranslation()
  const [docs, setDocs] = useState<HashDoc[]>([])
  const [selected, setSelected] = useState<HashDoc | null>(null)
  const [breakcrumbItems, setBreakcrumbItems] = useState<BreakcrumbItem[]>([])

  function pushd(dir: HashDoc) {
    if (dir.type === 'DocumentType' || undefined === dir.children) {
      return
    }

    const items: BreakcrumbItem[] = [...breakcrumbItems, { title: dir.name, docs: dir.children }]

    setBreakcrumbItems(items)
    setSelected(null)
    setDocs(dir.children)
  }

  function popd(toIndex?: number) {
    toIndex = toIndex || 0
    const items: BreakcrumbItem[] = [...breakcrumbItems]

    while (items.length > toIndex + 1) {
      items.pop()
    }

    const item = items[items.length - 1]

    setBreakcrumbItems(items)
    setSelected(null)
    setDocs(item.docs)
  }

  useEffect(() => {
    listDocuments()
      .then((response) => {
        const data = response.data as { Entries: HashDoc[]; Trash: HashDoc[] }

        setDocs(data.Entries)
        setBreakcrumbItems([{ title: t('nav.documents'), docs: data.Entries }])

        return response
      })
      .catch((err) => {
        throw err
      })
  }, [t])

  const children = docs.map((doc) => {
    return (
      <TreeElement
        key={doc.id || uuidv4()}
        className={
          selected && selected.id === doc.id
            ? 'bg-slate-800 fill-neutral-200 text-neutral-200'
            : 'fill-neutral-400'
        }
        doc={doc}
        onClickDoc={(doc) => {
          if (doc.type === 'DocumentType') {
            if (selected && selected.id === doc.id) {
              setSelected(null)
            } else {
              setSelected(doc)
            }

            return
          }

          // Handle folder
          if (undefined !== doc.children) {
            pushd(doc)
          }
        }}
      />
    )
  })

  function removeDoc(doc: HashDoc) {
    const newDocs: HashDoc[] = []

    for (const docEntry of docs) {
      if (doc.id === docEntry.id) {
        continue
      }
      newDocs.push(docEntry)
    }

    setDocs(newDocs)
  }

  const onDocDeleted = (doc: HashDoc) => {
    setSelected(null)
    removeDoc(doc)
  }

  return (
    <>
      <Breadcrumbs
        className="sticky top-0 mt-8 border-b border-slate-100/10 bg-slate-900 py-4"
        items={breakcrumbItems}
        onClickBreadcrumb={(_item, index) => popd(index)}
      />
      <div className="divide-y divide-slate-800">{children}</div>
      <FileMenu
        doc={selected}
        onDocDeleted={onDocDeleted}
      />
    </>
  )
}
