<script setup lang="ts">
import { RouterView } from 'vue-router'
import { Eye } from 'lucide-vue-next'
import { useUiStore } from '@/stores/ui'

const ui = useUiStore()
</script>

<template>
  <RouterView />
  <button
    type="button"
    class="vision-toggle fixed right-4 bottom-4 z-[70] flex h-12 min-w-12 items-center justify-center gap-2 rounded-xl border border-cyan-400/40 bg-[#1E1E1E] px-3 font-extrabold text-cyan-400 shadow-lg transition-colors hover:bg-cyan-400 hover:text-black"
    :class="{ active: ui.accessibleVision }"
    :aria-pressed="ui.accessibleVision"
    :title="ui.accessibleVision ? 'Disable accessible vision mode' : 'Enable accessible vision mode'"
    @click="ui.toggleAccessibleVision"
  >
    <Eye class="h-5 w-5" />
    A+
  </button>
</template>
