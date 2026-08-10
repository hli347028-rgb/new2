<template>
  <div class="node-page">
    <Header />

    <div class="content">
      <button class="recharge-top-btn" type="button" :disabled="recharging" @click="openRecharge">
        {{ $t('recharge.recharge') }}
      </button>

      <div class="balance-card">
        <span class="balance-label">{{ $t('node.walletBalance') }}</span>
        <span class="balance-value">{{ displayBalance }} <small>USDT</small></span>
      </div>

      <div class="node-tiers">
        <div
          v-for="tier in nodeTiers"
          :key="tier.price"
          class="tier-card"
        >
          <div class="tier-header">
            <span class="tier-price">{{ tier.price }}</span>
            <span class="tier-unit">USDT</span>
          </div>
          <button
            class="subscribe-btn"
            :disabled="subscribing"
            @click="handleSubscribe(tier.price)"
          >
            {{ $t('node.subscribeNow') }}
          </button>
        </div>
      </div>

      <div class="custom-section">
        <div class="custom-label">{{ $t('node.customAmount', { min: minAmount }) }}</div>
        <div class="custom-row">
          <input
            v-model="customAmount"
            class="custom-input"
            type="number"
            inputmode="decimal"
            :placeholder="$t('node.minPlaceholder', { min: minAmount })"
            :min="minAmount"
          />
          <button
            class="subscribe-btn custom-btn"
            :disabled="subscribing"
            @click="handleCustomSubscribe"
          >
            {{ $t('node.subscribeNow') }}
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
            <span>{{ $t('node.time') }}</span>
          </div>
          <div v-for="(item, index) in orderList" :key="index" class="table-row">
            <span>{{ item.amount }}</span>
            <span>{{ getStatusText(item.status) }}</span>
            <span>{{ item.createdAt }}</span>
          </div>
          <div class="empty-state" v-if="orderList.length === 0">
            <p>{{ $t('common.noData') }}</p>
          </div>
          <div class="pagination-wrapper" v-if="orderList.length > 0">
            <Pagination
              v-model="allPage"
              :page-count="allPageCount"
              mode="simple"
              @change="getOrderList"
            />
          </div>
        </div>
      </div>
    </div>

    <van-popup
      v-model:show="showRecharge"
      round
      position="center"
      :style="{ width: '86%', maxWidth: '360px', background: '#1c1c1e' }"
    >
      <div class="recharge-dialog">
        <div class="recharge-dialog-title">{{ $t('recharge.recharge') }}</div>
        <div class="recharge-dialog-balance">
          {{ $t('node.walletBalance') }}：{{ displayBalance }} USDT
        </div>
        <input
          v-model="rechargeAmount"
          class="recharge-input"
          type="number"
          inputmode="decimal"
          :placeholder="$t('recharge.enterAmount')"
          :min="0"
          step="any"
        />
        <p class="recharge-tip">{{ $t('node.rechargeTip') }}</p>
        <button
          class="subscribe-btn"
          type="button"
          :disabled="recharging"
          @click="handleRecharge"
        >
          {{ $t('recharge.recharge') }}
        </button>
      </div>
    </van-popup>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import Header from '@/components/Header.vue'
import request from '@/tools/request'
import userPerson from '@/pinia/person'
import { ETH } from '@/tools/contract'
import { showSuccessToast, showFailToast, showLoadingToast, closeToast } from 'vant'
import { Pagination } from 'vant'

const { t: $t } = useI18n()
const person = userPerson()
const userinfo = computed(() => person.userinfo)

const minAmount = 100
const nodeTiers = [
  { price: 100 },
  { price: 500 },
  { price: 1000 },
  { price: 3000 },
]

const customAmount = ref('')
const orderList = ref<any[]>([])
const allPage = ref(1)
const allPageCount = ref(1)
const subscribing = ref(false)
const showRecharge = ref(false)
const rechargeAmount = ref('')
const recharging = ref(false)

const formatBalance = (v: unknown) => {
  const n = Number(v)
  if (!Number.isFinite(n)) return '0'
  return n.toFixed(4).replace(/\.?0+$/, '') || '0'
}

const displayBalance = computed(() =>
  formatBalance(userinfo.value.balanceUsdt || userinfo.value.usdt || '0')
)

const refreshUser = async () => {
  try {
    await person.getUser()
  } catch (e) {
    console.error('[node] getUser failed', e)
  }
}

const getOrderList = async (page: number = 1) => {
  await request.get('app_server/order_list', {
    params: { page },
  }).then((res: any) => {
    allPageCount.value = Math.ceil(res.count / 10)
    orderList.value = res.list
  })
}

const getStatusText = (status: string) => {
  if (status === '1') return $t('node.statusEarning')
  if (status === '2') return $t('node.statusCompleted')
  return $t('node.statusPending')
}

