import axios, { type AxiosRequestConfig } from 'axios'
import { showLoadingToast, showFailToast, closeToast } from 'vant'
import userPerson from '../pinia/person'
import lang from '@/i18n/index'
import { adaptRequest } from './backendAdapter'

export interface RequestConfig extends AxiosRequestConfig {
  silent?: boolean
}

interface RequestClient {
  get<T = any>(url: string, config?: RequestConfig): Promise<T>
  post<T = any>(url: string, data?: any, config?: RequestConfig): Promise<T>
  put<T = any>(url: string, data?: any, config?: RequestConfig): Promise<T>
  delete<T = any>(url: string, config?: RequestConfig): Promise<T>
}

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
  config?: RequestConfig,
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
    if (!data?.noMsg && !config?.params?.noMsg && !config?.silent) {
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

const request: RequestClient = {
  get(url: string, config?: RequestConfig) {
    return dispatch('get', url, undefined, config)
  },
  post(url: string, data?: any, config?: RequestConfig) {
    return dispatch('post', url, data, config)
  },
  put(url: string, data?: any, config?: RequestConfig) {
    return dispatch('put', url, data, config)
  },
  delete(url: string, config?: RequestConfig) {
    return dispatch('delete', url, undefined, config)
  },
}

export default request
