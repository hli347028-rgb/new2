<template>
  <Teleport to="body">
    <div v-if="visible" class="announcement-overlay" @click="handleClose">
      <div class="announcement-modal" @click.stop>
        <div class="modal-close" @click="handleClose">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M18 6L6 18M6 6l12 12" />
          </svg>
        </div>

        <h1 class="modal-title">{{ $t('announcement.title') }}</h1>

        <div class="modal-content">
         <p>{{ $t('announcement.content') }}</p>
        </div>

        <div class="modal-footer">
          <button class="confirm-btn" @click="handleClose">{{ $t('announcement.gotIt') }}</button>
          <p class="no-remind" @click="handleNoRemind">{{ $t('announcement.noRemind') }}</p>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'

const visible = ref(false)
const STORAGE_KEY = 'announcement_dismissed'

onMounted(() => {
  const dismissed = localStorage.getItem(STORAGE_KEY)
  if (!dismissed) {
    visible.value = true
  }
})

watch(() => visible.value, (val) => {
  document.body.style.overflow = val ? 'hidden' : ''
})

const handleClose = () => {
  visible.value = false
}

const handleNoRemind = () => {
  localStorage.setItem(STORAGE_KEY, 'true')
  visible.value = false
}
</script>

<style lang="scss" scoped>
@use '@/style/variables.scss' as *;

.announcement-overlay {
  position: fixed;
  inset: 0;
  z-index: 2000;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.8);
  padding: 20px;
  backdrop-filter: blur(8px);
}

.announcement-modal {
  position: relative;
  width: 90%;
  max-width: 380px;
  max-height: 70vh;
  background: linear-gradient(145deg, rgba(20, 20, 20, 0.98) 0%, rgba(30, 30, 30, 0.95) 100%);
  border-radius: 20px;
  padding: 28px 24px;
  display: flex;
  flex-direction: column;
  border: 1px solid rgba(212, 175, 55, 0.3);
  box-shadow: 0 0 24px rgba(212, 175, 55, 0.2), 0 8px 32px rgba(0, 0, 0, 0.6);
  animation: modalSlideIn 0.3s ease;

  .modal-close {
    position: absolute;
    top: 16px;
    right: 16px;
    width: 28px;
    height: 28px;
    display: flex;
    align-items: center;
    justify-content: center;
    background: rgba(255, 255, 255, 0.05);
    border-radius: 50%;
    color: rgba(255, 255, 255, 0.6);
    cursor: pointer;
    transition: all 0.2s ease;

    &:hover {
      background: rgba(255, 255, 255, 0.1);
      color: #fff;
    }
  }

  .modal-title {
    font-size: 20px;
    font-weight: 700;
    color: #fff;
    text-align: center;
    margin: 0 0 20px;
    letter-spacing: 0.5px;
    background: linear-gradient(135deg, #fff 0%, $brand-gold 100%);
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
    background-clip: text;
  }

  .modal-subtitle {
    font-size: 14px;
    font-weight: normal;
    color: $brand-gold;
    text-align: center;
    margin: 0 0 16px;
  }

  .modal-content {
    flex: 1;
    overflow-y: auto;
    max-height: 50vh;
    padding: 0 4px;
    margin-bottom: 20px;

    &::-webkit-scrollbar {
      width: 4px;
    }

    &::-webkit-scrollbar-thumb {
      background: rgba(212, 175, 55, 0.3);
      border-radius: 2px;
    }

    p {
      margin: 0 0 12px;
      font-size: 14px;
      color: rgba(255, 255, 255, 0.85);
      line-height: 1.7;
      white-space: pre-wrap;
      text-align: center;
    }

    strong {
      color: $brand-gold;
      font-weight: 600;
    }
  }

  .modal-footer {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 12px;

    .confirm-btn {
      width: 100%;
      padding: 14px 0;
      background: linear-gradient(135deg, $brand-gold 0%, #d4a837 100%);
      color: #000;
      font-size: 16px;
      font-weight: 600;
      letter-spacing: 0.5px;
      border: none;
      border-radius: 12px;
      cursor: pointer;
      transition: all 0.3s ease;
      box-shadow: 0 4px 16px rgba(212, 175, 55, 0.3);

      &:hover {
        transform: translateY(-2px);
        box-shadow: 0 6px 24px rgba(212, 175, 55, 0.5);
      }

      &:active {
        transform: translateY(0);
      }
    }

    .no-remind {
      margin: 0;
      font-size: 13px;
      color: rgba(255, 255, 255, 0.5);
      cursor: pointer;
      transition: color 0.2s ease;

      &:hover {
        color: rgba(255, 255, 255, 0.8);
      }
    }
  }
}

@keyframes modalSlideIn {
  from {
    transform: translateY(-20px) scale(0.95);
    opacity: 0;
  }
  to {
    transform: translateY(0) scale(1);
    opacity: 1;
  }
}
</style>
