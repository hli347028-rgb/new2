<template>
<div class='page'>
  <van-nav-bar
    :title="$t('tab.hashrateMall')"
    left-arrow
    :border="false"
    fixed
    @click-left="handleBack"
  />
  <div class="page-main">
  </div>
</div>
</template>
<script setup>
import userPerson from "@/pinia/person";
import { useRouter } from 'vue-router'
import lang from '@/i18n/index'
import { useI18n } from 'vue-i18n'

const router = useRouter()
const person = userPerson();
const { t: $t } = useI18n()

const handleBack = () => {
  router.back()
}

</script>
<style lang='less' scoped>
  .page {
    min-height: 100vh;
    padding-top: 50px;
  }
</style>
