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
import { PulseLoader } from 'react-spinners'
import * as Yup from 'yup'
import { Formik, Form, Field, ErrorMessage } from 'formik'

import { EpubIcon, PDFIcon } from '../../utils/icons'
import { listDocuments } from '../../api'
import { deleteDocument, renameDocument } from '../../api/document'
import { ConfirmationDialog } from '../Dialog'
import { HashDoc } from '../../utils/models'
import { inputClassName } from '../../utils/form'

type HashDocElementProp = {
  doc: HashDoc
  onClickDoc?: (doc: HashDoc) => void
  onDocEditingDiscard?: (doc: HashDoc) => void
  onDocRenamed?: (doc: HashDoc, newName: string) => void
} & React.HTMLAttributes<HTMLDivElement>

function FileElement(params: HashDocElementProp) {
  const { doc, onClickDoc, onDocEditingDiscard, onDocRenamed, className, ...remainParams } = params
  const [unmountForm, setUnmountForm] = useState(false)
  const { preMode } = doc
  let { mode } = doc

  if (!mode) {
    mode = 'display'
  }
  const { t } = useTranslation()

  const formFadeout = () => {
    setTimeout(() => {
      setUnmountForm(true)
    }, 500)
  }

  const editingToDisplay = mode === 'display' && preMode === 'editing'

  if (editingToDisplay) {
    formFadeout()
  }

  const validationSchema = Yup.object().shape({
    name: Yup.string().required(t('documents.rename_form.name.required'))
  })

  const formDom = (
    <Formik
      initialValues={{ name: doc.name }}
      validationSchema={validationSchema}
      onSubmit={(values, { setSubmitting }) => {
        setSubmitting(true)

        renameDocument(doc.id, values.name)
          .then(() => {
            toast.success(t('notifications.document_renamed'))
            onDocRenamed && onDocRenamed(doc, values.name)

            return 'ok'
          })
          .catch((err) => {
            throw err
          })
          .finally(() => {
            setSubmitting(false)
          })
      }}
    >
      {({ isSubmitting, handleSubmit, errors, touched }) => (
        <Form
          className={`w-full overflow-hidden ${
            editingToDisplay ? 'animate-roll-up' : 'animate-roll-down'
          }`}
          onSubmit={handleSubmit}
        >
          <div className="mb-4">
            <label className="mb-2 block font-bold text-neutral-400">
              {t('documents.rename_form.name.label')}
            </label>
            <Field
              autoFocus={true}
              className={inputClassName(errors.name && touched.name)}
              name="name"
              type="text"
            />
            <ErrorMessage
              className="mt-2 text-xs text-red-600"
              component="div"
              name="name"
            />
          </div>
          <div className="flex">
            <button
              className="mr-2 w-full basis-1/2 rounded border border-slate-600 py-3 font-bold text-neutral-200 focus:outline-none"
              type="button"
              onClick={(e) => {
                e.stopPropagation()
                setUnmountForm(false)
                onDocEditingDiscard && onDocEditingDiscard(doc)
              }}
            >
              {t('documents.rename_form.cancel-btn')}
            </button>
            <button
              className="w-full basis-1/2 rounded bg-blue-700 py-3 font-bold text-neutral-200 hover:bg-blue-600 focus:outline-none disabled:bg-blue-500"
              disabled={isSubmitting}
              type="submit"
            >
              {isSubmitting ? (
                <PulseLoader
                  color="#e5e5e5"
                  cssOverride={{ lineHeight: 0, padding: '6px 0' }}
                  size={8}
                  speedMultiplier={0.8}
                />
              ) : (
                t('documents.rename_form.submit-btn')
              )}
            </button>
          </div>
        </Form>
      )}
    </Formik>
  )
  let innerDom: JSX.Element

  if (mode === 'editing') {
    innerDom = formDom
  } else {
    innerDom = (
      <>
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
      </>
    )
  }

  return (
    <>
      <div
        className={`flex cursor-pointer select-none py-6 ${
          editingToDisplay ? 'animate-fadein' : ''
        } ${className || ''}`}
        {...remainParams}
        onClick={() => {
          onClickDoc && onClickDoc(doc)
        }}
      >
        {innerDom}
      </div>
      {editingToDisplay && !unmountForm ? formDom : <></>}
    </>
  )
}

function DirElement(params: HashDocElementProp) {
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  const { doc, onClickDoc, onDocEditingDiscard, onDocRenamed, className, ...remainParams } = params

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
            ? '-mx-4 bg-slate-800 fill-neutral-200 px-4 text-neutral-200'
            : 'fill-neutral-400'
        }
        doc={doc}
        onClickDoc={(doc) => {
          if (doc.mode === 'editing') {
            return
          }
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
        onDocEditingDiscard={(doc) => {
          const newDocs = docs.map((entity) => {
            if (doc.id === entity.id) {
              entity.preMode = entity.mode
              entity.mode = 'display'
              setSelected(entity)
            }

            return entity
          })

          setDocs(newDocs)
        }}
        onDocRenamed={(doc, newName) => {
          setSelected(null)
          const newDocs = docs.map((entity) => {
            if (doc.id === entity.id) {
              entity.preMode = entity.mode
              entity.mode = 'display'
              entity.name = newName
            }

            return entity
          })

          setDocs(newDocs)
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

  const onDocEditing = (doc: HashDoc) => {
    const newDocs = docs.map((entity) => {
      entity.mode = 'display'
      if (doc.id === entity.id) {
        entity.preMode = entity.mode
        entity.mode = 'editing'
      }

      return entity
    })

    setDocs(newDocs)
    setSelected(null)
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
        onDocEditing={onDocEditing}
      />
    </>
  )
}
