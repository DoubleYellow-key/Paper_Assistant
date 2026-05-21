import axios from 'axios'
import { ElMessage } from 'element-plus'
import { useUserStore } from '@/stores/user'
import router from '@/router'

const request = axios.create({
  baseURL: '/api/v1',
  timeout: 30000
})

request.interceptors.request.use(
  (config) => {
    const userStore = useUserStore()
    if (userStore.token) {
      config.headers.Authorization = `Bearer ${userStore.token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

request.interceptors.response.use(
  (response) => {
    const res = response.data
    // 如果是二进制流或者其它格式，直接返回
    if (response.config.responseType === 'blob') {
      return response
    }
    if (res.code !== 0) {
      if (res.code !== 40901) {
        ElMessage.error(res.message || '请求失败')
      }
      if (res.code === 40101) {
        const userStore = useUserStore()
        userStore.logout()
        router.push('/login')
      }
      const error: any = new Error(res.message || 'Error')
      error.code = res.code
      return Promise.reject(error)
    }
    return res.data
  },
  (error) => {
    const res = error.response?.data
    const msg = res?.message || error.message || '系统错误'
    if (res?.code !== 40901) {
      ElMessage.error(msg)
    }
    if (res?.code === 40101 || error.response?.status === 401) {
      const userStore = useUserStore()
      userStore.logout()
      router.push('/login')
    }
    return Promise.reject(error)
  }
)

export default request