/* eslint-disable tailwindcss/no-custom-classname */

import { useEffect, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { PulseLoader } from 'react-spinners'
import { v4 as uuidv4 } from 'uuid'
import PQueue from 'p-queue'
import { toast } from 'react-toastify'

import { listDocuments } from '../../api'
import { moveDocumentTo } from '../../api/document'
import { HashDoc } from '../../utils/models'

import Breadcrumbs, { BreakcrumbItem } from './breadcrumb'
import FileMenu from './menu'
import TreeElement from './treeElement'
import MovingDocumentsFoldersContainer from './movingDocumentsFoldersContainer'
import ContextMenu from './contextMenu'
import RemoveDocDialog from './removeDocDialog'
import Sticky from 'react-stickynode'

export default function FileTreeView({ reloadCnt }: { reloadCnt?: number }) {
  const { t } = useTranslation()
  // Docs to rendered in the file tree view
  const [docs, setDocs] = useState<HashDoc[]>([])
  // Selected doc when click on a file
  const [selected, setSelected] = useState<HashDoc | null>(null)
  // Breadcrumbs
  const [breadcrumbItems, setBreakcrumbItems] = useState<BreakcrumbItem[]>([])
  // Breadcrumbs mode switch animation between normal/moving documents
  const [breadcrumbsAnimationClass, setBreadcrumbsAnimationClass] = useState('')
  // Whether show docs reloading component
  const [isLoading, setIsLoading] = useState(false)
  // Text for docs reloading component
  const [loadingText, setLoadingText] = useState('')
  // Whether show moving documents checkboxes
  const [isMovingDocuments, setIsMovingDocuments] = useState(false)
  // Checked documents when moving documents
  const [checkedDocs, setCheckedDocs] = useState<{ doc: HashDoc; index: number }[]>([])
  // Whether show moving documents to folder container
  const [isShowMovingDocumentsFolders, setIsShowMovingDocumentsFolders] = useState(false)
  // Hack for trigger reload docs
  const [innerReloadCnt, setInnerReloadCnt] = useState(0)
  // Selected folder when right clicked or long pressed
  const [contextMenuDoc, setContextMenuDoc] = useState<{
    doc: HashDoc
    index: number
    x: number
    y: number
  } | null>(null)
  // Remembered position to scroll to after docs changed
  const [rememberedPos, setRememberedPos] = useState<{ x: number; y: number } | null>(null)
  // Folder to be removed when click "Remove Folder" on context menu
  const [removingFolder, setRemovingFolder] = useState<HashDoc | null>(null)
  // Folder enter/exit animation
  const [treeAnimation, setTreeAnimation] = useState('')

  // Long press timeout id
  const pressTimer = useRef<NodeJS.Timeout | null>(null)

  function pushd(dir: HashDoc) {
    if (dir.type === 'DocumentType' || undefined === dir.children) {
      return
    }

    if (breadcrumbItems.length > 0) {
      breadcrumbItems[breadcrumbItems.length - 1].posX = window.pageXOffset
      breadcrumbItems[breadcrumbItems.length - 1].posY = window.pageYOffset
    }

    const items: BreakcrumbItem[] = [
      ...breadcrumbItems,
      { id: dir.id, title: dir.name, docs: dir.children }
    ]

    setBreakcrumbItems(items)
    setSelected(null)
    setContextMenuDoc(null)
    setDocs(dir.children)
    if (pressTimer.current) {
      clearTimeout(pressTimer.current)
    }
    setTreeAnimation('animate-[slidein_150ms_ease-out]')
    setTimeout(() => setTreeAnimation(''), 150)
  }

  function popd(toIndex?: number) {
    toIndex = toIndex || 0
    const items: BreakcrumbItem[] = [...breadcrumbItems]

    while (items.length > toIndex + 1) {
      items.pop()
    }

    const item = items[items.length - 1]

    setBreakcrumbItems(items)
    setSelected(null)
    setDocs(
      item.docs.map((doc) => {
        return { ...doc, preMode: undefined, mode: 'display' }
      })
    )
    setTreeAnimation('animate-[slideout_150ms_ease-in]')
    setTimeout(() => setTreeAnimation(''), 150)
    if (item.posX || item.posY) {
      setRememberedPos({
        x: item.posX || 0,
        y: item.posY || 0
      })
    }
  }

  useEffect(() => {
    setIsLoading(true)

    listDocuments()
      .then((response) => {
        const data = response.data as { Entries: HashDoc[]; Trash: HashDoc[] }

        const newBreadcrumbItems: BreakcrumbItem[] =
          breadcrumbItems.length === 0
            ? [{ title: t('nav.documents'), docs: data.Entries }]
            : [{ ...breadcrumbItems[0], docs: data.Entries }]

        let docs = data.Entries

        for (let i = 1; i < breadcrumbItems.length; i++) {
          let found = false
          const breadcrumbItem = breadcrumbItems[i]

          for (const entry of docs) {
            if (entry.id === breadcrumbItem.id) {
              newBreadcrumbItems.push({
                ...breadcrumbItem,
                id: entry.id,
                title: entry.name,
                docs: entry.children || []
              })
              docs = entry.children || []
              found = true
              break
            }
          }
          if (!found) {
            break
          }
        }

        setBreakcrumbItems(newBreadcrumbItems)
        setDocs(docs)
        const last = newBreadcrumbItems[newBreadcrumbItems.length - 1]

        setRememberedPos({
          x: last.posX || 0,
          y: last.posY || 0
        })

        return response
      })
      .catch((err) => {
        throw err
      })
      .finally(() => {
        setIsLoading(false)
      })
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [t, reloadCnt, innerReloadCnt])

  // Remember scroll position
  useEffect(() => {
    function handler() {
      setBreakcrumbItems((items) => {
        return items.map((item, i) => {
          if (i < items.length - 1) {
            return item
          }

          return {
            ...item,
            posX: window.pageXOffset,
            posY: window.pageYOffset
          }
        })
      })
    }

    window.addEventListener('scroll', handler)

    return function () {
      window.removeEventListener('scroll', handler)
    }
  }, [])

  useEffect(() => {
    if (isMovingDocuments) {
      setBreadcrumbsAnimationClass('animate-flip-x')

      return
    }

    if (breadcrumbsAnimationClass.length > 0) {
      setBreadcrumbsAnimationClass('animate-flip-x-reverse')
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isMovingDocuments])

  useEffect(() => {
    if (rememberedPos) {
      window.scroll({
        top: rememberedPos.y,
        left: rememberedPos.x,
        behavior: 'auto'
      })
      setRememberedPos(null)
    }
  }, [rememberedPos])

  const children = docs.map((doc, i) => {
    function isActivedOrNext(): boolean {
      if (!selected) {
        return false
      }

      if (i > 0) {
        const prev = docs[i - 1]

        if (prev.id === selected.id) {
          return true
        }
      }

      if (doc.id === selected.id) {
        return true
      }

      return false
    }

    return (
      <TreeElement
        key={doc.id || uuidv4()}
        className={`${
          selected && selected.id === doc.id
            ? '-mx-4 bg-slate-800 fill-neutral-200 px-4 text-neutral-200'
            : 'fill-neutral-400'
        } ${
          isActivedOrNext() ? 'mt-px' : 'border-t border-slate-800'
        } md:hover:-mx-4 md:hover:bg-slate-800 md:hover:fill-neutral-200 md:hover:px-4 md:hover:text-neutral-200`}
        doc={doc}
        index={i}
        multiple={isMovingDocuments}
        onCheckBoxChanged={({ doc, index, checked }) => {
          setCheckedDocs((prev) => {
            const i = prev.findIndex((item) => {
              return item.doc.id === doc.id
            })

            if (checked && i === -1) {
              prev.push({ doc, index })
            }

            if (!checked && i !== -1) {
              prev.splice(i, 1)
            }

            return [...prev]
          })
        }}
        onClickDoc={(doc) => {
          if (isMovingDocuments) {
            return
          }
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
        onContextMenu={(e) => {
          if (doc.type === 'DocumentType') {
            return
          }
          if (isMovingDocuments) {
            return
          }
          e.preventDefault()

          setSelected(null)
          setContextMenuDoc({ doc, index: i, x: e.clientX, y: e.clientY })
        }}
        onDocEditingDiscard={(doc) => {
          const newDocs = docs.map((entity) => {
            if (doc.id === entity.id) {
              entity.preMode = entity.mode
              entity.mode = 'display'
              if (entity.type === 'DocumentType') {
                setSelected(entity)
              }
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

          setBreakcrumbItems((items) => {
            return items.map((item, i) => {
              if (i < items.length - 1) {
                return item
              }

              return {
                ...item,
                docs: newDocs
              }
            })
          })
          setDocs(newDocs)
        }}
        onFolderCreated={() => {
          setInnerReloadCnt(innerReloadCnt + 1)
        }}
        onFolderCreationDiscarded={(_doc, i) => {
          setDocs((prevDocs) => {
            return prevDocs.filter((_hashDoc, index) => index !== i)
          })
        }}
        onTouchEnd={() => {
          pressTimer.current && clearTimeout(pressTimer.current)
        }}
        onTouchMove={() => {
          pressTimer.current && clearTimeout(pressTimer.current)
          setContextMenuDoc(null)
        }}
        onTouchStart={(e) => {
          if (doc.type === 'DocumentType') {
            return
          }
          if (isMovingDocuments) {
            return
          }
          pressTimer.current = setTimeout(() => {
            setSelected(null)
            setContextMenuDoc({ doc, index: i, x: e.touches[0].clientX, y: e.touches[0].clientY })
          }, 1000)
        }}
      />
    )
  })

  const onDocDeleted = () => {
    setSelected(null)
    setRemovingFolder(null)
    setInnerReloadCnt(innerReloadCnt + 1)
  }

  const onDocEditing = (doc: HashDoc) => {
    const newDocs = docs.map((entity) => {
      if (entity.mode !== 'creating') {
        entity.mode = 'display'
      }
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
      <Sticky>
        <Breadcrumbs
          checkedDocCount={checkedDocs.length}
          className={`z-10 mx-4 border-b border-slate-100/10 bg-slate-900 py-4 ${breadcrumbsAnimationClass}`}
          isMovingDocuments={isMovingDocuments}
          items={breadcrumbItems}
          onClickBreadcrumb={(_item, index) => popd(index)}
          onClickMoveDocuments={() => {
            setIsMovingDocuments(true)
            setSelected(null)
          }}
          onClickNewFolder={() => {
            setDocs((prevDocs) => {
              const newFolder: HashDoc = {
                id: uuidv4(),
                name: '',
                type: 'CollectionType',
                parent:
                  breadcrumbItems.length > 1
                    ? breadcrumbItems[breadcrumbItems.length - 1].id
                    : undefined,
                size: 0,
                LastModified: '',
                mode: 'creating'
              }

              return [newFolder, ...prevDocs]
            })
          }}
          onDiscardMovingDocuments={() => {
            setCheckedDocs([])
            setIsMovingDocuments(false)
          }}
          onMovingDocumentsSubmit={() => {
            setIsShowMovingDocumentsFolders(true)
          }}
        />
      </Sticky>
      {isLoading ? (
        <div className="relative mt-24 text-center">
          <PulseLoader
            color="#e5e5e5"
            cssOverride={{ lineHeight: 0, padding: '6px 0' }}
            size={8}
            speedMultiplier={0.8}
          />
          {loadingText.length > 0 ? (
            <p className="mt-2 overflow-hidden text-ellipsis whitespace-nowrap text-sm">
              {loadingText}
            </p>
          ) : (
            <></>
          )}
        </div>
      ) : children.length > 0 ? (
        <div className={`mx-4 min-h-[calc(100vh-475px)] md:min-h-screen ${treeAnimation}`}>
          {children}
        </div>
      ) : (
        <div className={`relative mt-20 text-center text-slate-100/10 ${treeAnimation}`}>
          <svg
            className="mx-auto h-16 w-16"
            fill="none"
            stroke="currentColor"
            strokeWidth={1.5}
            viewBox="0 0 24 24"
            xmlns="http://www.w3.org/2000/svg"
          >
            <path
              d="M3.75 9.776c.112-.017.227-.026.344-.026h15.812c.117 0 .232.009.344.026m-16.5 0a2.25 2.25 0 00-1.883 2.542l.857 6a2.25 2.25 0 002.227 1.932H19.05a2.25 2.25 0 002.227-1.932l.857-6a2.25 2.25 0 00-1.883-2.542m-16.5 0V6A2.25 2.25 0 016 3.75h3.879a1.5 1.5 0 011.06.44l2.122 2.12a1.5 1.5 0 001.06.44H18A2.25 2.25 0 0120.25 9v.776"
              strokeLinecap="round"
              strokeLinejoin="round"
            />
          </svg>
          <p className="font-bold">{t('documents.empty')}</p>
        </div>
      )}
      <FileMenu
        doc={selected}
        onDocDeleted={onDocDeleted}
        onDocEditing={onDocEditing}
      />
      <MovingDocumentsFoldersContainer
        show={isShowMovingDocumentsFolders}
        onDiscardMove={() => {
          setIsShowMovingDocumentsFolders(false)
        }}
        onMoveTo={(folder) => {
          setIsShowMovingDocumentsFolders(false)
          setIsLoading(true)
          let finishedOrUpdating = 0
          let current = checkedDocs.length > 0 ? checkedDocs[0] : null
          let succeededCount = 0
          let failedCount = 0
          const updateLoadingText = () => {
            setLoadingText(
              `[${finishedOrUpdating}/${checkedDocs.length}]${t(
                'documents.breadcrumbs.move_documents_container.moving'
              )} ${current?.doc.name}`
            )
          }

          updateLoadingText()

          const queue = new PQueue({
            concurrency: 1,
            autoStart: false,
            timeout: 30 * 1000 // 30s per operation
          })

          for (const item of checkedDocs) {
            queue.add(() => {
              return moveDocumentTo(item.doc.id, folder?.id)
            })
          }

          queue.on('active', () => {
            current = checkedDocs[finishedOrUpdating]
            finishedOrUpdating++
            updateLoadingText()
          })

          queue.on('completed', () => {
            succeededCount++
          })

          queue.on('error', () => {
            failedCount++
          })

          queue.on('empty', () => {
            if (succeededCount + failedCount === checkedDocs.length) {
              const msg = t('notifications.documents_moved', {
                total: checkedDocs.length,
                succeeded: succeededCount,
                failed: failedCount
              })

              if (succeededCount === 0) {
                toast.error(msg)
              } else if (failedCount > 0) {
                toast.warning(msg)
              } else {
                toast.success(msg)
              }

              setIsLoading(false)
              setLoadingText('')
              setIsMovingDocuments(false)
              setCheckedDocs([])
              setInnerReloadCnt(innerReloadCnt + 1)
            }
          })

          queue.start()
        }}
      />
      <ContextMenu
        {...contextMenuDoc}
        onClickMenuItem={({ doc, menuItem }) => {
          if (!doc) {
            return
          }

          if (menuItem === 'rename') {
            setDocs((prevDocs) => {
              return prevDocs.map((prevDoc) => {
                prevDoc.preMode = prevDoc.mode
                if (prevDoc.id === doc.id) {
                  prevDoc.mode = 'editing'
                } else {
                  prevDoc.mode = 'display'
                }

                return prevDoc
              })
            })
            setContextMenuDoc(null)
          }

          if (menuItem === 'remove') {
            setRemovingFolder(doc)
          }
        }}
        onDismissMenu={() => {
          setContextMenuDoc(null)
        }}
      />
      <RemoveDocDialog
        doc={removingFolder}
        onDismissDialog={() => {
          setRemovingFolder(null)
        }}
        onDocDeleted={onDocDeleted}
      />
    </>
  )
}
