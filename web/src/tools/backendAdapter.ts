import axios from 'axios'

const raw = axios.create({
  baseURL: '',
  timeout: 30000,
})

function getToken() {
  return localStorage.getItem('token') || ''
}

function authHeaders() {
  const token = getToken()
  return token ? { Authorization: `Bearer ${token}` } : {}
}

function clearAuth() {
  localStorage.removeItem('token')
  localStorage.removeItem('account')
}

function isUnauthorized(err: any) {
  return err?.response?.status === 401 || err?.response?.data?.reason === 'UNAUTHORIZED'
}

async function authGet(url: string, params?: Record<string, any>) {
  const token = getToken()
  if (!token) {
    const err: any = new Error('请先登录')
    err.response = { status: 401, data: { message: '请先登录', reason: 'UNAUTHORIZED' } }
    throw err
  }
  try {
    return await raw.get(url, { headers: authHeaders(), params: params || {} })
  } catch (err: any) {
    if (isUnauthorized(err)) {
      clearAuth()
    }
    throw err
  }
}

async function authPost(url: string, data?: Record<string, any>) {
  const token = getToken()
  if (!token) {
    const err: any = new Error('请先登录')
    err.response = { status: 401, data: { message: '请先登录', reason: 'UNAUTHORIZED' } }
    throw err
  }
  try {
    // token 仅放在 Authorization 请求头，不再写入 body
    return await raw.post(url, { ...(data || {}) }, { headers: authHeaders() })
  } catch (err: any) {
    if (isUnauthorized(err)) {
      clearAuth()
    }
    throw err
  }
}

function pathOnly(url: string) {
  return url.split('?')[0]
}

function queryParams(url: string) {
  const idx = url.indexOf('?')
  if (idx < 0) return {}
  const qs = new URLSearchParams(url.slice(idx + 1))
  const out: Record<string, string> = {}
  qs.forEach((v, k) => {
    out[k] = v
  })
  return out
}

async function login(address: string, inviteCode: string, signature: string) {
  try {
    const res = await raw.post('/v1/auth/login', {
      address,
      signature,
      invite_code: inviteCode || '',
    })
    return res.data
  } catch (err: any) {
    const msg = err?.response?.data?.message || err?.message || 'login failed'
    throw new Error(msg)
  }
}

function formatUnixTime(v: any) {
  const n = Number(v)
  if (!n) return ''
  const d = new Date(n * 1000)
  const pad = (x: number) => String(x).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
}

function mapGoods(products: any[]) {
  return (products || []).map((item) => ({
    id: item.id,
    one: item.name || '',
    two: item.description || '',
    three: '',
    four: String(item.price || '0'),
    amount: String(item.price || '0'),
    name: item.name || '',
    status: item.status != null ? String(item.status) : '1',
  }))
}

function mapClaimRecords(list: any[]) {
  return (list || []).map((item) => {
    const createdAt = formatUnixTime(item.created_at)
    return {
      amount: String(item.amount || '0'),
      amountTwo: String(item.amount || '0'),
      reward: String(item.amount || '0'),
      address: '',
      createdAt,
      created_at: createdAt,
      type: 'claim',
    }
  })
}

function calcOrderDailyRelease(o: any, withdrawReset = false) {
  const principal = numOrZero(o.total_amount)
  const released = numOrZero(o.released_amount)
  let exitTarget = numOrZero(o.exit_target)
  const mul = numOrZero(o.exit_multiplier) || 1
  if (exitTarget <= 0) exitTarget = principal * mul
  const status = String(o.status || '').toLowerCase()
  if (status === 'completed' || status === '2') return 0
  if (released >= exitTarget) return 0
  let rate = 1.4
  if (!withdrawReset) {
    const idx = Math.max(0, Math.min(16, Number(o.cycle_day ?? 0)))
    rate = 0.6 + idx * 0.05
  }
  const want = principal * (rate / 100) // 每日静态释放 = 释放系数 × 认购金额；系数 = 日比例%/100
  const remaining = Math.max(0, exitTarget - released)
  return Math.min(want, remaining)
}

