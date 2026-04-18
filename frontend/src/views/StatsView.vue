<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { statsApi, type RoleCount } from '@/api/client'
import { Users, UserCheck, BarChart3, RefreshCw } from 'lucide-vue-next'

const roles = ref<RoleCount[]>([])
const totalUsers = ref(0)
const totalOnline = ref(0)
const loading = ref(true)
const error = ref('')
let interval: ReturnType<typeof setInterval> | null = null

const roleConfig: Record<string, { label: string; color: string; bg: string }> = {
  admin: { label: 'Admins', color: 'text-red-400', bg: 'bg-red-400/10' },
  manager: { label: 'Managers', color: 'text-purple-400', bg: 'bg-purple-400/10' },
  teacher: { label: 'Teachers', color: 'text-blue-400', bg: 'bg-blue-400/10' },
  student: { label: 'Students', color: 'text-cyan-400', bg: 'bg-cyan-400/10' },
  parent: { label: 'Parents', color: 'text-green-400', bg: 'bg-green-400/10' },
}

const sortedRoles = computed(() =>
  [...roles.value].sort((a, b) => {
    const order = ['admin', 'manager', 'teacher', 'student', 'parent']
    return order.indexOf(a.role) - order.indexOf(b.role)
  }),
)

async function fetchStats() {
  try {
    const res = await statsApi.get()
    roles.value = res.data.roles || []
    totalUsers.value = res.data.total_users
    totalOnline.value = res.data.total_online
    error.value = ''
  } catch (e: any) {
    error.value = e.response?.data?.error || 'Failed to load stats'
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchStats()
  interval = setInterval(fetchStats, 30000)
})

onUnmounted(() => {
  if (interval) clearInterval(interval)
})
</script>

<template>
  <div>
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-6">
      <h1 class="text-2xl font-extrabold text-cyan-400">Statistics</h1>
      <button
        @click="fetchStats"
        :disabled="loading"
        class="flex items-center gap-2 px-4 py-2.5 bg-[#1E1E1E] border border-white/10 text-white/60 font-semibold text-sm rounded-xl hover:bg-white/5 hover:text-white transition-colors disabled:opacity-50"
      >
        <RefreshCw class="w-4 h-4" :class="{ 'animate-spin': loading }" />
        Refresh
      </button>
    </div>

    <div v-if="loading && roles.length === 0" class="flex justify-center py-16">
      <div class="w-8 h-8 border-2 border-cyan-400/30 border-t-cyan-400 rounded-full animate-spin" />
    </div>

    <div v-else-if="error" class="text-center py-16">
      <p class="text-red-400 mb-4">{{ error }}</p>
      <button @click="fetchStats" class="px-4 py-2 bg-cyan-400 text-black rounded-lg font-semibold text-sm">
        Retry
      </button>
    </div>

    <template v-else>
      <!-- Summary cards -->
      <div class="grid grid-cols-1 sm:grid-cols-2 gap-4 mb-8">
        <div class="bg-[#1E1E1E] rounded-2xl p-5 border border-white/5">
          <div class="flex items-center gap-4">
            <div class="w-12 h-12 rounded-xl bg-cyan-400/10 flex items-center justify-center">
              <Users class="w-6 h-6 text-cyan-400" />
            </div>
            <div>
              <p class="text-3xl font-extrabold text-white">{{ totalUsers }}</p>
              <p class="text-sm text-white/50">Total Registered</p>
            </div>
          </div>
        </div>

        <div class="bg-[#1E1E1E] rounded-2xl p-5 border border-white/5">
          <div class="flex items-center gap-4">
            <div class="w-12 h-12 rounded-xl bg-green-400/10 flex items-center justify-center">
              <UserCheck class="w-6 h-6 text-green-400" />
            </div>
            <div>
              <p class="text-3xl font-extrabold text-white">{{ totalOnline }}</p>
              <p class="text-sm text-white/50">Online Now</p>
            </div>
          </div>
        </div>
      </div>

      <!-- Role breakdown -->
      <div class="bg-[#1E1E1E] rounded-2xl border border-white/5 overflow-hidden">
        <div class="flex items-center gap-2 px-5 py-4 border-b border-white/5">
          <BarChart3 class="w-5 h-5 text-cyan-400" />
          <h2 class="text-lg font-bold text-cyan-400">Users by Role</h2>
        </div>

        <div v-if="sortedRoles.length === 0" class="p-8 text-center text-white/40">
          No users registered yet
        </div>

        <div v-else class="divide-y divide-white/5">
          <div
            v-for="rc in sortedRoles"
            :key="rc.role"
            class="flex items-center gap-4 px-5 py-4"
          >
            <div
              class="w-11 h-11 rounded-xl flex items-center justify-center shrink-0"
              :class="roleConfig[rc.role]?.bg ?? 'bg-white/10'"
            >
              <Users class="w-5 h-5" :class="roleConfig[rc.role]?.color ?? 'text-white/60'" />
            </div>

            <div class="flex-1 min-w-0">
              <div class="flex items-center justify-between mb-2">
                <span class="text-sm font-bold text-white">
                  {{ roleConfig[rc.role]?.label ?? rc.role }}
                </span>
                <span class="text-sm text-white/50">
                  {{ rc.total }} registered
                </span>
              </div>
              <!-- Progress bar -->
              <div class="h-2 bg-white/5 rounded-full overflow-hidden">
                <div
                  class="h-full rounded-full transition-all duration-500"
                  :class="{
                    'bg-red-400': rc.role === 'admin',
                    'bg-purple-400': rc.role === 'manager',
                    'bg-blue-400': rc.role === 'teacher',
                    'bg-cyan-400': rc.role === 'student',
                    'bg-green-400': rc.role === 'parent',
                    'bg-white/30': !roleConfig[rc.role],
                  }"
                  :style="{ width: totalUsers ? (rc.total / totalUsers * 100) + '%' : '0%' }"
                />
              </div>
            </div>

            <div class="text-right shrink-0 ml-2">
              <span
                class="inline-flex items-center gap-1 px-2.5 py-1 rounded-lg text-xs font-bold"
                :class="rc.online > 0 ? 'bg-green-400/10 text-green-400' : 'bg-white/5 text-white/30'"
              >
                <span class="w-1.5 h-1.5 rounded-full" :class="rc.online > 0 ? 'bg-green-400' : 'bg-white/20'" />
                {{ rc.online }} online
              </span>
            </div>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>
