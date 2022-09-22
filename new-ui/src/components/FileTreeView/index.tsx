/* eslint-disable tailwindcss/no-custom-classname */

import { useEffect, useState } from 'react'
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

export default function FileTreeView({ reloadCnt }: { reloadCnt?: number }) {
  const { t } = useTranslation()
  const [docs, setDocs] = useState<HashDoc[]>([])
  const [selected, setSelected] = useState<HashDoc | null>(null)
  const [breakcrumbItems, setBreakcrumbItems] = useState<BreakcrumbItem[]>([])
  const [breadcrumbsAnimationClass, setBreadcrumbsAnimationClass] = useState('')
  const [isLoading, setIsLoading] = useState(false)
  const [loadingText, setLoadingText] = useState('')
  const [isMovingDocuments, setIsMovingDocuments] = useState(false)
  const [checkedDocs, setCheckedDocs] = useState<{ doc: HashDoc; index: number }[]>([])
  const [isShowMovingDocumentsFolders, setIsShowMovingDocumentsFolders] = useState(false)
  const [innerReloadCnt, setInnerReloadCnt] = useState(0)
  const [stayInFolder, setStayInFolder] = useState<HashDoc | null>(null)

  function pushd(dir: HashDoc) {
    if (dir.type === 'DocumentType' || undefined === dir.children) {
      return
    }

    const items: BreakcrumbItem[] = [...breakcrumbItems, { title: dir.name, docs: dir.children }]

    setBreakcrumbItems(items)
    setSelected(null)
    setDocs(dir.children)
    setStayInFolder(dir)
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
    if (toIndex === 0) {
      setStayInFolder(null)
    }
  }

  useEffect(() => {
    setIsLoading(true)

    listDocuments()
      .then((response) => {
        const data = response.data as { Entries: HashDoc[]; Trash: HashDoc[] }

        const breadcrumbItems = [{ title: t('nav.documents'), docs: data.Entries }]
        let docs = data.Entries

        if (stayInFolder) {
          data.Entries.forEach((entry) => {
            if (entry.id === stayInFolder.id) {
              breadcrumbItems.push({
                title: entry.name,
                docs: entry.children || []
              })
              docs = entry.children || []
            }
          })
        }

        setBreakcrumbItems(breadcrumbItems)
        setDocs(docs)

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
        } ${isActivedOrNext() ? 'mt-px' : 'border-t border-slate-800'}`}
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
        onFolderCreated={(doc, i) => {
          setDocs((prevDocs) => {
            return prevDocs.map((entity, index) => {
              if (index === i) {
                return {
                  ...doc,
                  preMode: 'creating',
                  mode: 'display'
                }
              }

              return entity
            })
          })
        }}
        onFolderCreationDiscarded={(_doc, i) => {
          setDocs((prevDocs) => {
            return prevDocs.filter((_hashDoc, index) => index !== i)
          })
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
      <Breadcrumbs
        checkedDocCount={checkedDocs.length}
        className={`sticky top-0 mt-8 border-b border-slate-100/10 bg-slate-900 py-4 z-10 ${breadcrumbsAnimationClass}`}
        isMovingDocuments={isMovingDocuments}
        items={breakcrumbItems}
        onClickBreadcrumb={(_item, index) => popd(index)}
        onClickMoveDocuments={() => {
          setIsMovingDocuments(true)
          setSelected(null)
        }}
        onClickNewFolder={() => {
          setDocs((prevDocs) => {
            const newFolder: HashDoc = {
              id: '',
              name: '',
              type: 'CollectionType',
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
        <div>{children}</div>
      ) : (
        <div className="relative mt-20 text-center text-slate-100/10">
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
    </>
  )
}
