import axios from 'axios'
import { StatusCodes } from 'http-status-codes'
import { toast } from 'react-toastify'

import i18n from '../i18n'

const axiosInstance = axios.create({
  timeout: 30000 // 30s
})

axiosInstance.interceptors.response.use(
  (response) => {
    return response
  },
  (error) => {
    const { response } = error

    if (response) {
      if (response.status === StatusCodes.UNAUTHORIZED && window.location.pathname !== '/login') {
        window.location.href = '/login'
      }
      if (response.status === StatusCodes.BAD_REQUEST && response.data && response.data.error) {
        toast.error(response.data.error)
      }
      if (response.status === StatusCodes.FORBIDDEN) {
        toast.error(i18n.t('notifications.forbidden').toString())
      }
    }

    return Promise.reject(error)
  }
)

export default axiosInstance
