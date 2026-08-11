<template>
  <div class="transfer-page">
    <van-nav-bar
      :title="$t('transfer.title')"
      left-arrow
      :border="false"
      fixed
      @click-left="router.back()"
    />

    <main class="page-main">
      <div class="transfer-mode" role="tablist" :aria-label="$t('transfer.type')">
        <button
          type="button"
          role="tab"
          :aria-selected="mode === 'self'"
          :class="{ active: mode === 'self' }"
          @click="changeMode('self')"
        >
          {{ $t('transfer.toRewardWallet') }}
        </button>
        <button
          type="button"
          role="tab"
          :aria-selected="mode === 'user'"
          :class="{ active: mode === 'user' }"
          @click="changeMode('user')"
        >
          {{ $t('transfer.toUser') }}
        </button>
      </div>

      <section v-if="mode === 'self'" class="wallet-flow" :aria-label="$t('transfer.direction')">
        <div class="wallet-summary">
          <strong>{{ sourceWalletName }}</strong>
          <span class="wallet-balance">{{ sourceBalance }} USDT</span>
        </div>
        <van-icon class="flow-icon" name="arrow" aria-hidden="true" />
        <div class="wallet-summary wallet-summary-target">
          <strong>{{ targetWalletName }}</strong>
          <span class="wallet-balance">{{ mode === 'self' ? `${rewardBalance} USDT` : '' }}</span>
        </div>
      </section>
      <div v-else class="reward-balance-row">
        <span>{{ $t('transfer.rewardBalance') }}</span>
        <strong>{{ rewardBalance }} USDT</strong>
      </div>

      <section class="transfer-form">
        <template v-if="mode === 'user'">
          <label class="field-label" for="transfer-recipient">{{ $t('transfer.recipientAddress') }}</label>
          <div class="input-shell">
            <van-icon name="contact-o" aria-hidden="true" />
            <input
              id="transfer-recipient"
              v-model.trim="recipient"
              type="text"
              autocomplete="off"
              spellcheck="false"
              :placeholder="$t('transfer.recipientPlaceholder')"
            />
          </div>
        </template>

        <div class="amount-heading">
          <label class="field-label" for="transfer-amount">{{ $t('transfer.amount') }}</label>
          <button type="button" class="all-btn" @click="fillAll">{{ $t('transfer.all') }}</button>
        </div>
        <div class="input-shell amount-shell">
          <input
            id="transfer-amount"
            v-model="amount"
            type="text"
            inputmode="decimal"
            placeholder="0.00"
            @input="normalizeAmount"
          />
          <span class="currency">USDT</span>
        </div>

        <button
          type="button"
          class="submit-btn"
          :disabled="!canSubmit || loading"
          @click="submitTransfer"
        >
          {{ loading ? $t('transfer.processing') : $t('transfer.confirm') }}
        </button>
      </section>

      <p class="security-note">
        <van-icon name="shield-o" aria-hidden="true" />
        {{ transferHint }}
      </p>

      <section class="record-section">
        <div class="section-title-wrap">
          <div class="title-bar"></div>
          <h3 class="section-title">{{ $t('transfer.records') }}</h3>
          <div v-if="mode === 'user'" class="direction-filter" :aria-label="$t('transfer.recordDirection')">
            <button
              v-for="item in directionOptions"
              :key="item.value"
              type="button"
              :class="{ active: direction === item.value }"
              @click="changeDirection(item.value)"
            >
              {{ item.label }}
            </button>
          </div>
        </div>

        <div class="table-card" :aria-busy="recordLoading">
          <div class="table-header">
            <span>{{ mode === 'self' ? $t('transfer.walletDirection') : $t('transfer.directionAndUser') }}</span>
            <span>{{ $t('transfer.amountColumn') }}</span>
            <span>{{ $t('transfer.time') }}</span>
          </div>

          <div v-if="recordLoading" class="record-loading">
            <van-loading type="spinner" color="#1597E5" />
          </div>
          <template v-else-if="records.length > 0">
            <div class="order-list" v-for="item in records" :key="item.id">
              <div class="table-row">
                <span v-if="mode === 'self'" class="wallet-direction">{{ $t('transfer.rechargeToReward') }}</span>
                <span v-else class="counterparty-cell">
                  <strong>{{ item.direction === 'in' ? $t('transfer.in') : $t('transfer.out') }} · {{ relationshipText(item.relationship) }}</strong>
                  <small>{{ formatAddress(item.counterparty_address) }}</small>
                </span>
                <span class="amount-cell" :class="item.direction === 'in' ? 'income' : mode === 'user' ? 'outcome' : ''">
                  {{ signedAmount(item) }} USDT
                </span>
                <span class="time-cell">{{ formatUnixTime(item.created_at) }}</span>
              </div>
            </div>
          </template>
          <div v-else class="empty-state">
            <p>{{ recordsUnavailable ? $t('transfer.recordsUnavailable') : $t('transfer.noRecords') }}</p>
          </div>

          <div class="pagination-wrapper" v-if="!recordLoading && recordPageCount > 1">
            <Pagination
              v-model="recordPage"
              :page-count="recordPageCount"
              mode="simple"
              @change="loadTransferRecords"
            />
          </div>
        </div>
      </section>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { Pagination, showFailToast, showSuccessToast, showToast } from 'vant'
