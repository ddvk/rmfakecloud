import { HashDoc } from '../../utils/models'

export type HashDocElementProp = {
  doc: HashDoc
  index: number
  onClickDoc?: (doc: HashDoc) => void
  onDocEditingDiscard?: (doc: HashDoc) => void
  onDocRenamed?: (doc: HashDoc, newName: string) => void
  onFolderCreated?: (doc: HashDoc, index: number) => void
  onFolderCreationDiscarded?: (doc: HashDoc, index: number) => void
} & React.HTMLAttributes<HTMLDivElement>