function sumOrderReleaseStats(orderList: any[], withdrawReset = false) {
  let exitTotal = 0
  let releasedTotal = 0
  let dailyReleaseTotal = 0
  let unexitedTotal = 0
  let exitCount = 0
  for (const o of orderList || []) {
    const principal = numOrZero(o.total_amount)
    const released = numOrZero(o.released_amount)
    let exitTarget = numOrZero(o.exit_target)
    const mul = numOrZero(o.exit_multiplier) || 1
    if (exitTarget <= 0) exitTarget = principal * mul
    exitTotal += exitTarget
    releasedTotal += released
    unexitedTotal += Math.max(0, exitTarget - released)
    dailyReleaseTotal += calcOrderDailyRelease(o, withdrawReset)
    if (o.status === 'completed' || o.status === '2') exitCount += 1
  }
  return {
    exitTotal,
    releasedTotal,
    unexitedTotal,
    pendingTotal: dailyReleaseTotal,
    exitCount,
    totalNodes: (orderList || []).length,
  }
}

async function fetchUserInfo() {
  if (!getToken()) {
    throw new Error('请先登录')
  }
  try {
    // 核心接口：失败才整体失败；releases 仅用于静态收益，失败不阻断
    const [profileRes, balanceRes, ordersRes] = await Promise.all([
      authGet('/v1/auth/profile'),
      authGet('/v1/wallet/balance'),
      authGet('/v1/wallet/orders'),
    ])
    let releaseList: any[] = []
    let referralList: any[] = []
    // 后端暂未提供生态奖励流水；累计值使用 profile.eco_reward_total。
    const ecoList: any[] = []
    let referralLoaded = false
    try {
      const releasesRes = await authGet('/v1/wallet/releases')
      releaseList = apiBody(releasesRes).records || []
    } catch {
      releaseList = []
    }
    try {
      const referralsRes = await authGet('/v1/wallet/referral-rewards')
      referralList = apiBody(referralsRes).rewards || []
      referralLoaded = true
    } catch {
      referralList = []
    }
    const p = apiBody(profileRes)
    const b = apiBody(balanceRes)
    // AIX 模式不提供传统商城商品；保留空数组兼容上游页面字段。
    const goods: any[] = []
    const orderList = apiBody(ordersRes).orders || []
    const orderStats = sumOrderReleaseStats(orderList)

    const buyTotal = orderList.reduce((sum: number, o: any) => sum + numOrZero(o.total_amount), 0)
    const inviteeCount = Number(p.invitee_count || 0)
    const totalDownline = Number(p.total_downline_count || 0)

    // 奖励结算：可领取=释放余额池；已领取=领取到钱包累计（claim_records 累加）
    const poolBalance = numOrZero(b.released_balance)
    const pending = numOrZero(b.pending_amount)
    const releasedOnOrders = orderStats.releasedTotal
    const claimedTotal = numOrZero(b.claimed_amount)
    const totalNodes = Number(b.total_nodes || orderStats.totalNodes || 0)
    const exitCount = orderStats.exitCount
    const exitTotal = numOrZero(orderStats.exitTotal)
    const claimableAmount = Math.max(
      0,
      b.claimable_amount != null && b.claimable_amount !== ''
        ? numOrZero(b.claimable_amount)
        : poolBalance
    )
    const staticFromReleases = releaseList.reduce((sum: number, r: any) => sum + numOrZero(r.amount), 0)
    // 无释放流水时，用订单已释放金额作为静态收益展示
    const staticTotal = staticFromReleases > 0 ? staticFromReleases : releasedOnOrders
    const serverTime = Number(b.server_time || Math.floor(Date.now() / 1000))

    // 直推=1代推荐奖；团队收益=2代及以上；平级=eco equal；全网=eco total
    let directProfit = 0
    let teamProfit = 0
    for (const r of referralList) {
      const amt = numOrZero(r.reward_amount)
      if (Number(r.generation) > 1) teamProfit += amt
      else directProfit += amt
    }
    const equalProfit = ecoList.reduce((sum: number, r: any) => sum + numOrZero(r.equal_reward), 0)
    let ecoBaseProfit = ecoList.reduce((sum: number, r: any) => sum + numOrZero(r.base_reward), 0)
    const ecoTotalFromList = ecoList.reduce((sum: number, r: any) => sum + numOrZero(r.total_reward), 0)
    // 兼容旧数据：仅有 total_reward 时，基础奖 = 合计 - 平级
    if (ecoBaseProfit <= 0 && ecoTotalFromList > 0) {
      ecoBaseProfit = Math.max(0, ecoTotalFromList - equalProfit)
    }
    const ecoTotal = ecoTotalFromList || numOrZero(p.eco_reward_total)
    const generationProfit = referralLoaded ? directProfit + teamProfit : numOrZero(p.share_profit_total)
    // 收益明细：最近结算日当日值（每天结算后变化）
    const dayReward = latestDayRewardBreakdown(releaseList, referralList, ecoList)

    return {
      status: 'ok',
      level: p.community_level || '0',
      communityLevel: (() => {
        const lv = String(p.community_level || '').trim().toUpperCase()
        if (!lv || lv === '0') return ''
        return lv.startsWith('V') ? lv : `V${lv}`
      })(),
      locationNum: String(inviteeCount),
      communityStake: p.community_stake || '0',
      // 总业绩=伞下；小区=community_stake；大区=总-小区；等级按小区
      total: p.team_stake || '0',
      max: String(Math.max(0, numOrZero(p.team_stake) - numOrZero(p.community_stake))),
      min: p.community_stake || '0',
      inviteUserAddress: p.inviter_address || '',
      inviteUrl: p.address || '',
      recommendNum: inviteeCount,
      recommendTeamNum: totalDownline,
      LocationList: p.address ? [{ address: p.address }] : [],
      buy: buyTotal.toFixed(2),
      // 可实际划入钱包的释放余额池（提现操作以此为准）
      amountGetSub: String(poolBalance),
      // 我的资产：已产量 = 所有订单已释放金额合计
      amountGet: String(releasedOnOrders),
      // 待释放金额
      amountLast: String(pending),
      // 出局次数 = 已出局订单数
      outNum: String(exitCount),
      time: serverTime,
      nextReleaseAt: Number(b.next_release_at || 0),
      // 奖励结算页字段：可领取 = 释放余额池（可划入钱包）
      claimableAmount: String(claimableAmount),
      pendingAmount: String(pending),
      claimedAmount: String(claimedTotal),
      // 未出局金额 = Σ max(出局目标 - 已释放, 0)
      unexitedAmount: String(
        Math.max(
          0,
          b.unexited_amount != null && b.unexited_amount !== ''
            ? numOrZero(b.unexited_amount)
            : orderStats.unexitedTotal
        )
      ),
      totalNodes: String(totalNodes),
      exitTotal: String(exitTotal),
      // 静态收益（累计，资产页用）
      location: String(staticTotal),
      recommend: referralLoaded ? String(directProfit) : p.share_profit_total || '0',
      recommendTwo: '0.00',
      team: referralLoaded ? String(teamProfit) : '0',
      teamTwo: String(equalProfit),
      // 代数奖励合计（1代+≥2代，累计）
      generationReward: String(generationProfit),
      // 社区基础奖累计（不含平级）
      ecoBase: String(ecoBaseProfit),
      // 收益明细（最近结算日）
      rewardDetailDate: dayReward.date,
      dailyStatic: String(dayReward.staticIncome),
      dailyGeneration: String(dayReward.generation),
      dailyEcoBase: String(dayReward.ecoBase),
      dailyEqual: String(dayReward.equal),
      dailyRewardTotal: String(dayReward.total),
      all: String(ecoTotal),
      usdt: b.balance || '0',
      amountUsdt: b.balance || '0',
      balanceUsdt: b.balance || '0',
      balanceBiw: '0',
      biwPrice: '0',
      exchangeRate: '0',
      withdrawRate: 0.06,
      withdrawMin: 10,
      raw: '0.00',
      withdrawRateTwo: 0,
      withdrawMinTwo: 0,
      notice: '',
      goods,
      goodsThree: goods, // 兼容旧商城页字段
      userAddress: [],
      one: '',
      two: '',
      three: '',
      four: '',
      five: '',
      six: '',
      seven: '',
    }
  } catch (err: any) {
    if (isUnauthorized(err)) {
      throw new Error(err?.response?.data?.message || '登录过期')
    }
    throw err
  }
}

