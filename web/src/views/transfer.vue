<template>
  <div class="page">
    <Header />
    <div class="main">
      <van-nav-bar title="划转" left-arrow :border="false" @click-left="$router.back()" />

      <section class="tabs">
        <button :class="{ on: mode === 'inner' }" @click="mode = 'inner'">充值→奖励</button>
        <button :class="{ on: mode === 'peer' }" @click="mode = 'peer'">上下级转账</button>
      </section>

      <section v-if="mode === 'inner'" class="card">
        <div class="bal">充值钱包 {{ fmt(p.usdt_recharge) }} → 奖励钱包 {{ fmt(p.usdt_reward) }}</div>
        <van-field v-model="innerAmount" type="number" label="金额" placeholder="USDT" />
        <p class="tip">同账户划转不产生直推；划出后不再具备可发直推属性。</p>
        <button class="cta" @click="doInner">确认转入奖励钱包</button>
      </section>

      <section v-else class="card">
        <van-field v-model="toAddress" label="对方地址" placeholder="上下级钱包地址" />
        <van-field v-model="peerAmount" type="number" label="金额" />
        <div class="bal">奖励钱包余额：{{ fmt(p.usdt_reward) }} USDT</div>
        <div v-if="false" class="pay-from">
          <button :class="{ on: asset === 'USDT' }" @click="asset = 'USDT'">USDT</button>
          <button :class="{ on: asset === 'AIX' }" @click="asset = 'AIX'">AIX</button>
        </div>
        <div v-if="false" class="pay-from">
          <button :class="{ on: payFrom === 'recharge' }" @click="payFrom = 'recharge'">扣充值钱包</button>
          <button :class="{ on: payFrom === 'reward' }" @click="payFrom = 'reward'">扣奖励钱包</button>
        </div>
        <p class="tip">收款方 USDT 只进奖励钱包；永不触发直推。</p>
        <button class="cta" @click="doPeer">确认转账</button>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { showFailToast, showSuccessToast, showLoadingToast, closeToast } from 'vant'
import Header from '@/components/Header.vue'
import userPerson from '@/pinia/person'
import { rechargeToReward, transfer, errMsg } from '@/api/aix'

const person = userPerson()
const p = computed(() => person.profile)
const mode = ref<'inner' | 'peer'>('inner')
const innerAmount = ref('')
const peerAmount = ref('')
const toAddress = ref('')
const asset = ref<'USDT' | 'AIX'>('USDT')
const payFrom = ref<'recharge' | 'reward'>('reward')

const fmt = (v: any) => {
  const n = Number(v || 0)
  if (Number.isNaN(n)) return '0'
  return n.toLocaleString(undefined, { maximumFractionDigits: 6 })
}

const doInner = async () => {
  const n = Number(innerAmount.value)
  if (!innerAmount.value || Number.isNaN(n) || n <= 0) {
    showFailToast('请输入有效金额')
    return
  }
  showLoadingToast({ message: '处理中', duration: 0, forbidClick: true })
  try {
    await rechargeToReward(String(innerAmount.value))
    showSuccessToast('已转入奖励钱包')
    innerAmount.value = ''
    await person.refreshProfile()
  } catch (e) {
    showFailToast(errMsg(e))
  } finally {
    closeToast()
  }
}

const doPeer = async () => {
  const n = Number(peerAmount.value)
  if (!toAddress.value || !peerAmount.value || Number.isNaN(n) || n <= 0) {
    showFailToast('请填写地址与金额')
    return
  }
  showLoadingToast({ message: '处理中', duration: 0, forbidClick: true })
  try {
    await transfer({
      to_address: toAddress.value.trim(),
      asset: 'USDT',
      amount: String(peerAmount.value),
      pay_from: 'reward',
    })
    showSuccessToast('转账成功')
    peerAmount.value = ''
    await person.refreshProfile()
  } catch (e) {
    showFailToast(errMsg(e))
  } finally {
    closeToast()
  }
}
</script>

<style scoped lang="scss">
@use '@/style/variables.scss' as *;

.page { min-height: 100vh; background: $bg-main; color: #fff; }
.main { padding: 56px 16px 24px; }
:deep(.van-nav-bar) {
  background: transparent;
  .van-nav-bar__title, .van-icon { color: #fff; }
}
.tabs {
  display: flex;
  gap: 8px;
  margin-bottom: 12px;
  button {
    flex: 1;
    height: 36px;
    border-radius: 8px;
    border: 1px solid $border-color;
    background: $bg-card;
    color: rgba(255, 255, 255, 0.7);
    &.on { color: $brand-gold; border-color: $brand-gold; }
  }
}
.card {
  background: $bg-card;
  border: 1px solid $border-color;
  border-radius: 12px;
  padding: 16px;
}
.bal { margin-bottom: 12px; font-size: 13px; color: rgba(255, 255, 255, 0.75); }
.tip { font-size: 12px; color: rgba(255, 255, 255, 0.45); }
.pay-from {
  display: flex;
  gap: 8px;
  margin: 10px 0;
  button {
    flex: 1;
    height: 34px;
    border-radius: 8px;
    border: 1px solid $border-color;
    background: transparent;
    color: rgba(255, 255, 255, 0.7);
    font-size: 12px;
    &.on { color: $brand-gold; border-color: $brand-gold; }
  }
}
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
</style>
