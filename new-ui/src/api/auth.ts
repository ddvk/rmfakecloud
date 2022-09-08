import { AxiosResponse } from 'axios'

import requests from '../utils/request'

export function login(email: string, password: string): Promise<AxiosResponse> {
  return requests.post('/ui/api/login', { email, password })
}

export function logout(): Promise<AxiosResponse> {
  return requests.get('/ui/api/logout')
}