function mapWithdrawals(list: any[]) {
  return list.map((item) => {
    const createdAt = formatUnixTime(item.created_at)
    return {
      amount: item.amount,
      fee: item.fee,
      net_amount: item.net_amount,
      relAmount: item.net_amount,
      asset: item.asset || 'AIX',
      status: item.status,
      tx_hash: item.tx_hash,
      createdAt,
      created_at: createdAt,
      to_address: item.to_address,
    }
  })
}

function mapRecharges(list: any[]) {
  return list.map((item) => {
    const createdAt = formatUnixTime(item.created_at)
    return {
      amount: item.amount,
      status: item.status,
      tx_hash: item.tx_hash,
      createdAt,
      created_at: createdAt,
    }
  })
}

function mapOrders(list: any[]) {
  return (list || []).map((item) => {
    const createdAt = formatUnixTime(item.created_at)
    const total = numOrZero(item.total_amount)
    const released = numOrZero(item.released_amount)
    let exitTarget = numOrZero(item.exit_target)
    const mul = numOrZero(item.exit_multiplier) || 1
    if (exitTarget <= 0) {
      exitTarget = total * mul
    }
    const pending = Math.max(0, exitTarget - released)
    const name = item.product_name || '认购'
    // 认购订单记录：对接首页「认购记录」展示字段
    return {
      id: item.id,
      amount: String(item.total_amount ?? '0'),
      amountGet: String(item.released_amount ?? '0'), // 已释放（订单累计释放进度）
      amountLast: String(Number(pending.toFixed(8))), // 待释放（出局目标 - 已释放）
      product_name: name,
      four: name, // 名称
      three: '-', // 收货人（链上认购无收货信息）
      one: '-', // 收货地址
      released_amount: item.released_amount,
      exit_target: item.exit_target,
      exit_multiplier: item.exit_multiplier,
      // 状态：1=收益中 2=已出局；倍率位前端复用 status，同步给出倍数文案字段
      status: item.status === 'completed' ? '2' : '1',
      rate: String(item.exit_multiplier || mul),
      release_day: item.release_day,
      created_at: createdAt,
      createdAt,
    }
  })
}

