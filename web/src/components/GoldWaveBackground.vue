<template>
  <div ref="containerRef" class="gold-wave-bg"></div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import * as THREE from 'three'

const containerRef = ref<HTMLDivElement>()

let renderer: THREE.WebGLRenderer | null = null
let camera: THREE.PerspectiveCamera | null = null
let scene: THREE.Scene | null = null
let particles: THREE.Points | null = null
let rafId = 0
let count = 0

const SEPARATION = 35
const AMOUNTX = 65
const AMOUNTY = 40

onMounted(() => {
  const container = containerRef.value
  if (!container) return

  const w = container.clientWidth
  const h = container.clientHeight

  // 相机
  camera = new THREE.PerspectiveCamera(75, w / h, 1, 10000)
  camera.position.z = 1000
  camera.position.y = 400
  camera.lookAt(new THREE.Vector3(0, 150, 0))

  // 场景
  scene = new THREE.Scene()

  // 粒子网格
  const numParticles = AMOUNTX * AMOUNTY
  const positions = new Float32Array(numParticles * 3)

  let i = 0
  for (let ix = 0; ix < AMOUNTX; ix++) {
    for (let iy = 0; iy < AMOUNTY; iy++) {
      // 使用非线性分布，让上方（小 iy）密集，下方（大 iy）稀疏
      const normalizedY = iy / AMOUNTY
      const nonLinearY = 1 - (1 - normalizedY) * (1 - normalizedY) // 反向平方函数
      const zPos = nonLinearY * (AMOUNTY * SEPARATION) - ((AMOUNTY * SEPARATION) / 2)

      positions[i] = ix * SEPARATION - ((AMOUNTX * SEPARATION) / 2)
      positions[i + 1] = 0
      positions[i + 2] = zPos
      i += 3
    }
  }

  const geometry = new THREE.BufferGeometry()
  geometry.setAttribute('position', new THREE.BufferAttribute(positions, 3))

  // 金黄色粒子材质
  const material = new THREE.PointsMaterial({
    color: 0xECD0A5,
    size: 3.5,
    transparent: true,
    opacity: 1
  })

  particles = new THREE.Points(geometry, material)
  scene.add(particles)

  // 渲染器
  renderer = new THREE.WebGLRenderer({ antialias: true, alpha: true })
  renderer.setSize(w, h)
  renderer.setPixelRatio(Math.min(window.devicePixelRatio, 2))
  container.appendChild(renderer.domElement)

  // 窗口变化
  const onResize = () => {
    if (!camera || !renderer || !container) return
    const cw = container.clientWidth
    const ch = container.clientHeight
    camera.aspect = cw / ch
    camera.updateProjectionMatrix()
    renderer.setSize(cw, ch)
  }
  window.addEventListener('resize', onResize)

  // 动画
  const animate = () => {
    rafId = requestAnimationFrame(animate)
    if (!particles) return

    const posArray = particles.geometry.attributes.position.array as Float32Array
    let j = 0
    for (let ix = 0; ix < AMOUNTX; ix++) {
      for (let iy = 0; iy < AMOUNTY; iy++) {
        posArray[j + 1] =
          Math.sin((ix + count) * 0.3) * 50 +
          Math.sin((iy + count) * 0.5) * 50
        j += 3
      }
    }
    particles.geometry.attributes.position.needsUpdate = true
    renderer?.render(scene!, camera!)
    count += 0.05
  }
  animate()

  // 卸载时清理
  onUnmounted(() => {
    window.removeEventListener('resize', onResize)
    cancelAnimationFrame(rafId)
    renderer?.dispose()
    geometry.dispose()
    material.dispose()
    if (container.contains(renderer!.domElement)) {
      container.removeChild(renderer!.domElement)
    }
  })
})
</script>

<style scoped>
.gold-wave-bg {
  position: absolute;
  inset: 0;
  z-index: 0;
  pointer-events: none;
}
.gold-wave-bg canvas {
  display: block;
  width: 100% !important;
  height: 100% !important;
}
</style>
