import { HashDoc } from '../../utils/models'

export type HashDocElementProp = {
  doc: HashDoc
  index: number
  multiple?: boolean
  onClickDoc?: (doc: HashDoc) => void
  onDocEditingDiscard?: (doc: HashDoc) => void
  onDocRenamed?: (doc: HashDoc, newName: string) => void
  onFolderCreated?: (doc: HashDoc, index: number) => void
  onFolderCreationDiscarded?: (doc: HashDoc, index: number) => void
  onCheckBoxChanged?: (obj: { doc: HashDoc; index: number; checked: boolean }) => void
} & React.HTMLAttributes<HTMLDivElement>