function numOrZero(v: any) {
  const n = Number(v)
  return Number.isFinite(n) ? n : 0
}

function mapReleaseRecords(records: any[]) {
  return (records || []).map((item) => {
    const createdAt = formatUnixTime(item.created_at)
    const releaseAmount = String(item.amount || '0')
    return {
      // 静态释放：右侧「奖励」展示释放金额；数量可展示日比例
      amount: item.rate != null && item.rate !== '' ? String(item.rate) : releaseAmount,
      amountTwo: releaseAmount,
      reward: releaseAmount,
      num: '',
      name: '静态收益',
      rate: item.rate,
      address: '',
      createdAt,
      created_at: createdAt,
      settlement_date: item.settlement_date,
      type: 'release',
    }
  })
}

function mapReferralRecords(records: any[]) {
  return (records || []).map((item) => {
    const createdAt = formatUnixTime(item.created_at)
    const orderIndex = item.order_index != null && item.order_index !== 0 ? String(item.order_index) : ''
    const referralAmt = String(item.reward_amount || '0')
    return {
      amount: orderIndex || String(item.base_amount || '0'),
      amountTwo: referralAmt,
      reward: referralAmt,
      num: String(item.generation || ''),
      name: '代数奖励',
      rate: item.rate,
      address: item.source_address || '',
      createdAt,
      created_at: createdAt,
      settlement_date: item.settlement_date,
      type: 'referral',
    }
  })
}

