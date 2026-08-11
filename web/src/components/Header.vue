<template>
  <header class="app-header">
    <div class="header-container">
      <div class="header-left" @click="handleSidebarClick">
        <img src="/assets/logo.jpg" alt="Logo" class="logo" />
      </div>
      <div class="header-right">
        <button class="connect-btn" @click="handleWalletClick">
          {{ address ? formatAddress(address) : $t('common.connectWallet') }}
        </button>
        <div class="language-selector" @click="handleLanguageClick">
          <img src="/assets/language.png" alt="Language" class="language-icon" />
        </div>
      </div>
    </div>
  </header>

  <Teleport to="body">
    <div v-if="showLangDrawer" class="lang-drawer-overlay" @click="showLangDrawer = false">
      <div class="lang-drawer" @click.stop>
        <div class="lang-drawer-title">{{ $t('common.selectLanguage') }}</div>
        <div class="lang-list">
          <div
            v-for="lang in languages"
            :key="lang.code"
            class="lang-item"
            :class="{ active: currentLanguage === lang.code }"
            @click="selectLanguage(lang.code)"
          >
            {{ lang.name }}
          </div>
        </div>
      </div>
    </div>

    <Sidebar
      :visible="showSidebar"
      @close="showSidebar = false"
      @language-click="handleLanguageClick"
    />
  </Teleport>
</template>

<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { Teleport } from 'vue'
import { useI18n } from 'vue-i18n'
import userPerson from '@/pinia/person'
import Sidebar from '@/components/Sidebar.vue'

const person = userPerson()
const { locale } = useI18n()

const currentLanguage = ref(locale.value)
const showLangDrawer = ref(false)
const showSidebar = ref(false)

const address = computed(() => person.address)

const formatAddress = (value: string) => {
  const frontSix = value.slice(0, 6)
  const backSix = value.slice(-6)
  const middle = '...'
  return frontSix + middle + backSix
}

const languages = [
  { code: 'zh', name: '简体中文' },
  { code: 'zh-tw', name: '繁體中文' },
  { code: 'en', name: 'English' },
  { code: 'ja', name: '日本語' },
  { code: 'ko', name: '한국어' },
  { code: 'vi', name: 'Tiếng Việt' }
]

watch(locale, (val) => {
  currentLanguage.value = val
})

const handleSidebarClick = () => {
  showSidebar.value = true
}

const handleLanguageClick = () => {
  showLangDrawer.value = true
}

const selectLanguage = (code: string) => {
  locale.value = code
  localStorage.setItem('lan', code)
  showLangDrawer.value = false
}

const handleWalletClick = () => {
  // if (walletStore.isConnected) {
  //   walletStore.disconnect()
  // } else {
  //   walletStore.connect()
  // }
  console.log('Wallet click - hidden')
}
</script>

<style lang="scss" scoped>
@use '@/style/variables.scss' as *;

.app-header {
  position: fixed;
  top: 0;
  left: 50%;
  transform: translateX(-50%);
  max-width: 414px;
  width: 100%;
  z-index: 999;
  background: rgba(5, 5, 5, 0.95);
  backdrop-filter: blur(10px);
  border-bottom: 1px solid $border-color;

  .header-container {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 4px 20px;
    max-width: 1200px;
    margin: 0 auto;

    .header-left {
      display: flex;
      align-items: center;
      gap: $spacing-sm;

      .logo {
        height: 44px;
        width: auto;
        object-fit: contain;
      }

      span {
        font-size: 18px;
        font-weight: bold;
        color: #fff;
        letter-spacing: -0.5px;
      }
    }

    .header-right {
      display: flex;
      align-items: center;
      gap: 12px;

      .connect-btn {
        padding: 8px 16px;
        background-color: rgba(18, 40, 58, var(--un-bg-opacity, 1));
        color: rgba(143, 223, 255, var(--un-text-opacity, 1));
        border: 1px solid #88888844;
        border-radius: $radius-sm;
        font-size: 12px;
        font-weight: 500;
        cursor: pointer;
        transition: all $transition-fast;
      }

      .language-selector {
        display: flex;
        align-items: center;
        gap: 4px;
        cursor: pointer;

        .language-icon {
          width: 24px;
          height: 24px;
          object-fit: contain;
        }
      }
    }
  }
}

.lang-drawer-overlay {
  position: fixed;
  inset: 0;
  z-index: 1000;
  background: rgba(0, 0, 0, 0.5);

  .lang-drawer {
    position: absolute;
    top: 0;
    right: 0;
    width: 70%;
    max-width: 280px;
    height: 100%;
    background: $bg-card;
    padding: 20px;

    .lang-drawer-title {
      font-size: 16px;
      font-weight: bold;
      color: #fff;
      text-align: center;
      margin-bottom: 30px;
      padding-top: 10px;
    }

    .lang-list {
      .lang-item {
        padding: 14px 0;
        font-size: 14px;
        color: #ccc;
        border-bottom: 1px solid rgba(255, 255, 255, 0.06);
        cursor: pointer;
        transition: color 0.2s ease;

        &.active {
          color: $brand-primary;
        }

        &:hover {
          color: #fff;
        }
      }
    }
  }
}
</style>
