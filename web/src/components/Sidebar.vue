<template>
  <Teleport to="body">
    <div v-if="visible" class="sidebar-overlay" @click="handleOverlayClick">
      <div class="sidebar" @click.stop>
        <!-- 关闭按钮 -->
        <button class="close-btn" type="button" :aria-label="$t('common.close')" @click="close">×</button>

        <!-- Logo -->
        <div class="sidebar-logo">
          <img src="/assets/logo.png" alt="Logo" class="logo-img" />
          <div class="brand-copy">
            <strong>AIX</strong>
            <span>{{ $t('common.protocol') }}</span>
          </div>
        </div>

        <div class="user-address" @click="handleCopyAddress">
          <div class="address-value">{{ userAddress }}</div>
          <span class="copy-mark" aria-hidden="true"></span>
        </div>

        <!-- 菜单项：仅金牛主路径；未接能力不展示入口 -->
        <nav class="sidebar-nav">
          <div class="nav-item" :class="{ active: isActive('/') }" @click="go('/')">
            {{ $t('tab.home') }}
          </div>
          <div class="nav-item" :class="{ active: isActive('/recharge') }" @click="go('/recharge')">
            {{ $t('recharge.recharge') }}
          </div>
          <!-- <div class="nav-item" :class="{ active: isActive('/withdrawal') }" @click="go('/withdrawal')">
            {{ $t('tab.withdraw') }}
          </div> -->
          <div class="nav-item" :class="{ active: isActive('/node') }" @click="go('/node')">
            报单专区
          </div>
          <div class="nav-item" :class="{ active: isActive('/community') }" @click="go('/community')">
            {{ $t('tab.myTeam') }}
          </div>
          <div class="nav-item" :class="{ active: isActive('/wallet') }" @click="go('/wallet')">
            {{ $t('tab.myAssets') }}
          </div>
          <div class="nav-item external" @click="showToast($t('common.comingSoon'))">
            {{ $t('tab.globalLaunch') }}
          </div>
          <div class="nav-item upcoming" @click="showToast($t('common.comingSoon'))">
            {{ $t('tab.networkHashrate') }}
          </div>
          <div class="nav-item upcoming" @click="showToast($t('common.comingSoon'))">
            {{ $t('tab.walletDownload') }}
          </div>

        </nav>

      </div>
    </div>

    <Modal
      :visible="showModal"
      :message="modalMessage"
      :confirm-text="$t('common.confirm')"
      @close="handleModalClose"
      @confirm="handleModalClose"
    />
  </Teleport>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { showToast } from 'vant'
import userPerson from '@/pinia/person'
import Modal from '@/components/Modal.vue'

const router = useRouter()
const route = useRoute()
const person = userPerson()
const { t: $t } = useI18n()

const props = defineProps({
  visible: Boolean
})

const emit = defineEmits(['close', 'languageClick'])

const showModal = ref(false)
const modalMessage = ref('')

const address = computed(() => person.address)

const formatAddress = (value) => {
  const frontSix = value.slice(0, 6)
  const backFour = value.slice(-4)
  return frontSix + '...' + backFour
}

const userAddress = computed(() => {
  return formatAddress(address.value || '0x0000000000000000000000000000000000000000')
})

const isActive = (path) => {
  return route.path === path
}

const go = (path) => {
  if (route.path !== path) {
    router.push(path)
  }
  close()
}

const close = () => {
  emit('close')
}

const handleOverlayClick = () => {
  close()
}

const handleDownload = () => {
  const link = document.createElement('a')
  link.href = '/base.apk'
  link.download = 'base.apk'
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
}

const handleExternalLink = () => {
  router.push('/recharge')
  close()
}

const handleComingSoon = () => {
  showModal.value = true
  modalMessage.value = $t('common.comingSoon')
}

const handleModalClose = () => {
  showModal.value = false
}