function apiBody(res: any) {
  return res?.data ?? res ?? {}
}

/** 取用户最近一个结算日的收益拆分（收益明细按日变化） */
function latestDayRewardBreakdown(releases: any[], referrals: any[], ecos: any[]) {
  let latest = ''
  const touch = (d: any) => {
    const key = String(d || '').trim()
    if (key && key > latest) latest = key
  }
  for (const r of releases || []) touch(r.settlement_date)
  for (const r of referrals || []) touch(r.settlement_date)
  for (const r of ecos || []) touch(r.settlement_date)

  if (!latest) {
    return { date: '', staticIncome: 0, generation: 0, ecoBase: 0, equal: 0, total: 0 }
  }

  let staticIncome = 0
  let generation = 0
  let ecoBase = 0
  let equal = 0
  for (const r of releases || []) {
    if (String(r.settlement_date || '') !== latest) continue
    staticIncome += numOrZero(r.amount)
  }
  for (const r of referrals || []) {
    if (String(r.settlement_date || '') !== latest) continue
    generation += numOrZero(r.reward_amount)
  }
  for (const r of ecos || []) {
    if (String(r.settlement_date || '') !== latest) continue
    const base = numOrZero(r.base_reward)
    const eq = numOrZero(r.equal_reward)
    const total = numOrZero(r.total_reward)
    equal += eq
    if (base > 0) {
      ecoBase += base
    } else if (total > 0) {
      ecoBase += Math.max(0, total - eq)
    }
  }
  return {
    date: latest,
    staticIncome,
    generation,
    ecoBase,
    equal,
    total: staticIncome + generation + ecoBase + equal,
  }
}

/** 我的团队：按日一条。数量=进行中认购订单数（购买+1/结束-1），代数=几代人，奖励=本人静态+每代+社区 */
function mapTeamRewardRecords(releases: any[], referrals: any[], ecos: any[], orders: any[] = []) {
  type DayAgg = {
    staticIncome: number
    referral: number
    eco: number
    generations: Set<number>
    createdAt: string
    sortKey: number
    settlementDate: string
  }
  const byDate = new Map<string, DayAgg>()

  const ensure = (date: string, createdAtRaw: any): DayAgg => {
    const key = String(date || formatUnixTime(createdAtRaw).slice(0, 10) || '')
    let day = byDate.get(key)
    if (!day) {
      const createdAt = date ? `${date} 00:00:00` : formatUnixTime(createdAtRaw)
      day = {
        staticIncome: 0,
        referral: 0,
        eco: 0,
        generations: new Set(),
        createdAt,
        sortKey: date ? Date.parse(`${date}T00:00:00+08:00`) || 0 : Number(createdAtRaw) || 0,
        settlementDate: key,
      }
      byDate.set(key, day)
    }
    return day
  }

  const touchTime = (day: DayAgg, createdAtRaw: any) => {
    const ts = Number(createdAtRaw) || day.sortKey
    if (ts > day.sortKey) {
      day.sortKey = ts
      day.createdAt = formatUnixTime(createdAtRaw)
    }
  }

  for (const item of releases || []) {
    const day = ensure(item.settlement_date, item.created_at)
    // 静态 = 金额列（释放数量 * 出局倍数）
    const staticAmt =
      numOrZero(item.money) ||
      numOrZero(item.amount) * (numOrZero(item.exit_multiplier) || 1) ||
      numOrZero(item.amount)
    day.staticIncome += staticAmt
    day.referral += numOrZero(item.referral_distributed)
    if (numOrZero(item.referral_distributed) > 0) day.generations.add(1)
    touchTime(day, item.created_at)
  }
  for (const item of referrals || []) {
    const day = ensure(item.settlement_date, item.created_at)
    day.referral += numOrZero(item.reward_amount)
    const gen = Number(item.generation)
    if (Number.isFinite(gen) && gen > 0) day.generations.add(gen)
    touchTime(day, item.created_at)
  }
  for (const item of ecos || []) {
    const day = ensure(item.settlement_date, item.created_at)
    day.eco += numOrZero(item.total_reward)
    touchTime(day, item.created_at)
  }

  return [...byDate.entries()]
    .filter(([key]) => key)
    .sort((a, b) => b[1].sortKey - a[1].sortKey)
    .map(([, day]) => {
      // 奖励 = 本人静态收益 + 每代奖励 + 社区奖励
      const total = day.staticIncome + day.referral + day.eco
      return {
        amount: String(activeSubscribeCountOnDate(orders, day.settlementDate)),
        amountTwo: total.toFixed(4),
        reward: total.toFixed(4),
        num: String(day.generations.size || (day.referral > 0 ? 1 : 0)),
        name: '团队奖励',
        address: '',
        createdAt: day.createdAt,
        created_at: day.createdAt,
        type: 'team',
        staticAmount: day.staticIncome.toFixed(4),
        referralAmount: day.referral.toFixed(4),
        ecoAmount: day.eco.toFixed(4),
      }
    })
}

