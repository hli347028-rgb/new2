<template>
  <div class="exchange-page">
    <van-nav-bar
      :title="$t('exchange.title')"
      left-arrow
      :border="false"
      fixed
      @click-left="router.back()"
    />

    <main class="page-main">
      <section class="balance-card">
        <div>
          <span>{{ $t('exchange.availableAix') }}</span>
          <strong>{{ displayAmount(aixBalance) }}</strong>
        </div>
        <van-icon name="arrow" />
        <div class="win-balance">
          <span>{{ $t('exchange.currentWin') }}</span>
          <strong>{{ displayAmount(winBalance) }}</strong>
        </div>
      </section>

      <section class="exchange-form">
        <div class="amount-heading">
          <label for="exchange-amount">{{ $t('exchange.exchangeAmount') }}</label>
          <button type="button" @click="fillAll">{{ $t('exchange.all') }}</button>
        </div>
        <div class="input-shell" :class="{ invalid: amountError }">
          <input
            id="exchange-amount"
            v-model="amount"
            inputmode="decimal"
            type="text"
            autocomplete="off"
            placeholder="0.00"
            @input="normalizeAmount"
          />
          <span>AIX</span>
        </div>
        <p v-if="amountError" class="error-text">{{ amountError }}</p>

        <div class="rate-row">
          <span>{{ $t('exchange.currentRate') }}</span>
          <strong v-if="unitRate">1 AIX = {{ displayAmount(unitRate.winNet) }} WIN</strong>
          <strong v-else>{{ $t('exchange.priceUnavailable') }}</strong>
        </div>
        <div v-if="hasRate" class="rate-row rate-sub">
          <span>{{ $t('exchange.aixPrice') }}</span>
          <strong>{{ displayAmount(aixPrice) }} USDT</strong>
        </div>
        <div v-if="hasRate" class="rate-row rate-sub">
          <span>{{ $t('exchange.winPrice') }}</span>
          <strong>{{ displayAmount(winPrice) }} USDT</strong>
        </div>
        <!-- <div v-if="unitRate" class="rate-row rate-sub">
          <span>{{ $t('exchange.rateDetail') }}</span>
          <strong>{{ $t('exchange.rateBreakdown', {
            gross: displayAmount(unitRate.winGross),
            fee: feeRateText,
            net: displayAmount(unitRate.winNet),
          }) }}</strong>
        </div> -->
        <div class="rate-row">
          <span>{{ $t('exchange.feeRate') }}</span>
          <strong>{{ feeRateText }}</strong>
        </div>
        <p v-if="preview" class="estimate-box">
          {{ $t('exchange.estimate', {
            gross: displayAmount(preview.winGross),
            net: displayAmount(preview.winNet),
            fee: displayAmount(preview.fee),
          }) }}
        </p>
        <p class="rate-note"><van-icon name="info-o" /> {{ $t('exchange.rateHint') }}</p>

        <button
          type="button"
          class="submit-btn"
          :disabled="!canSubmit || submitting"
          @click="submitExchange"
        >
          {{ submitting ? $t('exchange.processing') : $t('exchange.confirm') }}
        </button>
      </section>

      <section class="record-section">
        <h2>{{ $t('exchange.records') }}</h2>
        <div class="record-card" :aria-busy="recordLoading">
          <div v-if="recordLoading" class="state-box"><van-loading color="#1597e5" /></div>
          <van-empty v-else-if="records.length === 0" :description="$t('exchange.noRecords')" :image="emptyImage" />
          <div v-else class="record-list">
            <article v-for="item in records" :key="item.id">
              <div class="record-assets">
                <strong>{{ displayAmount(item.from_amount) }} AIX</strong>
                <van-icon name="arrow" />
                <strong class="win-value">{{ displayAmount(item.to_amount) }} WIN</strong>
              </div>
              <div class="record-meta">
                <span>{{ formatTime(item.created_at) }}</span>
                <span>{{ $t('exchange.rateShort') }} 1 AIX = {{ recordUnitRate(item) }} WIN</span>
              </div>
              <div v-if="item.fee_amount" class="record-fee">
                {{ $t('exchange.feeDeducted', { fee: displayAmount(item.fee_amount) }) }}
              </div>
            </article>
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
import { showFailToast, showSuccessToast } from 'vant'
import userPerson from '@/pinia/person'
import { exchangeAixToWin, getAixWinExchangeRecords } from '@/api/aix'
import type { AixWinExchangeRecord } from '@/api/aix'
import emptyImage from '@/assets/images/custom-empty-image.png'
import {
  calcExchangePreview,
  calcUnitExchangeRate,
  compareDecimals,
  displayDecimal,
  divDecimal,
  formatFeeRate,
  isPositiveDecimal,
  resolveAixToWinRate,
} from '@/tools/decimal'

