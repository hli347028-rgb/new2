import axios from 'axios'
import { showLoadingToast, showFailToast, closeToast } from 'vant'
import userPerson from '../pinia/person'
import lang from '@/i18n/index'
import { adaptRequest } from './backendAdapter'

const instance = axios.create({
  baseURL: String(import.meta.env.VITE_API || '').replace(/\/+$/, ''),
  timeout: 30000,
})

function isLegacyApi(url = '') {
  return url.startsWith('app_server/')
}

async function dispatch(
  method: 'get' | 'post' | 'put' | 'delete',
  url: string,
  data?: any,
  config?: any,
) {
  const loading = data?.loading || config?.params?.loading
  if (loading) {
    showLoadingToast({
      message: lang('加载中'),
      duration: 0,
      overlay: true,
      overlayStyle: { background: 'transparent' },
    })
  }
  try {
    if (isLegacyApi(url)) {
      return await adaptRequest(method, url, data, config?.params)
    }
    const token = localStorage.getItem('token')
    const headers = { ...(config?.headers || {}) }
    if (token) headers.Authorization = `Bearer ${token}`
    const response = await instance.request({ method, url, data, params: config?.params, headers })
    return response.data
  } catch (error: any) {
    const message = error?.response?.data?.message || error?.message
    const status = error?.response?.status
    if (!data?.noMsg && !config?.params?.noMsg) {
      showFailToast(message || error?.response?.statusText || lang('操作失败'))
    }
    if (status === 401) {
      const person = userPerson()
      person.outLogin()
    }
    throw error
  } finally {
    if (loading) closeToast()
  }
}

const request = {
  get(url: string, config?: any) {
    return dispatch('get', url, undefined, config)
  },
  post(url: string, data?: any, config?: any) {
    return dispatch('post', url, data, config)
  },
  put(url: string, data?: any, config?: any) {
    return dispatch('put', url, data, config)
  },
  delete(url: string, config?: any) {
    return dispatch('delete', url, undefined, config)
  },
}

export default request as typeof axios
