<template>
  <div class="page">
    <Header />
    <div class="main">
      <section class="hero-card">
        <div class="item">
          <label>充值钱包</label>
          <b>{{ fmt(p.usdt_recharge) }} <small>USDT</small></b>
        </div>
        <div class="item">
          <label>奖励钱包</label>
          <b>{{ fmt(p.usdt_reward) }} <small>USDT</small></b>
        </div>
        <div class="item">
          <label>AIX 代币数</label>
          <b class="gold">{{ fmt(p.aix_balance) }}</b>
        </div>
        <div class="item">
          <label>静态总收益</label>
          <b>{{ fmt(p.static_usdt_total) }} <small>USDT</small></b>
        </div>
      </section>

      <section class="actions">
        <button @click="$router.push('/recharge')">充值</button>
        <button @click="$router.push('/transfer')">划转</button>
        <button @click="$router.push('/withdraw')">提现 AIX</button>
      </section>

      <section class="card">
        <div class="row"><span>剩余出局额度</span><b>{{ fmt(p.unexited_amount) }} U</b></div>
        <div class="row"><span>订单数</span><b>{{ p.total_nodes || 0 }}</b></div>
        <div class="row"><span>管理级别</span><b>W{{ p.mgmt_level || 0 }}</b></div>
        <div class="row"><span>预估日静态</span><b>{{ fmt(p.pending_amount) }} U</b></div>
      </section>

      <section class="card">
        <h3>奖励流水</h3>
        <div v-if="!rewards.length" class="empty">暂无流水</div>
        <div v-for="r in rewards" :key="r.id" class="reward">
          <div class="left">
            <div class="type">{{ typeText(r.type) }}</div>
            <div class="time">{{ formatTime(r.created_time) }}</div>
          </div>
          <div class="amt">+{{ fmt(r.amount) }} {{ r.asset }}</div>
        </div>
      </section>
    </div>
    <Tabbar />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import Header from '@/components/Header.vue'
import Tabbar from '@/components/Tabbar.vue'
import userPerson from '@/pinia/person'
import { listRewards } from '@/api/aix'
import dayjs from 'dayjs'

const person = userPerson()
const p = computed(() => person.profile)
const rewards = ref<any[]>([])

const fmt = (v: any) => {
  const n = Number(v || 0)
  if (Number.isNaN(n)) return '0'
  return n.toLocaleString(undefined, { maximumFractionDigits: 6 })
}

const typeText = (t: string) =>
  ({
    static_aix: '静态释放',
    dynamic_usdt: '直推奖',
    mgmt: '管理奖',
    transfer_in: '转入',
    transfer_out: '转出',
  } as Record<string, string>)[t] || t

const formatTime = (ts: number) => (ts ? dayjs.unix(ts).format('MM-DD HH:mm') : '-')

onMounted(async () => {
  await person.refreshProfile()
  try {
    const res = await listRewards()
    rewards.value = res.rewards || []
  } catch {
    rewards.value = []
  }
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
.hero-card {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10px;
  background: linear-gradient(160deg, rgba(212, 175, 55, 0.18), $bg-card 55%);
  border: 1px solid $border-color;
  border-radius: 14px;
  padding: 16px;
  margin-bottom: 12px;
  .item label {
    display: block;
    font-size: 11px;
    color: rgba(255, 255, 255, 0.55);
    margin-bottom: 4px;
  }
  b { font-size: 16px; }
  .gold { color: $brand-gold; }
  small { font-size: 11px; color: rgba(255, 255, 255, 0.45); font-weight: 400; }
}
.actions {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 8px;
  margin-bottom: 12px;
  button {
    height: 40px;
    border-radius: 10px;
    border: 1px solid $border-color;
    background: $bg-card;
    color: $brand-gold-light;
    font-size: 13px;
  }
}
.card {
  background: $bg-card;
  border: 1px solid $border-color;
  border-radius: 12px;
  padding: 14px 16px;
  margin-bottom: 12px;
  h3 { margin: 0 0 10px; color: $brand-gold-light; font-size: 15px; }
  .row {
    display: flex;
    justify-content: space-between;
    padding: 8px 0;
    border-bottom: 1px solid $border-light;
    font-size: 13px;
    span { color: rgba(255, 255, 255, 0.55); }
    &:last-child { border-bottom: none; }
  }
}
.empty { text-align: center; color: rgba(255, 255, 255, 0.4); padding: 16px 0; font-size: 13px; }
.reward {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 0;
  border-top: 1px solid $border-light;
  .type { font-size: 13px; }
  .time { font-size: 11px; color: rgba(255, 255, 255, 0.4); margin-top: 2px; }
  .amt { color: $brand-gold; font-weight: 600; font-size: 13px; }
}
</style>
