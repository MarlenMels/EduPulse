<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { auditApi } from '@/api/client'
import { ShieldCheck } from 'lucide-vue-next'

const logs = ref<any[]>([])
const loading = ref(true)
const error = ref('')

async function fetchLogs() {
  loading.value = true
  error.value = ''
  try {
    const res = await auditApi.list({ limit: 50 })
    logs.value = res.data.items || []
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

onMounted(fetchLogs)
</script>

<template>
  <div>
    <h1 class="text-2xl font-extrabold text-cyan-400 mb-6">Audit Logs</h1>

    <div v-if="loading" class="flex justify-center py-16">
      <div class="w-8 h-8 border-2 border-cyan-400/30 border-t-cyan-400 rounded-full animate-spin" />
    </div>

    <div v-else-if="error" class="text-center py-16">
      <p class="text-red-400 mb-4">{{ error }}</p>
      <button @click="fetchLogs" class="px-4 py-2 bg-cyan-400 text-black rounded-lg font-semibold text-sm">Retry</button>
    </div>

    <div v-else-if="logs.length === 0" class="text-center py-16 text-white/40">
      <ShieldCheck class="w-16 h-16 mx-auto mb-4 opacity-30" />
      <p>No logs yet</p>
    </div>

    <div v-else class="bg-[#1E1E1E] rounded-2xl border border-white/5 overflow-hidden">
      <div class="overflow-x-auto">
        <table class="w-full text-sm">
          <thead>
            <tr class="border-b border-white/5">
              <th class="px-5 py-3 text-left text-white/50 font-semibold">ID</th>
              <th class="px-5 py-3 text-left text-white/50 font-semibold">Action</th>
              <th class="px-5 py-3 text-left text-white/50 font-semibold">User</th>
              <th class="px-5 py-3 text-left text-white/50 font-semibold">Date</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-white/5">
            <tr v-for="log in logs" :key="log.id" class="hover:bg-white/5 transition-colors">
              <td class="px-5 py-3 text-white/40">{{ log.id }}</td>
              <td class="px-5 py-3 text-white font-semibold">{{ log.action }}</td>
              <td class="px-5 py-3 text-white/60">{{ log.user_id }}</td>
              <td class="px-5 py-3 text-white/40">{{ formatDate(log.created_at) }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>
