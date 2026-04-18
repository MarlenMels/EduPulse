<script setup lang="ts">
import { ref, onBeforeUnmount } from 'vue'
import { Upload, CheckCircle2, XCircle, Loader2 } from 'lucide-vue-next'
import { videoApi } from '@/api/client'

const props = defineProps<{ lessonId: number }>()
const emit = defineEmits<{
  uploaded: [hlsUrl: string]
  failed: [error: string]
}>()

type Status = 'idle' | 'uploading' | 'processing' | 'ready' | 'failed'

const status = ref<Status>('idle')
const progress = ref(0)
const errorMsg = ref('')
const fileName = ref('')
const dragOver = ref(false)
const inputEl = ref<HTMLInputElement | null>(null)
let pollTimer: number | null = null

const ALLOWED = ['.mp4', '.mov', '.mkv']
const MAX_SIZE = 500 * 1024 * 1024

function validate(f: File): string {
  const lower = f.name.toLowerCase()
  if (!ALLOWED.some((e) => lower.endsWith(e))) {
    return 'Формат не поддерживается (mp4, mov, mkv)'
  }
  if (f.size > MAX_SIZE) return 'Файл больше 500MB'
  return ''
}

async function handleFile(f: File) {
  const v = validate(f)
  if (v) {
    status.value = 'failed'
    errorMsg.value = v
    emit('failed', v)
    return
  }
  fileName.value = f.name
  errorMsg.value = ''
  status.value = 'uploading'
  progress.value = 0
  try {
    await videoApi.upload(props.lessonId, f, (pct) => { progress.value = pct })
    status.value = 'processing'
    startPolling()
  } catch (e: any) {
    status.value = 'failed'
    const msg = e?.response?.data?.error || 'Ошибка загрузки'
    errorMsg.value = msg
    emit('failed', msg)
  }
}

function startPolling() {
  stopPolling()
  pollTimer = window.setInterval(async () => {
    try {
      const res = await videoApi.status(props.lessonId)
      const s = res.data.status
      if (s === 'ready') {
        status.value = 'ready'
        stopPolling()
        emit('uploaded', res.data.hls_path)
      } else if (s === 'failed') {
        status.value = 'failed'
        errorMsg.value = res.data.error_message || 'Ошибка конвертации'
        stopPolling()
        emit('failed', errorMsg.value)
      }
    } catch {
      // ignore transient errors
    }
  }, 3000)
}

function stopPolling() {
  if (pollTimer !== null) {
    clearInterval(pollTimer)
    pollTimer = null
  }
}

onBeforeUnmount(stopPolling)

function onDrop(e: DragEvent) {
  e.preventDefault()
  dragOver.value = false
  const f = e.dataTransfer?.files?.[0]
  if (f) handleFile(f)
}

function onSelect(e: Event) {
  const f = (e.target as HTMLInputElement).files?.[0]
  if (f) handleFile(f)
}

function pickFile() {
  inputEl.value?.click()
}

function reset() {
  status.value = 'idle'
  progress.value = 0
  errorMsg.value = ''
  fileName.value = ''
}
</script>

<template>
  <div
    class="bg-[#1E1E1E] border border-dashed rounded-2xl p-6 transition-colors"
    :class="dragOver ? 'border-cyan-400/70' : 'border-white/10 hover:border-cyan-400/50'"
    @dragover.prevent="dragOver = true"
    @dragleave.prevent="dragOver = false"
    @drop="onDrop"
  >
    <input
      ref="inputEl"
      type="file"
      accept=".mp4,.mov,.mkv,video/mp4,video/quicktime,video/x-matroska"
      class="hidden"
      @change="onSelect"
    />

    <div v-if="status === 'idle'" class="flex flex-col items-center gap-3 text-center">
      <div class="w-12 h-12 rounded-xl bg-[#2D2D2D] flex items-center justify-center">
        <Upload class="w-6 h-6 text-cyan-400" />
      </div>
      <div>
        <div class="text-white font-semibold">Перетащите видео или выберите файл</div>
        <div class="text-white/40 text-sm">mp4, mov, mkv до 500MB</div>
      </div>
      <button
        type="button"
        class="bg-cyan-400 hover:bg-cyan-500 text-black rounded-xl px-4 py-2 font-semibold"
        @click="pickFile"
      >
        Выбрать файл
      </button>
    </div>

    <div v-else-if="status === 'uploading'" class="space-y-3">
      <div class="flex items-center gap-2 text-white/80">
        <Loader2 class="w-4 h-4 text-cyan-400 animate-spin" />
        <span class="text-sm truncate">Загрузка: {{ fileName }}</span>
        <span class="ml-auto text-white/60 text-sm">{{ progress }}%</span>
      </div>
      <div class="h-2 rounded-full bg-white/10 overflow-hidden">
        <div class="h-full bg-cyan-400 transition-all" :style="{ width: progress + '%' }" />
      </div>
    </div>

    <div v-else-if="status === 'processing'" class="flex items-center gap-3 text-white/80">
      <Loader2 class="w-5 h-5 text-cyan-400 animate-spin" />
      <div>
        <div class="font-semibold text-white">Обработка видео…</div>
        <div class="text-white/40 text-sm">Конвертация в HLS может занять несколько минут</div>
      </div>
    </div>

    <div v-else-if="status === 'ready'" class="flex items-center gap-3">
      <CheckCircle2 class="w-5 h-5 text-cyan-400" />
      <div class="flex-1">
        <div class="text-white font-semibold">Готово</div>
        <div class="text-white/40 text-sm truncate">{{ fileName }}</div>
      </div>
      <button
        type="button"
        class="text-white/60 hover:text-white text-sm rounded-xl px-3 py-1.5 bg-white/5 hover:bg-white/10"
        @click="reset"
      >
        Загрузить ещё
      </button>
    </div>

    <div v-else-if="status === 'failed'" class="flex items-center gap-3">
      <XCircle class="w-5 h-5 text-red-400" />
      <div class="flex-1">
        <div class="text-white font-semibold">Ошибка</div>
        <div class="text-white/60 text-sm">{{ errorMsg }}</div>
      </div>
      <button
        type="button"
        class="text-white/60 hover:text-white text-sm rounded-xl px-3 py-1.5 bg-white/5 hover:bg-white/10"
        @click="reset"
      >
        Повторить
      </button>
    </div>
  </div>
</template>
