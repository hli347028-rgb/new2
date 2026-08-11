<template>
  <div class="index-page">
    <Header />

    <div class="content">
      <div class="welcome-card">
        <AixWaveBackground />
        <div class="welcome-content">
          <img src="/assets/logo.png" alt="Logo" class="logo active-logo" />
          <div class="welcome-subtitle">Aix Protocol</div>
          <!-- <div class="welcome-description">{{ $t('index.welcomeDescription2') }}</div> -->
        </div>
      </div>

      <div class="features">
        <div class="feature-item" @click="handleExternalLink">
          <div class="feature-bg">
            <img src="/static/output-icon-bg-DIHWgt3w.png" alt="internationalPayment" class="bg-image" />
            <div class="feature-content">
              <img src="/static/output-icon.png" alt="internationalPayment" class="feature-icon" />
              <div class="feature-title">{{ $t('recharge.recharge') }}</div>
            </div>
          </div>
        </div>
        <div class="feature-item" @click="handleDownload">
          <div class="feature-bg">
            <img src="/static/module-icon-bg-weJSvE3-.png" alt="chainGame" class="bg-image" />
            <div class="feature-content">
              <img src="/static/module-icon.png" alt="chainGame" class="feature-icon" />
              <div class="feature-title">{{ $t('tab.walletDownload') }}</div>
            </div>
          </div>
        </div>
        <div class="feature-item" @click="showToast($t('common.comingSoon'))">
          <div class="feature-bg">
            <img src="/static/chart-icon-bg-yYEAC2yW.png" alt="aixChain" class="bg-image" />
            <div class="feature-content">
              <img src="/static/chart-icon.png" alt="aixChain" class="feature-icon" />
              <div class="feature-title">{{ $t('tab.aixChain') }}</div>
            </div>
          </div>
        </div>
      </div>
      <button class="start-button" @click="handleStart">{{ $t('index.startNow') }}</button>

      <div class="advantage-section">
        <!-- <div class="ido-header">
          <img src="/static/node-project.png" alt="" class="ido-icon" />
          <h2 class="ido-title">{{ $t('index.idoRecruitPlan') }}</h2>
        </div>

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
        <p class="ido-limit-text">{{ $t('index.accountLimit', { count: 1 }) }}</p> -->

        <div class="project-info">
          <div class="info-section">
            <h4 class="section-title">{{ $t('index.projectInfoTitle1') }}</h4>

            <h4 class="section-title">{{ $t('index.predictionExchangeTitle') }}</h4>
            <div class="info-content exchange-info">
              <p class="exchange-subtitle">{{ $t('index.aiAgentBot') }}</p>

              <div class="token-grid">
                <div class="token-item">
                  <span class="token-label">{{ $t('index.tokenNameLabel') }}</span>
                  <strong>AIX</strong>
                </div>
                <div class="token-item">
                  <span class="token-label">{{ $t('index.tokenSupplyLabel') }}</span>
                  <strong>{{ $t('index.tokenSupplyValue') }}</strong>
                </div>
                <div class="token-item token-pair">
                  <span class="token-label">{{ $t('index.tradingPairLabel') }}</span>
                  <strong>WIN / AIX</strong>
                </div>
              </div>

              <div class="slippage-card">
                <div class="slippage-title">{{ $t('index.slippageRules') }}</div>
                <div class="rule-list">
                  <p class="rule-item"><span></span>{{ $t('index.buyDisabled') }}</p>
                  <p class="rule-item"><span></span>{{ $t('index.sellFee') }}</p>
                  <p class="rule-item emphasis"><span></span>{{ $t('index.priceDrop2Rule') }}</p>
                  <p class="rule-item emphasis"><span></span>{{ $t('index.priceDrop10Rule') }}</p>
                </div>
              </div>
            </div>
          </div>

          <div class="info-section">
            <h4 class="section-title">{{ $t('index.rulesSummaryTitle') }}</h4>
            <div class="info-content exchange-info">
              <div class="slippage-card rules-summary-card">
                <div class="rule-list">
                  <p class="rule-item"><span></span>{{ $t('index.staticIncomeRule') }}</p>
                  <p class="rule-item"><span></span>{{ $t('index.directReferralRule') }}</p>
                  <p class="rule-item"><span></span>{{ $t('index.managementRewardRule') }}</p>
                  <p class="rule-item emphasis"><span></span>{{ $t('index.exitRule') }}</p>
                </div>
              </div>
            </div>
          </div>

        </div>

        <!-- <button class="ido-btn" @click="handleGoToNode">{{ $t('index.becomeIdo') }}</button> -->
      </div>

      <StrengthSection />

      <h2 class="partners-title">{{ $t('index.partners') }}</h2>
      <PartnersWall />
    </div>
  </div>

</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { showToast } from 'vant'
import Header from '@/components/Header.vue'
import AixWaveBackground from '@/components/AixWaveBackground.vue'
import StrengthSection from '@/components/StrengthSection.vue'
import PartnersWall from '@/components/PartnersWall.vue'

const router = useRouter()
const { t: $t } = useI18n()

const handleStart = () => {
  router.push('/node')
}

const handleGoToNode = () => {
  router.push('/node')
}

const handleDownload = () => {
  window.open('https://testnet.wallet.eoeo.info/06bx', '_blank', 'noopener,noreferrer')
}

const handleExternalLink = () => {
  router.push('/recharge')
}

</script>

<style lang="scss" scoped>
@use '@/style/variables.scss' as *;

