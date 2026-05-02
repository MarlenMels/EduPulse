<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, watch } from 'vue'
import Hls from 'hls.js'
import { resolveMediaUrl } from '@/api/client'

const props = defineProps<{ src: string; poster?: string }>()
const videoEl = ref<HTMLVideoElement | null>(null)
let hls: Hls | null = null

function absoluteSrc(s: string): string {
  return resolveMediaUrl(s)
}

function attach() {
  if (!videoEl.value || !props.src) return
  if (hls) { hls.destroy(); hls = null }
  const src = absoluteSrc(props.src)
  if (!src.endsWith('.m3u8')) {
    videoEl.value.src = src
  } else if (Hls.isSupported()) {
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
