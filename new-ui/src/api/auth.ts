import { AxiosResponse } from 'axios'

import requests from '../utils/request'

export function login(email: string, password: string): Promise<AxiosResponse> {
  return requests.post('/ui/api/login', { email, password })
}

export function logout(): Promise<AxiosResponse> {
  return requests.get('/ui/api/logout')
}

export function getUserProfile(): Promise<AxiosResponse> {
  return requests.get('/ui/api/profile')
}

export function resetPassword(data: {
  userid: string
  currentPassword: string
  newPassword: string
}): Promise<AxiosResponse> {
  return requests.post('/ui/api/changePassword', data)
}