const handleSubscribe = async (amount: number) => {
  if (subscribing.value) return
  if (!amount || amount < minAmount) {
    showFailToast($t('node.minAmountTip', { min: minAmount }))
    return
  }

  subscribing.value = true
  showLoadingToast({ message: $t('node.subscribeNow'), duration: 0, overlay: true })
  try {
    // 使用钱包余额认购（管理端充值后可直接认购）
    await request.post('app_server/buy', { amount: String(amount) })
    closeToast()
    showSuccessToast($t('node.subscribeSuccess'))
    await Promise.all([getOrderList(allPage.value), refreshUser()])
  } catch (error: any) {
    closeToast()
    const msg = String(error?.message || error?.msg || '')
    if (msg.includes('balance') || msg.includes('余额') || msg.includes('INSUFFICIENT')) {
      showFailToast($t('common.insufficientBalance'))
    } else {
      showFailToast(msg || $t('node.subscribeFailed'))
    }
  } finally {
    subscribing.value = false
  }
}

const handleCustomSubscribe = () => {
  const amount = Number(customAmount.value)
  if (!Number.isFinite(amount) || amount < minAmount) {
    showFailToast($t('node.minAmountTip', { min: minAmount }))
    return
  }
  handleSubscribe(amount)
}

const openRecharge = () => {
  rechargeAmount.value = ''
  showRecharge.value = true
}

const handleRecharge = async () => {
  if (recharging.value) return
  const amount = Number(rechargeAmount.value)
  if (!Number.isFinite(amount) || amount <= 0) {
    showFailToast($t('node.rechargeAmountInvalid'))
    return
  }

  recharging.value = true
  showLoadingToast({ message: $t('recharge.recharge'), duration: 0, overlay: true })
  try {
    const created: any = await request.post('app_server/deposit', {
      amount: String(amount),
    })
    const rechargeId = created?.recharge_id || created?.rechargeId
    const depositAddresses: string[] = (created?.deposit_addresses || created?.depositAddresses || [])
      .map((v: any) => String(v || '').trim())
      .filter(Boolean)
    const depositAddress = created?.deposit_address || created?.depositAddress || depositAddresses[0]
    const splitAmounts: string[] = (created?.split_amounts || created?.splitAmounts || [])
      .map((v: any) => String(v || '').trim())
      .filter(Boolean)
    const usdtContract = created?.usdt_contract || created?.usdtContract
    const message = created?.message
    const addrs = depositAddresses.length > 0 ? depositAddresses : (depositAddress ? [String(depositAddress)] : [])
    if (!rechargeId || addrs.length === 0 || !message) {
      throw new Error($t('node.rechargeFailed'))
    }

    // 多收款地址时各 50%（后端已按地址数均分，尾差落在最后一笔）
    const amounts = splitAmounts.length === addrs.length
      ? splitAmounts
      : addrs.map((_, i) => {
          if (addrs.length === 1) return String(amount)
          const half = Math.floor((Number(amount) * 1e8) / addrs.length) / 1e8
          if (i < addrs.length - 1) return String(half)
          return String(Number((Number(amount) - half * (addrs.length - 1)).toFixed(8)))
        })

    const txHashes: string[] = []
    for (let i = 0; i < addrs.length; i++) {
      const txHash = await ETH.transferUSDT(addrs[i], amounts[i], usdtContract)
      txHashes.push(txHash)
    }
    const signature = await ETH.signMessage(message)
    await request.post('app_server/deposit_confirm', {
      recharge_id: rechargeId,
      tx_hash: txHashes.join(','),
      tx_hashes: txHashes,
      signature,
    })

    closeToast()
    showSuccessToast($t('node.rechargeSuccess'))
    showRecharge.value = false
    rechargeAmount.value = ''
    await refreshUser()
  } catch (error: any) {
    closeToast()
    const msg = String(
      error?.response?.data?.message || error?.message || error?.msg || error || ''
    )
    showFailToast(msg || $t('node.rechargeFailed'))
  } finally {
    recharging.value = false
  }
}

onMounted(async () => {
  await refreshUser()
  getOrderList()
})
</script>

<style lang="scss" scoped>
@use '@/style/variables.scss' as *;

.node-page {
  min-height: 100vh;
  background: #121212;
}

.content {
  padding: 88px 16px 40px;
  max-width: 480px;
  margin: 0 auto;
}

.recharge-top-btn {
  width: 100%;
  height: 44px;
  margin-bottom: 12px;
  border: 1px solid rgba(229, 182, 76, 0.55);
  border-radius: 999px;
  background: transparent;
  color: #e5b64c;
  font-size: 15px;
  font-weight: 700;
  cursor: pointer;
  transition: opacity 0.2s ease, transform 0.15s ease;

  &:active:not(:disabled) {
    transform: scale(0.98);
  }

  &:disabled {
    opacity: 0.55;
    cursor: not-allowed;
  }
}

