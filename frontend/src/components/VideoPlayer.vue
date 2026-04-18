<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, watch } from 'vue'
import Hls from 'hls.js'

const props = defineProps<{ src: string; poster?: string }>()
const videoEl = ref<HTMLVideoElement | null>(null)
let hls: Hls | null = null

function absoluteSrc(s: string): string {
  if (!s) return ''
  if (s.startsWith('http://') || s.startsWith('https://')) return s
  return s
}

function attach() {
  if (!videoEl.value || !props.src) return
  if (hls) { hls.destroy(); hls = null }
  const src = absoluteSrc(props.src)
  if (Hls.isSupported()) {
    hls = new Hls()
    hls.loadSource(src)
    hls.attachMedia(videoEl.value)
  } else if (videoEl.value.canPlayType('application/vnd.apple.mpegurl')) {
    videoEl.value.src = src
  }
}

onMounted(attach)
watch(() => props.src, attach)
onBeforeUnmount(() => { if (hls) { hls.destroy(); hls = null } })
</script>

<template>
  <video
    ref="videoEl"
    :poster="poster"
    controls
    class="w-full aspect-video rounded-2xl bg-black"
  />
</template>
