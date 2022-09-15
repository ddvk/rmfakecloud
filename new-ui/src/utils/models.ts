export interface User {
  userid: string
  name: string
  email?: string
  CreatedAt?: string
  integrations?: string[]
}

type HashDocMode = 'display' | 'editing'

export interface HashDoc {
  id: string
  name: string
  type: 'DocumentType' | 'CollectionType'
  size: number
  extension?: string
  children?: HashDoc[]
  LastModified: string

  preMode?: HashDocMode
  mode?: HashDocMode
}
