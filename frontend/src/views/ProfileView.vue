<script setup lang="ts">
import { useAuthStore } from '@/stores/auth'
import { computed, ref } from 'vue'
import { KeyRound, Mail, Shield, LogOut } from 'lucide-vue-next'
import { useRouter } from 'vue-router'
import { usersApi } from '@/api/client'

const auth = useAuthStore()
const router = useRouter()
const passwordForm = ref({ current_password: '', new_password: '', confirm_password: '' })
const passwordError = ref('')
const passwordSuccess = ref('')
const changingPassword = ref(false)

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

async function changePassword() {
  passwordError.value = ''
  passwordSuccess.value = ''
  if (!passwordForm.value.current_password || !passwordForm.value.new_password) {
    passwordError.value = 'Fill in current and new password'
    return
  }
  if (passwordForm.value.new_password.length < 6) {
    passwordError.value = 'New password must be at least 6 characters'
    return
  }
  if (passwordForm.value.new_password !== passwordForm.value.confirm_password) {
    passwordError.value = 'Passwords do not match'
    return
  }
  changingPassword.value = true
  try {
    await usersApi.changePassword({
      current_password: passwordForm.value.current_password,
      new_password: passwordForm.value.new_password,
    })
    passwordForm.value = { current_password: '', new_password: '', confirm_password: '' }
    passwordSuccess.value = 'Password changed'
  } catch (e: any) {
    if (e.response?.data?.error) {
      passwordError.value = e.response.data.error
    } else if (e.code === 'ERR_NETWORK' || e.message === 'Network Error') {
      passwordError.value = 'Cannot reach server. Try again.'
    } else {
      passwordError.value = `Failed to change password (HTTP ${e.response?.status ?? '?'})`
    }
  } finally {
    changingPassword.value = false
  }
}
</script>

<template>
  <div class="max-w-2xl">
    <h1 class="text-2xl font-extrabold text-cyan-400 mb-6">Profile</h1>

    <div class="flex items-center gap-5 mb-8">
      <div class="w-20 h-20 rounded-full bg-cyan-400/20 border-2 border-cyan-400/40 flex items-center justify-center">
        <span class="text-cyan-400 text-2xl font-extrabold">{{ initials }}</span>
      </div>
      <div>
        <h2 class="text-xl font-extrabold text-white">{{ auth.displayName }}</h2>
        <p class="text-white/50 text-sm">{{ auth.userEmail || '—' }}</p>
      </div>
    </div>

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

      <div class="bg-[#1E1E1E] rounded-2xl p-5 border border-white/5">
        <div class="flex items-center gap-3 mb-3">
          <KeyRound class="w-5 h-5 text-cyan-400" />
          <span class="text-sm font-bold text-white/70">Change Password</span>
        </div>
        <form class="space-y-3" @submit.prevent="changePassword">
          <p v-if="passwordError" class="rounded-xl bg-red-400/10 px-4 py-3 text-sm text-red-400">{{ passwordError }}</p>
          <p v-if="passwordSuccess" class="rounded-xl bg-green-400/10 px-4 py-3 text-sm text-green-400">{{ passwordSuccess }}</p>
          <input v-model="passwordForm.current_password" type="password" placeholder="Current password" class="w-full px-4 py-3 bg-[#2D2D2D] rounded-xl text-white text-sm placeholder-white/30 border border-transparent focus:border-cyan-400 focus:outline-none" />
          <input v-model="passwordForm.new_password" type="password" placeholder="New password" class="w-full px-4 py-3 bg-[#2D2D2D] rounded-xl text-white text-sm placeholder-white/30 border border-transparent focus:border-cyan-400 focus:outline-none" />
          <input v-model="passwordForm.confirm_password" type="password" placeholder="Confirm new password" class="w-full px-4 py-3 bg-[#2D2D2D] rounded-xl text-white text-sm placeholder-white/30 border border-transparent focus:border-cyan-400 focus:outline-none" />
          <button type="submit" :disabled="changingPassword" class="w-full py-3 rounded-xl bg-cyan-400 text-black font-semibold text-sm hover:bg-cyan-300 disabled:opacity-50 transition-colors">
            {{ changingPassword ? 'Saving...' : 'Save Password' }}
          </button>
        </form>
      </div>
    </div>

    <button
      @click="handleLogout"
      class="w-full flex items-center justify-center gap-3 py-4 rounded-xl bg-red-500/10 border border-red-500/20 text-red-400 font-bold hover:bg-red-500/20 transition-colors"
    >
      <LogOut class="w-5 h-5" />
      Sign Out
    </button>
  </div>
</template>
