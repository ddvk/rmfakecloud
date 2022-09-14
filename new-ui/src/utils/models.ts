export interface User {
  userid: string
  name: string
  email?: string
  CreatedAt?: string
  integrations?: string[]
}

export interface HashDoc {
  id: string
  name: string
  type: 'DocumentType' | 'CollectionType'
  size: number
  extension?: string
  children?: HashDoc[]
  LastModified: string
}
