<template>
  <header class="app-header">
    <div class="header-container">
      <div class="header-left" @click="showSidebar = true">
        <img src="/assets/logo.png" alt="Logo" class="logo" />
        <span>AIX</span>
      </div>
      <div class="header-right">
        <button class="connect-btn">
          {{ address ? formatAddress(address) : '连接钱包' }}
        </button>
      </div>
    </div>
  </header>

  <Teleport to="body">
    <Sidebar :visible="showSidebar" @close="showSidebar = false" />
  </Teleport>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import userPerson from '@/pinia/person'
import Sidebar from '@/components/Sidebar.vue'

const person = userPerson()
const showSidebar = ref(false)
const address = computed(() => person.address || '')

const formatAddress = (value: string) => {
  if (!value || value.length < 12) return value
  return value.slice(0, 6) + '...' + value.slice(-4)
}
</script>

<style scoped lang="scss">
@use '@/style/variables.scss' as *;

.app-header {
  position: fixed;
  top: 0;
  left: 50%;
  transform: translateX(-50%);
  width: 100%;
  max-width: 414px;
  z-index: 100;
  background: rgba(5, 5, 5, 0.92);
  border-bottom: 1px solid $border-color;
  backdrop-filter: blur(8px);
}

.header-container {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 56px;
  padding: 0 16px;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  color: $brand-gold;
  font-weight: 700;
  font-size: 18px;
  letter-spacing: 0.08em;

  .logo {
    width: 28px;
    height: 28px;
    object-fit: contain;
  }
}

.connect-btn {
  border: 1px solid $border-color;
  background: $bg-card;
  color: $brand-gold-light;
  border-radius: 20px;
  padding: 6px 12px;
  font-size: 12px;
}
</style>
