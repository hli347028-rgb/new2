<template>
  <div class="node-page">
    <Header />

    <div class="content">
      <div class="page-header">
        <div class="mode-tabs" role="radiogroup">
          <label class="mode-option" :class="{ active: activeMode === 'recharge', disabled: submitting }">
            <input
              type="radio"
              name="subscribe-mode"
              value="recharge"
              :checked="activeMode === 'recharge'"
              :disabled="submitting"
              @change="switchMode('recharge')"
            />
            <span class="radio-dot" aria-hidden="true"></span>
            <strong>{{ $t('node.reportOrder') }}</strong>
          </label>
          <label class="mode-option" :class="{ active: activeMode === 'reward', disabled: submitting }">
            <input
              type="radio"
              name="subscribe-mode"
              value="reward"
              :checked="activeMode === 'reward'"
              :disabled="submitting"
              @change="switchMode('reward')"
            />
            <span class="radio-dot" aria-hidden="true"></span>
            <strong>{{ $t('node.reinvest') }}</strong>
          </label>
          <label class="mode-option" :class="{ active: activeMode === 'win', disabled: submitting }">
            <input
              type="radio"
              name="subscribe-mode"
              value="win"
              :checked="activeMode === 'win'"
              :disabled="submitting"
              @change="switchMode('win')"
            />
            <span class="radio-dot" aria-hidden="true"></span>
            <strong>{{ $t('node.winPay') }}</strong>
          </label>
        </div>

        <div class="balance-card">
          <span>{{ balanceLabel }}</span>
          <strong>{{ accountBalance }} <small>{{ balanceUnit }}</small></strong>
        </div>
        <p v-if="activeMode === 'win' && winPrice > 0" class="mode-tip win-cost-tip">
          {{ $t('node.winPayHint', { price: winPrice, cost: estimatedWinCost }) }}
        </p>
        <p class="mode-tip">{{ modeTip }}</p>
      </div>

      <div class="node-tiers">
        <div
          v-for="tier in nodeTiers"
          :key="tier.price"
          class="tier-card"
          :class="{ active: selectedTier === tier.price }"
        >
          <div class="tier-header">
            <div class="tier-price">{{ tier.price }}</div>
            <div class="tier-unit">USDT</div>
          </div>
          <button class="subscribe-btn" :disabled="submitting" @click="handleSubscribe(tier.price)">
            {{ actionText }}
          </button>
        </div>
      </div>

      <div class="custom-amount">
        <p class="custom-hint">{{ $t('node.customAmountHint', { amount: minSubscribe }) }}</p>
        <div class="custom-row">
          <input
            v-model="customAmount"
            class="custom-input"
            type="number"
            :min="minSubscribe"
            step="1"
            :placeholder="$t('node.minPlaceholder', { amount: minSubscribe })"
          />
          <button class="subscribe-btn custom-btn" :disabled="submitting" @click="handleCustomSubscribe">
            {{ actionText }}
          </button>
        </div>
      </div>

      <div class="record-section">
        <div class="section-title-wrap">
          <div class="title-bar"></div>
          <h3 class="section-title">{{ $t('node.orderList') }}</h3>
        </div>

        <div class="table-card">
          <div class="table-header">
            <span>{{ $t('node.amount') }}</span>
            <span>{{ $t('node.status') }}</span>
            <span>{{ $t('node.fundingSource') }}</span>
            <span>{{ $t('node.time') }}</span>
          </div>
          <div class="order-list" v-for="(item, index) in orderList" :key="index">
            <div class="table-row">
              <span>{{ item.total_amount ?? item.amount }}</span>
              <span>{{ orderStatusText(item.status) }}</span>
              <span>{{ fundingSourceText(item.product_name) }}</span>
              <span>{{ item.created_at ?? item.createdAt }}</span>
            </div>
          </div>
          <div class="empty-state" v-if="orderList.length === 0">
            <p>{{ $t('common.noData') }}</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import Header from '@/components/Header.vue'
