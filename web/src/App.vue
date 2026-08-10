<template>
  <a-config-provider :theme="theme">
    <div id="app">
      <router-view v-if="person.isLogin" />
      <Login v-else />
    </div>
  </a-config-provider>
</template>

<script setup lang="ts">
import userPerson from '@/pinia/person'
import { theme as antdTheme } from 'ant-design-vue'
import { onMounted, nextTick } from 'vue'

const theme = { algorithm: antdTheme.darkAlgorithm }
const person = userPerson()

onMounted(async () => {
  await nextTick()
  person.init()
})
</script>

<style>
#app {
  width: 100%;
  min-height: 100%;
  max-width: 414px;
  margin: 0 auto;
  background: #050505;
}
</style>
