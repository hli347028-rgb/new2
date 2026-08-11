<template>
  <div class="tabbar">
    <div
      v-for="tab in tabs"
      :key="tab.path"
      class="tab-item"
      :class="{ active: isActive(tab.path) }"
      @click="handleTabClick(tab.path)"
    >
      <img :src="isActive(tab.path) ? tab.activeIcon : tab.icon" :alt="$t(tab.name)" />
      <span class="tab-label">{{ $t(tab.name) }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useRoute, useRouter } from 'vue-router'

interface TabItem {
  path: string
  name: string
  icon: string
  activeIcon: string
}

/** 底栏仅主路径：首页 / 认购 / 社区 / 资产 */
const tabs: TabItem[] = [
  {
    path: '/',
    name: 'tab.home',
    icon: '/assets/tabbar/home.svg',
    activeIcon: '/assets/tabbar/home-active.svg'
  },
  {
    path: '/node',
    name: 'tab.nodeSubscription',
    icon: '/assets/tabbar/pledge.svg',
    activeIcon: '/assets/tabbar/pledge-active.svg'
  },
  {
    path: '/community',
    name: 'tab.community',
    icon: '/assets/tabbar/community.svg',
    activeIcon: '/assets/tabbar/community-active.svg'
  },
  {
    path: '/wallet',
    name: 'tab.myAssets',
    icon: '/assets/tabbar/mine.svg',
    activeIcon: '/assets/tabbar/mine-active.svg'
  }
]

const route = useRoute()
const router = useRouter()

const isActive = (path: string) => {
  return route.path === path
}

const handleTabClick = (path: string) => {
  if (route.path !== path) {
    router.push(path)
  }
}
</script>

<style lang="scss" scoped>
@use '@/style/variables.scss' as *;

.tabbar {
  position: fixed;
  bottom: 0;
  left: 50%;
  transform: translateX(-50%);
  max-width: 414px;
  width: 100%;
  display: flex;
  justify-content: space-around;
  align-items: center;
  background: $bg-main;
  border-top: 1px solid rgba(255, 255, 255, 0.08);
  padding: 8px 0 calc(8px + env(safe-area-inset-bottom));
  z-index: 100;
}

.tab-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
  cursor: pointer;
  opacity: 0.6;

  &.active {
    opacity: 1;
  }

  img {
    width: 24px;
    height: 24px;
  }

  .tab-label {
    font-size: 11px;
    color: #fff;
  }
}
</style>