import request from '@/tools/request'
import { showSuccessToast, showFailToast, showLoadingToast, closeToast } from 'vant'
import userPerson from '@/pinia/person'
import { errMsg, listAixOrders, subscribeAix } from '@/api/aix'

const { t: $t } = useI18n()
const person = userPerson()

type SubscribeMode = 'recharge' | 'reward' | 'win'

const activeMode = ref<SubscribeMode>('recharge')
const winPrice = computed(() => Number(person.profile?.win_price || 0))
const accountBalance = computed(() => {
  const profile = person.profile
  if (activeMode.value === 'win') return profile?.win_balance || '0.00'
  if (activeMode.value === 'recharge') return profile?.usdt_recharge || '0.00'
  return profile?.usdt_reward || '0.00'
})
const balanceLabel = computed(() => {
  if (activeMode.value === 'win') return $t('node.winWalletBalance')
  if (activeMode.value === 'recharge') return $t('node.rechargeWalletBalance')
  return $t('node.rewardWalletBalance')
})
const balanceUnit = computed(() => (activeMode.value === 'win' ? 'WIN' : 'USDT'))
const modeTip = computed(() => {
  if (activeMode.value === 'win') return $t('node.winReferralTip')
  return $t('node.referralRewardTip')
})
const estimatedWinCost = computed(() => {
  const amount = selectedTier.value || Number(customAmount.value) || minSubscribe.value
  if (!winPrice.value || winPrice.value <= 0) return '-'
  return (amount / winPrice.value).toFixed(4)
})
const actionText = computed(() => {
  if (activeMode.value === 'reward') return $t('node.reinvestNow')
  if (activeMode.value === 'win') return $t('node.winPayNow')
  return $t('node.reportNow')
})

interface NodeTier {
  price: number
}

const selectedTier = ref<number | null>(null)
const orderList = ref<any[]>([])
const nodeTiers = ref<NodeTier[]>([])
const minSubscribe = ref(100)
const customAmount = ref('')
const submitting = ref(false)

const switchMode = (mode: SubscribeMode) => {
  if (submitting.value) return
  activeMode.value = mode
  selectedTier.value = null
  customAmount.value = ''
}

const getSubscribeTiers = async () => {
  try {
    const res: any = await request.get('app_server/subscribe_tiers')
    minSubscribe.value = Math.max(100, Number(res.min_subscribe_amount || 100))
    nodeTiers.value = (res.tiers || []).map((v: string | number) => ({
      price: Number(v),
    })).filter((t: NodeTier) => t.price > 0)
  } catch {
    nodeTiers.value = [
      { price: 100 },
      { price: 500 },
      { price: 1000 },
      { price: 3000 },
    ]
    minSubscribe.value = 100
  }
}

const getOrderList = async () => {
  try {
    const res: any = await listAixOrders()
    orderList.value = res.orders || []
  } catch {
    orderList.value = []
  }
}

const fundingSourceText = (source: string) => {
  if (source === 'reward') return $t('node.rewardSource')
  if (source === 'win') return $t('node.winSource')
  return $t('node.rechargeSource')
}

const orderStatusText = (status: string | number) => {
  const s = String(status)
  if (s === '1' || s === 'active') return $t('node.statusActive')
  if (s === '2' || s === 'exited') return $t('node.statusExited')
  return s || '-'
}

