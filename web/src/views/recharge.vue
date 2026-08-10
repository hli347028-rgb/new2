<template>
  <div class="page">
    <Header />
    <div class="main">
      <van-nav-bar title="充值 USDT" left-arrow :border="false" @click-left="$router.back()" />
      <section class="card">
        <div class="bal">充值钱包：{{ fmt(p.usdt_recharge) }} U</div>
        <van-field v-model="amount" type="number" label="金额" placeholder="输入充值金额" />
        <p class="tip">将向平台地址转入 BSC USDT，确认后入账充值钱包。</p>
        <button class="cta" :disabled="loading" @click="doRecharge">发起充值</button>
      </section>

      <section class="card">
        <h3>充值记录</h3>
        <div v-if="!list.length" class="empty">暂无记录</div>
        <div v-for="item in list" :key="item.id" class="row-item">
          <div>
            <div>{{ fmt(item.amount) }} U</div>
            <div class="sub">{{ formatTime(item.created_at || item.createdAt) }}</div>
          </div>
          <div class="status">{{ item.status }}</div>
        </div>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { showFailToast, showSuccessToast, showLoadingToast, closeToast, showConfirmDialog } from 'vant'
import Header from '@/components/Header.vue'
import userPerson from '@/pinia/person'
import { createRecharge, confirmRecharge, listRecharges, errMsg } from '@/api/aix'
import { ETH } from '@/tools/contract'
import dayjs from 'dayjs'

const person = userPerson()
const p = computed(() => person.profile)
const amount = ref('')
const loading = ref(false)
const list = ref<any[]>([])

const fmt = (v: any) => {
  const n = Number(v || 0)
  if (Number.isNaN(n)) return '0'
  return n.toLocaleString(undefined, { maximumFractionDigits: 6 })
}
const formatTime = (ts: any) => {
  if (!ts) return '-'
  if (typeof ts === 'string' && ts.includes('-')) return ts
  return dayjs.unix(Number(ts)).format('YYYY-MM-DD HH:mm')
}

const loadList = async () => {
  try {
    const res = await listRecharges()
    list.value = res.recharges || res.list || []
  } catch {
    list.value = []
  }
}

const doRecharge = async () => {
  const n = Number(amount.value)
  if (!amount.value || Number.isNaN(n) || n <= 0) {
    showFailToast('请输入有效金额')
    return
  }
  loading.value = true
  showLoadingToast({ message: '创建订单', duration: 0, forbidClick: true })
  try {
    const order = await createRecharge(String(amount.value))
    const deposit =
      (order.deposit_addresses && order.deposit_addresses[0]) || order.deposit_address
    if (!deposit) throw new Error('平台收款地址未配置')

    closeToast()
    await showConfirmDialog({
      title: '确认转账',
      message: `向 ${deposit.slice(0, 8)}...${deposit.slice(-6)} 转入 ${amount.value} USDT？`,
    })

    showLoadingToast({ message: '请在钱包确认', duration: 0, forbidClick: true })
    const usdt = order.usdt_contract || import.meta.env.VITE_USDT
    const txHash = await ETH.transferUSDT(deposit, String(amount.value), usdt)
    await confirmRecharge({
      recharge_id: order.recharge_id,
      tx_hash: txHash,
      signature: '',
    })
    showSuccessToast('充值已提交')
    amount.value = ''
    await person.refreshProfile()
    await loadList()
  } catch (e: any) {
    if (e !== 'cancel' && e?.message !== 'cancel') showFailToast(errMsg(e))
  } finally {
    closeToast()
    loading.value = false
  }
}

onMounted(async () => {
  await person.refreshProfile()
  await loadList()
})
</script>

<style scoped lang="scss">
@use '@/style/variables.scss' as *;

.page {
  min-height: 100vh;
  background: $bg-main;
  color: #fff;
}
.main { padding: 56px 16px 24px; }
:deep(.van-nav-bar) {
  background: transparent;
  .van-nav-bar__title, .van-icon { color: #fff; }
}
.card {
  background: $bg-card;
  border: 1px solid $border-color;
  border-radius: 12px;
  padding: 16px;
  margin-bottom: 12px;
  h3 { margin: 0 0 12px; color: $brand-gold-light; }
}
.bal { margin-bottom: 12px; color: $brand-gold; }
.tip { font-size: 12px; color: rgba(255, 255, 255, 0.45); }
.cta {
  width: 100%;
  height: 42px;
  margin-top: 8px;
  border: none;
  border-radius: 10px;
  background: $gradient-gold;
  color: #111;
  font-weight: 700;
}
:deep(.van-cell) {
  background: $bg-main;
  color: #fff;
  border-radius: 8px;
}
.empty { text-align: center; color: rgba(255, 255, 255, 0.4); padding: 16px; font-size: 13px; }
.row-item {
  display: flex;
  justify-content: space-between;
  padding: 10px 0;
  border-top: 1px solid $border-light;
  font-size: 13px;
  .sub { font-size: 11px; color: rgba(255, 255, 255, 0.4); margin-top: 2px; }
  .status { color: $brand-gold-light; }
}
</style>
