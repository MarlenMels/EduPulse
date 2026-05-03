<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { homeworkApi, sessionsApi, uploadsApi } from '@/api/client'
import { useAuthStore } from '@/stores/auth'
import { BookOpen, Send, CheckCircle, Clock, AlertCircle, X, Paperclip, Upload } from 'lucide-vue-next'

const auth = useAuthStore()
const homework = ref<any[]>([])
const sessions = ref<any[]>([])
const loading = ref(true)
const error = ref('')
const showSubmitModal = ref(false)
const submitting = ref(false)
const uploadingFiles = ref(false)
const attachments = ref<any[]>([])

const newHomework = ref({ session_id: '', content: '', attachments: '' })

async function fetchSessions() {
  try {
    const res = await sessionsApi.list({ limit: 100 })
    sessions.value = res.data.items || []
  } catch (e: any) {
    console.error('Failed to load sessions:', e)
  }
}

async function fetchHomework() {
  loading.value = true
  error.value = ''
  try {
    if (auth.isStudent || auth.isParent) {
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

async function handleFileUpload(event: Event) {
  const target = event.target as HTMLInputElement
  if (!target.files || target.files.length === 0) return
  
  uploadingFiles.value = true
  const files = Array.from(target.files)
  
  try {
    for (const file of files) {
      const res = await uploadsApi.upload(file)
      attachments.value.push({
        url: res.data.url,
        name: res.data.name,
        size: res.data.size
      })
    }
  } catch (e: any) {
    error.value = e.response?.data?.error || 'Failed to upload files'
  } finally {
    uploadingFiles.value = false
    target.value = ''
  }
}

function removeAttachment(index: number) {
  attachments.value.splice(index, 1)
}

async function submitHomework() {
  if (!newHomework.value.content.trim() || !newHomework.value.session_id) return
  submitting.value = true
  try {
    const attachmentsJson = attachments.value.length > 0 ? JSON.stringify(attachments.value) : ''
    await homeworkApi.submit({
      session_id: Number(newHomework.value.session_id),
      content: newHomework.value.content,
      attachments: attachmentsJson
    })
    showSubmitModal.value = false
    newHomework.value = { session_id: '', content: '', attachments: '' }
    attachments.value = []
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
    case 'accepted': return 'text-green-400 bg-green-400/10'
    case 'rejected': return 'text-red-400 bg-red-400/10'
    case 'pending': return 'text-yellow-400 bg-yellow-400/10'
    case 'submitted': return 'text-yellow-400 bg-yellow-400/10'
    case 'needs_fix': return 'text-orange-400 bg-orange-400/10'
    default: return 'text-white/40 bg-white/5'
  }
}

function statusLabel(status: string) {
  switch (status) {
    case 'approved': return 'Approved'
    case 'accepted': return 'Accepted'
    case 'rejected': return 'Rejected'
    case 'pending': return 'Pending'
    case 'submitted': return 'Submitted'
    case 'needs_fix': return 'Needs Fix'
    default: return status
  }
}

onMounted(async () => {
  await fetchSessions()
  await fetchHomework()
})

async function openSubmitModal() {
  await fetchSessions()
  showSubmitModal.value = true
}
</script>

<template>
  <div>
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-6">
      <h1 class="text-2xl font-extrabold text-cyan-400">Homework</h1>
      <button
        v-if="auth.isStudent"
        @click="openSubmitModal"
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
            <div v-if="hw.attachments" class="mt-2">
              <div class="flex flex-wrap gap-2">
                <span 
                  v-for="(attachment, index) in JSON.parse(hw.attachments || '[]')" 
                  :key="index"
                  class="inline-flex items-center gap-1 px-2 py-1 bg-[#2D2D2D] rounded-lg text-xs text-cyan-400"
                >
                  <Paperclip class="w-3 h-3" />
                  {{ attachment.name }}
                </span>
              </div>
            </div>
            <p class="text-xs text-white/30 mt-2">Session: {{ hw.session_id }} · Student: {{ hw.student_id }}</p>
          </div>

          <!-- Teacher actions -->
          <div v-if="auth.isTeacher && (hw.status === 'pending' || hw.status === 'submitted')" class="flex gap-2 shrink-0">
            <button
              @click="updateStatus(hw.id, 'accepted')"
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
                <label class="block text-sm font-semibold text-white/70 mb-1.5">Session</label>
                <select v-model="newHomework.session_id" class="w-full px-4 py-3 bg-[#2D2D2D] rounded-xl text-white text-sm border border-transparent focus:border-cyan-400 focus:outline-none">
                  <option value="" disabled>Select a session</option>
                  <option v-for="session in sessions" :key="session.id" :value="session.id">
                    {{ session.title }}
                  </option>
                </select>
              </div>
              <div>
                <label class="block text-sm font-semibold text-white/70 mb-1.5">Content</label>
                <textarea v-model="newHomework.content" rows="4" placeholder="Describe your completed work..." class="w-full px-4 py-3 bg-[#2D2D2D] rounded-xl text-white text-sm placeholder-white/30 border border-transparent focus:border-cyan-400 focus:outline-none resize-none" />
              </div>
              
              <div>
                <label class="block text-sm font-semibold text-white/70 mb-1.5">Attachments</label>
                <div class="space-y-2">
                  <div class="relative">
                    <input 
                      type="file" 
                      multiple 
                      @change="handleFileUpload"
                      :disabled="uploadingFiles"
                      accept=".jpg,.jpeg,.png,.gif,.pdf,.doc,.docx,.txt,.zip"
                      class="w-full px-4 py-3 bg-[#2D2D2D] rounded-xl text-white text-sm file:mr-4 file:py-2 file:px-4 file:rounded-lg file:border-0 file:text-sm file:font-semibold file:bg-cyan-400 file:text-black hover:file:bg-cyan-300 disabled:opacity-50 cursor-pointer"
                    />
                    <div v-if="uploadingFiles" class="absolute inset-0 flex items-center justify-center bg-[#2D2D2D]/80 rounded-xl">
                      <div class="w-6 h-6 border-2 border-cyan-400/30 border-t-cyan-400 rounded-full animate-spin" />
                    </div>
                  </div>
                  
                  <div v-if="attachments.length > 0" class="space-y-2">
                    <div 
                      v-for="(attachment, index) in attachments" 
                      :key="index"
                      class="flex items-center gap-3 p-3 bg-[#2D2D2D] rounded-lg"
                    >
                      <Paperclip class="w-4 h-4 text-cyan-400" />
                      <span class="flex-1 text-sm text-white truncate">{{ attachment.name }}</span>
                      <button
                        @click="removeAttachment(index)"
                        class="w-6 h-6 rounded-lg bg-red-400/10 flex items-center justify-center hover:bg-red-400/20 transition-colors"
                        title="Remove file"
                      >
                        <X class="w-3 h-3 text-red-400" />
                      </button>
                    </div>
                  </div>
                  
                  <p class="text-xs text-white/40">Supported formats: Images, PDF, DOC, TXT, ZIP (max 50MB per file)</p>
                </div>
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
