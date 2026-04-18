<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { sessionsApi } from '@/api/client'
import { useAuthStore } from '@/stores/auth'
import { CalendarDays, Plus, Clock, X } from 'lucide-vue-next'

const auth = useAuthStore()
const sessions = ref<any[]>([])
const loading = ref(true)
const error = ref('')
const showCreateModal = ref(false)
const creating = ref(false)

const newSession = ref({
  title: '',
  start_time: '',
})

async function fetchSessions() {
  loading.value = true
  error.value = ''
  try {
    const res = await sessionsApi.list({ limit: 50 })
    sessions.value = res.data.items || []
  } catch (e: any) {
    error.value = e.response?.data?.error || 'Failed to load'
  } finally {
    loading.value = false
  }
}

async function createSession() {
  if (!newSession.value.title.trim() || !newSession.value.start_time) return
  creating.value = true
  try {
    await sessionsApi.create({
      title: newSession.value.title,
      start_time: new Date(newSession.value.start_time).toISOString(),
    })
    showCreateModal.value = false
    newSession.value = { title: '', start_time: '' }
    await fetchSessions()
  } catch (e: any) {
    error.value = e.response?.data?.error || 'Failed to create'
  } finally {
    creating.value = false
  }
}

function formatDate(dateStr: string) {
  if (!dateStr) return '—'
  return new Date(dateStr).toLocaleString('en-US', {
    day: '2-digit',
    month: 'short',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

onMounted(fetchSessions)
</script>

<template>
  <div>
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-6">
      <h1 class="text-2xl font-extrabold text-cyan-400">Sessions</h1>
      <button
        v-if="auth.isAdmin || auth.isManager || auth.isTeacher"
        @click="showCreateModal = true"
        class="flex items-center gap-2 px-4 py-2.5 bg-cyan-400 text-black font-semibold text-sm rounded-xl hover:bg-cyan-300 transition-colors"
      >
        <Plus class="w-4 h-4" />
        New Session
      </button>
    </div>

    <div v-if="loading" class="flex justify-center py-16">
      <div class="w-8 h-8 border-2 border-cyan-400/30 border-t-cyan-400 rounded-full animate-spin" />
    </div>

    <div v-else-if="error" class="text-center py-16">
      <p class="text-red-400 mb-4">{{ error }}</p>
      <button @click="fetchSessions" class="px-4 py-2 bg-cyan-400 text-black rounded-lg font-semibold text-sm">
        Retry
      </button>
    </div>

    <div v-else-if="sessions.length === 0" class="text-center py-16 text-white/40">
      <CalendarDays class="w-16 h-16 mx-auto mb-4 opacity-30" />
      <p>No sessions yet</p>
    </div>

    <div v-else class="space-y-3">
      <div
        v-for="session in sessions"
        :key="session.id"
        class="bg-[#1E1E1E] rounded-2xl p-5 border border-white/5 hover:border-cyan-400/30 transition-all"
      >
        <div class="flex items-start gap-4">
          <div class="w-11 h-11 rounded-xl bg-blue-400/10 flex items-center justify-center shrink-0">
            <CalendarDays class="w-5 h-5 text-blue-400" />
          </div>
          <div class="flex-1 min-w-0">
            <h3 class="font-bold text-white">{{ session.title }}</h3>
            <div class="flex flex-wrap items-center gap-3 mt-2 text-xs text-white/40">
              <span class="flex items-center gap-1">
                <Clock class="w-3 h-3" />
                {{ formatDate(session.start_time) }}
              </span>
              <span>Teacher: {{ session.teacher_id || '—' }}</span>
            </div>
          </div>
          <span class="text-xs text-white/30 shrink-0">#{{ session.id }}</span>
        </div>
      </div>
    </div>

    <!-- Create modal -->
    <Teleport to="body">
      <Transition name="fade">
        <div v-if="showCreateModal" class="fixed inset-0 z-50 flex items-end sm:items-center justify-center bg-black/60 p-4" @click.self="showCreateModal = false">
          <div class="bg-[#1E1E1E] rounded-t-3xl sm:rounded-2xl w-full max-w-md p-6">
            <div class="flex items-center justify-between mb-6">
              <h2 class="text-lg font-extrabold text-white">New Session</h2>
              <button @click="showCreateModal = false" class="text-white/40 hover:text-white">
                <X class="w-5 h-5" />
              </button>
            </div>

            <form @submit.prevent="createSession" class="space-y-4">
              <div>
                <label class="block text-sm font-semibold text-white/70 mb-1.5">Title</label>
                <input
                  v-model="newSession.title"
                  placeholder="Session title"
                  class="w-full px-4 py-3 bg-[#2D2D2D] rounded-xl text-white text-sm placeholder-white/30 border border-transparent focus:border-cyan-400 focus:outline-none"
                />
              </div>
              <div>
                <label class="block text-sm font-semibold text-white/70 mb-1.5">Date & Time</label>
                <input
                  v-model="newSession.start_time"
                  type="datetime-local"
                  class="w-full px-4 py-3 bg-[#2D2D2D] rounded-xl text-white text-sm border border-transparent focus:border-cyan-400 focus:outline-none"
                />
              </div>
              <div class="flex gap-3 pt-2">
                <button type="button" @click="showCreateModal = false" class="flex-1 py-3 rounded-xl text-white/60 font-semibold text-sm hover:bg-white/5 transition-colors">
                  Cancel
                </button>
                <button type="submit" :disabled="creating" class="flex-1 py-3 rounded-xl bg-cyan-400 text-black font-semibold text-sm hover:bg-cyan-300 disabled:opacity-50 transition-colors">
                  {{ creating ? 'Creating...' : 'Create' }}
                </button>
              </div>
            </form>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<style scoped>
.fade-enter-active, .fade-leave-active { transition: opacity 0.2s ease; }
.fade-enter-from, .fade-leave-to { opacity: 0; }
</style>
