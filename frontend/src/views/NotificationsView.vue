<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { notificationsApi } from '@/api/client'
import { Bell } from 'lucide-vue-next'

const notifications = ref<any[]>([])
const loading = ref(true)
const error = ref('')

async function fetchNotifications() {
  loading.value = true
  error.value = ''
  try {
    const res = await notificationsApi.list({ limit: 50 })
    notifications.value = res.data.items || []
  } catch (e: any) {
    error.value = e.response?.data?.error || 'Failed to load'
  } finally {
    loading.value = false
  }
}

function formatDate(dateStr: string) {
  if (!dateStr) return '—'
  return new Date(dateStr).toLocaleString('en-US', {
    day: '2-digit', month: 'short', year: 'numeric', hour: '2-digit', minute: '2-digit',
  })
}

onMounted(fetchNotifications)
</script>

<template>
  <div>
    <h1 class="text-2xl font-extrabold text-cyan-400 mb-6">Notifications</h1>

    <div v-if="loading" class="flex justify-center py-16">
      <div class="w-8 h-8 border-2 border-cyan-400/30 border-t-cyan-400 rounded-full animate-spin" />
    </div>

    <div v-else-if="error" class="text-center py-16">
      <p class="text-red-400 mb-4">{{ error }}</p>
      <button @click="fetchNotifications" class="px-4 py-2 bg-cyan-400 text-black rounded-lg font-semibold text-sm">Retry</button>
    </div>

    <div v-else-if="notifications.length === 0" class="text-center py-16 text-white/40">
      <Bell class="w-16 h-16 mx-auto mb-4 opacity-30" />
      <p>No notifications yet</p>
    </div>

    <div v-else class="space-y-3">
      <div
        v-for="notif in notifications"
        :key="notif.id"
        class="bg-[#1E1E1E] rounded-2xl p-5 border border-white/5"
      >
        <div class="flex items-start gap-4">
          <div class="w-10 h-10 rounded-xl bg-purple-400/10 flex items-center justify-center shrink-0">
            <Bell class="w-5 h-5 text-purple-400" />
          </div>
          <div class="flex-1 min-w-0">
            <p class="text-sm font-bold text-white">{{ notif.message || notif.title || `Notification #${notif.id}` }}</p>
            <p class="text-xs text-white/40 mt-1">{{ formatDate(notif.created_at) }}</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
