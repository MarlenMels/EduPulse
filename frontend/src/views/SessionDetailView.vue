<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { sessionsApi } from '@/api/client'
import { CalendarDays, ArrowLeft, Users, Clock } from 'lucide-vue-next'

const route = useRoute()
const router = useRouter()

const session = ref<any>(null)
const loading = ref(true)
const error = ref('')

onMounted(async () => {
  const sessionId = Number(route.params.id)
  if (isNaN(sessionId)) {
    error.value = 'Invalid session ID'
    loading.value = false
    return
  }

  try {
    const res = await sessionsApi.get(sessionId)
    session.value = res.data
  } catch (e: any) {
    error.value = e.response?.data?.error || 'Failed to load session'
  } finally {
    loading.value = false
  }
})

function goBack() {
  router.push('/sessions')
}
</script>

<template>
  <div>
    <div class="flex items-center gap-4 mb-6">
      <button
        @click="goBack"
        class="flex items-center gap-2 px-3 py-2 bg-white/10 rounded-lg text-white/70 hover:bg-white/20 transition-colors"
      >
        <ArrowLeft class="w-4 h-4" />
        Back
      </button>
      <h1 class="text-2xl font-extrabold text-white">Session Details</h1>
    </div>

    <div v-if="loading" class="flex justify-center py-16">
      <div class="w-8 h-8 border-2 border-cyan-400/30 border-t-cyan-400 rounded-full animate-spin" />
    </div>

    <div v-else-if="error" class="text-center py-16">
      <p class="text-red-400 mb-4">{{ error }}</p>
      <button @click="goBack" class="px-4 py-2 bg-cyan-400 text-black rounded-lg font-semibold text-sm">Go Back</button>
    </div>

    <div v-else-if="session" class="bg-[#1E1E1E] rounded-2xl p-6 border border-white/5">
      <div class="flex items-start gap-4">
        <div class="w-12 h-12 rounded-xl bg-blue-400/10 flex items-center justify-center shrink-0">
          <CalendarDays class="w-6 h-6 text-blue-400" />
        </div>
        <div class="flex-1">
          <h2 class="text-xl font-bold text-white mb-2">{{ session.title }}</h2>
          <div class="space-y-2 text-sm text-white/60">
            <div class="flex items-center gap-2">
              <Users class="w-4 h-4" />
              <span>Course ID: {{ session.course_id }}</span>
            </div>
            <div class="flex items-center gap-2">
              <Clock class="w-4 h-4" />
              <span>Start: {{ new Date(session.start_time).toLocaleString() }}</span>
            </div>
            <div class="flex items-center gap-2">
              <CalendarDays class="w-4 h-4" />
              <span>Created: {{ new Date(session.created_at).toLocaleString() }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
