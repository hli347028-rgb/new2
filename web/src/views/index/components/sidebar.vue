<template>
  <ul class="sidebar">
    <li @click="go('/')"><i class="nav-icon-10" />{{ lang('首页') }}</li>
    <li @click="go('/recharge')"><i class="nav-icon-10" />{{ lang('充值') }}</li>
    <li @click="go('/node')"><i class="nav-icon-2" />{{ lang('认购') }}</li>
    <li @click="go('/community')"><i class="nav-icon-7" />{{ lang('社群建设') }}</li>
    <li @click="go('/wallet')"><i class="nav-icon-5" />{{ lang('我的资产') }}</li>
    <li @click="go('/withdrawal')"><i class="nav-icon-11" />{{ lang('提现') }}</li>
    <li @click="go('/count')"><i class="nav-icon-4" />{{ lang('community.rewardRecord') }}</li>
  </ul>
</template>
<script setup lang="ts">
import { useRouter } from 'vue-router'
import lang from '@/i18n/index'

const router = useRouter()
const go = (path: string) => {
  router.push(path)
}
</script>