const handleSubscribe = async (amount: number) => {
  if (submitting.value) return

  if (!amount || amount < minSubscribe.value) {
    showFailToast($t('node.minSubscribeAmount', { amount: minSubscribe.value }))
    return
  }
  const mode = activeMode.value
  const bal = Number(accountBalance.value)
  if (mode === 'win') {
    if (!winPrice.value || winPrice.value <= 0) {
      showFailToast($t('node.winPriceMissing'))
      return
    }
    const needWin = amount / winPrice.value
    if (Number.isFinite(bal) && needWin > bal) {
      showFailToast($t('common.insufficientBalance'))
      return
    }
  } else if (Number.isFinite(bal) && amount > bal) {
    showFailToast($t('common.insufficientBalance'))
    return
  }
  selectedTier.value = amount
  submitting.value = true
  showLoadingToast({ message: $t('common.loading'), duration: 0 })
  try {
    await subscribeAix(String(amount), mode)
    closeToast()
    const okMsg = mode === 'reward'
      ? $t('node.reinvestSuccess')
      : mode === 'win'
        ? $t('node.winPaySuccess')
        : $t('node.reportSuccess')
    showSuccessToast(okMsg)
    customAmount.value = ''
    await Promise.all([person.refreshProfile(), getOrderList()])
  } catch (error: any) {
    closeToast()
    const failMsg = mode === 'reward'
      ? $t('node.reinvestFailed')
      : mode === 'win'
        ? $t('node.winPayFailed')
        : $t('node.reportFailed')
    showFailToast(errMsg(error, failMsg))
  } finally {
    submitting.value = false
  }
}

const handleCustomSubscribe = () => {
  const amount = Number(customAmount.value)
  if (!amount || Number.isNaN(amount)) {
    showFailToast($t('node.enterSubscribeAmount'))
    return
  }
  handleSubscribe(amount)
}

onMounted(async () => {
  await Promise.all([
    getSubscribeTiers(),
    getOrderList(),
    person.refreshProfile()
  ])
})
</script>

<style lang="scss" scoped>
@use '@/style/variables.scss' as *;

.node-page {
  min-height: 100vh;
  background: linear-gradient(180deg, #030A11 0%, #0D1B2A 100%);
}

.content {
  padding: 60px 20px 40px;
  max-width: 1200px;
  margin: 0 auto;
}

.page-header {
  margin-bottom: 20px;

  .mode-tabs {
    width: 100%;
    height: 40px;
    box-sizing: border-box;
    margin: 0 auto 18px;
    padding: 3px;
    border: 1px solid $border-color;
    border-radius: $radius-md;
    background: rgba(3, 10, 17, 0.72);
    display: grid;
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }

  .mode-option {
    position: relative;
    display: flex;
    min-width: 0;
    align-items: center;
    justify-content: center;
    height: 100%;
    box-sizing: border-box;
    padding: 0 10px;
    background: transparent;
    border: 0;
    border-radius: $radius-sm;
    color: $text-muted;
    cursor: pointer;
    transition: background $transition-fast, color $transition-fast;

    input {
      position: absolute;
      width: 1px;
      height: 1px;
      opacity: 0;
      pointer-events: none;
    }

    strong {
      min-width: 0;
      font-size: 13px;
      font-weight: 400;
      white-space: nowrap;
      overflow: hidden;
      text-overflow: ellipsis;
    }

    .radio-dot {
      display: none;
    }

    &.active {
      background: $gradient-primary;
      color: $text-inverse;

      strong {
        font-weight: 600;
      }
    }

    &.disabled {
      cursor: wait;
      opacity: 0.65;
    }
  }

  .balance-card {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-top: 10px;
    padding: 11px 14px;
    background: rgba(21, 151, 229, 0.07);
    border: 1px solid rgba(21, 151, 229, 0.14);
    border-radius: 12px;

    > span {
      color: $text-muted;
      font-size: 11px;
    }

    strong {
      color: $brand-primary-light;
      font-size: 16px;
      font-weight: 600;

      small {
        font-size: 10px;
        font-weight: 500;
        opacity: 0.7;
      }
    }
  }

  .mode-tip {
    margin: 8px 2px 0;
    color: #fff;
    font-size: 10px;
    line-height: 1.55;
  }
}

.node-tiers {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 10px;
  margin-bottom: 16px;
}

.custom-amount {
  margin-bottom: 40px;

  .custom-hint {
    font-size: 13px;
    color: rgba(255, 255, 255, 0.7);
    margin-bottom: 10px;
  }

  .custom-row {
    display: flex;
    gap: 12px;
    align-items: center;
  }

  .custom-input {
    flex: 1;
    height: 44px;
    padding: 0 14px;
    border: 1px solid rgba(255, 255, 255, 0.2);
    border-radius: 8px;
    background: rgba(0, 0, 0, 0.25);
    color: #fff;
    font-size: 15px;
    outline: none;

    &:focus {
      border-color: $brand-primary;
    }
  }

  .custom-btn {
    flex-shrink: 0;
    min-width: 120px;
    width: auto;
    height: 44px;
    padding: 0 20px;
  }
}

.subscribe-btn {
  padding: 8px 20px;
  background: $gradient-primary;
  color: $text-inverse;
  border: none;
  border-radius: 12px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s ease;
  width: 100%;

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
        font-size: 14px;
        color: $text-primary;
      }
    }
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

  .pagination-wrapper {
    padding: 16px 0;
    display: flex;
    justify-content: center;
  }
}

