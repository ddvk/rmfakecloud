import { AxiosResponse } from 'axios'

import requests from '../utils/request'

export function listUsers(): Promise<AxiosResponse> {
  return requests.get('/ui/api/users')
}

export function removeUser(id: string): Promise<AxiosResponse> {
  return requests.delete(`/ui/api/users/${id}`)
}

export function getUser(id: string): Promise<AxiosResponse> {
  return requests.get(`/ui/api/users/${id}`)
}

export function updateUser({
  user_id,
  email,
  password
}: {
  user_id: string
  email: string
  password: string
}): Promise<AxiosResponse> {
  return requests.put('/ui/api/users', { userid: user_id, email, newpassword: password })
}

export function createUser({
  user_id,
  email,
  password
}: {
  user_id: string
  email: string
  password: string
}): Promise<AxiosResponse> {
  return requests.post('/ui/api/users', { userid: user_id, email, newpassword: password })
}
