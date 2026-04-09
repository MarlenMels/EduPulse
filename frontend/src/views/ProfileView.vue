<script setup lang="ts">
import { useAuthStore } from '@/stores/auth'
import { computed } from 'vue'
import { User, Mail, Shield, LogOut } from 'lucide-vue-next'
import { useRouter } from 'vue-router'

const auth = useAuthStore()
const router = useRouter()

const initials = computed(() => {
  const name = auth.displayName
  const parts = name.split(/[\s_.-]+/)
  if (parts.length === 1) return parts[0]!.charAt(0).toUpperCase()
  return (parts[0]!.charAt(0) + parts[1]!.charAt(0)).toUpperCase()
})

function handleLogout() {
  auth.logout()
  router.push('/')
}
</script>

<template>
  <div class="max-w-2xl">
    <h1 class="text-2xl font-extrabold text-cyan-400 mb-6">Profile</h1>

    <!-- Avatar + name -->
    <div class="flex items-center gap-5 mb-8">
      <div class="w-20 h-20 rounded-full bg-cyan-400/20 border-2 border-cyan-400/40 flex items-center justify-center">
        <span class="text-cyan-400 text-2xl font-extrabold">{{ initials }}</span>
      </div>
      <div>
        <h2 class="text-xl font-extrabold text-white">{{ auth.displayName }}</h2>
        <p class="text-white/50 text-sm">{{ auth.userEmail || '—' }}</p>
      </div>
    </div>

    <!-- Info cards -->
    <div class="space-y-3 mb-8">
      <div class="bg-[#1E1E1E] rounded-2xl p-5 border border-white/5">
        <div class="flex items-center gap-3 mb-3">
          <Mail class="w-5 h-5 text-cyan-400" />
          <span class="text-sm font-bold text-white/70">Email</span>
        </div>
        <p class="text-white font-bold">{{ auth.userEmail || '—' }}</p>
      </div>

      <div class="bg-[#1E1E1E] rounded-2xl p-5 border border-white/5">
        <div class="flex items-center gap-3 mb-3">
          <Shield class="w-5 h-5 text-cyan-400" />
          <span class="text-sm font-bold text-white/70">Role</span>
        </div>
        <div class="flex items-center gap-2">
          <span class="px-3 py-1 rounded-full text-xs font-bold bg-cyan-400/10 text-cyan-400 border border-cyan-400/30">
            {{ auth.userRole || 'Not assigned' }}
          </span>
        </div>
      </div>

      <div v-if="auth.user" class="bg-[#1E1E1E] rounded-2xl p-5 border border-white/5">
        <div class="flex items-center gap-3 mb-3">
          <User class="w-5 h-5 text-cyan-400" />
          <span class="text-sm font-bold text-white/70">User ID</span>
        </div>
        <p class="text-white font-bold">{{ auth.user.id }}</p>
      </div>
    </div>

    <!-- Logout -->
    <button
      @click="handleLogout"
      class="w-full flex items-center justify-center gap-3 py-4 rounded-xl bg-red-500/10 border border-red-500/20 text-red-400 font-bold hover:bg-red-500/20 transition-colors"
    >
      <LogOut class="w-5 h-5" />
      Sign Out
    </button>
  </div>
</template>
