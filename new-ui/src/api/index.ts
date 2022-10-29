import { AxiosResponse } from 'axios'

import requests from '../utils/request'

export function listDocuments(): Promise<AxiosResponse> {
  return requests.get('/ui/api/documents')
}
