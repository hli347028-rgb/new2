import axios from 'axios'

const raw = axios.create({ baseURL: '', timeout: 30000 })

function token() {
  return localStorage.getItem('token') || ''
}

function authHeaders() {
  const t = token()
  return t ? { Authorization: `Bearer ${t}` } : {}
}

export function clearAuth() {
  localStorage.removeItem('token')
  localStorage.removeItem('account')
}

async function get<T = any>(url: string, params?: Record<string, any>, opts?: { clearOn401?: boolean }): Promise<T> {
  const t = token()
  const mergedParams = { ...(params || {}) }
  // 自定义路由兼容：query token 兜底
  if (t && mergedParams.token == null) {
    mergedParams.token = t
  }
  try {
    const res = await raw.get(url, { headers: authHeaders(), params: mergedParams })
    return res.data
  } catch (err: any) {
    if (opts?.clearOn401 !== false && err?.response?.status === 401) {
      // 登录流程里由调用方决定是否清理，避免误清刚写入的 token
    }
    throw err
  }
}

async function post<T = any>(url: string, data?: Record<string, any>): Promise<T> {
  const body = { ...(data || {}) }
  const t = token()
  if (t && !(body as any).token) {
    ;(body as any).token = t
  }
  try {
    const res = await raw.post(url, body, { headers: authHeaders() })
    return res.data
  } catch (err: any) {
    throw err
  }
}

export function errMsg(err: any, fallback = '请求失败') {
  return err?.response?.data?.message || err?.message || fallback
}

export async function fetchChallenge(address: string) {
  return get<{ address: string; message: string; expire_at: number }>(
    '/v1/auth/challenge',
    { address },
    { clearOn401: false },
  )
}

export async function login(address: string, signature: string, inviteCode: string) {
  const res = await post<any>('/v1/auth/login', {
    address,
    signature,
    invite_code: inviteCode || '',
  })
  const tok = res?.token || res?.Token || res?.data?.token
  if (!tok) {
    throw new Error(res?.message || '登录成功但未返回 token')
  }
  return { ...res, token: tok }
}

export async function getAuthProfile() {
  return get('/v1/auth/profile')
}

export async function listInvitees() {
  return get('/v1/auth/invitees')
}

export async function getAixProfile() {
  return get('/v1/wallet/aix-profile')
}

export async function getBalance() {
  return get('/v1/wallet/balance')
}

export async function getAixPrice(date?: string) {
  return get('/v1/wallet/aix-price', date ? { date } : undefined, { clearOn401: false })
}

export async function subscribeAix(amount: string, payFrom: 'recharge' | 'reward') {
  return post('/v1/wallet/subscribe-aix', { amount, pay_from: payFrom })
}

export async function listOrders() {
  return get('/v1/wallet/orders')
}

export async function createRecharge(amount: string) {
  return post('/v1/wallet/recharge', { amount })
}

export async function confirmRecharge(payload: {
  recharge_id: number
  tx_hash?: string
  tx_hashes?: string[]
  signature?: string
}) {
  return post('/v1/wallet/recharge/confirm', payload)
}

export async function listRecharges() {
  return get('/v1/wallet/recharges')
}

export async function rechargeToReward(amount: string) {
  return post('/v1/wallet/recharge-to-reward', { amount })
}

export async function transfer(payload: {
  to_address: string
  asset: string
  amount: string
  pay_from?: string
}) {
  return post('/v1/wallet/transfer', payload)
}

export async function withdrawAix(amount: string, toAddress?: string) {
  return post('/v1/wallet/withdraw-aix', { amount, to_address: toAddress || '' })
}

export async function listWithdrawals() {
  return get('/v1/wallet/withdrawals')
}

export async function listRewards() {
  return get('/v1/wallet/rewards')
}