/** 购买增加、出局结束减少：结算日仍有效的认购订单数量 */
function activeSubscribeCountOnDate(orders: any[], settlementDate: string): number {
  if (!orders?.length) return 0
  const dayEnd = Date.parse(`${settlementDate}T23:59:59+08:00`) || Number.MAX_SAFE_INTEGER
  return orders.filter((o) => {
    let createdMs = numOrZero(o.created_at)
    if (createdMs > 0 && createdMs < 1e12) createdMs *= 1000
    if (createdMs > dayEnd) return false
    const status = String(o.status || '').toLowerCase()
    if (status === 'completed' || status === '2') return false
    return true
  }).length
}

function mapEcoRecords(
  records: any[],
  opts?: { rewardField?: 'total_reward' | 'equal_reward' | 'base_reward'; name?: string; type?: string }
) {
  const field = opts?.rewardField || 'total_reward'
  const name = opts?.name || '社区奖励'
  const type = opts?.type || 'eco'
  return (records || []).map((item) => {
    const createdAt = formatUnixTime(item.created_at)
    const reward = String(item[field] || '0')
    return {
      amount: String(item.base_amount || item.community_stake || '0'),
      amountTwo: reward,
      reward,
      num: item.community_level || '',
      name,
      rate: item.base_rate,
      address: '',
      createdAt,
      created_at: createdAt,
      settlement_date: item.settlement_date,
      type,
    }
  })
}

function mapBuyRecords(orders: any[]) {
  return mapOrders(orders).map((item) => ({
    amount: item.amount,
    amountTwo: item.amount,
    amountGet: item.amountGet,
    amountLast: item.amountLast,
    reward: item.amount,
    num: '',
    name: item.four || item.product_name || '认购',
    four: item.four,
    three: item.three,
    one: item.one,
    address: '',
    createdAt: item.createdAt,
    created_at: item.createdAt,
    type: 'buy',
    status: item.status,
    rate: item.rate,
  }))
}

function mapInvitees(items: any[]) {
  return (items || []).map((item) => {
    // 数量 = 用户出局金额（订单出局目标合计）
    const exitAmount =
      item.exit_amount != null && item.exit_amount !== ''
        ? numOrZero(item.exit_amount)
        : numOrZero(item.team_stake)
    const level = String(item.community_level || '').trim()
    return {
      address: item.address,
      amount: exitAmount.toFixed(4),
      level: level && level !== '0' ? level.toUpperCase() : '',
      countLow: item.direct_count || 0,
    }
  })
}

