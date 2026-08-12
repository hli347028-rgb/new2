<template>
  <div class="withdrawal-page">
    <Header />

    <div class="content">
      <div class="page-header">
        <h1 class="page-title">{{ $t('withdraw.winSubtitle') }}</h1>
        <p class="page-balance">{{ $t('withdraw.availableBalance') }}: {{ displayAmount(winBalance) }} WIN</p>
        <p class="page-hint">
          {{ $t('withdraw.aixExchangeHint') }}
          <button type="button" class="link-btn" @click="router.push('/exchange')">{{ $t('wallet.exchange') }}</button>
        </p>
      </div>

      <div class="withdraw-form">
        <div class="form-hint-row">
          <p class="form-hint">{{ $t('withdraw.amount') }}</p>
          <button type="button" class="all-btn" @click="handleAllAmount">
            {{ $t('withdraw.all') }}
          </button>
        </div>
        <div class="form-row">
          <input
            class="form-input"
            v-model="amountWin"
            @input="checkWinAmount"
            type="text"
            inputmode="decimal"
            :placeholder="$t('withdraw.enterAmount')"
          />
          <span class="asset-tag">WIN</span>
        </div>
        <p v-if="amountError" class="error-text">{{ amountError }}</p>

        <button
          class="subscribe-btn custom-btn"
          :disabled="!canSubmit || loading"
          @click="handleWithdrawal"
        >
          {{ loading ? $t('withdraw.processing') : $t('withdraw.confirm') }}
        </button>

        <div class="form-info">
          <p>{{ $t('withdraw.fee') }}: 0 WIN</p>
        </div>
      </div>

      <div class="record-section">
        <div class="section-title-wrap">
          <div class="title-bar"></div>
          <h3 class="section-title">{{ $t('withdraw.details') }}</h3>
        </div>

        <div class="table-card">
          <div class="table-header table-header-4">
            <span>{{ $t('node.amount') }}</span>
            <span>{{ $t('withdraw.received') }}</span>
            <span>{{ $t('withdraw.toAddressShort') }}</span>
            <span>{{ $t('withdraw.status') }}</span>
          </div>
          <div class="order-list" v-for="item in amountList" :key="item.id">
            <div class="table-row table-row-4">
              <span>{{ displayAmount(item.amount) }} WIN</span>
              <span>{{ displayAmount(item.net_amount) }} WIN</span>
              <span class="address-cell">{{ formatAddress(item.to_address) }}</span>
              <span class="status-cell">
                {{ withdrawStatusText(item.status) }}
                <small v-if="item.tx_hash" class="tx-hint">{{ String(item.tx_hash).slice(0, 10) }}…</small>
                <small class="muted">{{ formatTime(item.created_at) }}</small>
              </span>
            </div>
          </div>
          <div class="empty-state" v-if="!recordLoading && amountList.length === 0">
            <p>{{ $t('withdraw.noRecords') }}</p>
          </div>
          <div class="state-box" v-if="recordLoading">
            <van-loading color="#1597e5" />
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import Header from '@/components/Header.vue'
import userPerson from '@/pinia/person'
import { getWinWithdrawRecords, withdrawWin } from '@/api/aix'
import type { WinWithdrawRecord } from '@/api/aix'
import { compareDecimals, displayDecimal, isPositiveDecimal } from '@/tools/decimal'
import { showToast } from 'vant'
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'

const router = useRouter()
const { t: $t } = useI18n()
const person = userPerson()

const amountWin = ref('')
const loading = ref(false)
const recordLoading = ref(false)
const amountList = ref<WinWithdrawRecord[]>([])
const pollTimer = ref<ReturnType<typeof setInterval> | null>(null)

const hasPendingRecords = computed(() => amountList.value.some((item) => item.status === 'pending'))

const winBalance = computed(() => String(person.profile?.win_balance || '0'))

const amountError = computed(() => {
  if (!amountWin.value) return ''
  if (!isPositiveDecimal(amountWin.value)) return $t('withdraw.enterAmount')
  if (compareDecimals(amountWin.value, winBalance.value) > 0) return $t('withdraw.insufficientWin')
  return ''
})

const canSubmit = computed(() => Boolean(amountWin.value) && !amountError.value)

const withdrawStatusText = (status: string) => {
  switch (status) {
    case 'pending': return $t('withdraw.statusPending')
    case 'completed': return $t('withdraw.statusCompleted')
    case 'failed': return $t('withdraw.statusFailed')
    default: return status || '-'
  }
}

