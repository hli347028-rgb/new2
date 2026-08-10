<template>
  <Teleport to="body">
    <Transition name="modal-fade">
      <div v-if="visible" class="modal-overlay" @click="handleOverlayClick">
        <div class="modal" @click.stop>
          <div class="modal-body">
            <p class="modal-message">{{ message }}</p>
            <div class="input-wrapper">
              <input
                v-model="inputValue"
                type="text"
                :placeholder="placeholder"
                class="modal-input"
              />
            </div>
          </div>
          <div class="modal-footer">
            <button class="modal-btn confirm-btn" @click="handleConfirm">{{ confirmText }}</button>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'

const props = withDefaults(
  defineProps<{
    visible: boolean
    message?: string
    placeholder?: string
    confirmText?: string
    closeOnOverlay?: boolean
  }>(),
  {
    message: '',
    placeholder: '',
    confirmText: '',
    closeOnOverlay: true
  }
)

const emit = defineEmits<{
  close: []
  confirm: [value: string]
}>()

const inputValue = ref('')

watch(
  () => props.visible,
  (newVal) => {
    if (newVal) {
      inputValue.value = ''
    }
  }
)

const close = () => {
  emit('close')
}

const handleConfirm = () => {
  emit('confirm', inputValue.value)
  close()
}

const handleOverlayClick = () => {
  if (props.closeOnOverlay) {
    close()
  }
}
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
  max-width: 360px;
  backdrop-filter: blur(10px);
  animation: modalSlideIn 0.3s ease;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px 24px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);

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
    font-size: 14px;
    color: rgba(255, 255, 255, 0.8);
    line-height: 1.6;
    margin: 0 0 16px 0;
  }

  .input-wrapper {
    .modal-input {
      width: 100%;
      padding: 12px 16px;
      background: rgba(255, 255, 255, 0.05);
      border: 1px solid rgba(255, 255, 255, 0.1);
      border-radius: 12px;
      color: #fff;
      font-size: 14px;
      outline: none;
      transition: all 0.3s ease;

      &::placeholder {
        color: rgba(255, 255, 255, 0.4);
      }

      &:focus {
        border-color: $brand-gold;
        background: rgba(255, 255, 255, 0.08);
      }
    }
  }
}

.modal-footer {
  padding: 16px 24px 24px;
  display: flex;
  gap: 12px;

  .modal-btn {
    flex: 1;
    padding: 12px 0;
    border: none;
    border-radius: 12px;
    font-size: 16px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.3s ease;
  }

  .cancel-btn {
    background: rgba(255, 255, 255, 0.1);
    color: #fff;

    &:hover {
      background: rgba(255, 255, 255, 0.15);
    }
  }

  .confirm-btn {
    background: linear-gradient(135deg, $brand-gold 0%, #d4a837 100%);
    color: #000;

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