.index-page {
  min-height: 100vh;
  position: relative;
  overflow-x: hidden;
  padding-top: 64px;
  background: radial-gradient(circle at 50% 0%, #0D2A42 0%, $bg-main 36%, #020508 100%);

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
  height: 300px;
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
      margin-top: 10px;
      width: 220px;
      height: auto;
    }

    .welcome-subtitle {
      font-size: 32px;
      font-weight: 800;
      background: linear-gradient(135deg, rgba(143, 223, 255, 0.8) 0%, #FFFFFF 50%, rgba(143, 223, 255, 0.8) 100%);
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
      color: rgba(143, 223, 255, 1);
      background: linear-gradient(90deg, transparent, rgba(21, 151, 229, 0.3), transparent);
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
  position: relative;
  display: block;
  box-sizing: border-box;
  width: calc(100% - 40px);
  min-height: 42px;
  margin: 6px auto 4px;
  padding: 10px 20px;
  overflow: hidden;
  background: linear-gradient(135deg, rgba(21, 151, 229, 0.12), rgba(143, 223, 255, 0.045));
  border: 1px solid rgba(143, 223, 255, 0.28);
  border-radius: 12px;
  box-shadow:
    0 6px 18px rgba(0, 76, 135, 0.12),
    inset 0 1px 0 rgba(255, 255, 255, 0.08);
  color: $brand-primary-light;
  font-size: 14px;
  font-weight: 600;
  line-height: 1.4;
  letter-spacing: 0.6px;
  text-align: center;
  cursor: pointer;
  backdrop-filter: blur(8px);
  transition: all $transition-fast;

  &::before {
    content: '';
    position: absolute;
    top: 0;
    left: 18%;
    right: 18%;
    height: 1px;
    background: linear-gradient(90deg, transparent, rgba(143, 223, 255, 0.62), transparent);
    opacity: 0.65;
  }

  &:hover {
    color: $text-primary;
    background: linear-gradient(135deg, rgba(21, 151, 229, 0.2), rgba(143, 223, 255, 0.08));
    border-color: rgba(143, 223, 255, 0.48);
    transform: translateY(-1px);
    box-shadow:
      0 8px 22px rgba(0, 91, 156, 0.18),
      inset 0 1px 0 rgba(255, 255, 255, 0.12);
  }

  &:active {
    background: rgba(21, 151, 229, 0.16);
    transform: translateY(0) scale(0.99);
    box-shadow: inset 0 2px 5px rgba(2, 7, 12, 0.2);
  }

  &:focus-visible {
    outline: 1px solid $brand-primary-light;
    outline-offset: 2px;
  }
}

.advantage-section {
  padding: 24px 20px;
  background: rgba(8, 19, 30, 0.6);
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
    background: rgba(21, 151, 229, 0.08);
    border-radius: 12px;
    border: 1px solid rgba(21, 151, 229, 0.25);
    margin-bottom: 24px;

    .ido-label {
      font-size: 16px;
      font-weight: bold;
      color: #fff;
    }

    .ido-limit {
      font-size: 14px;
      color: $brand-primary;
    }
  }

  .ido-price {
    text-align: center;
    margin-bottom: 4px;

    .price-number {
      font-size: 36px;
      font-weight: bold;
      color: $brand-primary;
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
        color: $brand-primary;
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

        &.exchange-info {
          .exchange-subtitle {
            margin: 0 0 14px;
            padding: 9px 12px;
            color: $brand-primary-light;
            font-size: 12px;
            font-weight: 500;
            line-height: 1.5;
            background: linear-gradient(90deg, rgba(21, 151, 229, 0.13), transparent);
            border-left: 2px solid $brand-accent;
            border-radius: 0 8px 8px 0;
          }

          .token-grid {
            display: grid;
            grid-template-columns: repeat(2, minmax(0, 1fr));
            gap: 8px;
            margin-bottom: 12px;
          }

          .token-item {
            display: flex;
            flex-direction: column;
            gap: 5px;
            min-width: 0;
            padding: 11px 12px;
            background: rgba(143, 223, 255, 0.035);
            border: 1px solid rgba(143, 223, 255, 0.08);
            border-radius: 10px;

            &.token-pair {
              grid-column: 1 / -1;
              flex-direction: row;
              align-items: center;
              justify-content: space-between;
            }

            .token-label {
              color: $text-muted;
              font-size: 11px;
            }

            strong {
              color: $text-primary;
              font-size: 13px;
              font-weight: 600;
              letter-spacing: 0.3px;
            }
          }

          .slippage-card {
            padding: 13px 14px;
            background: linear-gradient(145deg, rgba(21, 151, 229, 0.09), rgba(8, 19, 30, 0.6));
            border: 1px solid rgba(21, 151, 229, 0.16);
            border-radius: 12px;
          }

          .slippage-title {
            margin-bottom: 10px;
            color: $brand-primary;
            font-size: 12px;
            font-weight: 600;
            letter-spacing: 0.4px;
          }

          .rule-list {
            display: flex;
            flex-direction: column;
            gap: 7px;
          }

          .rule-item {
            display: flex;
            align-items: flex-start;
            gap: 8px;
            margin: 0;
            color: rgba(244, 250, 255, 0.9);
            font-size: 11px;
            line-height: 1.55;

            > span {
              flex: 0 0 auto;
              width: 4px;
              height: 4px;
              margin-top: 6px;
              background: $brand-primary;
              border-radius: 50%;
              box-shadow: 0 0 6px rgba(33, 182, 234, 0.55);
            }

            &.emphasis {
              color: $brand-primary-light;
            }
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
    border: 1px solid $brand-primary;
    border-radius: 24px;
    cursor: pointer;
    text-align: center;
    transition: all 0.3s ease;
    background: $gradient-primary;
    color: $text-inverse; 
  }
}

.nft-section {
  padding: 24px 20px;
  background: rgba(8, 19, 30, 0.6);
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
    color: $brand-primary;
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
