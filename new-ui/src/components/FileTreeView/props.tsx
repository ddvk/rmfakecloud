import { HashDoc } from '../../utils/models'

export type HashDocElementProp = {
  doc: HashDoc
  onClickDoc?: (doc: HashDoc) => void
  onDocEditingDiscard?: (doc: HashDoc) => void
  onDocRenamed?: (doc: HashDoc, newName: string) => void
} & React.HTMLAttributes<HTMLDivElement>
