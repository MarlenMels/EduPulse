<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { coursesApi, sessionsApi, notificationsApi } from '@/api/client'
import { GraduationCap, CalendarDays, Bell, TrendingUp } from 'lucide-vue-next'

const auth = useAuthStore()

const stats = ref({
  courses: 0,
  sessions: 0,
  notifications: 0,
})
const recentSessions = ref<any[]>([])
const loading = ref(true)

onMounted(async () => {
  try {
    const [courseRes, sessionRes] = await Promise.all([
      coursesApi.list({ limit: 100 }),
      sessionsApi.list({ limit: 10 }),
    ])
    stats.value.courses = courseRes.data.count || 0
    stats.value.sessions = sessionRes.data.count || 0
    recentSessions.value = sessionRes.data.items || []

    if (auth.isAdmin || auth.isManager) {
      const notifRes = await notificationsApi.list({ limit: 100 })
      stats.value.notifications = notifRes.data.count || 0
    }
  } catch {
    // silently fail
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div>
    <div class="mb-8">
      <h1 class="text-2xl md:text-3xl font-extrabold text-white">
        Hello, <span class="text-cyan-400">{{ auth.displayName }}</span>
      </h1>
      <p class="mt-1 text-white/50 text-sm">Welcome to the EDpulse dashboard</p>
    </div>

    <!-- Stats -->
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4 mb-8">
      <div class="bg-[#1E1E1E] rounded-2xl p-5 border border-white/5">
        <div class="flex items-center gap-4">
          <div class="w-12 h-12 rounded-xl bg-cyan-400/10 flex items-center justify-center">
            <GraduationCap class="w-6 h-6 text-cyan-400" />
          </div>
          <div>
            <p class="text-2xl font-extrabold text-white">{{ loading ? '...' : stats.courses }}</p>
            <p class="text-sm text-white/50">Courses</p>
          </div>
        </div>
      </div>

      <div class="bg-[#1E1E1E] rounded-2xl p-5 border border-white/5">
        <div class="flex items-center gap-4">
          <div class="w-12 h-12 rounded-xl bg-blue-400/10 flex items-center justify-center">
            <CalendarDays class="w-6 h-6 text-blue-400" />
          </div>
          <div>
            <p class="text-2xl font-extrabold text-white">{{ loading ? '...' : stats.sessions }}</p>
            <p class="text-sm text-white/50">Sessions</p>
          </div>
        </div>
      </div>

      <div
        v-if="auth.isAdmin || auth.isManager"
        class="bg-[#1E1E1E] rounded-2xl p-5 border border-white/5"
      >
        <div class="flex items-center gap-4">
          <div class="w-12 h-12 rounded-xl bg-purple-400/10 flex items-center justify-center">
            <Bell class="w-6 h-6 text-purple-400" />
          </div>
          <div>
            <p class="text-2xl font-extrabold text-white">{{ loading ? '...' : stats.notifications }}</p>
            <p class="text-sm text-white/50">Notifications</p>
          </div>
        </div>
      </div>
    </div>

    <!-- Recent sessions -->
    <div class="bg-[#1E1E1E] rounded-2xl border border-white/5 overflow-hidden">
      <div class="flex items-center justify-between px-5 py-4 border-b border-white/5">
        <h2 class="text-lg font-bold text-cyan-400 flex items-center gap-2">
          <TrendingUp class="w-5 h-5" />
          Recent Sessions
        </h2>
      </div>

      <div v-if="loading" class="p-8 flex justify-center">
        <div class="w-8 h-8 border-2 border-cyan-400/30 border-t-cyan-400 rounded-full animate-spin" />
      </div>

      <div v-else-if="recentSessions.length === 0" class="p-8 text-center text-white/40">
        No sessions yet
      </div>

      <div v-else class="divide-y divide-white/5">
        <div
          v-for="session in recentSessions"
          :key="session.id"
          class="flex items-center gap-4 px-5 py-4 hover:bg-white/5 transition-colors cursor-pointer"
          @click="$router.push(`/sessions/${session.id}`)"
        >
          <div class="w-10 h-10 rounded-lg bg-blue-400/10 flex items-center justify-center shrink-0">
            <CalendarDays class="w-5 h-5 text-blue-400" />
          </div>
          <div class="flex-1 min-w-0">
            <p class="text-sm font-bold text-white truncate">{{ session.title }}</p>
            <p class="text-xs text-white/40">ID: {{ session.id }}</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
