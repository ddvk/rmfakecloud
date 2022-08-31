import axios from 'axios'
import { StatusCodes } from 'http-status-codes'

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
    }

    return Promise.reject(error)
  }
)

export default axiosInstance
