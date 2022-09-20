import DirElement from './dirElement'
import FileElement from './fileElement'
import { HashDocElementProp } from './props'

export default function TreeElement(params: HashDocElementProp) {
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
