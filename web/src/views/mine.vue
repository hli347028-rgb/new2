<template>
  <div class="mine-page">
    <Header />
    <div class="container">
      <div class="airdrop-card">
        <div class="card-bg"></div>
        <div class="card-content">
          <div class="card-header">
            <img src="/assets/logo.png" alt="logo" class="card-logo" />
            <h2 class="card-title">{{ $t('mine.tokenAirdrop') }}</h2>
          </div>

          <div class="stats-list">
            <div class="stats-row">
              <span class="stats-name">{{ $t('mine.claimableAmount') }}</span>
              <span class="stats-value green">0</span>
            </div>
            <div class="stats-row">
              <span class="stats-name">{{ $t('mine.pendingAmount') }}</span>
              <span class="stats-value accent">0</span>
            </div>
            <div class="stats-row">
              <span class="stats-name">{{ $t('mine.claimedAmount') }}</span>
              <span class="stats-value green">0</span>
            </div>
            <div class="stats-row">
              <span class="stats-name">{{ $t('mine.totalNodes') }}</span>
              <span class="stats-value accent">0</span>
            </div>
            <div class="stats-row">
              <span class="stats-name">{{ $t('mine.releaseCountdown') }}</span>
              <span class="stats-value purple">--</span>
            </div>
          </div>

          <button class="claim-btn">{{ $t('mine.claimAirdrop') }}</button>
        </div>
        <div class="deco-circle top"></div>
        <div class="deco-circle bottom"></div>
      </div>

      <div class="record-section">
        <div class="section-title-wrap">
          <div class="title-bar"></div>
          <h3 class="section-title">{{ $t('mine.claimRecord') }}</h3>
        </div>

        <div class="table-card">
          <div class="table-header">
            <span>{{ $t('mine.time') }}</span>
            <span>{{ $t('mine.amount') }}</span>
          </div>
          <div class="empty-state">
            <p>{{ $t('common.noData') }}</p>
          </div>
        </div>
      </div>

      <div class="safe-bottom"></div>
    </div>
  </div>
</template>

<script setup lang="ts">
import Header from '@/components/Header.vue'
</script>

<style lang="scss" scoped>
@use '@/style/variables.scss' as *;

.mine-page {
  min-height: 100vh;
  padding-top: 64px;
}

.container {
  padding: 0 15px;
}

.airdrop-card {
  position: relative;
  margin-top: 25px;
  overflow: hidden;
  border-radius: 12px;

  .card-bg {
    position: absolute;
    inset: 0;
    border-radius: 12px;
    background: rgba(8, 19, 30, 0.6);
    backdrop-filter: blur(10px);
  }

  .card-content {
    position: relative;
    z-index: 10;
    padding: 20px;
  }

  .card-header {
    display: flex;
    align-items: center;
    margin-bottom: 15px;

    .card-logo {
      width: 25px;
      height: 25px;
      border-radius: 10px;
    }

    .card-title {
      margin: 0 0 0 8px;
      font-size: 16px;
      font-weight: bold;
      color: #fff;
    }
  }

  .stats-list {
    margin-bottom: 10px;

    .stats-row {
      display: flex;
      align-items: center;
      justify-content: space-between;
      margin-bottom: 12px;
      padding: 10px;
      border-radius: 8px;
      background: rgba(255, 255, 255, 0.05);

      &:last-child {
        margin-bottom: 0;
      }

      .stats-name {
        font-size: 14px;
        color: #fff;
      }

      .stats-value {
        font-size: 16px;
        font-weight: bold;

        &.green {
          color: $brand-primary;
        }

        &.accent {
          color: $brand-primary-light;
        }

        &.purple {
          color: $brand-primary;
        }
      }
    }
  }

  .claim-btn {
    width: 100%;
    margin-top: 10px;
    padding: 11px 0;
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

  .deco-circle {
    position: absolute;
    border-radius: 50%;
    background: rgba(255, 255, 255, 0.05);

    &.top {
      top: 0;
      right: 0;
      width: 100px;
      height: 100px;
      transform: translate(50%, -50%);
    }

    &.bottom {
      bottom: 0;
      left: 0;
      width: 75px;
      height: 75px;
      transform: translate(-50%, 50%);
      background: rgba(255, 255, 255, 0.03);
    }
  }
}

.record-section {
  margin-top: 20px;

  .section-title-wrap {
    position: relative;
    margin-bottom: 10px;
    margin-left: 10px;
    display: flex;
    align-items: center;

    .title-bar {
      position: absolute;
      left: -10px;
      top: 50%;
      width: 4px;
      height: 16px;
      border-radius: 2px;
      background: linear-gradient(180deg, #1597E5 0%, #075FB8 100%);
      transform: translateY(-50%);
    }

    .section-title {
      margin: 0 0 0 8px;
      font-size: 16px;
      font-weight: bold;
      color: #fff;
    }
  }
}

.table-card {
  margin-top: 10px;
  min-height: 300px;
  overflow: hidden;
  border: 1px solid $border-color;
  border-radius: 11px;
  background: rgba(8, 19, 30, 0.6);
  backdrop-filter: blur(10px);
  padding: 11px 0;

  .table-header {
    display: flex;
    align-items: center;
    background: #030A11;
    padding: 8px 0;
    margin: -11px 0 0;

    span {
      flex: 1;
      text-align: center;
      font-size: 10px;
      color: $text-muted;
    }
  }

  .empty-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    height: 250px;

    p {
      margin-top: 8px;
      font-size: 12px;
      color: $text-muted;
    }
  }
}

.safe-bottom {
  height: 50px;
}
</style>
