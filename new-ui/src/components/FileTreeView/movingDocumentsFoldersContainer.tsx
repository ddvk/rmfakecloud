import { Transition } from '@headlessui/react'
import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { ChevronLeftIcon } from '@heroicons/react/outline'
import { PulseLoader } from 'react-spinners'

import { HashDoc } from '../../utils/models'
import { listDocuments } from '../../api'
import { MD_SCREEN_SIZE } from '../../utils/size'

import DirElement from './dirElement'
import Breadcrumbs, { BreakcrumbItem } from './breadcrumb'

interface MovingDocumentsFoldersContainerProps {
  show: boolean
  onMoveTo?: (folder: HashDoc | null) => void
  onDiscardMove?: () => void
}

export default function MovingDocumentsFoldersContainer(
  props: MovingDocumentsFoldersContainerProps
) {
  const { show, onMoveTo, onDiscardMove } = props
  const { t } = useTranslation()

  const isScrollable = document.body.offsetHeight > window.innerHeight
  const isAtLeastMiddleScreen = window.innerWidth >= MD_SCREEN_SIZE

  const [isLoading, setIsLoading] = useState(false)
  const [folders, setFolders] = useState<HashDoc[]>([])
  const [selected, setSelected] = useState<HashDoc | null>(null)
  const [breadcrumbItems, setBreakcrumbItems] = useState<BreakcrumbItem[]>([])

  function pushd(dir: HashDoc) {
    if (dir.type === 'DocumentType' || undefined === dir.children) {
      return
    }
    const folders = (dir.children || []).filter((child) => child.type !== 'DocumentType')

    setSelected(dir)
    setFolders(folders)
    setBreakcrumbItems((items) => {
      return [...items, { id: dir.id, title: dir.name, docs: folders, folder: dir }]
    })
  }

  function popd(toIndex?: number) {
    const maxIndex =
      toIndex !== undefined ? toIndex : breadcrumbItems.length > 1 ? breadcrumbItems.length - 2 : 0
    const items: BreakcrumbItem[] = [...breadcrumbItems].filter((_doc, i) => i <= maxIndex)
    const lastItem = items[items.length - 1]

    if (!lastItem.folder) {
      setSelected(null)
    } else {
      setSelected(lastItem.folder)
    }
    setFolders(lastItem.docs)
    setBreakcrumbItems(items)
  }

  useEffect(() => {
    if (show) {
      document.body.classList.add('overflow-hidden')
      if (isScrollable && isAtLeastMiddleScreen) {
        document.body.classList.add('md:pr-4')
      }
    } else {
      document.body.classList.remove('overflow-hidden')
      if (isScrollable && isAtLeastMiddleScreen) {
        document.body.classList.remove('md:pr-4')
      }
    }

    if (!show) {
      setTimeout(() => setSelected(null), 300)
    }
  }, [show, isAtLeastMiddleScreen, isScrollable])

  useEffect(() => {
    if (!show) {
      return
    }

    setIsLoading(true)

    listDocuments()
      .then((response) => {
        const data = response.data as { Entries: HashDoc[]; Trash: HashDoc[] }
        const folders: HashDoc[] = []

        for (const entry of data.Entries) {
          if (entry.type !== 'DocumentType') {
            folders.push(entry)
          }
        }
        setFolders(folders)
        setBreakcrumbItems([{ title: t('nav.documents'), docs: folders }])

        return true
      })
      .catch((err) => {
        throw err
      })
      .finally(() => {
        setIsLoading(false)
      })
  }, [show, t])

  const children = folders.map((folder, i) => {
    function isActivedOrNext(): boolean {
      if (!selected) {
        return false
      }

      if (i > 0) {
        const prev = folders[i - 1]

        if (prev.id === selected.id) {
          return true
        }
      }

      if (folder.id === selected.id) {
        return true
      }

      return false
    }

    return (
      <DirElement
        key={folder.id}
        className={`${
          selected && selected.id === folder.id
            ? '-mx-4 bg-slate-800 fill-neutral-200 px-4 text-neutral-200'
            : 'fill-neutral-400'
        } ${isActivedOrNext() ? 'mt-px' : i > 0 ? 'border-t border-slate-800' : ''}`}
        doc={folder}
        index={i}
        onClickDoc={(doc) => {
          pushd(doc)
        }}
      />
    )
  })

  const discardFn = () => {
    onDiscardMove && onDiscardMove()
    setTimeout(() => setSelected(null), 300)
  }

  return (
    <Transition
      as="div"
      show={show}
    >
      <Transition.Child
        enter="transition-opacity duration-300"
        enterFrom="opacity-0"
        enterTo="opacity-100"
        leave="transition-opacity duration-300"
        leaveFrom="opacity-100"
        leaveTo="opacity-0"
      >
        <div
          className="fixed inset-0 z-[100] bg-neutral-800/50 backdrop-blur-sm"
          onClick={discardFn}
        />
      </Transition.Child>
      <Transition.Child
        as="div"
        className={`fixed ${
          isScrollable && isAtLeastMiddleScreen ? 'left-0 right-[16px]' : 'inset-x-0'
        } bottom-0 top-[15%] z-[100]`}
        enter="transition-translate transform duration-300 ease-out"
        enterFrom="translate-y-[100%]"
        enterTo="translate-y-[0%]"
        leave="transition-translate transform duration-300 ease-in"
        leaveFrom="translate-y-[0%]"
        leaveTo="translate-y-[100%]"
      >
        <div className="mx-auto h-full max-w-4xl">
          <div className="relative h-full overflow-y-auto rounded-t-3xl bg-slate-900 shadow-lg shadow-slate-900/50 md:mx-4">
            <div className="h-auto min-h-full px-4 pb-[80px]">
              {isLoading ? (
                <Loader />
              ) : (
                <>
                  <div className="sticky top-0 mt-4 border-b border-slate-800 bg-slate-900 pt-6 text-neutral-200">
                    <div className="flex items-center">
                      <button
                        className="relative left-[-6px] h-6 w-6 shrink-0"
                        onClick={() => {
                          if (breadcrumbItems.length > 1) {
                            popd()
                          } else {
                            discardFn()
                          }
                        }}
                      >
                        <ChevronLeftIcon className="mx-auto h-6 w-6" />
                      </button>
                      <h1 className="overflow-hidden text-ellipsis whitespace-nowrap font-bold">
                        {t('documents.breadcrumbs.move_documents_container.title', {
                          folder:
                            selected?.name ||
                            t('documents.breadcrumbs.move_documents_container.root_folder')
                        })}
                      </h1>
                    </div>
                    <Breadcrumbs
                      className="mt-2 mb-4"
                      hideMoreMenu={true}
                      items={breadcrumbItems}
                      onClickBreadcrumb={(_item, toIndex) => popd(toIndex)}
                    />
                  </div>
                  {children}
                </>
              )}
            </div>
            <div className="fixed inset-x-0 bottom-0 py-4 text-center backdrop-blur-sm">
              <button
                className="w-56 rounded-3xl bg-blue-700 py-3 font-bold text-neutral-200 hover:bg-blue-600 disabled:bg-blue-500/90"
                disabled={isLoading}
                onClick={() => {
                  onMoveTo && onMoveTo(selected)
                  setSelected(null)
                }}
              >
                {t('documents.breadcrumbs.move_documents_container.submit')}
              </button>
            </div>
          </div>
        </div>
      </Transition.Child>
    </Transition>
  )
}

function Loader() {
  return (
    <div className="absolute inset-x-0 top-[50%] translate-y-[-50%] text-center">
      <PulseLoader
        color="#e5e5e5"
        size={8}
        speedMultiplier={0.8}
      />
    </div>
  )
}
