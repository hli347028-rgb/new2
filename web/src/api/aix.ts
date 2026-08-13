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
  return get<AixProfile>('/v1/wallet/aix-profile')
}

export interface AixProfile {
  address?: string
  aix_balance?: string
  win_balance?: string
  aix_price?: number
  win_price?: number
  aix_to_win_rate?: number
  exchange_fee_rate?: number
  pending_mgmt_reward?: string
  overflow_reward?: string
  mgmt_reward_total?: string
  usdt_recharge?: string
  usdt_reward?: string
  [key: string]: unknown
}

export function getAixBalance() {
  return get('/v1/wallet/balance')
}

export interface AixWinExchangeResult {
  record_id: number
  from_asset: 'AIX'
  from_amount: string
  to_asset: 'WIN'
  to_amount: string
  exchange_price: string
  exchange_fee_rate: number
  status: 'completed'
  aix_balance: string
  win_balance: string
  created_at: number
}

export interface AixWinExchangeRecord {
  id: number
  from_asset: 'AIX'
  from_amount: string
  to_asset: 'WIN'
  to_amount: string
  fee_amount: string
  fee_rate: string
  exchange_price: string
  status: string
  remark: string
  created_at: number
}

export interface WinWithdrawResult {
  withdraw_id: number
  asset: 'WIN'
  amount: string
  to_address: string
  status: 'pending' | 'completed' | 'failed'
  tx_hash: string
  win_balance: string
  win_contract: string
}

export interface WinWithdrawRecord {
  id: number
  asset: 'WIN'
  amount: string
  fee: string
  net_amount: string
  to_address: string
  status: 'pending' | 'completed' | 'failed'
  tx_hash: string
  remark: string
  created_at: number
  updated_at: number
}

export function exchangeAixToWin(aixAmount: string) {
  return post<AixWinExchangeResult>('/v1/wallet/exchange-aix-to-win', { aix_amount: aixAmount })
}

export function getAixWinExchangeRecords() {
  return get<{ records: AixWinExchangeRecord[] }>('/v1/wallet/exchange-records')
}

export function withdrawWin(amount: string, toAddress?: string) {
  const payload: Record<string, string> = { amount }
  if (toAddress?.trim()) payload.to_address = toAddress.trim()
  return post<WinWithdrawResult>('/v1/wallet/withdraw-win', payload)
}

export function getWinWithdrawRecords() {
  return get<{ records: WinWithdrawRecord[] }>('/v1/wallet/withdraw-records')
}

export function subscribeAix(amount: string, payFrom: 'recharge' | 'reward' | 'win') {
  return post('/v1/wallet/subscribe-aix', { amount, pay_from: payFrom })
}

export interface WinRechargeCreateResult {
  recharge_id: number
  asset: 'WIN'
  amount: string
  deposit_address: string
  deposit_addresses: string[]
  win_contract: string
  win_decimals: number
  token_symbol: 'WIN'
  message: string
  expire_at: number
  dev_mode: boolean
  win_price: number
}

export interface WinRechargeConfirmResult {
  asset: 'WIN'
  amount: string
  win_balance: string
}

export interface WinRechargeRecord {
  id: number
  asset: 'WIN' | string
  amount: string
  tx_hash: string
  status: string
  created_at: number
  confirmed_at?: number
}

/** 创建 WIN 充值单：链上向 deposit_address 转入 WIN 后调用 confirm */
export function createWinRecharge(amount: string) {
  return post<WinRechargeCreateResult>('/v1/wallet/recharge-win', { amount })
}

/** 确认 WIN 充值：校验链上 Transfer 后入账 win_balance */
export function confirmWinRecharge(rechargeId: number, txHash: string, signature: string) {
  return post<WinRechargeConfirmResult>('/v1/wallet/recharge-win/confirm', {
    recharge_id: rechargeId,
    tx_hash: txHash,
    signature,
  })
}

export function listWinRecharges() {
  return get<{ recharges: WinRechargeRecord[] }>('/v1/wallet/recharges-win')
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