import userPerson from '@/pinia/person'
import request from '@/tools/request'

type TransferMode = 'self' | 'user'
type TransferDirection = 'all' | 'in' | 'out'

interface TransferRecord {
  id: number
  asset: 'USDT'
  amount: string
  from_wallet: 'recharge' | 'reward'
  to_wallet: 'reward'
  created_at: number
  direction?: 'in' | 'out'
  relationship?: 'upline' | 'downline'
  counterparty_address?: string
}

const router = useRouter()
const { t: $t, locale } = useI18n()
const person = userPerson()

const mode = ref<TransferMode>('self')
const recipient = ref('')
const amount = ref('')
const loading = ref(false)
const records = ref<TransferRecord[]>([])
const recordPage = ref(1)
const recordPageCount = ref(1)
const recordLoading = ref(false)
const recordsUnavailable = ref(false)
const direction = ref<TransferDirection>('all')
let recordRequestId = 0

const directionOptions = computed<Array<{ label: string; value: TransferDirection }>>(() => [
  { label: $t('transfer.all'), value: 'all' },
  { label: $t('transfer.in'), value: 'in' },
  { label: $t('transfer.out'), value: 'out' },
])

const rechargeBalance = computed(() => String(person.userinfo?.usdt || person.profile?.usdt_recharge || '0'))
const rewardBalance = computed(() => String((person.userinfo as any)?.reward || person.profile?.usdt_reward || '0'))
const sourceBalance = computed(() => mode.value === 'self' ? rechargeBalance.value : rewardBalance.value)
const sourceWalletName = computed(() => mode.value === 'self' ? $t('transfer.rechargeWallet') : $t('transfer.rewardWallet'))
const targetWalletName = computed(() => mode.value === 'self' ? $t('transfer.myRewardWallet') : $t('transfer.userRewardWallet'))
const transferHint = computed(() => mode.value === 'self'
  ? $t('transfer.selfHint')
  : $t('transfer.userHint')
)

const isPositiveAmount = (value: string) => /^\d+(?:\.\d+)?$/.test(value) && /[1-9]/.test(value)

const compareDecimalStrings = (left: string, right: string) => {
  const normalize = (value: string) => {
    const [integer = '0', fraction = ''] = String(value || '0').split('.')
    return {
      integer: integer.replace(/^0+(?=\d)/, '') || '0',
      fraction: fraction.replace(/0+$/, ''),
    }
  }
  const a = normalize(left)
  const b = normalize(right)
  if (a.integer.length !== b.integer.length) return a.integer.length > b.integer.length ? 1 : -1
  if (a.integer !== b.integer) return a.integer > b.integer ? 1 : -1
  const fractionLength = Math.max(a.fraction.length, b.fraction.length)
  const aFraction = a.fraction.padEnd(fractionLength, '0')
  const bFraction = b.fraction.padEnd(fractionLength, '0')
  return aFraction === bFraction ? 0 : aFraction > bFraction ? 1 : -1
}

const canSubmit = computed(() => {
  const recipientReady = mode.value === 'self' || recipient.value.length > 0
  return recipientReady && isPositiveAmount(amount.value)
})

const changeMode = (nextMode: TransferMode) => {
  if (mode.value === nextMode) return
  mode.value = nextMode
  recipient.value = ''
  amount.value = ''
  direction.value = 'all'
  loadTransferRecords(1)
}

const changeDirection = (nextDirection: TransferDirection) => {
  if (direction.value === nextDirection) return
  direction.value = nextDirection
  loadTransferRecords(1)
}

