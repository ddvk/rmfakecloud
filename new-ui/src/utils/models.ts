export interface User {
  userid: string
  name: string
  email?: string
  CreatedAt?: string
  integrations?: string[]
}

type HashDocMode = 'display' | 'editing' | 'creating'

type HashDocType = 'DocumentType' | 'CollectionType'

export interface HashDoc {
  id: string
  name: string
  type: HashDocType
  size: number
  extension?: string
  parent?: string
  children?: HashDoc[]
  LastModified: string

  preMode?: HashDocMode
  mode?: HashDocMode
}

export interface HashDocMetadata {
  ID: string
  Type: HashDocType
  VissibleName: string
  Version?: number
  Message?: string
  Success?: boolean
  BlobURLGet?: string
  BlobURLGetExpires?: string
  ModifiedClient: string
  CurrentPage?: number
  Bookmarked?: boolean
  Parent?: string
}

export interface AppUser {
  userid: string
  email?: string
  name?: string
  CreatedAt?: string
}