const router = useRouter()
const { t: $t } = useI18n()
const person = userPerson()
const amount = ref('')
const submitting = ref(false)
const recordLoading = ref(false)
const records = ref<AixWinExchangeRecord[]>([])

const aixBalance = computed(() => String(person.profile?.aix_balance || person.userinfo?.amountGet || '0'))
const winBalance = computed(() => String(person.profile?.win_balance || '0'))
const aixPrice = computed(() => Number(person.profile?.aix_price || 0))
const winPrice = computed(() => Number(person.profile?.win_price || 0))
const exchangeFeeRate = computed(() => Number(person.profile?.exchange_fee_rate ?? 0.05))
const aixToWinRate = computed(() => resolveAixToWinRate({
  aix_to_win_rate: person.profile?.aix_to_win_rate,
  aix_price: person.profile?.aix_price,
  win_price: person.profile?.win_price,
}))
const feeRateText = computed(() => formatFeeRate(exchangeFeeRate.value))
const hasRate = computed(() => Boolean(aixToWinRate.value))
const unitRate = computed(() => (
  aixToWinRate.value ? calcUnitExchangeRate(aixToWinRate.value, exchangeFeeRate.value) : null
))
const preview = computed(() => {
  if (!amount.value || !aixToWinRate.value) return null
  return calcExchangePreview(amount.value, aixToWinRate.value, exchangeFeeRate.value)
})
const amountError = computed(() => {
  if (!amount.value) return ''
  if (!isPositiveDecimal(amount.value)) return $t('exchange.positiveAmount')
  if (compareDecimals(amount.value, aixBalance.value) > 0) return $t('exchange.insufficientAix')
  if (preview.value && !isPositiveDecimal(preview.value.winNet)) return $t('exchange.netAmountTooSmall')
  return ''
})
const canSubmit = computed(() => Boolean(amount.value) && !amountError.value && hasRate.value)

function normalizeAmount(event: Event) {
  const input = event.target as HTMLInputElement
  let value = input.value.replace(/[^\d.]/g, '')
  const dot = value.indexOf('.')
  if (dot >= 0) value = value.slice(0, dot + 1) + value.slice(dot + 1).replace(/\./g, '').slice(0, 18)
  value = value.replace(/^0+(?=\d)/, '')
  amount.value = value
  input.value = value
}

function fillAll() {
  amount.value = aixBalance.value
}

function displayAmount(value: unknown) {
  return displayDecimal(value)
}

function recordUnitRate(item: AixWinExchangeRecord) {
  if (isPositiveDecimal(item.from_amount) && isPositiveDecimal(item.to_amount)) {
    return displayAmount(divDecimal(item.to_amount, item.from_amount))
  }
  return '-'
}

