<template>
  <Teleport to="body">
    <div v-if="visible" class="sidebar-overlay" @click="close">
      <div class="sidebar" @click.stop>
        <button class="close-btn" @click="close">×</button>
        <div class="sidebar-logo">
          <img src="/assets/logo.png" alt="Logo" class="logo-img" />
          <div class="brand">AIX</div>
        </div>
        <div class="user-address" @click="copyAddress">{{ shortAddr }}</div>
        <nav class="sidebar-nav">
          <div class="nav-item" :class="{ active: isActive('/') }" @click="go('/')">首页</div>
          <div class="nav-item" :class="{ active: isActive('/order') }" @click="go('/order')">报单</div>
          <div class="nav-item" :class="{ active: isActive('/mine') }" @click="go('/mine')">资产</div>
          <div class="nav-item" :class="{ active: isActive('/team') }" @click="go('/team')">团队</div>
        </nav>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { showSuccessToast, showFailToast } from 'vant'
import userPerson from '@/pinia/person'
import copy from 'copy-to-clipboard'

defineProps<{ visible: boolean }>()
const emit = defineEmits(['close'])

const router = useRouter()
const route = useRoute()
const person = userPerson()

const shortAddr = computed(() => {
  const a = person.address || ''
  if (!a) return '-'
  return a.slice(0, 6) + '...' + a.slice(-4)
})

const isActive = (path: string) => route.path === path
const close = () => emit('close')
const go = (path: string) => {
  router.push(path)
  close()
}

const copyAddress = () => {
  if (!person.address) return
  if (copy(person.address)) showSuccessToast('地址已复制')
  else showFailToast('复制失败')
}
</script>

<style scoped lang="scss">
@use '@/style/variables.scss' as *;

.sidebar-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.55);
  z-index: 2000;
}

.sidebar {
  width: 260px;
  height: 100%;
  background: $bg-card;
  border-right: 1px solid $border-color;
  padding: 24px 16px;
  position: relative;
}

.close-btn {
  position: absolute;
  right: 12px;
  top: 8px;
  background: transparent;
  border: none;
  color: #fff;
  font-size: 28px;
  line-height: 1;
}

.sidebar-logo {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-top: 12px;

  .logo-img {
    width: 40px;
    height: 40px;
  }
  .brand {
    color: $brand-gold;
    font-size: 22px;
    font-weight: 700;
    letter-spacing: 0.12em;
  }
}

.user-address {
  margin: 20px 0;
  color: rgba(255, 255, 255, 0.7);
  font-size: 13px;
  padding: 10px 12px;
  background: $bg-main;
  border-radius: 8px;
  border: 1px solid $border-light;
}

.nav-item {
  padding: 14px 12px;
  color: rgba(255, 255, 255, 0.85);
  border-radius: 8px;
  margin-bottom: 6px;
  font-size: 15px;

  &.active,
  &:active {
    background: rgba(212, 175, 55, 0.12);
    color: $brand-gold;
  }
}
</style>
