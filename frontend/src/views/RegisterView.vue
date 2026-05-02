<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { Mail, Lock, Eye, EyeOff, ArrowLeft, UserPlus } from 'lucide-vue-next'
 
const router = useRouter()
const auth = useAuthStore()
 
const email = ref('')
const password = ref('')
const confirmPassword = ref('')
const role = ref('student')
const showPassword = ref(false)
const showConfirm = ref(false)
const submitting = ref(false)
const errorMsg = ref('')

const roles = [
  { value: 'student', label: 'Student' },
  { value: 'parent', label: 'Parent' },
]
 
async function handleSubmit() {
  errorMsg.value = ''
 
  if (!email.value.trim() || !password.value.trim() || !confirmPassword.value.trim()) {
    errorMsg.value = 'Please fill in all required fields'
    return
  }
  if (password.value.length < 6) {
    errorMsg.value = 'Password must be at least 6 characters'
    return
  }
  if (password.value !== confirmPassword.value) {
    errorMsg.value = 'Passwords do not match'
    return
  }
 
  submitting.value = true
  try {
    await auth.register(email.value.trim(), password.value.trim(), role.value)
    router.push('/dashboard')
  } catch (e: any) {
    errorMsg.value = e.response?.data?.error || 'Registration failed'
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
        <h1 class="text-2xl font-extrabold text-[#0A84FF]">Sign Up</h1>
      </div>
 
      <div class="flex-1 flex flex-col justify-center px-6 max-w-md mx-auto w-full">
        <div class="animate-fade-in">
          <h2 class="text-2xl font-extrabold text-white leading-tight">Create Account</h2>
          <p class="mt-2 text-sm text-[#B0B0B0]">Fill in your details to get started</p>
        </div>
 
        <form @submit.prevent="handleSubmit" class="mt-8 space-y-4 animate-fade-in-delay">
          <div v-if="errorMsg" class="p-3 rounded-xl bg-red-500/10 border border-red-500/30 text-red-400 text-sm">
            {{ errorMsg }}
          </div>
 
          <div>
            <label class="block text-sm font-semibold text-white mb-2 pl-1">Email *</label>
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
            <label class="block text-sm font-semibold text-white mb-2 pl-1">Role *</label>
            <div class="relative">
              <UserPlus class="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-[#8E8E93]" />
              <select
                v-model="role"
                class="w-full pl-12 pr-4 py-4 bg-[#2D2D2D] rounded-2xl text-white text-sm border-[1.5px] border-[#0A84FF]/70 focus:border-[#0A84FF] focus:outline-none transition-colors appearance-none cursor-pointer"
              >
                <option v-for="r in roles" :key="r.value" :value="r.value" class="bg-[#2D2D2D]">{{ r.label }}</option>
              </select>
            </div>
          </div>
 
          <div>
            <label class="block text-sm font-semibold text-white mb-2 pl-1">Password *</label>
            <div class="relative">
              <Lock class="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-[#8E8E93]" />
              <input
                v-model="password"
                :type="showPassword ? 'text' : 'password'"
                placeholder="At least 6 characters"
                class="w-full pl-12 pr-12 py-4 bg-[#2D2D2D] rounded-2xl text-white text-sm placeholder-[#8E8E93] border-[1.5px] border-[#0A84FF]/70 focus:border-[#0A84FF] focus:outline-none transition-colors"
              />
              <button type="button" @click="showPassword = !showPassword" class="absolute right-4 top-1/2 -translate-y-1/2 text-[#8E8E93] hover:text-white transition-colors">
                <EyeOff v-if="showPassword" class="w-5 h-5" />
                <Eye v-else class="w-5 h-5" />
              </button>
            </div>
          </div>
 
          <div>
            <label class="block text-sm font-semibold text-white mb-2 pl-1">Confirm Password *</label>
            <div class="relative">
              <Lock class="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-[#8E8E93]" />
              <input
                v-model="confirmPassword"
                :type="showConfirm ? 'text' : 'password'"
                placeholder="Repeat password"
                class="w-full pl-12 pr-12 py-4 bg-[#2D2D2D] rounded-2xl text-white text-sm placeholder-[#8E8E93] border-[1.5px] border-white/10 focus:border-[#0A84FF] focus:outline-none transition-colors"
                @keyup.enter="handleSubmit"
              />
              <button type="button" @click="showConfirm = !showConfirm" class="absolute right-4 top-1/2 -translate-y-1/2 text-[#8E8E93] hover:text-white transition-colors">
                <EyeOff v-if="showConfirm" class="w-5 h-5" />
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
            <span v-else>Sign Up</span>
          </button>
 
          <p class="text-center text-sm text-white/50 pt-2">
            Already have an account?
            <button type="button" @click="router.push('/login')" class="text-cyan-400 font-semibold hover:underline ml-1">
              Sign In
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
  from { opacity: 0; transform: translateY(16px); }
  to { opacity: 1; transform: translateY(0); }
}
</style>
 