const formatAddress = (value?: string) => {
  if (!value) return '-'
  return value.length > 14 ? `${value.slice(0, 6)}...${value.slice(-4)}` : value
}

const formatUnixTime = (timestamp: number) => {
  if (!timestamp) return '-'
  const dateLocales: Record<string, string> = {
    zh: 'zh-CN',
    'zh-tw': 'zh-TW',
    en: 'en-US',
    ja: 'ja-JP',
    ko: 'ko-KR',
    vi: 'vi-VN',
  }
  return new Date(timestamp * 1000).toLocaleString(dateLocales[locale.value] || locale.value, { hour12: false })
}

const relationshipText = (relationship?: TransferRecord['relationship']) => {
  if (relationship === 'upline') return $t('transfer.upline')
  if (relationship === 'downline') return $t('transfer.downline')
  return '-'
}

const signedAmount = (item: TransferRecord) => {
  if (mode.value === 'self') return item.amount
  return `${item.direction === 'in' ? '+' : '-'}${item.amount}`
}

const loadTransferRecords = async (page = 1) => {
  const requestId = ++recordRequestId
  const requestedMode = mode.value
  const requestedDirection = direction.value
  recordLoading.value = true
  recordsUnavailable.value = false

  try {
    const url = requestedMode === 'self'
      ? '/v1/wallet/transfer-records/self'
      : '/v1/wallet/transfer-records/lineal'
    const params: Record<string, string | number> = {
      page,
      page_size: 10,
    }
    if (requestedMode === 'user') params.direction = requestedDirection

    const result: any = await request.get(url, { params, silent: true })
    if (
      requestId !== recordRequestId ||
      requestedMode !== mode.value ||
      requestedDirection !== direction.value
    ) return

    records.value = Array.isArray(result?.list) ? result.list : []
    recordPage.value = Number(result?.page || page)
    recordPageCount.value = Math.max(1, Math.ceil(Number(result?.total || 0) / Number(result?.page_size || 10)))
  } catch (error: any) {
    if (requestId !== recordRequestId) return
    records.value = []
    recordPage.value = 1
    recordPageCount.value = 1
    recordsUnavailable.value = error?.response?.status === 404
    if (!recordsUnavailable.value && error?.response?.status !== 401) {
      showFailToast(error?.response?.data?.message || error?.message || $t('transfer.fetchRecordsFailed'))
    }
  } finally {
    if (requestId === recordRequestId) recordLoading.value = false
  }
}

const normalizeAmount = () => {
  let value = amount.value.replace(/[^\d.]/g, '')
  const parts = value.split('.')
  if (parts.length > 2) value = `${parts[0]}.${parts.slice(1).join('')}`
  if (parts[1]?.length > 8) value = `${parts[0]}.${parts[1].slice(0, 8)}`
  amount.value = value
}

const fillAll = () => {
  if (!isPositiveAmount(sourceBalance.value)) {
    showToast({ message: $t('transfer.insufficientBalance', { wallet: sourceWalletName.value }), position: 'middle' })
    return
  }
  amount.value = sourceBalance.value
}

const submitTransfer = async () => {
  if (loading.value) return
  if (!isPositiveAmount(amount.value)) {
    showFailToast($t('transfer.amountMustBePositive'))
    return
  }
  if (compareDecimalStrings(amount.value, sourceBalance.value) > 0) {
    showFailToast($t('transfer.insufficientBalance', { wallet: sourceWalletName.value }))
    return
  }
  if (mode.value === 'user') {
    if (!/^0x[a-fA-F0-9]{40}$/.test(recipient.value)) {
      showFailToast($t('transfer.invalidRecipient'))
      return
    }
    if (recipient.value.toLowerCase() === String(person.address || '').toLowerCase()) {
      showFailToast($t('transfer.cannotTransferToSelf'))
      return
    }
  }

  loading.value = true
  try {
    if (mode.value === 'self') {
      await request.post('/v1/wallet/recharge-to-reward', {
        amount: amount.value,
      })
    } else {
      await request.post('/v1/wallet/transfer', {
        to_address: recipient.value,
        asset: 'USDT',
        amount: amount.value,
        pay_from: 'reward',
      })
    }
    amount.value = ''
    recipient.value = ''
    await Promise.all([
      person.getUser?.(),
      person.refreshProfile?.(),
    ])
    await loadTransferRecords(1)
    showSuccessToast($t('transfer.success'))
  } catch (error: any) {
    // request 已优先展示后端 message，此处仅兜底非 Axios 错误。
    if (!error?.response) showFailToast(error?.message || $t('transfer.failed'))
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  await Promise.all([
    person.getUser?.(),
    person.refreshProfile?.(),
  ])
  await loadTransferRecords(1)
})
</script>