const handleCopyAddress = async () => {
  const addr = userAddress.value
  if (!addr) return
  const copyText = `${window.location.origin}/?code=${addr}`
  try {
    await navigator.clipboard.writeText(copyText)
    showToast($t('common.inviteLinkCopied'))
  } catch (err) {
    console.error('复制失败:', err)
    showToast($t('common.copyFailed'))
  }
}
</script>

<style lang="scss" scoped>
@use '@/style/variables.scss' as *;

.sidebar-overlay {
  position: fixed;
  inset: 0;
  z-index: 2000;
  background: rgba(1, 5, 9, 0.72);
  backdrop-filter: blur(6px);
  animation: fadeIn 0.25s ease;
}

.sidebar {
  position: absolute;
  top: 0;
  left: max(0px, calc(50% - 207px));
  bottom: 0;
  width: min(88vw, 360px);
  box-sizing: border-box;
  padding: 0 22px 32px;
  background:
    radial-gradient(circle at 16% 5%, rgba(33, 182, 234, 0.16), transparent 30%),
    linear-gradient(165deg, rgba(13, 31, 47, 0.99) 0%, rgba(3, 10, 17, 0.99) 48%, rgba(2, 7, 12, 1) 100%);
  box-shadow: 18px 0 48px rgba(0, 0, 0, 0.48);
  animation: slideInLeft 0.28s cubic-bezier(0.22, 1, 0.36, 1);
  display: flex;
  flex-direction: column;
  align-items: center;
  overflow-y: auto;
  overscroll-behavior: contain;
  scrollbar-width: none;

  &::before {
    content: '';
    position: absolute;
    top: 0;
    left: 0;
    width: 2px;
    height: 34%;
    background: linear-gradient(180deg, $brand-accent, rgba(21, 151, 229, 0));
    box-shadow: 0 0 18px rgba(33, 182, 234, 0.7);
    pointer-events: none;
  }

  &::-webkit-scrollbar {
    display: none;
  }
}

.close-btn {
  position: sticky;
  top: 16px;
  z-index: 2;
  align-self: flex-end;
  flex: 0 0 auto;
  width: 36px;
  height: 36px;
  margin-right: -4px;
  background: rgba(143, 223, 255, 0.06);
  border: 1px solid rgba(143, 223, 255, 0.14);
  border-radius: 50%;
  color: $text-muted;
  font-size: 26px;
  font-weight: 300;
  line-height: 1;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all $transition-fast;

  &:hover {
    color: $text-primary;
    border-color: rgba(33, 182, 234, 0.45);
    background: rgba(21, 151, 229, 0.14);
    transform: rotate(8deg);
  }
}

.sidebar-logo {
  width: 100%;
  margin: 16px 0 22px;
  display: flex;
  flex-direction: column;
  align-items: center;

  .logo-img {
    width: 108px;
    height: 108px;
    object-fit: cover;
    border-radius: 28px;
    border: 1px solid rgba(143, 223, 255, 0.22);
    box-shadow: 0 14px 42px rgba(0, 0, 0, 0.42), 0 0 30px rgba(21, 151, 229, 0.12);
  }

  .brand-copy {
    margin-top: 12px;
    display: flex;
    align-items: baseline;
    gap: 7px;

    strong {
      color: $text-primary;
      font-size: 22px;
      line-height: 1;
      letter-spacing: 4px;
    }

    span {
      color: $brand-primary-light;
      font-size: 9px;
      letter-spacing: 2.4px;
      opacity: 0.72;
    }
  }
}