.tier-card {
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 14px;
  padding: 12px 14px;
  cursor: pointer;
  transition: all 0.3s ease;
  position: relative;
  overflow: hidden;
  display: flex;
  flex-direction: row;
  align-items: center;
  justify-content: space-between;
  gap: 8px;

  &:hover {
    background: rgba(255, 255, 255, 0.08);
    border-color: rgba(21, 151, 229, 0.3);
    transform: translateY(-4px);
  }

  &.active {
    background: rgba(21, 151, 229, 0.1);
    border-color: $brand-primary;
    box-shadow: 0 0 20px rgba(21, 151, 229, 0.2);
  }

  .tier-header {
    display: flex;
    align-items: baseline;
    gap: 4px;
    min-width: 0;
    
    .tier-price {
      font-size: 20px;
      font-weight: 500;
      color: $brand-primary;
      line-height: 1;
    }

    .tier-unit {
      font-size: 12px;
      color: rgba(255, 255, 255, 0.6);
    }
  }

  .subscribe-btn {
    flex: 0 0 auto;
    width: auto;
    min-width: 104px;
    min-height: 34px;
    padding: 6px 14px;
    border-radius: 9px;
    font-size: 13px;
  }

  .tier-divider {
    height: 1px;
    background: rgba(255, 255, 255, 0.1);
    margin-bottom: 16px;
  }

  .tier-info {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 16px;

    .info-label {
      font-size: 14px;
      color: rgba(255, 255, 255, 0.6);
    }

    .info-value {
      font-size: 18px;
      font-weight: 500;
      color: #fff;
    }
  }

  .tier-status {
    position: absolute;
    top: 16px;
    right: 16px;
    padding: 4px 12px;
    border-radius: 12px;
    font-size: 12px;
    font-weight: 500;

    &.available {
      background: rgba(21, 151, 229, 0.2);
      color: $brand-primary;
    }

    &.full {
      background: rgba(255, 59, 59, 0.2);
      color: #ff3b3b;
    }

    &.coming {
      background: rgba(255, 255, 255, 0.1);
      color: rgba(255, 255, 255, 0.6);
    }
  }
}

.action-section {
  display: flex;
  justify-content: center;
  padding: 20px 0;

  .subscribe-btn {
    padding: 16px 48px;
    background: linear-gradient(135deg, $brand-primary 0%, #087BC1 100%);
    color: #000;
    border: none;
    border-radius: 32px;
    font-size: 18px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.3s ease;

    &:hover:not(:disabled) {
      transform: translateY(-2px);
      box-shadow: 0 8px 24px rgba(21, 151, 229, 0.3);
    }

    &:disabled {
      opacity: 0.5;
      cursor: not-allowed;
      background: rgba(255, 255, 255, 0.1);
      color: rgba(255, 255, 255, 0.5);
    }
  }
}

@keyframes gradient-move {
  0% {
    background-position: 100% 50%;
  }
  50% {
    background-position: 0% 50%;
  }
  100% {
    background-position: 100% 50%;
  }
}
</style>
