<template>
  <nav class="tabbar">
    <div class="item" :class="{ active: path === '/' }" @click="$router.push('/')">首页</div>
    <div class="item" :class="{ active: path === '/order' }" @click="$router.push('/order')">报单</div>
    <div class="item" :class="{ active: path === '/mine' }" @click="$router.push('/mine')">资产</div>
    <div class="item" :class="{ active: path === '/team' }" @click="$router.push('/team')">团队</div>
  </nav>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
const route = useRoute()
const path = computed(() => route.path)
</script>

<style scoped lang="scss">
@use '@/style/variables.scss' as *;

.tabbar {
  position: fixed;
  bottom: 0;
  left: 50%;
  transform: translateX(-50%);
  width: 100%;
  max-width: 414px;
  height: 56px;
  display: flex;
  background: rgba(18, 18, 18, 0.96);
  border-top: 1px solid $border-color;
  z-index: 100;
}

.item {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  color: rgba(255, 255, 255, 0.55);
  font-size: 13px;

  &.active {
    color: $brand-gold;
    font-weight: 600;
  }
}
</style>
