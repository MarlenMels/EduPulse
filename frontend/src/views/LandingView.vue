<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'

const router = useRouter()
const canvas = ref<HTMLCanvasElement | null>(null)
const titleVisible = ref(false)
const buttonsVisible = ref(false)

const title = 'EDpulse'
let animId = 0

interface Node {
  x: number
  y: number
  vx: number
  vy: number
}

const nodes: Node[] = []
const NODE_COUNT = 50

for (let i = 0; i < NODE_COUNT; i++) {
  nodes.push({
    x: Math.random(),
    y: Math.random(),
    vx: (Math.random() - 0.5) * 0.002,
    vy: (Math.random() - 0.5) * 0.002,
  })
}

function drawNetwork() {
  const c = canvas.value
  if (!c) return
  const ctx = c.getContext('2d')
  if (!ctx) return

  c.width = c.offsetWidth * window.devicePixelRatio
  c.height = c.offsetHeight * window.devicePixelRatio
  ctx.scale(window.devicePixelRatio, window.devicePixelRatio)

  const w = c.offsetWidth
  const h = c.offsetHeight

  ctx.clearRect(0, 0, w, h)

  for (const n of nodes) {
    n.x += n.vx
    n.y += n.vy
    if (n.x < 0 || n.x > 1) n.vx *= -1
    if (n.y < 0 || n.y > 1) n.vy *= -1
    n.x = Math.max(0, Math.min(1, n.x))
    n.y = Math.max(0, Math.min(1, n.y))
  }

  for (let i = 0; i < nodes.length; i++) {
    for (let j = i + 1; j < nodes.length; j++) {
      const ni = nodes[i]!
      const nj = nodes[j]!
      const p1x = ni.x * w
      const p1y = ni.y * h
      const p2x = nj.x * w
      const p2y = nj.y * h
      const dist = Math.hypot(p1x - p2x, p1y - p2y)
      if (dist < 250) {
        const alpha = (1 - dist / 250) * 0.4
        ctx.strokeStyle = `rgba(0, 255, 255, ${alpha})`
        ctx.lineWidth = 1
        ctx.beginPath()
        ctx.moveTo(p1x, p1y)
        ctx.lineTo(p2x, p2y)
        ctx.stroke()
      }
    }
  }

  for (const n of nodes) {
    ctx.fillStyle = 'rgba(0, 255, 255, 0.8)'
    ctx.beginPath()
    ctx.arc(n.x * w, n.y * h, 4, 0, Math.PI * 2)
    ctx.fill()
  }

  animId = requestAnimationFrame(drawNetwork)
}

onMounted(() => {
  drawNetwork()
  setTimeout(() => (titleVisible.value = true), 200)
  setTimeout(() => (buttonsVisible.value = true), 1500)
})

onUnmounted(() => {
  cancelAnimationFrame(animId)
})
</script>

<template>
  <div class="relative min-h-screen overflow-hidden bg-gradient-to-br from-[#0D0D1A] to-[#1B1B2F]">
    <canvas
      ref="canvas"
      class="absolute inset-0 w-full h-full"
    />

    <div class="relative z-10 flex flex-col items-center justify-center min-h-screen px-6">
      <div
        class="flex gap-1 transition-all duration-1000"
        :class="titleVisible ? 'opacity-100 translate-y-0' : 'opacity-0 translate-y-8'"
      >
        <span
          v-for="(char, i) in title"
          :key="i"
          class="text-7xl md:text-9xl font-extrabold bg-gradient-to-r from-cyan-400 via-blue-500 to-purple-500 bg-clip-text text-transparent"
          :style="{ transitionDelay: `${i * 120}ms` }"
        >
          {{ char }}
        </span>
      </div>

      <p
        class="mt-6 text-white/60 text-xl md:text-2xl transition-all duration-700 delay-700"
        :class="titleVisible ? 'opacity-100' : 'opacity-0'"
      >
        Educational Platform
      </p>

      <div
        class="mt-16 flex flex-col gap-4 w-full max-w-sm transition-all duration-700"
        :class="buttonsVisible ? 'opacity-100 translate-y-0' : 'opacity-0 translate-y-8'"
      >
        <button
          @click="router.push('/login')"
          class="w-full py-4 rounded-2xl bg-cyan-400 text-black font-bold text-lg hover:bg-cyan-300 transition-colors shadow-lg shadow-cyan-400/20"
        >
          Sign In
        </button>
        <button
          @click="router.push('/register')"
          class="w-full py-4 rounded-2xl border-2 border-cyan-400/50 text-cyan-400 font-bold text-lg hover:bg-cyan-400/10 transition-colors"
        >
          Sign Up
        </button>
      </div>
    </div>
  </div>
</template>