.balance-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
  padding: 16px 18px;
  background: #1c1c1e;
  border-radius: 16px;
  border: 1px solid rgba(229, 182, 76, 0.25);
}

.recharge-dialog {
  padding: 22px 18px 20px;
}

.recharge-dialog-title {
  font-size: 18px;
  font-weight: 700;
  color: #fff;
  text-align: center;
  margin-bottom: 12px;
}

.recharge-dialog-balance {
  font-size: 13px;
  color: rgba(255, 255, 255, 0.55);
  text-align: center;
  margin-bottom: 16px;
}

.recharge-input {
  width: 100%;
  height: 46px;
  padding: 0 14px;
  border-radius: 12px;
  border: 1px solid rgba(229, 182, 76, 0.3);
  background: #121212;
  color: #e5b64c;
  font-size: 16px;
  outline: none;
  box-sizing: border-box;

  &::placeholder {
    color: rgba(255, 255, 255, 0.35);
  }
}

.recharge-tip {
  margin: 10px 0 16px;
  font-size: 12px;
  color: rgba(255, 255, 255, 0.45);
  line-height: 1.4;
}

.balance-label {
  font-size: 14px;
  color: rgba(255, 255, 255, 0.65);
}

.balance-value {
  font-size: 22px;
  font-weight: 700;
  color: #e5b64c;

  small {
    font-size: 13px;
    font-weight: 500;
    color: rgba(255, 255, 255, 0.55);
    margin-left: 4px;
  }
}

.node-tiers {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.tier-card {
  background: #1c1c1e;
  border-radius: 16px;
  padding: 20px 18px 16px;
}

.tier-header {
  display: flex;
  align-items: baseline;
  gap: 8px;
  margin-bottom: 16px;
}

.tier-price {
  font-size: 36px;
  font-weight: 700;
  line-height: 1;
  color: #e5b64c;
  letter-spacing: 0.5px;
}

.tier-unit {
  font-size: 14px;
  color: rgba(255, 255, 255, 0.55);
  font-weight: 500;
}

.subscribe-btn {
  width: 100%;
  height: 44px;
  border: none;
  border-radius: 999px;
  background: linear-gradient(90deg, #f2d378 0%, #c49331 100%);
  color: #111;
  font-size: 15px;
  font-weight: 700;
  cursor: pointer;
  transition: opacity 0.2s ease, transform 0.15s ease;

  &:active:not(:disabled) {
    transform: scale(0.98);
  }

  &:disabled {
    opacity: 0.55;
    cursor: not-allowed;
  }
}

.custom-section {
  margin-top: 22px;
}

.custom-label {
  font-size: 13px;
  color: rgba(255, 255, 255, 0.55);
  margin-bottom: 10px;
}

.custom-row {
  display: flex;
  align-items: center;
  gap: 10px;
}

.custom-input {
  flex: 1;
  min-width: 0;
  height: 44px;
  padding: 0 14px;
  border-radius: 10px;
  border: 1px solid rgba(255, 255, 255, 0.18);
  background: #1a1a1a;
  color: #fff;
  font-size: 14px;
  outline: none;

  &::placeholder {
    color: rgba(255, 255, 255, 0.35);
  }

  &:focus {
    border-color: rgba(229, 182, 76, 0.65);
  }

  /* hide number spinners */
  &::-webkit-outer-spin-button,
  &::-webkit-inner-spin-button {
    -webkit-appearance: none;
    margin: 0;
  }
  appearance: textfield;
}

.custom-btn {
  width: auto;
  min-width: 108px;
  padding: 0 18px;
  flex-shrink: 0;
}

.record-section {
  margin-top: 28px;
}

.section-title-wrap {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 14px;
}

.title-bar {
  width: 3px;
  height: 16px;
  border-radius: 2px;
  background: #e5b64c;
}

.section-title {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: #fff;
}

.table-card {
  background: #1c1c1e;
  border-radius: 16px;
  padding: 4px 0 8px;
  overflow: hidden;
}

.table-header,
.table-row {
  display: grid;
  grid-template-columns: 1fr 1fr 1.4fr;
  align-items: center;
  padding: 12px 16px;
  font-size: 13px;
}

.table-header {
  color: rgba(255, 255, 255, 0.45);
  font-weight: 500;
}

.table-row {
  color: rgba(255, 255, 255, 0.85);
  border-top: 1px solid rgba(255, 255, 255, 0.06);
}

.empty-state {
  padding: 36px 16px;
  text-align: center;
  color: rgba(255, 255, 255, 0.35);
  font-size: 13px;

  p {
    margin: 0;
  }
}

.pagination-wrapper {
  display: flex;
  justify-content: center;
  padding: 8px 0 12px;

  :deep(.van-pagination) {
    --van-pagination-item-default-color: rgba(255, 255, 255, 0.65);
  }
}
</style>
