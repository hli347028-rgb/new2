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
          <strong>{{ aixBalance }}</strong>
        </div>
        <van-icon name="arrow" />
        <div class="win-balance">
          <span>{{ $t('exchange.currentWin') }}</span>
          <strong>{{ winBalance }}</strong>
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
          <span>{{ $t('exchange.currentPrice') }}</span>
          <strong v-if="hasPrice">1 WIN = {{ winPrice }} USDT</strong>
          <strong v-else>{{ $t('exchange.priceUnavailable') }}</strong>
        </div>
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
                <span>{{ $t('exchange.priceShort') }} {{ displayAmount(item.exchange_price) }}</span>
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

const router = useRouter()
const { t: $t } = useI18n()
const person = userPerson()
const amount = ref('')
const submitting = ref(false)
const recordLoading = ref(false)
const records = ref<AixWinExchangeRecord[]>([])

const aixBalance = computed(() => String(person.profile?.aix_balance || person.userinfo?.amountGet || '0'))
const winBalance = computed(() => String(person.profile?.win_balance || '0'))
const winPrice = computed(() => String(person.profile?.win_price || '0'))
const hasPrice = computed(() => isPositiveDecimal(winPrice.value))
const amountError = computed(() => {
  if (!amount.value) return ''
  if (!isPositiveDecimal(amount.value)) return $t('exchange.positiveAmount')
  if (compareDecimals(amount.value, aixBalance.value) > 0) return $t('exchange.insufficientAix')
  return ''
})
const canSubmit = computed(() => Boolean(amount.value) && !amountError.value && hasPrice.value)

function normalizeDecimal(value: string) {
  const clean = String(value || '0').trim().replace(/^\+/, '')
  if (!/^\d*(\.\d*)?$/.test(clean)) return null
  const [integer = '0', fraction = ''] = clean.split('.')
  return { integer: (integer.replace(/^0+(?=\d)/, '') || '0'), fraction: fraction.replace(/0+$/, '') }
}

function isPositiveDecimal(value: string) {
  const normalized = normalizeDecimal(value)
  return Boolean(normalized && (normalized.integer !== '0' || /[1-9]/.test(normalized.fraction)))
}

function compareDecimals(left: string, right: string) {
  const a = normalizeDecimal(left)
  const b = normalizeDecimal(right)
  if (!a || !b) return 0
  if (a.integer.length !== b.integer.length) return a.integer.length > b.integer.length ? 1 : -1
  if (a.integer !== b.integer) return a.integer > b.integer ? 1 : -1
  const width = Math.max(a.fraction.length, b.fraction.length)
  const af = a.fraction.padEnd(width, '0')
  const bf = b.fraction.padEnd(width, '0')
  return af === bf ? 0 : af > bf ? 1 : -1
}

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
  const text = String(value ?? '0')
  if (!text.includes('.')) return text
  const trimmed = text.replace(/0+$/, '').replace(/\.$/, '')
  return trimmed || '0'
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
    await loadRecords()
  } catch (error: any) {
    const code = error?.response?.data?.reason || error?.response?.data?.code
    const messageKey: Record<string, string> = {
      INVALID_AMOUNT: 'exchange.positiveAmount',
      INSUFFICIENT_AIX: 'exchange.insufficientAix',
      WIN_PRICE_NOT_CONFIGURED: 'exchange.priceUnavailable',
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
.rate-row span { color: #91a7b7; }
.rate-row strong { text-align: right; color: #dcebf4; }
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
</style>
