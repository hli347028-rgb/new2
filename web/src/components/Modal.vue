<template>
  <Transition name="modal-fade">
    <div v-if="visible" class="modal-overlay" @click="handleOverlayClick">
      <div class="modal" @click.stop>
        <div class="modal-header">
          <!-- <h3 class="modal-title">{{ title }}</h3> -->
          <button class="modal-close" @click="close">×</button>
        </div>
        <div class="modal-body">
          <p class="modal-message">{{ message }}</p>
        </div>
        <div class="modal-footer" v-if="showConfirmButton">
          <button class="modal-btn" @click="handleConfirm">{{ confirmText }}</button>
        </div>
      </div>
    </div>
  </Transition>
</template>

<script setup lang="ts">
import { watch } from 'vue'

const props = withDefaults(
  defineProps<{
    visible: boolean
    title?: string
    message?: string
    confirmText?: string
    closeOnOverlay?: boolean
    showConfirmButton?: boolean
    autoCloseDelay?: number
  }>(),
  {
    title: '',
    message: '',
    confirmText: '',
    closeOnOverlay: true,
    showConfirmButton: true,
    autoCloseDelay: 0
  }
)

const emit = defineEmits<{
  close: []
  confirm: []
}>()

const close = () => {
  emit('close')
}

const handleConfirm = () => {
  emit('confirm')
  close()
}

const handleOverlayClick = () => {
  if (props.closeOnOverlay) {
    close()
  }
}

watch(() => props.visible, (newVal) => {
  if (newVal && props.autoCloseDelay > 0) {
    setTimeout(() => {
      close()
    }, props.autoCloseDelay)
  }
})
</script>

<style lang="scss" scoped>
@use '@/style/variables.scss' as *;

.modal-overlay {
  position: fixed;
  inset: 0;
  z-index: 3000;
  background: rgba(0, 0, 0, 0.7);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px;
}

.modal {
  background: rgba(20, 20, 20, 0.95);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 20px;
  width: 100%;
  max-width: 320px;
  backdrop-filter: blur(10px);
  animation: modalSlideIn 0.3s ease;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  // padding: 20px 24px;
  // border-bottom: 1px solid rgba(255, 255, 255, 0.1);

  .modal-title {
    font-size: 18px;
    font-weight: 600;
    color: #fff;
    margin: 0;
  }

  .modal-close {
    width: 32px;
    height: 32px;
    background: transparent;
    border: none;
    color: rgba(255, 255, 255, 0.6);
    font-size: 24px;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: color 0.2s;

    &:hover {
      color: #fff;
    }
  }
}

.modal-body {
  padding: 24px;

  .modal-message {
    font-size: 15px;
    color: rgba(255, 255, 255, 0.8);
    line-height: 1.6;
    margin: 0;
    text-align: center;
  }
}

.modal-footer {
  padding: 16px 24px 24px;
  display: flex;
  justify-content: center;

  .modal-btn {
    padding: 12px 32px;
    background: linear-gradient(135deg, $brand-gold 0%, #d4a837 100%);
    color: #000;
    border: none;
    border-radius: 24px;
    font-size: 16px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.3s ease;

    &:hover {
      transform: translateY(-2px);
      box-shadow: 0 8px 24px rgba(212, 175, 55, 0.3);
    }

    &:active {
      transform: translateY(0);
    }
  }
}

.modal-fade-enter-active,
.modal-fade-leave-active {
  transition: opacity 0.3s ease;
}

.modal-fade-enter-from,
.modal-fade-leave-to {
  opacity: 0;
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
