<template>
  <div class="pledge-page">
    <Header />
    <div class="container">
      <img src="/static/pledge/hero.png" :alt="$t('pledge.pledge')" class="hero-image" />

      <div class="number-container">
        <!-- <template v-for="(item, index) in digitLayout" :key="index">
          <div v-if="item.type === 'digit'" class="digit-wrapper">
            <div class="digit-column" :style="{ transform: `translateY(-${item.value * 10}%)` }">
              <div v-for="num in digitNumbers" :key="num" class="digit-item">{{ num }}</div>
            </div>
          </div>
          <div v-else-if="item.type === 'decimal'" class="digit-wrapper decimal-point">
            <div class="decimal-item">{{ item.char }}</div>
          </div>
        </template> -->
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import Header from '@/components/Header.vue'

type DigitItem =
  | { type: 'digit'; value: number }
  | { type: 'decimal'; char: string }

const modalVisible = ref(false)
const amount = ref('')
const selectedCycle = ref(30)

const digitNumbers = [0, 1, 2, 3, 4, 5, 6, 7, 8, 9]

// 灵活配置：自由调整位数和小数点位置
// 当前显示 0.000000（1位整数 + 小数点 + 6位小数）
const digitLayout = ref<DigitItem[]>([
  { type: 'digit', value: 0 },
  { type: 'decimal', char: '.' },
  { type: 'digit', value: 0 },
  { type: 'digit', value: 0 },
  { type: 'digit', value: 0 },
  { type: 'digit', value: 0 },
  { type: 'digit', value: 0 },
  { type: 'digit', value: 0 }
])

const cycles = [
  { days: 30, rate: '1.3' },
  { days: 15, rate: '0.7' },
  { days: 7, rate: '0.3' }
]

const handleSubmit = () => {
  console.log('质押', { amount: amount.value, cycle: selectedCycle.value })
}
</script>

<style lang="scss" scoped>
@use '@/style/variables.scss' as *;

.pledge-page {
  min-height: 100vh;
  padding-top: 64px;
}

.container {
  padding: 0 20px;
  display: flex;
  flex-direction: column;
  align-items: center;

  .hero-image {
    width: 100%;
    max-width: 345px;
    height: auto;
    margin-top: 18px;
  }

  .number-container {
    display: flex;
    align-items: center;
    justify-content: center;
    margin-top: 22px;

    .digit-wrapper {
      height: 48px;
      overflow: hidden;

      &.decimal-point {
        display: flex;
        align-items: center;
        justify-content: center;
        height: 48px;
      }

      .digit-column {
        transition: transform 0.8s ease-out;

        .digit-item {
          height: 48px;
          line-height: 48px;
          font-size: 40px;
          font-weight: bold;
          color: #fff;
          text-align: center;
        }
      }

      .decimal-item {
        height: 48px;
        line-height: 48px;
        font-size: 40px;
        font-weight: bold;
        color: #fff;
        text-align: center;
      }
    }
  }

  .asset-label {
    margin-top: 4px;
    font-size: 12px;
    color: #fff;
    text-align: center;
  }

  .mining-btn {
    width: 100%;
    margin-top: 16px;
    padding: 12px 0;
    background: transparent;
    color: $brand-primary;
    font-size: 16px;
    font-weight: bold;
    border: 1px solid $brand-primary;
    border-radius: 24px;
    cursor: pointer;
    text-align: center;
    transition: all 0.3s ease;

    &:hover {
      background: $gradient-primary;
      color: $text-inverse;
    }

    &:active {
      transform: scale(0.95);
    }
  }
}

.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.6);
  display: flex;
  align-items: flex-end;
  justify-content: center;
  z-index: 1000;
}

.modal-content {
  width: 100%;
  max-width: 414px;
  background: rgba(8, 19, 30, 0.8);
  backdrop-filter: blur(10px);
  border-radius: 32px 32px 0 0;
  padding: 20px 24px 32px;
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 20px;

  .modal-title {
    font-size: 16px;
    font-weight: bold;
    color: #fff;
  }

  .modal-close {
    font-size: 24px;
    color: #fff;
    cursor: pointer;
    line-height: 1;
  }
}

.modal-body {
  .section-label {
    font-size: 13px;
    color: #fff;
    margin: 0 0 16px 0;
  }

  .cycle-list {
    display: flex;
    flex-direction: column;
    gap: 12px;
    margin-bottom: 20px;

    .cycle-card {
      display: flex;
      flex-direction: column;
      align-items: center;
      justify-content: center;
      min-height: 70px;
      padding: 10px 5px 15px;
      border: 0.5px solid rgba(21, 151, 229, 0.2);
      border-radius: 10px;
      background: $bg-card;
      cursor: pointer;
      transition: all 0.3s ease;

      &.active {
        background: $gradient-primary;
        border-color: transparent;
        transform: scale(1.05);

        .cycle-header span,
        .cycle-rate .rate-num,
        .cycle-rate .rate-label {
          color: $text-inverse;
        }
      }

      .cycle-header {
        display: flex;
        align-items: center;
        gap: 3px;
        font-size: 12px;
        color: $text-muted;
        margin-bottom: 5px;

        .cycle-icon {
          width: 12px;
          height: 12px;
          flex-shrink: 0;
        }
      }

      .cycle-rate {
        display: flex;
        flex-direction: column;
        align-items: center;
        margin-top: 5px;

        .rate-num {
          font-size: 23px;
          font-weight: bold;
          color: $text-muted;
        }

        .rate-label {
          font-size: 10px;
          color: $text-muted;
          margin-top: 2px;
        }
      }
    }
  }

  .balance-text {
    font-size: 13px;
    color: #fff;
    margin: 0 0 12px 0;
    font-weight: bold;
  }

  .input-row {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 8px;

    .input-wrap {
      flex: 1;
      background: #030A11;
      border: 0.5px solid rgba(21, 151, 229, 0.15);
      border-radius: 8px;
      padding: 10px 12px;

      .amount-input {
        width: 100%;
        background: transparent;
        border: none;
        color: #fff;
        font-size: 14px;
        outline: none;

        &::placeholder {
          color: $text-muted;
        }
      }
    }

    .max-btn {
      width: 54px;
      height: 41px;
      background: transparent;
      color: $brand-primary;
      font-size: 13px;
      font-weight: bold;
      border: 1px solid $brand-primary;
      border-radius: 10px;
      cursor: pointer;
      transition: all 0.3s ease;

      &:hover {
        background: $gradient-primary;
        color: $text-inverse;
      }
    }
  }

  .limit-row {
    display: flex;
    justify-content: space-between;
    font-size: 11px;
    color: #fff;
    margin-bottom: 20px;
  }

  .profit-row {
    display: flex;
    justify-content: space-between;
    font-size: 14px;
    color: #fff;
    margin-bottom: 20px;
  }

  .highlight {
    color: $brand-primary;
    font-weight: bold;
  }

  .submit-btn {
    width: 100%;
    padding: 12px 0;
    background: transparent;
    color: $brand-primary;
    font-size: 14px;
    font-weight: bold;
    border: 1px solid $brand-primary;
    border-radius: 24px;
    cursor: pointer;
    text-align: center;
    transition: all 0.3s ease;

    &:hover {
      background: $gradient-primary;
      color: $text-inverse;
    }

    &:active {
      transform: scale(0.95);
    }
  }
}
</style>
