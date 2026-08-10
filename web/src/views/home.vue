<template>
  <div class="page">
    <Header />
    <div class="main">
      <section class="hero">
        <p class="eyebrow">AIX PROTOCOL</p>
        <h1>AIX</h1>
        <p class="desc">USDT 报单 · 日静态释放 · 直推与管理奖 · 4 倍出局</p>
        <button class="cta" @click="$router.push('/order')">立即报单</button>
      </section>

      <section class="card">
        <div class="row">
          <span>AIX 价格</span>
          <strong>{{ price }} USDT / 枚</strong>
        </div>
        <p class="hint">价格=1 为 1:1；价格=2 为 USDT:AIX = 2:1</p>
      </section>

      <section class="card rules">
        <h3>规则摘要</h3>
        <ul>
          <li>静态：本金 × 0.5%/日，按价格换算入 AIX 代币</li>
          <li>直推：仅充值钱包报单产生，50% USDT 进奖励钱包</li>
          <li>管理奖：W1–W10 级差，USDT 进奖励钱包</li>
          <li>出局目标按 USDT；仅可提现 AIX</li>
        </ul>
      </section>
    </div>
    <Tabbar />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import Header from '@/components/Header.vue'
import Tabbar from '@/components/Tabbar.vue'
import { getAixPrice } from '@/api/aix'

const price = ref('1')

onMounted(async () => {
  try {
    const res = await getAixPrice()
    price.value = res.price || '1'
  } catch {
    price.value = '1'
  }
})
</script>

<style scoped lang="scss">
@use '@/style/variables.scss' as *;

.page {
  min-height: 100vh;
  background: radial-gradient(120% 80% at 50% -10%, rgba(212, 175, 55, 0.18), transparent 55%), $bg-main;
  color: #fff;
  padding-bottom: 72px;
}

.main {
  padding: 72px 16px 24px;
}

.hero {
  margin-bottom: 20px;

  .eyebrow {
    color: $brand-gold-light;
    letter-spacing: 0.2em;
    font-size: 12px;
    margin: 0 0 8px;
  }
  h1 {
    margin: 0;
    font-size: 48px;
    line-height: 1;
    background: $gradient-gold;
    -webkit-background-clip: text;
    color: transparent;
  }
  .desc {
    margin: 12px 0 20px;
    color: rgba(255, 255, 255, 0.72);
    font-size: 14px;
    line-height: 1.5;
  }
}

.cta {
  width: 100%;
  height: 44px;
  border: none;
  border-radius: 10px;
  background: $gradient-gold;
  color: #111;
  font-weight: 700;
  font-size: 15px;
}

.card {
  background: $bg-card;
  border: 1px solid $border-color;
  border-radius: 12px;
  padding: 16px;
  margin-bottom: 12px;

  .row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    span { color: rgba(255, 255, 255, 0.65); font-size: 13px; }
    strong { color: $brand-gold; font-size: 18px; }
  }
  .hint {
    margin: 8px 0 0;
    font-size: 12px;
    color: rgba(255, 255, 255, 0.45);
  }
}

.rules {
  h3 {
    margin: 0 0 10px;
    color: $brand-gold-light;
    font-size: 15px;
  }
  ul {
    margin: 0;
    padding-left: 18px;
    color: rgba(255, 255, 255, 0.75);
    font-size: 13px;
    line-height: 1.7;
  }
}
</style>
