<template>
  <div class="page">
    <Header />
    <div class="main">
      <section class="card">
        <h3>报单 / 复投</h3>
        <div class="balances">
          <div>充值钱包：{{ fmt(p.usdt_recharge) }} U</div>
          <div>奖励钱包：{{ fmt(p.usdt_reward) }} U</div>
        </div>
        <van-field v-model="amount" type="number" label="金额" placeholder="最低认购以配置为准" />
        <div class="pay-from">
          <button :class="{ on: payFrom === 'recharge' }" @click="payFrom = 'recharge'">充值钱包</button>
          <button :class="{ on: payFrom === 'reward' }" @click="payFrom = 'reward'">奖励钱包</button>
        </div>
        <p class="tip">仅「充值钱包」报单产生直推奖；奖励钱包报单不产生直推。</p>
        <button class="cta" :disabled="loading" @click="submit">确认报单</button>
      </section>

      <section class="card">
        <h3>我的订单</h3>
        <div v-if="!orders.length" class="empty">暂无订单</div>
        <div v-for="o in orders" :key="o.id" class="order-item">
          <div class="top">
            <span>#{{ o.id }}</span>
            <span :class="o.status">{{ statusText(o.status) }}</span>
          </div>
          <div class="grid">
            <div><label>本金</label><b>{{ fmt(o.total_amount) }}</b></div>
            <div><label>出局目标</label><b>{{ fmt(o.exit_target) }}</b></div>
            <div><label>已获收益</label><b>{{ fmt(o.released_amount) }}</b></div>
            <div><label>资金来源</label><b>{{ o.product_name === 'reward' ? '奖励' : '充值' }}</b></div>
          </div>
        </div>
      </section>
    </div>
    <Tabbar />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { showFailToast, showSuccessToast, showLoadingToast, closeToast } from 'vant'
import Header from '@/components/Header.vue'
import Tabbar from '@/components/Tabbar.vue'
import userPerson from '@/pinia/person'
import { subscribeAix, listOrders, errMsg } from '@/api/aix'

const person = userPerson()
const p = computed(() => person.profile)
const amount = ref('')
const payFrom = ref<'recharge' | 'reward'>('recharge')
const loading = ref(false)
const orders = ref<any[]>([])

const fmt = (v: any) => {
  const n = Number(v || 0)
  if (Number.isNaN(n)) return String(v || '0')
  return n.toFixed(4).replace(/\.?0+$/, (m) => (m.includes('.') ? '' : m))
}

const statusText = (s: string) => (s === 'active' ? '进行中' : s === 'exited' ? '已出局' : s)

const loadOrders = async () => {
  const res = await listOrders()
  orders.value = res.orders || []
}

const submit = async () => {
  const n = Number(amount.value)
  if (!amount.value || Number.isNaN(n) || n <= 0) {
    showFailToast('请输入有效金额')
    return
  }
  loading.value = true
  showLoadingToast({ message: '提交中', duration: 0, forbidClick: true })
  try {
    await subscribeAix(String(amount.value), payFrom.value)
    showSuccessToast('报单成功')
    amount.value = ''
    await person.refreshProfile()
    await loadOrders()
  } catch (e) {
    showFailToast(errMsg(e))
  } finally {
    closeToast()
    loading.value = false
  }
}

onMounted(async () => {
  await person.refreshProfile()
  await loadOrders()
})
</script>

<style scoped lang="scss">
@use '@/style/variables.scss' as *;

.page {
  min-height: 100vh;
  background: $bg-main;
  color: #fff;
  padding-bottom: 72px;
}
.main { padding: 72px 16px 24px; }
.card {
  background: $bg-card;
  border: 1px solid $border-color;
  border-radius: 12px;
  padding: 16px;
  margin-bottom: 12px;
  h3 { margin: 0 0 12px; color: $brand-gold-light; font-size: 16px; }
}
.balances {
  display: grid;
  gap: 6px;
  margin-bottom: 12px;
  font-size: 13px;
  color: rgba(255, 255, 255, 0.75);
}
.pay-from {
  display: flex;
  gap: 8px;
  margin: 12px 0;
  button {
    flex: 1;
    height: 36px;
    border-radius: 8px;
    border: 1px solid $border-color;
    background: transparent;
    color: rgba(255, 255, 255, 0.75);
    &.on {
      background: rgba(212, 175, 55, 0.15);
      color: $brand-gold;
      border-color: $brand-gold;
    }
  }
}
.tip { font-size: 12px; color: rgba(255, 255, 255, 0.45); margin: 0 0 12px; }
.cta {
  width: 100%;
  height: 42px;
  border: none;
  border-radius: 10px;
  background: $gradient-gold;
  color: #111;
  font-weight: 700;
}
:deep(.van-cell) {
  background: $bg-main;
  color: #fff;
  margin-bottom: 8px;
  border-radius: 8px;
  padding: 8px 12px;
}
:deep(.van-field__label) { color: rgba(255, 255, 255, 0.65); }
.empty { color: rgba(255, 255, 255, 0.4); font-size: 13px; text-align: center; padding: 20px 0; }
.order-item {
  border-top: 1px solid $border-light;
  padding: 12px 0;
  .top {
    display: flex;
    justify-content: space-between;
    margin-bottom: 8px;
    font-size: 13px;
    .active { color: #52c41a; }
    .exited { color: rgba(255, 255, 255, 0.45); }
  }
  .grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 8px;
    label { display: block; color: rgba(255, 255, 255, 0.45); font-size: 11px; }
    b { color: #fff; font-size: 13px; font-weight: 600; }
  }
}
</style>
