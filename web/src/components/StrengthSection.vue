<template>
  <div class="strength-section">
    <h2 class="section-title">{{ $t('advantage.title') }}</h2>
    <div ref="wrapperRef" class="carousel-wrapper">
      <div ref="trackRef" class="carousel-track" :style="trackStyle">
        <div
          v-for="(card, i) in clonedCards"
          :key="i"
          class="strength-card"
          :class="{ active: i === activeIndex }"
        >
          <div class="card-border">
            <div class="card-tag">{{ $t(card.tag) }}</div>
            <div class="card-inner">
              <img :src="card.illustration" :alt="$t(card.tag)" class="card-illustration" />
              <img :src="card.icon" :alt="$t(card.tag)" class="card-icon" />
              <p class="card-desc">{{ $t(card.desc) }}</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'

interface StrengthCard {
  tag: string
  illustration: string
  icon: string
  desc: string
}

const cards: StrengthCard[] = [
  {
    tag: 'advantage.revenueTag',
    illustration: '/static/advantage/illustration-revenue-growth.png',
    icon: '/static/advantage/icon-revenue-up.png',
    desc: 'advantage.revenueDesc'
  },
  {
    tag: 'advantage.transparentTag',
    illustration: '/static/advantage/illustration-data-transparent.png',
    icon: '/static/advantage/icon-transparent-link.png',
    desc: 'advantage.transparentDesc'
  },
  {
    tag: 'advantage.securityTag',
    illustration: '/static/advantage/illustration-security-shield.png',
    icon: '/static/advantage/icon-security-lock.png',
    desc: 'advantage.securityDesc'
  }
]

// 首尾克隆实现无限循环
const clonedCards = [
  cards[cards.length - 1],
  ...cards,
  cards[0]
]

const wrapperRef = ref<HTMLDivElement>()
const activeIndex = ref(1)
const isTransitioning = ref(true)
const translateX = ref(0)
const CARD_WIDTH_RATIO = 0.55
const GAP = 12

const updateTranslateX = () => {
  if (!wrapperRef.value) return
  const wrapperWidth = wrapperRef.value.clientWidth
  const cardWidth = wrapperWidth * CARD_WIDTH_RATIO
  const stagePadding = (wrapperWidth - cardWidth) / 2
  translateX.value = stagePadding - activeIndex.value * (cardWidth + GAP)
}

const trackStyle = computed(() => ({
  transform: `translateX(${translateX.value}px)`,
  transition: isTransitioning.value ? 'transform 0.5s ease' : 'none'
}))

let timer: ReturnType<typeof setInterval> | null = null
let resizeObserver: ResizeObserver | null = null

const goNext = () => {
  activeIndex.value++
  isTransitioning.value = true
  updateTranslateX()

  if (activeIndex.value >= clonedCards.length - 1) {
    setTimeout(() => {
      isTransitioning.value = false
      activeIndex.value = 1
      updateTranslateX()
    }, 500)
  }
}

onMounted(() => {
  updateTranslateX()

  resizeObserver = new ResizeObserver(updateTranslateX)
  if (wrapperRef.value) resizeObserver.observe(wrapperRef.value)

  timer = setInterval(goNext, 3000)
})

onUnmounted(() => {
  if (timer) clearInterval(timer)
  if (resizeObserver) resizeObserver.disconnect()
})
</script>

<style lang="scss" scoped>
@use '@/style/variables.scss' as *;

.strength-section {
  margin-bottom: 32px;

  .section-title {
    font-size: 21px;
    font-weight: bold;
    color: #fff;
    text-align: center;
    margin: 0 0 20px 0;
  }

  .carousel-wrapper {
    overflow: hidden;
    width: 100%;
  }

  .carousel-track {
    display: flex;
    gap: 12px;
    padding-top: 16px;
    will-change: transform;
  }

  .strength-card {
    flex: 0 0 55%;
    height: 280px;
    min-width: 0;

    .card-border {
      position: relative;
      padding: 2px;
      background: linear-gradient(180deg, #F1DC9B 0%, #D4AF37 100%);
      border-radius: 20px;
      height: 100%;
    }

    .card-tag {
      position: absolute;
      top: -12px;
      left: 50%;
      transform: translateX(-50%);
      padding: 6px 24px;
      background: linear-gradient(180deg, #F1DC9B 0%, #D4AF37 51%, #B8941F 100%);
      color: #000;
      font-size: 13px;
      font-weight: 800;
      border-radius: 20px;
      z-index: 2;
      white-space: nowrap;
    }

    .card-inner {
      display: flex;
      flex-direction: column;
      align-items: center;
      justify-content: flex-start;
      background: rgba(20, 20, 20, 0.6);
      backdrop-filter: blur(10px);
      border-radius: 18px;
      padding: 0 0 20px;
      overflow: hidden;
      height: calc(100% - 4px);

      .card-illustration {
        width: 100%;
        height: 140px;
        object-fit: contain;
        margin-bottom: 8px;
      }

      .card-icon {
        width: 48px;
        height: 48px;
        object-fit: contain;
        margin-bottom: 12px;
      }

      .card-desc {
        font-size: 12px;
        color: $text-muted;
        text-align: center;
        margin: 0;
        line-height: 1.5;
        padding: 0 8px;
      }
    }
  }
}
</style>