export async function adaptRequest(
  method: 'get' | 'post' | 'put' | 'delete',
  url: string,
  data?: any,
  params?: any
) {
  const path = pathOnly(url)
  const mergedParams = { ...queryParams(url), ...(params || {}) }

  switch (path) {
    case 'app_server/eth_authorize': {
      try {
        const loginRes = await login(data?.address, data?.code || '', data?.sign)
        if (!loginRes?.token) {
          return { status: 'fail', message: loginRes?.message || 'login failed' }
        }
        return { status: 'ok', token: loginRes.token }
      } catch (err: any) {
        const msg = err?.response?.data?.message || err?.message || 'login failed'
        if (/首次登录需要邀请码|请输入推荐码/.test(msg)) return { status: '请输入推荐码' }
        if (/邀请码/.test(msg)) return { status: '无效的推荐码' }
        return { status: 'fail', message: msg }
      }
    }
    case 'app_server/user_info':
      return fetchUserInfo()
    case 'app_server/subscribe_tiers':
      return {
        min_subscribe_amount: 100,
        tiers: [100, 500, 1000, 3000],
      }
    case 'app_server/recommend_update':
      return { status: 'ok', inviteUserAddress: mergedParams.code || '' }
    case 'app_server/withdraw_list': {
      const res = await authGet('/v1/wallet/withdrawals')
      const list = mapWithdrawals(res.data?.withdrawals || [])
      return { count: list.length, list }
    }
    case 'app_server/withdraw': {
      const profile = await authGet('/v1/auth/profile')
      const toAddress = data?.to_address || data?.address || profile.data?.address || ''
      const res = await authPost('/v1/wallet/withdraw-aix', {
        amount: String(data?.amount || ''),
        to_address: toAddress,
      })
      return { status: 'ok', ...res.data }
    }
    case 'app_server/deposit_list': {
      const res = await authGet('/v1/wallet/recharges')
      const list = mapRecharges(res.data?.recharges || [])
      return { count: list.length, list }
    }
    case 'app_server/deposit': {
      const res = await authPost('/v1/wallet/recharge', {
        amount: String(data?.amount || ''),
      })
      return { status: 'ok', ...res.data }
    }
    case 'app_server/deposit_confirm': {
      const hashes = Array.isArray(data?.tx_hashes)
        ? data.tx_hashes
        : Array.isArray(data?.txHashes)
          ? data.txHashes
          : String(data?.tx_hash || data?.txHash || '')
              .split(/[,;|]/)
              .map((v: string) => v.trim())
              .filter(Boolean)
      const res = await authPost('/v1/wallet/recharge/confirm', {
        recharge_id: Number(data?.recharge_id || data?.rechargeId || 0),
        tx_hash: String(data?.tx_hash || data?.txHash || hashes.join(',') || ''),
        tx_hashes: hashes,
        signature: String(data?.signature || data?.sign || ''),
      })
      return { status: 'ok', ...res.data }
    }
    case 'app_server/order_list':
    case 'app_server/order_two_list':
    case 'app_server/order_three_list':
    case 'app_server/order_four_list': {
      const res = await authGet('/v1/wallet/orders')
      const list = mapOrders(res.data?.orders || [])
      return { count: list.length, list }
    }
    case 'app_server/reward_list': {
      // 兼容 $ref / 数字 / 字符串
      const rawType = mergedParams.reqType
      const reqType = String(
        rawType && typeof rawType === 'object' && 'value' in rawType
          ? (rawType as any).value
          : rawType ?? '1'
      )
      const page = Math.max(1, Number(mergedParams.page || 1))
      const pageSize = 10
      let records: any[] = []

      // 1=认购 | 2=静态 | 3/4=直推(1代) | 5=团队(≥2代) | 6=平级 | 7=社区合计 | team=按日汇总 | 11=领取
      if (reqType === '1') {
        const res = await authGet('/v1/wallet/orders')
        records = mapBuyRecords(apiBody(res).orders || [])
      } else if (reqType === '2') {
        const res = await authGet('/v1/wallet/releases')
        records = mapReleaseRecords(apiBody(res).records || [])
      } else if (reqType === '3' || reqType === '4') {
        const res = await authGet('/v1/wallet/referral-rewards')
        const all = apiBody(res).rewards || []
        records = mapReferralRecords(all.filter((r: any) => Number(r.generation || 1) <= 1))
      } else if (reqType === 'gen' || reqType === 'referral') {
        const res = await authGet('/v1/wallet/referral-rewards')
        records = mapReferralRecords(apiBody(res).rewards || [])
      } else if (reqType === '5') {
        const res = await authGet('/v1/wallet/referral-rewards')
        const all = apiBody(res).rewards || []
        records = mapReferralRecords(all.filter((r: any) => Number(r.generation) > 1))
      } else if (reqType === '11') {
        records = []
      } else if (reqType === '6') {
        records = []
      } else if (reqType === 'eco_base') {
        records = []
      } else if (reqType === '7') {
        records = []
      } else if (reqType === 'team') {
        const [releases, referrals, orders] = await Promise.all([
          authGet('/v1/wallet/releases'),
          authGet('/v1/wallet/referral-rewards'),
          authGet('/v1/wallet/orders'),
        ])
        records = mapTeamRewardRecords(
          apiBody(releases).records || [],
          apiBody(referrals).rewards || [],
          [],
          apiBody(orders).orders || []
        )
      } else if (reqType === 'all' || reqType === '0') {
        const [releases, referrals] = await Promise.all([
          authGet('/v1/wallet/releases'),
          authGet('/v1/wallet/referral-rewards'),
        ])
        records = [
          ...mapReleaseRecords(apiBody(releases).records || []),
          ...mapReferralRecords(apiBody(referrals).rewards || []),
        ].sort((a, b) => String(b.createdAt).localeCompare(String(a.createdAt)))
      } else {
        records = []
      }

      const total = records.length
      const start = (page - 1) * pageSize
      return { count: total, list: records.slice(start, start + pageSize) }
    }
    case 'app_server/recommend_list': {
      const targetAddress = mergedParams.address || ''
      const res = await authGet('/v1/auth/invitees', targetAddress ? { address: targetAddress } : {})
      return { recommends: mapInvitees(res.data?.invitees || []) }
    }
    case 'app_server/amount_to': {
      const res = await authPost('/v1/wallet/claim', {
        amount: String(data?.amount || ''),
      })
      return { status: 'ok', ...res.data }
    }
    case 'app_server/buy':
    case 'app_server/buy_two':
    case 'app_server/buy_three':
    case 'app_server/buy_four': {
      const amount = data?.amount != null && data?.amount !== '' ? String(data.amount) : ''
      if (!amount) return { status: 'fail', message: '请输入认购金额' }
      const payFrom = data?.pay_from === 'reward' || data?.payFrom === 'reward'
        ? 'reward'
        : 'recharge'
      const res = await authPost('/v1/wallet/subscribe-aix', {
        amount,
        pay_from: payFrom,
      })
      return { status: 'ok', ...res.data }
    }
    case 'app_server/exchange':
    case 'app_server/set_info':
    case 'app_server/set_address':
    case 'app_server/delete_address':
      return { status: 'fail', message: '功能暂未开放' }
    default:
      return raw.request({ method, url, data, params, headers: authHeaders() }).then((r) => r.data)
  }
}

export async function fetchChallenge(address: string) {
  const res = await raw.get('/v1/auth/challenge', { params: { address } })
  return res.data
}
