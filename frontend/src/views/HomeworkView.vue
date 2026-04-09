<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { homeworkApi } from '@/api/client'
import { useAuthStore } from '@/stores/auth'
import { BookOpen, Send, CheckCircle, Clock, AlertCircle, X } from 'lucide-vue-next'

const auth = useAuthStore()
const homework = ref<any[]>([])
const loading = ref(true)
const error = ref('')
const showSubmitModal = ref(false)
const submitting = ref(false)

const newHomework = ref({ session_id: 1, content: '' })

async function fetchHomework() {
  loading.value = true
  error.value = ''
  try {
    if (auth.isStudent) {
      const res = await homeworkApi.mine({ limit: 50 })
      homework.value = res.data.items || []
    } else {
      const res = await homeworkApi.list({ limit: 50 })
      homework.value = res.data.items || []
    }
  } catch (e: any) {
    error.value = e.response?.data?.error || 'Failed to load'
  } finally {
    loading.value = false
  }
}

async function submitHomework() {
  if (!newHomework.value.content.trim()) return
  submitting.value = true
  try {
    await homeworkApi.submit(newHomework.value)
    showSubmitModal.value = false
    newHomework.value = { session_id: 1, content: '' }
    await fetchHomework()
  } catch (e: any) {
    error.value = e.response?.data?.error || 'Failed to submit'
  } finally {
    submitting.value = false
  }
}

async function updateStatus(id: number, status: string) {
  try {
    await homeworkApi.updateStatus(id, status)
    await fetchHomework()
  } catch (e: any) {
    error.value = e.response?.data?.error || 'Failed to update'
  }
}

function statusColor(status: string) {
  switch (status) {
    case 'approved': return 'text-green-400 bg-green-400/10'
    case 'rejected': return 'text-red-400 bg-red-400/10'
    case 'pending': return 'text-yellow-400 bg-yellow-400/10'
    default: return 'text-white/40 bg-white/5'
  }
}

function statusLabel(status: string) {
  switch (status) {
    case 'approved': return 'Approved'
    case 'rejected': return 'Rejected'
    case 'pending': return 'Pending'
    default: return status
  }
}

onMounted(fetchHomework)
</script>

<template>
  <div>
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-6">
      <h1 class="text-2xl font-extrabold text-cyan-400">Homework</h1>
      <button
        v-if="auth.isStudent"
        @click="showSubmitModal = true"
        class="flex items-center gap-2 px-4 py-2.5 bg-cyan-400 text-black font-semibold text-sm rounded-xl hover:bg-cyan-300 transition-colors"
      >
        <Send class="w-4 h-4" />
        Submit Work
      </button>
    </div>

    <div v-if="loading" class="flex justify-center py-16">
      <div class="w-8 h-8 border-2 border-cyan-400/30 border-t-cyan-400 rounded-full animate-spin" />
    </div>

    <div v-else-if="error" class="text-center py-16">
      <p class="text-red-400 mb-4">{{ error }}</p>
      <button @click="fetchHomework" class="px-4 py-2 bg-cyan-400 text-black rounded-lg font-semibold text-sm">Retry</button>
    </div>

    <div v-else-if="homework.length === 0" class="text-center py-16 text-white/40">
      <BookOpen class="w-16 h-16 mx-auto mb-4 opacity-30" />
      <p>No homework yet</p>
    </div>

    <div v-else class="space-y-3">
      <div
        v-for="hw in homework"
        :key="hw.id"
        class="bg-[#1E1E1E] rounded-2xl p-5 border border-white/5"
      >
        <div class="flex items-start gap-4">
          <div class="w-11 h-11 rounded-xl bg-purple-400/10 flex items-center justify-center shrink-0">
            <BookOpen class="w-5 h-5 text-purple-400" />
          </div>
          <div class="flex-1 min-w-0">
            <div class="flex items-center gap-2 flex-wrap">
              <h3 class="font-bold text-white">Homework #{{ hw.id }}</h3>
              <span class="px-2 py-0.5 rounded-full text-xs font-semibold" :class="statusColor(hw.status)">
                {{ statusLabel(hw.status) }}
              </span>
            </div>
            <p class="text-sm text-white/60 mt-1 line-clamp-2">{{ hw.content }}</p>
            <p class="text-xs text-white/30 mt-2">Session: {{ hw.session_id }} · Student: {{ hw.student_id }}</p>
          </div>

          <!-- Teacher actions -->
          <div v-if="auth.isTeacher && hw.status === 'pending'" class="flex gap-2 shrink-0">
            <button
              @click="updateStatus(hw.id, 'approved')"
              class="w-9 h-9 rounded-lg bg-green-400/10 flex items-center justify-center hover:bg-green-400/20 transition-colors"
              title="Approve"
            >
              <CheckCircle class="w-4 h-4 text-green-400" />
            </button>
            <button
              @click="updateStatus(hw.id, 'rejected')"
              class="w-9 h-9 rounded-lg bg-red-400/10 flex items-center justify-center hover:bg-red-400/20 transition-colors"
              title="Reject"
            >
              <AlertCircle class="w-4 h-4 text-red-400" />
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Submit modal -->
    <Teleport to="body">
      <Transition name="fade">
        <div v-if="showSubmitModal" class="fixed inset-0 z-50 flex items-end sm:items-center justify-center bg-black/60 p-4" @click.self="showSubmitModal = false">
          <div class="bg-[#1E1E1E] rounded-t-3xl sm:rounded-2xl w-full max-w-md p-6">
            <div class="flex items-center justify-between mb-6">
              <h2 class="text-lg font-extrabold text-white">Submit Work</h2>
              <button @click="showSubmitModal = false" class="text-white/40 hover:text-white">
                <X class="w-5 h-5" />
              </button>
            </div>
            <form @submit.prevent="submitHomework" class="space-y-4">
              <div>
                <label class="block text-sm font-semibold text-white/70 mb-1.5">Session ID</label>
                <input v-model.number="newHomework.session_id" type="number" class="w-full px-4 py-3 bg-[#2D2D2D] rounded-xl text-white text-sm border border-transparent focus:border-cyan-400 focus:outline-none" />
              </div>
              <div>
                <label class="block text-sm font-semibold text-white/70 mb-1.5">Content</label>
                <textarea v-model="newHomework.content" rows="4" placeholder="Describe your completed work..." class="w-full px-4 py-3 bg-[#2D2D2D] rounded-xl text-white text-sm placeholder-white/30 border border-transparent focus:border-cyan-400 focus:outline-none resize-none" />
              </div>
              <div class="flex gap-3 pt-2">
                <button type="button" @click="showSubmitModal = false" class="flex-1 py-3 rounded-xl text-white/60 font-semibold text-sm hover:bg-white/5 transition-colors">Cancel</button>
                <button type="submit" :disabled="submitting" class="flex-1 py-3 rounded-xl bg-cyan-400 text-black font-semibold text-sm hover:bg-cyan-300 disabled:opacity-50 transition-colors">
                  {{ submitting ? 'Submitting...' : 'Submit' }}
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
