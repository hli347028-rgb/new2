import { ethers } from 'ethers'

const UNIT = 18
const ONE = ethers.utils.parseUnits('1', UNIT)

export interface ExchangeRateInput {
  aix_to_win_rate?: number
  aix_price?: number
  win_price?: number
}

function sanitize(value: string | number): string {
  const text = String(value ?? '0').trim()
  if (!text || text === '.') return '0'
  if (!/^\d*(\.\d*)?$/.test(text)) return '0'
  const [integer = '0', fraction = ''] = text.split('.')
  const normalizedInteger = integer.replace(/^0+(?=\d)/, '') || '0'
  const normalizedFraction = fraction.slice(0, UNIT)
  return normalizedFraction ? `${normalizedInteger}.${normalizedFraction}` : normalizedInteger
}

function toUnits(value: string | number) {
  return ethers.utils.parseUnits(sanitize(value), UNIT)
}

function fromUnits(value: ethers.BigNumber) {
  return ethers.utils.formatUnits(value, UNIT)
}

export function displayDecimal(value: unknown, maxFraction = 8) {
  const text = fromUnits(toUnits(String(value ?? '0')))
  if (!text.includes('.')) return text
  const [integer, fraction = ''] = text.split('.')
  const trimmedFraction = fraction.replace(/0+$/, '').slice(0, maxFraction)
  return trimmedFraction ? `${integer}.${trimmedFraction}` : integer
}

export function isPositiveDecimal(value: string) {
  try {
    return toUnits(value).gt(0)
  } catch {
    return false
  }
}

export function compareDecimals(left: string, right: string) {
  try {
    const diff = toUnits(left).sub(toUnits(right))
    if (diff.isZero()) return 0
    return diff.gt(0) ? 1 : -1
  } catch {
    return 0
  }
}

export function divDecimal(left: string, right: string | number) {
  const divisor = toUnits(right)
  if (divisor.isZero()) return '0'
  return fromUnits(toUnits(left).mul(ONE).div(divisor))
}

export function mulDecimal(left: string, right: string | number) {
  return fromUnits(toUnits(left).mul(toUnits(right)).div(ONE))
}

export function subDecimal(left: string, right: string) {
  return fromUnits(toUnits(left).sub(toUnits(right)))
}

export function formatFeeRate(value: unknown) {
  const rate = Number(value)
  if (!Number.isFinite(rate) || rate <= 0) return '0%'
  const percent = rate * 100
  const text = Number.isInteger(percent) ? String(percent) : percent.toFixed(2).replace(/0+$/, '').replace(/\.$/, '')
  return `${text}%`
}

/** 优先用 aix_to_win_rate，否则用 aix_price / win_price 推算 */
export function resolveAixToWinRate(input: ExchangeRateInput): string | null {
  const direct = Number(input.aix_to_win_rate)
  if (Number.isFinite(direct) && direct > 0) return sanitize(String(direct))
  const aixPrice = Number(input.aix_price)
  const winPrice = Number(input.win_price)
  if (Number.isFinite(aixPrice) && aixPrice > 0 && Number.isFinite(winPrice) && winPrice > 0) {
    return divDecimal(String(aixPrice), String(winPrice))
  }
  return null
}

function calcWinFromAix(aixAmount: string, aixToWinRate: string, exchangeFeeRate: number) {
  const winGross = mulDecimal(aixAmount, aixToWinRate)
  const fee = mulDecimal(winGross, String(exchangeFeeRate || 0))
  const winNet = subDecimal(winGross, fee)
  return { winGross, fee, winNet }
}

/** 1 AIX 可兑换多少 WIN（毛量/净量） */
export function calcUnitExchangeRate(aixToWinRate: string, exchangeFeeRate: number) {
  if (!isPositiveDecimal(aixToWinRate)) return null
  return calcWinFromAix('1', aixToWinRate, exchangeFeeRate)
}

export function calcExchangePreview(aixAmount: string, aixToWinRate: string, exchangeFeeRate: number) {
  if (!isPositiveDecimal(aixAmount) || !isPositiveDecimal(aixToWinRate)) {
    return null
  }
  const result = calcWinFromAix(aixAmount, aixToWinRate, exchangeFeeRate)
  return {
    ...result,
    feeRateText: formatFeeRate(exchangeFeeRate),
  }
}
