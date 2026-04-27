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
const socialToast = ref('')

function handleSocialLogin(provider: string) {
  socialToast.value = `${provider} sign-in coming soon`
  setTimeout(() => (socialToast.value = ''), 2500)
}

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
    <!-- Decorative circles -->
    <div class="absolute -top-12 -right-12 w-48 h-48 rounded-full bg-white/5" />
    <div class="absolute -bottom-24 -left-12 w-60 h-60 rounded-full bg-white/5" />

    <div class="relative z-10 min-h-screen flex flex-col">
      <!-- Header -->
      <div class="flex items-center gap-4 px-6 py-4">
        <button
          @click="router.push('/')"
          class="w-10 h-10 flex items-center justify-center bg-white rounded-xl shadow-md hover:bg-gray-100 transition-colors"
        >
          <ArrowLeft class="w-4 h-4 text-black" />
        </button>
        <h1 class="text-2xl font-extrabold text-[#0A84FF]">Sign In</h1>
      </div>

      <!-- Content -->
      <div class="flex-1 flex flex-col justify-center px-6 max-w-md mx-auto w-full">
        <div class="animate-fade-in">
          <h2 class="text-2xl font-extrabold text-white leading-tight">Welcome Back</h2>
          <p class="mt-2 text-sm text-[#B0B0B0]">Sign in to continue</p>
        </div>

        <form @submit.prevent="handleSubmit" class="mt-8 space-y-5 animate-fade-in-delay">
          <!-- Error -->
          <div v-if="errorMsg" class="p-3 rounded-xl bg-red-500/10 border border-red-500/30 text-red-400 text-sm">
            {{ errorMsg }}
          </div>

          <!-- Email -->
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

          <!-- Password -->
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

          <!-- Submit -->
          <button
            type="submit"
            :disabled="submitting"
            class="w-full py-4 rounded-2xl bg-[#0A84FF] text-white font-bold text-base hover:bg-[#0A84FF]/90 disabled:opacity-50 disabled:cursor-not-allowed transition-all"
          >
            <span v-if="submitting" class="inline-block w-5 h-5 border-2 border-white/30 border-t-white rounded-full animate-spin" />
            <span v-else>Sign In</span>
          </button>

          <!-- Divider -->
          <div class="flex items-center gap-4 pt-1">
            <div class="flex-1 h-px bg-white/10" />
            <span class="text-xs text-white/30 font-medium">OR</span>
            <div class="flex-1 h-px bg-white/10" />
          </div>

          <!-- Social buttons -->
          <div class="flex gap-3">
            <button
              type="button"
              @click="handleSocialLogin('Google')"
              class="flex-1 flex items-center justify-center gap-3 py-3.5 rounded-2xl bg-white hover:bg-gray-100 transition-colors"
            >
              <svg class="w-5 h-5" viewBox="0 0 24 24">
                <path d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92a5.06 5.06 0 0 1-2.2 3.32v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.1z" fill="#4285F4"/>
                <path d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z" fill="#34A853"/>
                <path d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z" fill="#FBBC05"/>
                <path d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z" fill="#EA4335"/>
              </svg>
              <span class="text-sm font-semibold text-gray-700">Google</span>
            </button>

            <button
              type="button"
              @click="handleSocialLogin('Apple')"
              class="flex-1 flex items-center justify-center gap-3 py-3.5 rounded-2xl bg-white hover:bg-gray-100 transition-colors"
            >
              <svg class="w-5 h-5" viewBox="0 0 24 24" fill="black">
                <path d="M17.05 20.28c-.98.95-2.05.88-3.08.4-1.09-.5-2.08-.48-3.24 0-1.44.62-2.2.44-3.06-.4C2.79 15.25 3.51 7.59 9.05 7.31c1.35.07 2.29.74 3.08.8 1.18-.24 2.31-.93 3.57-.84 1.51.12 2.65.72 3.4 1.8-3.12 1.87-2.38 5.98.48 7.13-.57 1.5-1.31 2.99-2.54 4.09zM12.03 7.25c-.15-2.23 1.66-4.07 3.74-4.25.29 2.58-2.34 4.5-3.74 4.25z"/>
              </svg>
              <span class="text-sm font-semibold text-gray-700">Apple</span>
            </button>
          </div>

          <!-- Register link -->
          <p class="text-center text-sm text-white/50 pt-2">
            Don't have an account?
            <button type="button" @click="router.push('/register')" class="text-cyan-400 font-semibold hover:underline ml-1">
              Sign Up
            </button>
          </p>
        </form>

        <!-- Social toast -->
        <Transition name="toast">
          <div
            v-if="socialToast"
            class="fixed bottom-6 left-1/2 -translate-x-1/2 px-5 py-3 bg-[#2D2D2D] border border-white/10 rounded-2xl text-white text-sm font-semibold shadow-lg z-50"
          >
            {{ socialToast }}
          </div>
        </Transition>
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
.toast-enter-active { transition: all 0.3s ease; }
.toast-leave-active { transition: all 0.3s ease; }
.toast-enter-from { opacity: 0; transform: translate(-50%, 16px); }
.toast-leave-to { opacity: 0; transform: translate(-50%, 16px); }
</style>
