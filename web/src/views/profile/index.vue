<template>
  <div class="profile-page">
    <ChildrenHeader />
    <ul class="profile">
      <li>
        <label for="profile-language">{{ $t('common.languageSwitch') }}</label>
        <select id="profile-language" v-model="selectedLanguage" class="language-select" @change="onChangeLanguage">
          <option v-for="item in languageOptions" :key="item.code" :value="item.code">{{ item.name }}</option>
        </select>
      </li>
    </ul>
    <button class="switch-user-button" @click="switchUser">{{ $t('common.switchWallet') }}</button>
  </div>
</template>
<script setup lang="ts">
import ChildrenHeader from '../../components/header/childrenHeader.vue'
import { restartCurrentApp } from '@/tools/plaocRuntime'
import { useI18n } from 'vue-i18n'
import { ref } from 'vue'
const { t: $t, locale } = useI18n()

const selectedLanguage = ref(String(locale.value))
const languageOptions = [
  { code: 'zh', name: '简体中文' },
  { code: 'zh-tw', name: '繁體中文' },
  { code: 'en', name: 'English' },
  { code: 'ja', name: '日本語' },
  { code: 'ko', name: '한국어' },
  { code: 'vi', name: 'Tiếng Việt' },
]

const switchUser = async () => {
  localStorage.removeItem("token");
  localStorage.removeItem("account");
  await restartCurrentApp()
}

const onChangeLanguage = () => {
  locale.value = selectedLanguage.value
  localStorage.setItem('lan', selectedLanguage.value)
}

</script>
<style scoped lang="less">
@import "./styles/index.less";

.profile-page {
  width: 100%;
  min-height: 100vh;
}

.language-select {
  min-width: 132px;
  height: 34px;
  padding: 0 30px 0 10px;
  border: 1px solid rgba(143, 223, 255, 0.24);
  border-radius: 6px;
  background: #0d1c29;
  color: #fff;
}
</style>
