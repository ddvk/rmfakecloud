import { v4 as uuidv4 } from 'uuid'
import {
  DocumentTextIcon,
  SaveIcon,
  EyeIcon,
  PencilAltIcon,
  TrashIcon,
  FolderIcon
} from '@heroicons/react/solid'
import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Transition } from '@headlessui/react'
import { Link } from 'react-router-dom'

import { listDocuments } from '../../api'

export interface HashDoc {
  id: string
  name: string
  type: 'DocumentType' | 'CollectionType'
  size: number
  children?: HashDoc[]
  LastModified: string
}

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
      <DocumentTextIcon className="top-[-1px] mr-2 h-6 w-6" />
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
      <FolderIcon className="top-[-1px] mr-2 h-6 w-6" />
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

function FileMenu(params: { doc?: HashDoc | null }) {
  const { doc } = params
  const { t } = useTranslation()

  return (
    <Transition
      as="div"
      className="fixed inset-x-0 bottom-0 flex h-20 items-center justify-around border-t border-slate-100/10 bg-slate-900 md:mx-auto md:max-w-4xl md:border-x"
      enter="transition-translate-y duration-300"
      enterFrom="translate-y-full"
      enterTo="translate-y-0"
      leave="transition-translate-y duration-300"
      leaveFrom="translate-y-0"
      leaveTo="translate-y-full"
      show={new Boolean(doc).valueOf()}
    >
      <Link
        target="_blank"
        to={`/ui/api/documents/${doc?.id}`}
      >
        <div className="cursor-pointer p-4 hover:text-neutral-200">
          <SaveIcon className="mx-auto mb-1 h-6 w-6" />
          <p className="text-xs">{t('documents.file_tree_view.menu.download')}</p>
        </div>
      </Link>
      <div className="cursor-pointer p-4 hover:text-neutral-200">
        <EyeIcon className="mx-auto mb-1 h-6 w-6" />
        <p className="text-xs">{t('documents.file_tree_view.menu.view')}</p>
      </div>
      <div className="cursor-pointer p-4 hover:text-neutral-200">
        <PencilAltIcon className="mx-auto mb-1 h-6 w-6" />
        <p className="text-xs">{t('documents.file_tree_view.menu.rename')}</p>
      </div>
      <div className="cursor-pointer p-4 hover:text-neutral-200">
        <TrashIcon className="mx-auto mb-1 h-6 w-6" />
        <p className="text-xs">{t('documents.file_tree_view.menu.remove')}</p>
      </div>
    </Transition>
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
      .catch((err) => console.error(err))
  }, [])

  const children = docs.map((doc) => {
    return (
      <TreeElement
        key={doc.id || uuidv4()}
        className={selected && selected.id === doc.id ? 'bg-slate-800 text-neutral-200' : ''}
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

  return (
    <>
      <Breadcrumbs
        className="sticky top-0 mt-8 border-b border-slate-100/10 bg-slate-900 py-4"
        items={breakcrumbItems}
        onClickBreadcrumb={(_item, index) => popd(index)}
      />
      <div className="divide-y divide-slate-800">{children}</div>
      <FileMenu doc={selected} />
    </>
  )
}
