<template>
  <div class="page">
    <Header />
    <div class="main">
      <van-nav-bar title="提现 AIX" left-arrow :border="false" @click-left="$router.back()" />
      <section class="card">
        <div class="bal">可提现 AIX：{{ fmt(p.aix_balance) }}</div>
        <van-field v-model="amount" type="number" label="数量" placeholder="提现代币数量" />
        <van-field v-model="toAddress" label="收款地址" :placeholder="p.address || '默认本人地址'" />
        <p class="tip">仅支持提现 AIX，不支持提现 USDT。代币合约待配置，申请后可能为 pending。</p>
        <button class="cta" @click="submit">提交提现</button>
      </section>

      <section class="card">
        <h3>提现记录</h3>
        <div v-if="!list.length" class="empty">暂无记录</div>
        <div v-for="w in list" :key="w.id" class="row-item">
          <div>
            <div>{{ fmt(w.amount) }} AIX</div>
            <div class="sub">{{ formatTime(w.created_at) }}</div>
          </div>
          <div class="status">{{ w.status }}</div>
        </div>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { showFailToast, showSuccessToast, showLoadingToast, closeToast } from 'vant'
import Header from '@/components/Header.vue'
import userPerson from '@/pinia/person'
import { withdrawAix, listWithdrawals, errMsg } from '@/api/aix'
import dayjs from 'dayjs'

const person = userPerson()
const p = computed(() => person.profile)
const amount = ref('')
const toAddress = ref('')
const list = ref<any[]>([])

const fmt = (v: any) => {
  const n = Number(v || 0)
  if (Number.isNaN(n)) return '0'
  return n.toLocaleString(undefined, { maximumFractionDigits: 6 })
}
const formatTime = (ts: any) => (ts ? dayjs.unix(Number(ts)).format('YYYY-MM-DD HH:mm') : '-')

const loadList = async () => {
  try {
    const res = await listWithdrawals()
    list.value = res.withdrawals || []
  } catch {
    list.value = []
  }
}

const submit = async () => {
  const n = Number(amount.value)
  if (!amount.value || Number.isNaN(n) || n <= 0) {
    showFailToast('请输入有效数量')
    return
  }
  showLoadingToast({ message: '提交中', duration: 0, forbidClick: true })
  try {
    await withdrawAix(String(amount.value), toAddress.value.trim())
    showSuccessToast('提现已提交')
    amount.value = ''
    await person.refreshProfile()
    await loadList()
  } catch (e) {
    showFailToast(errMsg(e))
  } finally {
    closeToast()
  }
}

onMounted(async () => {
  await person.refreshProfile()
  await loadList()
})
</script>

<style scoped lang="scss">
@use '@/style/variables.scss' as *;

.page { min-height: 100vh; background: $bg-main; color: #fff; }
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
.bal { margin-bottom: 12px; color: $brand-gold; font-size: 16px; }
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
  margin-bottom: 8px;
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