.user-address {
  position: relative;
  width: 100%;
  box-sizing: border-box;
  padding: 13px 46px 13px 16px;
  background: linear-gradient(135deg, rgba(21, 151, 229, 0.13), rgba(143, 223, 255, 0.04));
  border-radius: 14px;
  cursor: pointer;
  transition: all $transition-base;
  border: 1px solid rgba(21, 151, 229, 0.22);
  text-align: left;

  &:hover {
    background: linear-gradient(135deg, rgba(21, 151, 229, 0.2), rgba(143, 223, 255, 0.07));
    border-color: rgba(33, 182, 234, 0.5);
    box-shadow: 0 8px 28px rgba(0, 80, 145, 0.18);
  }

  &:active {
    transform: scale(0.98);
  }

  .address-label {
    margin-bottom: 5px;
    color: $text-muted;
    font-size: 10px;
    letter-spacing: 0.8px;
    text-transform: uppercase;
  }

  .address-value {
    color: $brand-primary-light;
    font-size: 14px;
    font-weight: 600;
    letter-spacing: 0.4px;
    word-break: break-all;
  }

  .copy-mark {
    position: absolute;
    top: 50%;
    right: 16px;
    width: 14px;
    height: 14px;
    border: 1px solid rgba(143, 223, 255, 0.72);
    border-radius: 3px;
    transform: translateY(-35%);

    &::before {
      content: '';
      position: absolute;
      width: 10px;
      height: 10px;
      top: -5px;
      left: -5px;
      border: 1px solid rgba(143, 223, 255, 0.42);
      border-radius: 3px;
    }
  }
}

.sidebar-nav {
  width: 100%;
  margin-top: 22px;
  margin-bottom: 28px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.nav-item {
  position: relative;
  min-height: 46px;
  box-sizing: border-box;
  padding: 13px 42px 13px 20px;
  background: rgba(143, 223, 255, 0.035);
  border-radius: 13px;
  color: rgba(244, 250, 255, 0.78);
  font-size: 14px;
  font-weight: 500;
  text-align: left;
  cursor: pointer;
  transition: all $transition-fast;
  border: 1px solid rgba(143, 223, 255, 0.06);

  &::before {
    content: '';
    position: absolute;
    top: 50%;
    right: 20px;
    width: 5px;
    height: 5px;
    border-top: 1px solid currentColor;
    border-right: 1px solid currentColor;
    opacity: 0.42;
    transform: translateY(-50%) rotate(45deg);
  }

  &:hover {
    color: $text-primary;
    background: rgba(21, 151, 229, 0.1);
    border-color: rgba(21, 151, 229, 0.2);
    transform: translateX(3px);
  }

  &.active {
    color: $text-primary;
    background: linear-gradient(90deg, rgba(21, 151, 229, 0.3), rgba(33, 182, 234, 0.08));
    border-color: rgba(33, 182, 234, 0.48);
    box-shadow: inset 3px 0 0 $brand-accent, 0 8px 24px rgba(0, 91, 156, 0.16);

    &::before {
      color: $brand-primary-light;
      opacity: 1;
    }
  }

  &.external::after {
    content: '↗';
    position: absolute;
    right: 18px;
    top: 50%;
    color: $brand-primary-light;
    font-size: 13px;
    opacity: 0.62;
    transform: translateY(-50%);
  }

  &.external::before {
    display: none;
  }

  &.upcoming {
    color: rgba(159, 179, 200, 0.62);

    &::before {
      width: 4px;
      height: 4px;
      border: 0;
      border-radius: 50%;
      background: $brand-primary;
      box-shadow: 0 0 8px rgba(33, 182, 234, 0.7);
      transform: translateY(-50%);
    }
  }

  &.disabled {
    opacity: 0.5;
    cursor: not-allowed;

    &:hover {
      background: rgba(255, 255, 255, 0.05);
    }
  }
}

@keyframes fadeIn {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
}

@keyframes slideInLeft {
  from {
    transform: translateX(-100%);
  }
  to {
    transform: translateX(0);
  }
}

@media (max-width: 374px) {
  .sidebar {
    width: 92vw;
    padding-inline: 16px;
  }

  .sidebar-logo {
    margin-top: 8px;

    .logo-img {
      width: 92px;
      height: 92px;
    }
  }
}
</style>
