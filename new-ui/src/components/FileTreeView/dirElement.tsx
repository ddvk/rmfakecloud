import { FolderIcon } from '@heroicons/react/outline'

import { HashDocElementProp } from './props'

export default function DirElement(params: HashDocElementProp) {
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