<style lang="scss" scoped>
@use '@/style/variables.scss' as *;

.transfer-page {
  min-height: 100vh;
  background: linear-gradient(180deg, #030a11 0%, #071421 48%, #020508 100%);
}

.page-main {
  width: 100%;
  max-width: 520px;
  margin: 0 auto;
  padding: 76px 16px 40px;
  box-sizing: border-box;
}

.transfer-mode {
  width: 100%;
  height: 40px;
  margin-bottom: 18px;
  padding: 3px;
  border: 1px solid $border-color;
  border-radius: $radius-md;
  background: rgba(3, 10, 17, 0.72);
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  box-sizing: border-box;

  button {
    min-width: 0;
    border: 0;
    border-radius: $radius-sm;
    background: transparent;
    color: $text-muted;
    font-size: 13px;
    cursor: pointer;
    transition: background $transition-fast, color $transition-fast;

    &.active {
      background: $gradient-primary;
      color: $text-inverse;
      font-weight: 600;
    }
  }
}

.wallet-flow {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 28px minmax(0, 1fr);
  align-items: center;
  gap: 8px;
  margin-bottom: 24px;
}

.reward-balance-row {
  min-height: 36px;
  margin-bottom: 16px;
  padding: 0 14px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  box-sizing: border-box;
  color: $text-muted;
  font-size: 13px;

  strong {
    min-width: 0;
    color: $brand-primary-light;
    font-size: 14px;
    font-weight: 500;
    overflow-wrap: anywhere;
    text-align: right;
  }
}

.wallet-summary {
  min-width: 0;
  padding: 10px 14px;
  border: 1px solid $border-color;
  border-radius: $radius-md;
  background: rgba(8, 19, 30, 0.84);
  display: flex;
  flex-direction: column;
  justify-content: center;
  box-sizing: border-box;

  strong,
  span {
    overflow-wrap: anywhere;
  }

  strong {
    color: $text-primary;
    font-size: 15px;
    font-weight: 500;
  }
}

.wallet-summary-target {
  border-color: rgba(143, 223, 255, 0.3);
}

.wallet-label {
  color: $text-muted;
  font-size: 12px;
}

.wallet-balance {
  margin-top: 6px;
  color: $brand-primary-light;
  font-size: 12px;
}

.flow-icon {
  color: $brand-primary;
  font-size: 22px;
  text-align: center;
}

.transfer-form {
  padding: 20px 16px;
  border: 1px solid $border-color;
  border-radius: $radius-md;
  background: rgba(8, 19, 30, 0.9);
  box-shadow: $shadow-md;
}

.field-label {
  display: block;
  margin-bottom: 9px;
  color: $text-primary;
  font-size: 13px;
}

.amount-heading {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 20px;

  .field-label {
    margin-bottom: 9px;
  }
}

.transfer-form > .amount-heading:first-child {
  margin-top: 0;
}

.all-btn {
  padding: 0 0 9px;
  border: 0;
  background: transparent;
  color: $brand-primary-light;
  font-size: 13px;
  cursor: pointer;
}

.input-shell {
  width: 100%;
  height: 46px;
  padding: 0 13px;
  border: 1px solid $border-light;
  border-radius: $radius-md;
  background: rgba(3, 10, 17, 0.82);
  display: flex;
  align-items: center;
  gap: 9px;
  box-sizing: border-box;
  transition: border-color $transition-fast;

  &:focus-within {
    border-color: $brand-primary;
  }

  .van-icon {
    flex: 0 0 auto;
    color: $text-muted;
    font-size: 18px;
  }

  input {
    min-width: 0;
    flex: 1;
    border: 0;
    outline: 0;
    background: transparent;
    color: $text-primary;
    font-size: 14px;

    &::placeholder {
      color: rgba(159, 179, 200, 0.55);
    }
  }
}

.amount-shell input {
  font-size: 18px;
}

.currency {
  flex: 0 0 auto;
  color: $text-muted;
  font-size: 13px;
}

.submit-btn {
  width: 100%;
  height: 44px;
  margin-top: 24px;
  border: 0;
  border-radius: $radius-md;
  background: $gradient-primary;
  color: $text-inverse;
  font-size: 15px;
  font-weight: 600;
  cursor: pointer;
  transition: opacity $transition-fast, transform $transition-fast;

  &:active:not(:disabled) {
    transform: translateY(1px);
  }

  &:disabled {
    cursor: not-allowed;
    opacity: 0.4;
  }
}

.security-note {
  margin-top: 16px;
  color: $text-muted;
  display: flex;
  align-items: flex-start;
  justify-content: center;
  gap: 6px;
  font-size: 12px;
  line-height: 1.6;

  .van-icon {
    margin-top: 2px;
    color: $brand-primary;
    font-size: 14px;
  }
}

.record-section {
  margin-top: 28px;
}

.section-title-wrap {
  min-height: 30px;
  margin: 0 0 10px 10px;
  position: relative;
  display: flex;
  align-items: center;
  gap: 8px;

  .title-bar {
    position: absolute;
    left: -10px;
    top: 50%;
    width: 4px;
    height: 16px;
    border-radius: 2px;
    background: $gradient-primary-vertical;
    transform: translateY(-50%);
  }

  .section-title {
    margin: 0 0 0 8px;
    color: $text-primary;
    font-size: 16px;
    font-weight: 600;
  }
}

.direction-filter {
  height: 28px;
  margin-left: auto;
  padding: 2px;
  border: 1px solid $border-color;
  border-radius: $radius-sm;
  display: flex;
  box-sizing: border-box;

  button {
    min-width: 42px;
    padding: 0 8px;
    border: 0;
    border-radius: 3px;
    background: transparent;
    color: $text-muted;
    font-size: 11px;
    cursor: pointer;

    &.active {
      background: rgba(21, 151, 229, 0.2);
      color: $brand-primary-light;
    }
  }
}

.table-card {
  min-height: 300px;
  padding: 11px 0;
  border: 1px solid $border-color;
  border-radius: 8px;
  overflow: hidden;
  background: rgba(8, 19, 30, 0.6);
  backdrop-filter: blur(10px);

  .table-header,
  .table-row {
    display: grid;
    grid-template-columns: minmax(0, 1.35fr) minmax(82px, 0.85fr) minmax(0, 1fr);
    align-items: center;

    > span {
      min-width: 0;
      text-align: center;
    }
  }

  .table-header {
    margin: -11px 0 0;
    padding: 9px 6px;
    background: #030a11;

    span {
      color: $text-muted;
      font-size: 10px;
    }
  }

  .order-list {
    .table-row {
      min-height: 58px;
      padding: 10px 6px;
      border-bottom: 1px solid $border-light;
      box-sizing: border-box;

      > span {
        color: $text-primary;
        font-size: 12px;
      }
    }

    &:last-of-type .table-row {
      border-bottom: none;
    }
  }

  .counterparty-cell {
    padding: 0 4px;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 3px;

    strong {
      color: $text-primary;
      font-size: 12px;
      font-weight: 500;
    }

    small {
      max-width: 100%;
      color: $text-muted;
      font-size: 10px;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }
  }

  .wallet-direction {
    color: $brand-primary-light !important;
  }

  .amount-cell {
    padding: 0 3px;
    overflow-wrap: anywhere;

    &.income {
      color: #54d6a0;
    }

    &.outcome {
      color: #ff8d8d;
    }
  }

  .time-cell {
    padding: 0 4px;
    color: $text-muted !important;
    font-size: 10px !important;
    line-height: 1.4;
    overflow-wrap: anywhere;
  }

  .record-loading,
  .empty-state {
    height: 250px;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .empty-state p {
    margin: 0;
    color: $text-muted;
    font-size: 12px;
  }

  .pagination-wrapper {
    padding: 16px 0 5px;
    display: flex;
    justify-content: center;
  }
}

@media (max-width: 350px) {
  .page-main {
    padding-right: 12px;
    padding-left: 12px;
  }

  .wallet-summary {
    padding: 12px 10px;
  }

  .direction-filter button {
    min-width: 36px;
    padding: 0 5px;
  }

  .table-card {
    .table-header,
    .table-row {
      grid-template-columns: minmax(0, 1.2fr) minmax(72px, 0.8fr) minmax(0, 1fr);
    }
  }
}
</style>
