import { useTranslation } from 'react-i18next'
import { Fragment } from 'react'
import { DotsVerticalIcon, XIcon } from '@heroicons/react/solid'
import { Menu, Transition } from '@headlessui/react'
import { FolderAddIcon, CollectionIcon, ArrowsExpandIcon } from '@heroicons/react/outline'

import { HashDoc } from '../../utils/models'

export interface BreakcrumbItem {
  id?: string
  title: string
  docs: HashDoc[]
  posX?: number
  posY?: number
}

interface MovingDocumentsHeaderEventProps {
  onDiscardMovingDocuments?: () => void
  onMovingDocumentsSubmit?: () => void
}

export default function Breadcrumbs(
  params: {
    items: BreakcrumbItem[]
    className?: string
    isMovingDocuments?: boolean
    checkedDocCount?: number
    onClickBreadcrumb?: (item: BreakcrumbItem, index: number) => void
    onClickNewFolder?: () => void
    onClickMoveDocuments?: () => void
  } & MovingDocumentsHeaderEventProps
) {
  const {
    items,
    className,
    checkedDocCount,
    isMovingDocuments,
    onClickBreadcrumb,
    onClickNewFolder,
    onClickMoveDocuments,
    onDiscardMovingDocuments: discardMovingDocumentsFn,
    onMovingDocumentsSubmit
  } = params
  const { t } = useTranslation()

  const onDiscardMovingDocuments = () => {
    discardMovingDocumentsFn && discardMovingDocumentsFn()
  }

  const innerDom = items.map((item, i) => {
    return (
      <li
        key={`breakcrumb-item-${i}`}
        className="cursor-pointer whitespace-nowrap after:mx-2 after:text-neutral-400 after:content-['>'] last:after:hidden"
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
      {isMovingDocuments ? (
        <MovingDocumentsHeader
          count={checkedDocCount}
          onDiscardMovingDocuments={onDiscardMovingDocuments}
          onMovingDocumentsSubmit={onMovingDocumentsSubmit}
        />
      ) : (
        <div className="flex">
          <ul className="relative flex max-w-full flex-nowrap items-center overflow-hidden text-ellipsis text-sm font-semibold text-sky-600">
            {innerDom}
          </ul>
          <Menu
            as="div"
            className="relative ml-auto"
          >
            <Menu.Button>
              <DotsVerticalIcon className="relative top-[3px] h-6 w-6 transition-colors duration-300 hover:text-sky-600" />
            </Menu.Button>
            <Transition
              as={Fragment}
              enter="transition ease-out duration-100"
              enterFrom="transform opacity-0 scale-95"
              enterTo="transform opacity-100 scale-100"
              leave="transition ease-in duration-75"
              leaveFrom="transform opacity-100 scale-100"
              leaveTo="transform opacity-0 scale-95"
            >
              <Menu.Items className="absolute right-1 mt-1 w-56 origin-top-right divide-y divide-slate-100/20 rounded-md bg-slate-800 shadow-lg ring-1 ring-slate-800 focus:outline-none">
                <div className="p-2">
                  <Menu.Item>
                    {({ active }) => (
                      <button
                        className={`${
                          active ? 'bg-slate-900 text-sky-600' : 'text-neutral-400'
                        } group flex w-full items-center rounded-md p-2 text-sm font-bold disabled:text-neutral-400/20`}
                        onClick={() => {
                          onClickNewFolder && onClickNewFolder()
                        }}
                      >
                        <FolderAddIcon className="mr-2 h-5 w-5" />
                        {t('documents.breadcrumbs.new_folder')}
                      </button>
                    )}
                  </Menu.Item>
                </div>
                <div className="p-2">
                  <Menu.Item>
                    {({ active }) => (
                      <button
                        className={`${
                          active ? 'bg-slate-900 text-sky-600' : 'text-neutral-400'
                        } group flex w-full items-center rounded-md p-2 text-sm font-bold`}
                        onClick={() => {
                          onClickMoveDocuments && onClickMoveDocuments()
                        }}
                      >
                        <CollectionIcon className="mr-2 h-5 w-5" />
                        {t('documents.breadcrumbs.move_documents')}
                      </button>
                    )}
                  </Menu.Item>
                </div>
              </Menu.Items>
            </Transition>
          </Menu>
        </div>
      )}
    </div>
  )
}

function MovingDocumentsHeader(
  props: {
    count?: number
  } & MovingDocumentsHeaderEventProps
) {
  const { count, onDiscardMovingDocuments, onMovingDocumentsSubmit } = props
  const { t } = useTranslation()

  return (
    <div className="flex items-center text-neutral-200">
      <button
        className="mr-2 h-6 w-6 shrink-0 rounded-full bg-slate-100/10"
        onClick={() => {
          onDiscardMovingDocuments && onDiscardMovingDocuments()
        }}
      >
        <XIcon className="mx-auto h-4 w-4" />
      </button>
      <div className="flex-1 font-bold">
        {t('documents.breadcrumbs.move_documents_selected_tip', { count: count || 0 })}
      </div>
      <div className="ml-auto shrink-0 py-[3px]">
        <button
          className="flex h-6 items-center rounded font-bold transition-colors duration-200 hover:text-sky-700 disabled:text-neutral-200/40"
          disabled={count === 0}
          onClick={() => {
            onMovingDocumentsSubmit && onMovingDocumentsSubmit()
          }}
        >
          <ArrowsExpandIcon className="mr-2 h-4 w-4" />
          {t('documents.breadcrumbs.move_documents_submit')}
        </button>
      </div>
    </div>
  )
}
