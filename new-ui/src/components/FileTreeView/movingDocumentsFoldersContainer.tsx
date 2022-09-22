import { Transition } from '@headlessui/react'
import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { XIcon } from '@heroicons/react/outline'
import { PulseLoader } from 'react-spinners'

import { HashDoc } from '../../utils/models'
import { listDocuments } from '../../api'
import { MD_SCREEN_SIZE } from '../../utils/size'

import DirElement from './dirElement'

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
    setIsLoading(true)

    listDocuments()
      .then((response) => {
        const data = response.data as { Entries: HashDoc[]; Trash: HashDoc[] }
        const folders: HashDoc[] = []

        for (const entry of data.Entries) {
          if (entry.type !== 'DocumentType') {
            folders.push(entry)
          }
          setFolders(folders)
        }

        return true
      })
      .catch((err) => {
        throw err
      })
      .finally(() => {
        setIsLoading(false)
      })
  }, [])

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
        } ${isActivedOrNext() ? 'mt-px' : 'border-t border-slate-800'}`}
        doc={folder}
        index={i}
        onClickDoc={(doc) => {
          if (selected && doc.id === selected.id) {
            setSelected(null)

            return
          }
          setSelected(doc)
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
          <div className="relative h-full rounded-t-3xl bg-slate-900 shadow-lg shadow-slate-900/50 md:mx-4">
            <div className="h-full overflow-y-auto px-4">
              {isLoading ? (
                <Loader />
              ) : (
                <>
                  <div className="mt-4 flex items-center py-6 text-neutral-200">
                    <button
                      className="mr-2 h-6 w-6 shrink-0 rounded-full bg-slate-100/10"
                      onClick={discardFn}
                    >
                      <XIcon className="mx-auto h-4 w-4" />
                    </button>
                    <h1 className="overflow-hidden text-ellipsis whitespace-nowrap font-bold">
                      {t('documents.breadcrumbs.move_documents_container.title', {
                        folder:
                          selected?.name ||
                          t('documents.breadcrumbs.move_documents_container.root_folder')
                      })}
                    </h1>
                  </div>
                  {children}
                </>
              )}
            </div>
            <div className="fixed inset-x-0 bottom-0 my-4 text-center backdrop-blur-sm">
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
