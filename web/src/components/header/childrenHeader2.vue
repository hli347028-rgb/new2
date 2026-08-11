<template>
  <div class="children-header">
    <div class="children-header-main">
      <div class="children-header-left">
        <i class="return-button" @click="handleMenuChange ? handleMenuChange(1) : router.back()">{{ $t('common.back') }}</i>
      </div>
      <div class="children-header-center">
        <i class="logo">
        </i>
      </div>
    </div>
  </div>
</template>
<script setup lang="ts">
import logo from '../../assets/images/logo.png'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { ref } from 'vue'

const router = useRouter()
const { t: $t } = useI18n()

const props = defineProps({
  handleMenuChange: Function
})

</script>
<style scoped lang="less">
@import "./childrenHeader.less";
</style>
