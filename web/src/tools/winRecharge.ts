import { Contract, ETH } from '@/tools/contract'
import { compareDecimals } from '@/tools/decimal'
import { ethers } from 'ethers'

export interface WinChainRechargeRecord {
  id: number
  amount: string
  index: number
}

/** 从 BuySomething 合约读取当前用户的 WIN 充值记录（num 份数） */
export async function fetchWinRechargeRecords(address?: string): Promise<WinChainRechargeRecord[]> {
  await ETH.getAccount()
  const userAddress = (address || ETH.account || '').toLowerCase()
  if (!userAddress) return []

  const BUY = new Contract(import.meta.env.VITE_BUY, 'BUY')
  const length = Number(await BUY.call('getUserLength', []))
  if (!Number.isFinite(length) || length <= 0) return []

  const users: string[] = await BUY.call('getUsersByIndex', [0, length - 1])
  const amounts: ethers.BigNumber[] = await BUY.call('getUsersAmountByIndex', [0, length - 1])

  const records: WinChainRechargeRecord[] = []
  users.forEach((user, index) => {
    if (String(user).toLowerCase() !== userAddress) return
    records.push({
      id: index,
      index: index + 1,
      amount: amounts[index]?.toString?.() || String(amounts[index] ?? '0'),
    })
  })

  return records.reverse()
}

/** 链上充值成功后轮询 profile，等待 win_balance 入账 */
export async function pollWinBalance(
  refreshProfile: () => Promise<unknown>,
  beforeBalance: string,
  attempts = 5,
  intervalMs = 2000,
) {
  for (let i = 0; i < attempts; i += 1) {
    await new Promise((resolve) => setTimeout(resolve, intervalMs))
    const profile: any = await refreshProfile()
    const nextBalance = String(profile?.win_balance ?? '0')
    if (compareDecimals(nextBalance, beforeBalance) > 0) {
      return { updated: true, balance: nextBalance }
    }
  }
  return { updated: false, balance: beforeBalance }
}

export function mapWinRechargeError(error: any, t: (key: string, params?: Record<string, unknown>) => string) {
  const reason = String(error?.reason || error?.data?.message || error?.message || '').toLowerCase()
  if (/err num|num/.test(reason)) return t('recharge.minWinRechargeError')
  if (/bad value/.test(reason)) return t('recharge.winBadValueError')
  if (/balance|insufficient/.test(reason)) return t('recharge.winInsufficientNative')
  if (error?.code === 4001 || /user rejected|denied|cancel/i.test(reason)) return t('recharge.userCancelled')
  if (/chain|network|4902/.test(reason)) return t('recharge.wrongChain')
  return ''
}
