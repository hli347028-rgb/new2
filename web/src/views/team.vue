<template>
  <div class="page">
    <Header />
    <div class="main">
      <section class="card">
        <h3>我的邀请</h3>
        <div class="row"><span>管理级别</span><b>W{{ p.mgmt_level || 0 }}</b></div>
        <div class="row"><span>小区业绩</span><b>{{ fmt(p.small_area_perf) }}</b></div>
        <div class="row"><span>团队业绩</span><b>{{ fmt(p.team_perf) }}</b></div>
        <div class="row"><span>直推人数</span><b>{{ invitees.length }}</b></div>
      </section>

      <section class="card">
        <h3>邀请链接</h3>
        <div class="link">{{ inviteLink }}</div>
        <button class="cta" @click="copyLink">复制链接</button>
      </section>

      <section class="card">
        <h3>直推列表</h3>
        <div v-if="!invitees.length" class="empty">暂无直推</div>
        <div v-for="(u, i) in invitees" :key="i" class="member">
          <div class="addr">{{ short(u.address) }}</div>
          <div class="meta">业绩 {{ fmt(u.team_stake || u.team_perf || 0) }}</div>
        </div>
      </section>
    </div>
    <Tabbar />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { showSuccessToast, showFailToast } from 'vant'
import copy from 'copy-to-clipboard'
import Header from '@/components/Header.vue'
import Tabbar from '@/components/Tabbar.vue'
import userPerson from '@/pinia/person'
import { listInvitees, getAuthProfile } from '@/api/aix'

const person = userPerson()
const p = computed(() => person.profile)
const invitees = ref<any[]>([])

const inviteLink = computed(() => {
  const base = `${location.origin}${location.pathname}#/`
  const addr = person.address || ''
  return addr ? `${base}?inviteCode=${addr}` : base
})

const fmt = (v: any) => {
  const n = Number(v || 0)
  if (Number.isNaN(n)) return '0'
  return n.toLocaleString(undefined, { maximumFractionDigits: 4 })
}
const short = (a: string) => (!a ? '-' : a.slice(0, 8) + '...' + a.slice(-6))

const copyLink = () => {
  if (copy(inviteLink.value)) showSuccessToast('已复制')
  else showFailToast('复制失败')
}

onMounted(async () => {
  await person.refreshProfile()
  try {
    const res = await listInvitees()
    invitees.value = res.invitees || []
  } catch {
    try {
      const profile = await getAuthProfile()
      invitees.value = profile.downline_invitees || []
    } catch {
      invitees.value = []
    }
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
.card {
  background: $bg-card;
  border: 1px solid $border-color;
  border-radius: 12px;
  padding: 16px;
  margin-bottom: 12px;
  h3 { margin: 0 0 12px; color: $brand-gold-light; font-size: 15px; }
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
.link {
  word-break: break-all;
  font-size: 12px;
  color: rgba(255, 255, 255, 0.7);
  background: $bg-main;
  padding: 10px;
  border-radius: 8px;
  margin-bottom: 12px;
}
.cta {
  width: 100%;
  height: 40px;
  border: none;
  border-radius: 10px;
  background: $gradient-gold;
  color: #111;
  font-weight: 700;
}
.empty { text-align: center; color: rgba(255, 255, 255, 0.4); padding: 16px 0; font-size: 13px; }
.member {
  padding: 10px 0;
  border-top: 1px solid $border-light;
  .addr { font-size: 13px; }
  .meta { font-size: 11px; color: rgba(255, 255, 255, 0.45); margin-top: 4px; }
}
</style>