const displayAmount = (value: unknown) => displayDecimal(value)

const formatAddress = (value?: string) => {
  if (!value) return '-'
  return value.length > 14 ? `${value.slice(0, 6)}...${value.slice(-4)}` : value
}

const formatTime = (value: number) => {
  const date = new Date(Number(value) * 1000)
  if (Number.isNaN(date.getTime())) return ''
  const pad = (part: number) => String(part).padStart(2, '0')
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}`
}

const getAmountList = async () => {
  recordLoading.value = true
  try {
    const result = await getWinWithdrawRecords()
    amountList.value = result?.records || []
    syncWithdrawPolling()
  } catch {
    amountList.value = []
    stopWithdrawPolling()
  } finally {
    recordLoading.value = false
  }
}

const stopWithdrawPolling = () => {
  if (!pollTimer.value) return
  clearInterval(pollTimer.value)
  pollTimer.value = null
}

const syncWithdrawPolling = () => {
  if (!hasPendingRecords.value) {
    stopWithdrawPolling()
    return
  }
  if (pollTimer.value) return
  pollTimer.value = setInterval(async () => {
    try {
      const result = await getWinWithdrawRecords()
      amountList.value = result?.records || []
      await person.refreshProfile?.()
      if (!amountList.value.some((item) => item.status === 'pending')) {
        stopWithdrawPolling()
      }
    } catch {
      stopWithdrawPolling()
    }
  }, 10000)
}

const handleAllAmount = () => {
  if (!isPositiveDecimal(winBalance.value)) {
    showToast({ message: $t('withdraw.insufficientWin'), position: 'middle' })
    return
  }
  amountWin.value = winBalance.value
}

const checkWinAmount = (e: Event) => {
  const target = e.target as HTMLInputElement
  let raw = String(target?.value ?? amountWin.value ?? '')
  raw = raw.replace(/[^\d.]/g, '')
  const parts = raw.split('.')
  if (parts.length > 2) raw = parts[0] + '.' + parts.slice(1).join('')
  if (parts[1] != null && parts[1].length > 18) raw = parts[0] + '.' + parts[1].slice(0, 18)
  amountWin.value = raw
}

const handleWithdrawal = async () => {
  if (loading.value || !canSubmit.value) return
  loading.value = true
  try {
    const result = await withdrawWin(amountWin.value)
    person.profile = {
      ...person.profile,
      win_balance: result.win_balance,
    }
    showToast({
      message: $t('withdraw.submittedProcessing'),
      position: 'middle',
      duration: 2000,
    })
    amountWin.value = ''
    await Promise.allSettled([person.refreshProfile?.(), getAmountList()])
    syncWithdrawPolling()
  } catch (error: any) {
    const code = error?.response?.data?.reason || error?.response?.data?.code
    const messageKey: Record<string, string> = {
      INVALID_AMOUNT: 'withdraw.enterAmount',
      INSUFFICIENT_WIN: 'withdraw.insufficientWin',
      INVALID_ADDRESS: 'withdraw.invalidAddress',
      AIX_WITHDRAW_FORBIDDEN: 'withdraw.aixExchangeHint',
    }
    showToast({
      message: messageKey[code]
        ? $t(messageKey[code])
        : (error?.response?.data?.message || $t('withdraw.failed')),
      position: 'middle',
      duration: 2000,
    })
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  await Promise.allSettled([
    person.getUser?.(),
    person.refreshProfile?.(),
    getAmountList(),
  ])
})

onUnmounted(() => {
  stopWithdrawPolling()
})
</script>

<style lang="scss" scoped>
@use '@/style/variables.scss' as *;

.withdrawal-page {
  min-height: 100vh;
  background: linear-gradient(180deg, #030A11 0%, #0D1B2A 100%);
}

.content {
  padding: 90px 20px 40px;
  max-width: 1200px;
  margin: 0 auto;
}

.page-header {
  text-align: center;
  margin-bottom: 20px;

  .page-title {
    font-size: 16px;
    font-weight: bold;
    color: #fff;
    margin-bottom: 8px;
  }

  .page-subtitle {
    font-size: 12px;
    color: rgba(255, 255, 255, 0.6);
  }

  .page-balance {
    font-size: 14px;
    color: $brand-primary;
    margin-top: 8px;
  }

  .page-hint {
    margin: 10px 0 0;
    font-size: 12px;
    color: rgba(255, 255, 255, 0.55);
    line-height: 1.6;
  }

  .link-btn {
    margin-left: 4px;
    padding: 0;
    border: 0;
    background: transparent;
    color: $brand-primary;
    cursor: pointer;
  }
}

.withdraw-form {
  margin-bottom: 40px;

  .form-hint-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 10px;
  }

  .form-hint {
    margin: 0;
    font-size: 13px;
    color: rgba(255, 255, 255, 0.7);
  }

  .all-btn {
    padding: 0;
    border: none;
    background: transparent;
    color: $brand-primary;
    font-size: 13px;
    cursor: pointer;
  }

  .form-row {
    display: flex;
    gap: 12px;
    align-items: center;
  }

  .form-input {
    flex: 1;
    height: 44px;
    padding: 0 14px;
    border: 1px solid rgba(255, 255, 255, 0.2);
    border-radius: 8px;
    background: rgba(0, 0, 0, 0.25);
    color: #fff;
    font-size: 15px;
    outline: none;
    caret-color: $brand-primary;
    -webkit-text-fill-color: #fff;

    &::placeholder {
      color: rgba(255, 255, 255, 0.4);
      -webkit-text-fill-color: rgba(255, 255, 255, 0.4);
    }

    &:focus {
      border-color: $brand-primary;
    }
  }

  .asset-tag {
    flex-shrink: 0;
    color: #8ed5ff;
    font-weight: 600;
  }

  .error-text {
    margin: 8px 0 0;
    color: #f17b7b;
    font-size: 12px;
  }

  .form-info {
    margin-top: 12px;
    display: flex;
    flex-direction: column;
    gap: 4px;

    p {
      margin: 0;
      font-size: 12px;
      color: rgba(255, 255, 255, 0.5);
    }
  }
}

.subscribe-btn {
  width: 100%;
  margin-top: 18px;
  padding: 8px 20px;
  background: $gradient-primary;
  color: $text-inverse;
  border: none;
  border-radius: 12px;
  font-size: 14px;
  font-weight: 400;
  cursor: pointer;
  transition: all 0.3s ease;

  &:hover:not(:disabled) {
    background: linear-gradient(135deg, $brand-primary-light 0%, $brand-primary 100%);
    transform: translateY(-2px);
    box-shadow: 0 4px 12px rgba(21, 151, 229, 0.3);
  }

  &:disabled {
    opacity: 0.5;
    cursor: not-allowed;
    background: rgba(255, 255, 255, 0.1);
    color: rgba(255, 255, 255, 0.5);
  }
}

.record-section {
  margin-top: 20px;

  .section-title-wrap {
    position: relative;
    margin-bottom: 10px;
    margin-left: 10px;
    display: flex;
    align-items: center;

    .title-bar {
      position: absolute;
      left: -10px;
      top: 50%;
      width: 4px;
      height: 16px;
      border-radius: 2px;
      background: linear-gradient(180deg, #1597E5 0%, #075FB8 100%);
      transform: translateY(-50%);
    }

    .section-title {
      margin: 0 0 0 8px;
      font-size: 16px;
      font-weight: bold;
      color: #fff;
    }
  }
}

.table-card {
  margin-top: 10px;
  min-height: 300px;
  overflow: hidden;
  border: 1px solid $border-color;
  border-radius: 11px;
  background: rgba(8, 19, 30, 0.6);
  backdrop-filter: blur(10px);
  padding: 11px 0;

  .table-header {
    display: flex;
    align-items: center;
    background: #030A11;
    padding: 8px 0;
    margin: -11px 0 0;

    span {
      flex: 1;
      text-align: center;
      font-size: 10px;
      color: $text-muted;
    }
  }

  .order-list {
    .table-row {
      display: flex;
      align-items: center;
      padding: 12px 0;
      border-bottom: 1px solid $border-light;

      &:last-child {
        border-bottom: none;
      }

      span {
        flex: 1;
        text-align: center;
        font-size: 13px;
        color: $text-primary;
      }
    }
  }

  .address-cell {
    font-size: 11px !important;
    color: $text-muted !important;
  }

  .status-cell {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 2px;

    .tx-hint {
      font-size: 10px;
      color: $text-muted;
    }
  }

  .muted {
    font-size: 11px;
    color: $text-muted;
  }

  .empty-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    height: 250px;

    p {
      margin-top: 8px;
      font-size: 12px;
      color: $text-muted;
    }
  }

  .state-box {
    height: 120px;
    display: flex;
    align-items: center;
    justify-content: center;
  }
}
</style>
