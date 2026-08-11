import axios from 'axios'

const api = axios.create({
  baseURL: String(import.meta.env.VITE_API || '').replace(/\/+$/, ''),
  timeout: 30000,
})

const token = () => localStorage.getItem('token') || ''
const authHeaders = () => token() ? { Authorization: `Bearer ${token()}` } : {}

async function get<T = any>(url: string, params?: Record<string, any>): Promise<T> {
  const response = await api.get(url, { headers: authHeaders(), params: params || {} })
  return response.data
}

async function post<T = any>(url: string, data?: Record<string, any>): Promise<T> {
  const response = await api.post(url, data || {}, { headers: authHeaders() })
  return response.data
}

export function clearAuth() {
  localStorage.removeItem('token')
  localStorage.removeItem('account')
  localStorage.removeItem('sign')
}

export function errMsg(error: any, fallback = '请求失败') {
  return error?.response?.data?.message || error?.message || fallback
}

export function fetchChallenge(address: string) {
  return get<{ address: string; message: string; expire_at: number }>('/v1/auth/challenge', { address })
}

export function login(address: string, signature: string, inviteCode: string) {
  return post<any>('/v1/auth/login', {
    address,
    signature,
    invite_code: inviteCode || '',
  })
}

export function getAixProfile() {
  return get('/v1/wallet/aix-profile')
}

export function getAixBalance() {
  return get('/v1/wallet/balance')
}

export function subscribeAix(amount: string, payFrom: 'recharge' | 'reward') {
  return post('/v1/wallet/subscribe-aix', { amount, pay_from: payFrom })
}

function formatUnixTime(value: unknown): string {
  const timestamp = Number(value)
  if (!Number.isFinite(timestamp) || timestamp <= 0) return String(value || '')
  const date = new Date(timestamp < 1e12 ? timestamp * 1000 : timestamp)
  const pad = (part: number) => String(part).padStart(2, '0')
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}:${pad(date.getSeconds())}`
}

export async function listAixOrders() {
  const result: any = await get('/v1/wallet/orders')
  return {
    ...result,
    orders: (result?.orders || []).map((order: any) => {
      const createdAt = formatUnixTime(order.created_at ?? order.createdAt)
      return { ...order, created_at: createdAt, createdAt }
    }),
  }
}