function formatTime(value: number) {
  const date = new Date(Number(value) * 1000)
  if (Number.isNaN(date.getTime())) return ''
  const pad = (part: number) => String(part).padStart(2, '0')
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}`
}

async function loadRecords() {
  recordLoading.value = true
  try {
    const result = await getAixWinExchangeRecords()
    records.value = result?.records || []
  } catch {
    records.value = []
  } finally {
    recordLoading.value = false
  }
}

async function submitExchange() {
  if (!canSubmit.value || submitting.value) return
  submitting.value = true
  try {
    const result = await exchangeAixToWin(amount.value)
    person.profile = {
      ...person.profile,
      aix_balance: result.aix_balance,
      win_balance: result.win_balance,
    }
    amount.value = ''
    showSuccessToast($t('exchange.success', { amount: displayAmount(result.to_amount) }))
    await Promise.allSettled([person.refreshProfile(), loadRecords()])
  } catch (error: any) {
    const code = error?.response?.data?.reason || error?.response?.data?.code
    const messageKey: Record<string, string> = {
      INVALID_AMOUNT: 'exchange.positiveAmount',
      INSUFFICIENT_AIX: 'exchange.insufficientAix',
      WIN_PRICE_NOT_CONFIGURED: 'exchange.priceUnavailable',
      WIN_NET_AMOUNT_TOO_SMALL: 'exchange.netAmountTooSmall',
    }
    showFailToast(messageKey[code] ? $t(messageKey[code]) : (error?.response?.data?.message || $t('exchange.failed')))
  } finally {
    submitting.value = false
  }
}

onMounted(async () => {
  await Promise.allSettled([person.refreshProfile(), loadRecords()])
})
</script>

<style scoped lang="scss">
@use '@/style/variables.scss' as *;

.exchange-page { min-height: 100vh; background: url('@/assets/images/a3.png') no-repeat top / 100% auto; }
.page-main { padding: 70px 15px 30px; display: flex; flex-direction: column; gap: 20px; }
.balance-card { min-height: 80px; padding: 14px 20px; box-sizing: border-box; display: grid; grid-template-columns: 1fr auto 1fr; align-items: center; gap: 12px; border: 1px solid #1d4059; border-radius: 16px; background: linear-gradient(135deg, #102b40, #091722); box-shadow: 0 10px 30px rgba(0,0,0,.2); }
.balance-card > .van-icon { color: #1597e5; font-size: 24px; }
.balance-card div { min-width: 0; display: flex; flex-direction: column; gap: 8px; }
.balance-card span { color: #91a7b7; font-size: 12px; }
.balance-card .win-balance { text-align: right; }
.exchange-form, .record-card { padding: 20px; border: 1px solid #183247; border-radius: 16px; background: #0b1824; }
.amount-heading { display: flex; justify-content: space-between; align-items: center; margin-bottom: 10px; color: #d9e5ed; font-size: 14px; }
.amount-heading button { border: 0; color: #34aef7; background: transparent; font-size: 13px; }
.input-shell { height: 58px; padding: 0 16px; display: flex; align-items: center; gap: 10px; border: 1px solid #24455d; border-radius: 12px; background: #07131d; }
.input-shell:focus-within { border-color: #1597e5; box-shadow: 0 0 0 2px rgba(21,151,229,.12); }
.input-shell.invalid { border-color: #e65f5f; }
.input-shell input { min-width: 0; flex: 1; border: 0; outline: 0; color: #fff; background: transparent; font-size: 24px; }
.input-shell span { color: #8ed5ff; font-weight: 600; }
.error-text { margin: 7px 2px 0; color: #f17b7b; font-size: 12px; }
.rate-row { margin-top: 18px; display: flex; justify-content: space-between; gap: 12px; font-size: 13px; }
.rate-row + .rate-row { margin-top: 10px; }
.rate-row span { color: #91a7b7; }
.rate-row strong { text-align: right; color: #dcebf4; }
.rate-row.rate-sub { margin-top: 8px; }
.rate-row.rate-sub strong { font-size: 11px; color: #70899a; font-weight: 400; }
.estimate-box { margin: 12px 0 0; padding: 10px 12px; border-radius: 10px; background: rgba(21,151,229,.08); color: #8ed5ff; font-size: 12px; line-height: 1.6; }
.rate-note { margin: 10px 0 0; color: #70899a; font-size: 12px; line-height: 1.5; }
.submit-btn { width: 100%; height: 46px; margin-top: 22px; border: 0; border-radius: 24px; color: #fff; background: linear-gradient(90deg, #087bc1, #1597e5); font-size: 15px; font-weight: 600; }
.submit-btn:disabled { opacity: .4; }
.record-section h2 { margin: 0 0 12px; font-size: 17px; font-weight: 600; }
.record-card { padding: 0 16px; }
.state-box { height: 120px; display: flex; align-items: center; justify-content: center; }
.record-list article { padding: 16px 0; border-bottom: 1px solid #183247; }
.record-list article:last-child { border-bottom: 0; }
.record-assets { display: flex; align-items: center; gap: 10px; }
.record-assets .van-icon { color: #617b8d; }
.record-assets .win-value { color: #39b7ff; }
.record-meta { margin-top: 8px; display: flex; justify-content: space-between; color: #70899a; font-size: 11px; }
.record-fee { margin-top: 4px; color: #617b8d; font-size: 11px; }
</style>
