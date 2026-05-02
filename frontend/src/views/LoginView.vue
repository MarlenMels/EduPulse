<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { Mail, Lock, Eye, EyeOff, ArrowLeft } from 'lucide-vue-next'

const router = useRouter()
const auth = useAuthStore()

const email = ref('')
const password = ref('')
const showPassword = ref(false)
const submitting = ref(false)
const errorMsg = ref('')

async function handleSubmit() {
  errorMsg.value = ''
  if (!email.value.trim() || !password.value.trim()) {
    errorMsg.value = 'Please fill in all fields'
    return
  }
  submitting.value = true
  try {
    await auth.login(email.value.trim(), password.value.trim())
    router.push('/dashboard')
  } catch (e: any) {
    if (e.response?.data?.error) {
      errorMsg.value = e.response.data.error
    } else if (e.code === 'ERR_NETWORK' || e.message === 'Network Error') {
      errorMsg.value = 'Cannot reach server. Check your connection and try again.'
    } else {
      errorMsg.value = 'Login failed. Please try again.'
    }
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <div class="min-h-screen bg-gradient-to-br from-[#121212] via-[#1E1E1E] to-[#121212] relative overflow-hidden">
    <div class="absolute -top-12 -right-12 w-48 h-48 rounded-full bg-white/5" />
    <div class="absolute -bottom-24 -left-12 w-60 h-60 rounded-full bg-white/5" />

    <div class="relative z-10 min-h-screen flex flex-col">
      <div class="flex items-center gap-4 px-6 py-4">
        <button
          @click="router.push('/')"
          class="w-10 h-10 flex items-center justify-center bg-white rounded-xl shadow-md hover:bg-gray-100 transition-colors"
        >
          <ArrowLeft class="w-4 h-4 text-black" />
        </button>
        <h1 class="text-2xl font-extrabold text-[#0A84FF]">Sign In</h1>
      </div>

      <div class="flex-1 flex flex-col justify-center px-6 max-w-md mx-auto w-full">
        <div class="animate-fade-in">
          <h2 class="text-2xl font-extrabold text-white leading-tight">Welcome Back</h2>
          <p class="mt-2 text-sm text-[#B0B0B0]">Sign in to continue</p>
        </div>

        <form @submit.prevent="handleSubmit" class="mt-8 space-y-5 animate-fade-in-delay">
          <div v-if="errorMsg" class="p-3 rounded-xl bg-red-500/10 border border-red-500/30 text-red-400 text-sm">
            {{ errorMsg }}
          </div>

          <div>
            <label class="block text-sm font-semibold text-white mb-2 pl-1">Email</label>
            <div class="relative">
              <Mail class="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-[#8E8E93]" />
              <input
                v-model="email"
                type="email"
                placeholder="example@email.com"
                class="w-full pl-12 pr-4 py-4 bg-[#2D2D2D] rounded-2xl text-white text-sm placeholder-[#8E8E93] border-[1.5px] border-[#0A84FF]/70 focus:border-[#0A84FF] focus:outline-none transition-colors"
              />
            </div>
          </div>

          <div>
            <label class="block text-sm font-semibold text-white mb-2 pl-1">Password</label>
            <div class="relative">
              <Lock class="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-[#8E8E93]" />
              <input
                v-model="password"
                :type="showPassword ? 'text' : 'password'"
                placeholder="••••••••"
                class="w-full pl-12 pr-12 py-4 bg-[#2D2D2D] rounded-2xl text-white text-sm placeholder-[#8E8E93] border-[1.5px] border-[#0A84FF]/70 focus:border-[#0A84FF] focus:outline-none transition-colors"
                @keyup.enter="handleSubmit"
              />
              <button
                type="button"
                @click="showPassword = !showPassword"
                class="absolute right-4 top-1/2 -translate-y-1/2 text-[#8E8E93] hover:text-white transition-colors"
              >
                <EyeOff v-if="showPassword" class="w-5 h-5" />
                <Eye v-else class="w-5 h-5" />
              </button>
            </div>
          </div>

          <button
            type="submit"
            :disabled="submitting"
            class="w-full py-4 rounded-2xl bg-[#0A84FF] text-white font-bold text-base hover:bg-[#0A84FF]/90 disabled:opacity-50 disabled:cursor-not-allowed transition-all"
          >
            <span v-if="submitting" class="inline-block w-5 h-5 border-2 border-white/30 border-t-white rounded-full animate-spin" />
            <span v-else>Sign In</span>
          </button>

          <p class="text-center text-sm text-white/50 pt-2">
            Don't have an account?
            <button type="button" @click="router.push('/register')" class="text-cyan-400 font-semibold hover:underline ml-1">
              Sign Up
            </button>
          </p>
        </form>

      </div>
    </div>
  </div>
</template>

<style scoped>
.animate-fade-in {
  animation: fadeSlideIn 0.8s ease-out both;
}
.animate-fade-in-delay {
  animation: fadeSlideIn 0.8s ease-out 0.2s both;
}
@keyframes fadeSlideIn {
  from {
    opacity: 0;
    transform: translateY(16px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}
</style>
