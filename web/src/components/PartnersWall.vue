<template>
  <div class="partners-wall">
    <div class="wall-row" data-direction="normal">
      <div class="wall-track">
        <div v-for="i in row1Logos" :key="i" class="wall-item">
          <img :src="getPartnerImage(i)" :alt="`Partner ${i}`" />
        </div>
        <div v-for="i in row1Logos" :key="`dup-${i}`" class="wall-item">
          <img :src="getPartnerImage(i)" :alt="`Partner ${i}`" />
        </div>
      </div>
    </div>
    <div class="wall-row" data-direction="reverse">
      <div class="wall-track">
        <div v-for="i in row2Logos" :key="i" class="wall-item">
          <img :src="getPartnerImage(i)" :alt="`Partner ${i}`" />
        </div>
        <div v-for="i in row2Logos" :key="`dup-${i}`" class="wall-item">
          <img :src="getPartnerImage(i)" :alt="`Partner ${i}`" />
        </div>
      </div>
    </div>
    <div class="wall-row" data-direction="normal">
      <div class="wall-track">
        <div v-for="i in row3Logos" :key="i" class="wall-item">
          <img :src="getPartnerImage(i)" :alt="`Partner ${i}`" />
        </div>
        <div v-for="i in row3Logos" :key="`dup-${i}`" class="wall-item">
          <img :src="getPartnerImage(i)" :alt="`Partner ${i}`" />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
const row1Logos = [1, 2, 3, 4, 5, 6, 7, 8]
const row2Logos = [9, 10, 11, 12, 3, 5, 9, 13]
const row3Logos = [13, 14, 15, 16, 17, 18, 19, 20]

const getPartnerImage = (index: number) => {
  return `/static/partners/partner-${index}.png`
}
</script>

<style lang="scss" scoped>
.partners-wall {
  margin-top: 20px;
  overflow: hidden;

  .wall-row {
    overflow: hidden;
    margin-bottom: 12px;

    &:last-child {
      margin-bottom: 0;
    }

    .wall-track {
      display: flex;
      align-items: center;
      gap: 24px;
      width: max-content;
      animation: scroll-normal 20s linear infinite;
    }

    &[data-direction="reverse"] .wall-track {
      animation: scroll-reverse 25s linear infinite;
    }

    .wall-item {
      flex-shrink: 0;
      display: flex;
      align-items: center;
      justify-content: center;
      padding: 8px 12px;

      img {
        max-height: 30px;
        width: auto;
        object-fit: contain;
        transition: all 0.3s ease;
      }

      &:hover img {
        filter: brightness(1) grayscale(0);
        opacity: 1;
      }
    }
  }
}

@keyframes scroll-normal {
  from {
    transform: translateX(0);
  }
  to {
    transform: translateX(-50%);
  }
}

@keyframes scroll-reverse {
  from {
    transform: translateX(-50%);
  }
  to {
    transform: translateX(0);
  }
}
</style>
