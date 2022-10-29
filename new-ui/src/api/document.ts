import { AxiosResponse } from 'axios'

import requests from '../utils/request'

export interface UploadedFile {
  id: number | string
  file: File
  uploadedSize: number
  status: 'pending' | 'uploading' | 'uploaded'
  cancel: AbortController
}

export function getMetadata(id: string): Promise<AxiosResponse> {
  return requests.get(`/ui/api/documents/${id}/metadata`)
}

export function uploadDocument<T extends UploadedFile>(
  file: T,
  onChange?: (f: UploadedFile) => void
) {
  const data = new FormData()

  data.append('file', file.file)

  const promise = requests.postForm('/ui/api/documents/upload', data, {
    timeout: 5 * 60 * 1000, // 5mins
    signal: file.cancel.signal,
    onUploadProgress: (event) => {
      if (file.status === 'pending') {
        file.status = 'uploading'
      }
      const { loaded, total } = event

      file.uploadedSize = loaded
      if (loaded === total) {
        file.status = 'uploaded'
      }
      onChange && onChange(file)
    }
  })

  return promise
}

export function deleteDocument(id: string): Promise<AxiosResponse> {
  return requests.delete(`/ui/api/documents/${id}`)
}

export function renameDocument(id: string, name: string): Promise<AxiosResponse> {
  return requests.put(`/ui/api/documents/${id}`, { name })
}

export function moveDocumentTo(id: string, parentID?: string): Promise<AxiosResponse> {
  return requests.put(`/ui/api/documents/${id}`, {
    parentId: parentID,
    setParentToRoot: !parentID
  })
}

export function exportDocument(id: string): Promise<AxiosResponse> {
  return requests.get(`/ui/api/documents/${id}`, {
    responseType: 'blob',
    timeout: 1000 * 60 * 10 /* 10mins */
  })
}

export function createFolder(name: string, parent?: string): Promise<AxiosResponse> {
  return requests.post('/ui/api/folders', { name, parentId: parent || '' })
}
