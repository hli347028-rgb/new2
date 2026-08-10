<template>
  <div class="index-page">
    <Header />

    <div class="content">
      <div class="welcome-card">
        <GoldWaveBackground />
        <div class="welcome-content">
          <img src="/assets/logo.png" alt="Logo" class="logo active-logo" />
          <div class="welcome-subtitle">Taurus Protocol</div>
          <div class="welcome-description">{{ $t('index.welcomeDescription2') }}</div>
        </div>
      </div>

      <button class="start-button" @click="handleStart">{{ $t('index.startNow') }}</button>

      <div class="advantage-section">
        <div class="ido-header">
          <img src="/static/node-project.png" alt="" class="ido-icon" />
          <h2 class="ido-title">{{ $t('index.idoRecruitPlan') }}</h2>
        </div>
        <!-- <div class="ido-subtitle">成为Taurus IDO，享受提现手续费分红、升级级别等功能</div> -->

        <div class="ido-card">
          <div class="ido-card-left">
            <span class="ido-label">{{ $t('index.idoRights') }}</span>
          </div>

          <div class="ido-card-right">
            <span class="ido-limit">{{ $t('index.limited', { count: 160 }) }}</span>
          </div>
        </div>

        <div class="ido-price">
          <span class="price-number">500～3000</span>
          <span class="price-unit">{{ $t('index.usdtPerShare') }}</span>
        </div>
        <p class="ido-limit-text">{{ $t('index.accountLimit', { count: 1 }) }}</p>

        <div class="project-info">
          <div class="info-section">
            <h4 class="section-title">{{ $t('index.projectInfoTitle1') }}</h4>
            <div class="info-content">
              <p class="info-item">{{ $t('index.totalSupply') }}</p>
              <!-- <p class="info-item">2. {{ $t('index.nodeOutput') }}</p>
              <p class="info-subitem">{{ $t('index.blockInterval') }}</p>
              <p class="info-subitem">{{ $t('index.blocksPerDay') }}</p> -->
            </div>
          </div>

          <div class="info-section">
            <h4 class="section-title">{{ $t('index.scenariosTitle') }}</h4>
            <div class="info-content">
              <p class="info-item">1. {{ $t('index.chainLaunch') }}</p>
              <p class="info-item">2. {{ $t('index.miningLaunch', { month: 7 }) }}</p>
              <p class="info-item">3. {{ $t('index.paymentLaunch') }}</p>
              <p class="info-item">4. {{ $t('index.gameBeta') }}</p>
              <p class="info-item">5. {{ $t('index.mallLaunch') }}</p>
            </div>
          </div>

          <div class="info-section">
            <h4 class="section-title">{{ $t('index.techTitle') }}</h4>
            <div class="info-content">
              <p class="info-item">{{ $t('index.techDesc') }}</p>
            </div>
          </div>
        </div>

        <button class="ido-btn" @click="handleGoToNode">{{ $t('index.becomeIdo') }}</button>
      </div>

      <div class="nft-section">
        <div class="nft-header">
          <img src="/static/icon-nft-benefit.png" alt="" class="nft-icon" />
          <h2 class="nft-title">{{ $t('index.idoRecruitPlan') }}</h2>
        </div>
        <img src="/static/pledge/hero-CvxrCMfG.png" :alt="$t('index.nftRights')" class="nft-image" />
        <p class="nft-hint"></p>
        <p class="nft-desc">{{ $t('index.idoSubtitle') }}</p>
        <div class="nft-progress">
          <div class="progress-bar">
            <span class="progress-num"></span>
          </div>
          <span class="progress-text">{{ $t('index.idoRights') }}</span>
        </div>
      </div>

      <StrengthSection />

      <h2 class="partners-title">{{ $t('index.partners') }}</h2>
      <PartnersWall />
    </div>
  </div>

</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import Header from '@/components/Header.vue'
import GoldWaveBackground from '@/components/GoldWaveBackground.vue'
import StrengthSection from '@/components/StrengthSection.vue'
import PartnersWall from '@/components/PartnersWall.vue'

const router = useRouter()

const handleStart = () => {
  router.push('/node')
}

const handleGoToNode = () => {
  router.push('/node')
}

</script>

<style lang="scss" scoped>
@use '@/style/variables.scss' as *;

.index-page {
  min-height: 100vh;
  position: relative;
  overflow-x: hidden;
  padding-top: 64px;
  background-color: #000;

  &::before {
    content: '';
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: url('/static/chart-icon.png') no-repeat center;
    background-size: cover;
    opacity: 0.1;
    pointer-events: none;
  }
}

.content {
  position: relative;
  z-index: 1;
}

.welcome-card {
  width: 100%;
  height: 400px;
  margin-bottom: 32px;
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  animation: slideUp 0.3s ease;

  &::after {
    content: '';
    position: absolute;
    bottom: 0;
    left: 0;
    right: 0;
    height: 40%;
    // background: linear-gradient(to bottom, transparent, $bg-main);
    pointer-events: none;
    z-index: 0;
  }

  .welcome-content {
    position: relative;
    z-index: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    text-align: center;
    mix-blend-mode: lighten;

    .logo {
      margin-top: 50px;
      width: 200px;
      height: auto;
    }

    .welcome-subtitle {
      font-size: 32px;
      font-weight: 800;
      background: linear-gradient(135deg, rgba(236, 208, 165, 0.8) 0%, #FFFFFF 50%, rgba(236, 208, 165, 0.8) 100%);
      background-size: 200% 100%;
      animation: gradient-move 1s linear infinite;
      -webkit-background-clip: text;
      -webkit-text-fill-color: transparent;
      background-clip: text;
    }

    .welcome-description {
      margin-top: 16px;
      padding: 8px 24px;
      font-size: 16px;
      font-weight: 500;
      color: rgba(236, 208, 165, 1);
      background: linear-gradient(90deg, transparent, rgba(212, 175, 55, 0.3), transparent);
      display: inline-block;
      line-height: 1.5;
    }
  }
}

.features {
  padding: 0;
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
  margin-bottom: 32px;

  .feature-item {
    position: relative;
    cursor: pointer;

    .feature-bg {
      position: relative;
      width: 95px;
      margin: 0 auto;

      .bg-image {
        width: 100%;
        height: 100%;
        object-fit: cover;
        border-radius: 12px;
      }
    }

    .feature-content {
      position: absolute;
      top: 50%;
      left: 50%;
      z-index: 1;
      display: flex;
      flex-direction: column;
      align-items: center;
      transform: translate(-50%, calc(-50% - 8px));

      .feature-icon {
        width: 46px;
        height: 46px;
        object-fit: contain;
      }

      .feature-title {
        margin-top: 6px;
        color: #fff;
        font-size: 12px;
        font-weight: 600;
        line-height: 1;
        text-align: center;
        white-space: nowrap;
        text-shadow: 0 1px 3px rgba(0, 0, 0, 0.5);
      }
    }

  }
}

.start-button {
    width: 100%;
    padding: 12px 0;
    background: transparent;
    color: $brand-gold;
    font-size: 17px;
    font-weight: 500;
    border: 1px solid $brand-gold;
    border-radius: 24px;
    cursor: pointer;
    text-align: center;
    margin-top: 8px;
    margin-bottom: 20px;
    transition: all 0.3s ease;

    &:hover {
      background: $gradient-gold;
      color: $text-inverse;
    }
}

.advantage-section {
  padding: 24px 20px;
  background: rgba(20, 20, 20, 0.6);
  backdrop-filter: blur(10px);
  border-radius: 24px;
  margin-bottom: 32px;
  margin-top: 20px;
  border: 1px solid $border-color;

  .ido-header {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 8px;

    .ido-icon {
      width: 24px;
      height: 24px;
      object-fit: contain;
    }

    .ido-title {
      font-size: 18px;
      font-weight: bold;
      color: #fff;
      margin: 0;
    }
  }

  .ido-subtitle {
    font-size: 13px;
    color: rgba(255, 255, 255, 0.6);
    margin: 0 0 16px 0;
    line-height: 1.5;
  }

  .ido-card {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 16px;
    background: rgba(212, 175, 55, 0.08);
    border-radius: 12px;
    border: 1px solid rgba(212, 175, 55, 0.25);
    margin-bottom: 24px;

    .ido-label {
      font-size: 16px;
      font-weight: bold;
      color: #fff;
    }

    .ido-limit {
      font-size: 14px;
      color: $brand-gold;
    }
  }

  .ido-price {
    text-align: center;
    margin-bottom: 4px;

    .price-number {
      font-size: 36px;
      font-weight: bold;
      color: $brand-gold;
    }

    .price-unit {
      font-size: 16px;
      color: #fff;
      margin-left: 4px;
    }
  }

  .ido-limit-text {
    text-align: center;
    font-size: 13px;
    color: rgba(255, 255, 255, 0.6);
    margin: 0 0 24px 0;
  }

  .project-info {
    background: rgba(255, 255, 255, 0.03);
    border-radius: 16px;
    padding: 20px;
    margin-bottom: 24px;
    border: 1px solid rgba(255, 255, 255, 0.08);

    .info-section {
      margin-bottom: 20px;

      &:last-child {
        margin-bottom: 0;
      }

      .section-title {
        font-size: 14px;
        font-weight: 600;
        color: $brand-gold;
        margin: 0 0 12px 0;
        padding-bottom: 8px;
        border-bottom: 1px solid rgba(255, 255, 255, 0.1);
      }

      .info-content {
        .info-item {
          font-size: 13px;
          color: rgba(255, 255, 255, 0.85);
          margin: 0 0 8px 0;
          line-height: 1.6;

          &:last-child {
            margin-bottom: 0;
          }
        }

        .info-subitem {
          font-size: 12px;
          color: rgba(255, 255, 255, 0.6);
          margin: 0 0 6px 0;
          padding-left: 16px;
          line-height: 1.5;

          &:last-child {
            margin-bottom: 0;
          }
        }
      }
    }
  }

  .ido-benefits {
    display: flex;
    flex-direction: column;
    gap: 16px;
    margin-bottom: 24px;

    .benefit-item {
      display: flex;
      align-items: center;
      gap: 12px;

      .benefit-icon {
        width: 48px;
        height: 48px;
        object-fit: contain;
        flex-shrink: 0;
      }

      .benefit-info {
        .benefit-title {
          font-size: 16px;
          font-weight: bold;
          color: #fff;
          margin: 0 0 4px 0;
        }

        .benefit-desc {
          font-size: 13px;
          color: rgba(255, 255, 255, 0.6);
          margin: 0;
        }
      }
    }
  }

  .ido-btn {
    width: 100%;
    padding: 14px 0;
    font-size: 16px;
    font-weight: 500;
    border: 1px solid $brand-gold;
    border-radius: 24px;
    cursor: pointer;
    text-align: center;
    transition: all 0.3s ease;
    background: $gradient-gold;
    color: $text-inverse; 
  }
}

.nft-section {
  padding: 24px 20px;
  background: rgba(20, 20, 20, 0.6);
  backdrop-filter: blur(10px);
  border-radius: 24px;
  margin-bottom: 32px;
  border: 1px solid $border-color;

  .nft-header {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 16px;

    .nft-icon {
      width: 24px;
      height: 24px;
      object-fit: contain;
    }

    .nft-title {
      font-size: 18px;
      font-weight: bold;
      color: #fff;
      margin: 0;
    }
  }

  .nft-image {
    width: 100%;
    height: auto;
    border-radius: 12px;
    margin-bottom: 16px;
  }

  .nft-hint {
    text-align: center;
    font-size: 15px;
    font-weight: bold;
    color: $brand-gold;
    margin: 0 0 12px 0;
  }

  .nft-desc {
    padding: 12px 10px;
    background: rgba(255, 255, 255, 0.03);
    border-radius: 12px;
    font-size: 13px;
    color: $text-muted;
    line-height: 1.5;
    margin: 0 0 12px 0;
    text-align: center;
  }

  .nft-progress {
    display: flex;
    align-items: center;
    justify-content: space-between;

    .progress-text {
      font-size: 13px;
      color: $text-muted;
    }
  }
}

.partners-title {
  font-size: 21px;
  font-weight: bold;
  color: #fff;
  text-align: center;
  margin: 30px 0 16px 0;
}

@keyframes slideUp {
  from {
    transform: translateY(20px);
    opacity: 0;
  }
  to {
    transform: translateY(0);
    opacity: 1;
  }
}

@keyframes activeLogoBreathJump {
  0% {
    transform: translateY(0);
    opacity: 0.8;
  }
  10% {
    transform: translateY(-8px);
    opacity: 0.9;
  }
  25% {
    transform: translateY(-25px);
    opacity: 1;
  }
  40% {
    transform: translateY(-12px);
    opacity: 0.95;
  }
  50% {
    transform: translateY(0);
    opacity: 0.8;
  }
  60% {
    transform: translateY(-6px);
    opacity: 0.85;
  }
  75% {
    transform: translateY(-15px);
    opacity: 0.9;
  }
  85% {
    transform: translateY(-8px);
    opacity: 0.85;
  }
  100% {
    transform: translateY(0);
    opacity: 0.8;
  }
}

.active-logo {
  will-change: transform, opacity;
  animation: activeLogoBreathJump 10s infinite ease-in-out;
}

@keyframes gradient-move {
  0% {
    background-position: 100% 50%;
  }
  50% {
    background-position: 0% 50%;
  }
  100% {
    background-position: 100% 50%;
  }
}
</style>
