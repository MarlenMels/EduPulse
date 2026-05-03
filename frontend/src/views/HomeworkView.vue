<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import {
  assignmentsApi,
  homeworkApi,
  sessionsApi,
  type AssignmentRow,
  type SubmissionRow,
  type MineRow,
  type SessionRow,
} from '@/api/client'
import { useAuthStore } from '@/stores/auth'
import { BookOpen, Plus, X, Loader2, Eye } from 'lucide-vue-next'

const auth = useAuthStore()

const role = computed(() => auth.userRole)
const isStaff = computed(() => auth.isAdmin || auth.isManager || auth.isTeacher)
const isStudent = computed(() => auth.isStudent)

const loading = ref(true)
const error = ref('')

// Staff view: list of assignments
const assignments = ref<AssignmentRow[]>([])

// Student view: own submissions + assignments to choose from
const mine = ref<MineRow[]>([])
const studentAssignments = ref<AssignmentRow[]>([])

// Modal: create assignment (staff)
const showCreate = ref(false)
const sessions = ref<SessionRow[]>([])
const newAssignment = ref({ session_id: 0, title: '', description: '' })
const creating = ref(false)
const createError = ref('')

// Modal: view submissions (staff)
const showSubs = ref(false)
const subsAssignment = ref<AssignmentRow | null>(null)
const submissions = ref<SubmissionRow[]>([])
const subsLoading = ref(false)

// Modal: submit homework (student)
const showSubmit = ref(false)
const submitTarget = ref<AssignmentRow | null>(null)
const submission = ref({ content: '' })
const submitting = ref(false)
const submitError = ref('')

const statusStyle: Record<string, string> = {
  submitted: 'bg-cyan-400/10 text-cyan-400',
  accepted: 'bg-green-400/10 text-green-400',
  rejected: 'bg-red-400/10 text-red-400',
  needs_fix: 'bg-amber-400/10 text-amber-400',
}

async function fetchAll() {
  loading.value = true
  error.value = ''
  try {
    if (isStudent.value) {
      const [m, a] = await Promise.all([
        homeworkApi.mine({ limit: 100 }),
        assignmentsApi.list({ limit: 100 }),
      ])
      mine.value = m.data.items || []
      studentAssignments.value = a.data.items || []
    } else {
      const a = await assignmentsApi.list({ limit: 200 })
      assignments.value = a.data.items || []
    }
  } catch (e: any) {
    error.value = e.response?.data?.error || 'Failed to load'
  } finally {
    loading.value = false
  }
}

async function openCreate() {
  createError.value = ''
  newAssignment.value = { session_id: 0, title: '', description: '' }
  showCreate.value = true
  try {
    const res = await sessionsApi.list({ limit: 200 })
    sessions.value = res.data.items || []
  } catch {
    sessions.value = []
  }
}

async function createAssignment() {
  createError.value = ''
  if (!newAssignment.value.session_id) {
    createError.value = 'Choose a session'
    return
  }
  if (!newAssignment.value.title.trim()) {
    createError.value = 'Title is required'
    return
  }
  creating.value = true
  try {
    await assignmentsApi.create({
      session_id: newAssignment.value.session_id,
      title: newAssignment.value.title.trim(),
      description: newAssignment.value.description.trim(),
    })
    showCreate.value = false
    await fetchAll()
  } catch (e: any) {
    createError.value = e.response?.data?.error || 'Failed to create'
  } finally {
    creating.value = false
  }
}

async function openSubs(a: AssignmentRow) {
  subsAssignment.value = a
  showSubs.value = true
  subsLoading.value = true
  submissions.value = []
  try {
    const res = await assignmentsApi.submissions(a.id)
    submissions.value = res.data.items || []
  } catch (e: any) {
    error.value = e.response?.data?.error || 'Failed to load submissions'
  } finally {
    subsLoading.value = false
  }
}

async function setStatus(s: SubmissionRow, status: string) {
  try {
    await homeworkApi.updateStatus(s.id, status)
    s.status = status
  } catch (e: any) {
    alert(e.response?.data?.error || 'Failed to update')
  }
}

function openSubmit(a: AssignmentRow) {
  submitTarget.value = a
  submission.value = { content: '' }
  submitError.value = ''
  showSubmit.value = true
}

async function submitHomework() {
  if (!submitTarget.value) return
  submitError.value = ''
  if (!submission.value.content.trim()) {
    submitError.value = 'Answer is required'
    return
  }
  submitting.value = true
  try {
    await homeworkApi.submit({
      assignment_id: submitTarget.value.id,
      content: submission.value.content.trim(),
    })
    showSubmit.value = false
    await fetchAll()
  } catch (e: any) {
    submitError.value = e.response?.data?.error || 'Failed to submit'
  } finally {
    submitting.value = false
  }
}

const submittedAssignmentIds = computed(() => new Set(mine.value.map((m) => m.assignment_id)))

const pendingAssignments = computed(() =>
  studentAssignments.value.filter((a) => !submittedAssignmentIds.value.has(a.id)),
)

onMounted(fetchAll)
</script>

<template>
  <div>
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-6">
      <h1 class="text-2xl font-extrabold text-cyan-400">Homework</h1>
      <button
        v-if="isStaff"
        @click="openCreate"
        class="flex items-center gap-2 px-4 py-2.5 bg-cyan-400 text-[#121212] font-semibold text-sm rounded-xl hover:bg-cyan-300 transition-colors"
      >
        <Plus class="w-4 h-4" /> New assignment
      </button>
    </div>

    <p v-if="error" class="mb-4 p-3 rounded-xl bg-red-500/10 border border-red-500/30 text-red-400 text-sm">{{ error }}</p>

    <div v-if="loading" class="flex items-center gap-2 text-white/60 text-sm">
      <Loader2 class="w-4 h-4 animate-spin" /> Loading…
    </div>

    <!-- STAFF VIEW: assignments table -->
    <div v-else-if="isStaff" class="overflow-hidden bg-[#1E1E1E] border border-white/5 rounded-2xl">
      <table class="w-full text-sm">
        <thead class="bg-white/5 text-left text-xs uppercase tracking-wider text-white/40">
          <tr>
            <th class="px-5 py-3 font-semibold">Title</th>
            <th class="px-5 py-3 font-semibold">Course</th>
            <th class="px-5 py-3 font-semibold">Session</th>
            <th v-if="role === 'admin' || role === 'manager'" class="px-5 py-3 font-semibold">Teacher</th>
            <th class="px-5 py-3 font-semibold">Submissions</th>
            <th class="px-5 py-3"></th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="a in assignments" :key="a.id" class="border-t border-white/5 text-white/80">
            <td class="px-5 py-3 font-medium">{{ a.title }}</td>
            <td class="px-5 py-3">{{ a.course_title }}</td>
            <td class="px-5 py-3 text-white/60">{{ a.session_title }}</td>
            <td v-if="role === 'admin' || role === 'manager'" class="px-5 py-3 text-white/60">{{ a.creator_email }}</td>
            <td class="px-5 py-3">
              <span class="inline-flex items-center px-2.5 py-1 rounded-md text-xs font-semibold bg-white/10 text-white/70">
                {{ a.submission_count }}
              </span>
            </td>
            <td class="px-5 py-3 text-right">
              <button
                @click="openSubs(a)"
                class="inline-flex items-center gap-1.5 text-cyan-400 hover:text-cyan-300 text-xs font-semibold"
              >
                <Eye class="w-4 h-4" /> View
              </button>
            </td>
          </tr>
          <tr v-if="!assignments.length">
            <td :colspan="role === 'admin' || role === 'manager' ? 6 : 5" class="px-5 py-8 text-center text-white/40 text-sm">
              No assignments yet.
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- STUDENT VIEW -->
    <div v-else-if="isStudent" class="space-y-6">
      <div>
        <h2 class="text-sm font-bold uppercase tracking-wider text-white/50 mb-3">To do</h2>
        <div v-if="!pendingAssignments.length" class="text-white/40 text-sm">All caught up.</div>
        <div v-else class="grid sm:grid-cols-2 gap-3">
          <div
            v-for="a in pendingAssignments"
            :key="a.id"
            class="p-4 bg-[#1E1E1E] border border-white/5 rounded-2xl"
          >
            <div class="flex items-start gap-3">
              <div class="w-10 h-10 rounded-xl bg-cyan-400/10 flex items-center justify-center shrink-0">
                <BookOpen class="w-5 h-5 text-cyan-400" />
              </div>
              <div class="flex-1 min-w-0">
                <div class="font-semibold text-white">{{ a.title }}</div>
                <div class="text-xs text-white/40 mt-1">{{ a.course_title }} · {{ a.session_title }}</div>
                <p v-if="a.description" class="text-sm text-white/60 mt-2 line-clamp-2">{{ a.description }}</p>
                <button
                  @click="openSubmit(a)"
                  class="mt-3 px-3.5 py-1.5 rounded-lg bg-cyan-400 text-black font-semibold text-xs hover:bg-cyan-300 transition-colors"
                >
                  Submit
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div>
        <h2 class="text-sm font-bold uppercase tracking-wider text-white/50 mb-3">Submitted</h2>
        <div v-if="!mine.length" class="text-white/40 text-sm">Nothing submitted yet.</div>
        <div v-else class="space-y-2">
          <div
            v-for="m in mine"
            :key="m.id"
            class="p-4 bg-[#1E1E1E] border border-white/5 rounded-2xl flex items-start gap-3"
          >
            <div class="flex-1 min-w-0">
              <div class="font-semibold text-white">{{ m.assignment_title }}</div>
              <div class="text-xs text-white/40">{{ m.course_title }} · {{ m.session_title }}</div>
              <p class="text-sm text-white/70 mt-2 whitespace-pre-wrap">{{ m.content }}</p>
            </div>
            <span
              class="inline-flex items-center px-2.5 py-1 rounded-md text-xs font-semibold shrink-0"
              :class="statusStyle[m.status] || 'bg-white/10 text-white/60'"
            >{{ m.status }}</span>
          </div>
        </div>
      </div>
    </div>

    <div v-else class="text-white/40 text-sm">No homework view for your role.</div>

    <Teleport to="body">
      <!-- Create assignment modal -->
      <div v-if="showCreate" class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4" @click.self="showCreate = false">
        <div class="w-full max-w-lg bg-[#1E1E1E] border border-white/10 rounded-2xl p-6">
          <div class="flex items-center justify-between mb-5">
            <h2 class="text-lg font-extrabold text-white">New assignment</h2>
            <button @click="showCreate = false" class="text-white/40 hover:text-white"><X class="w-5 h-5" /></button>
          </div>
          <form @submit.prevent="createAssignment" class="space-y-4">
            <p v-if="createError" class="rounded-xl bg-red-400/10 px-4 py-3 text-sm text-red-400">{{ createError }}</p>
            <div>
              <label class="block text-xs font-semibold text-white/60 mb-1.5">Session</label>
              <select
                v-model.number="newAssignment.session_id"
                class="w-full px-4 py-3 bg-[#2D2D2D] rounded-xl text-white text-sm border border-white/10 focus:border-cyan-400 focus:outline-none"
              >
                <option :value="0" class="bg-[#2D2D2D]">Select a session…</option>
                <option v-for="s in sessions" :key="s.id" :value="s.id" class="bg-[#2D2D2D]">
                  {{ s.course_title }} — {{ s.title }}
                </option>
              </select>
              <p v-if="!sessions.length" class="mt-1.5 text-xs text-white/40">No sessions available.</p>
            </div>
            <div>
              <label class="block text-xs font-semibold text-white/60 mb-1.5">Title</label>
              <input
                v-model="newAssignment.title"
                maxlength="200"
                placeholder="e.g. Read chapter 4 and answer 3 questions"
                class="w-full px-4 py-3 bg-[#2D2D2D] rounded-xl text-white text-sm placeholder-white/30 border border-white/10 focus:border-cyan-400 focus:outline-none"
              />
            </div>
            <div>
              <label class="block text-xs font-semibold text-white/60 mb-1.5">Description (optional)</label>
              <textarea
                v-model="newAssignment.description"
                rows="3"
                maxlength="5000"
                placeholder="Optional details for the students"
                class="w-full px-4 py-3 bg-[#2D2D2D] rounded-xl text-white text-sm placeholder-white/30 border border-white/10 focus:border-cyan-400 focus:outline-none"
              />
            </div>
            <button type="submit" :disabled="creating" class="w-full py-3 rounded-xl bg-cyan-400 text-[#121212] font-bold text-sm hover:bg-cyan-300 disabled:opacity-50 transition-colors">
              {{ creating ? 'Creating…' : 'Create' }}
            </button>
          </form>
        </div>
      </div>

      <!-- View submissions modal -->
      <div v-if="showSubs" class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4" @click.self="showSubs = false">
        <div class="w-full max-w-2xl max-h-[85vh] overflow-auto bg-[#1E1E1E] border border-white/10 rounded-2xl p-6">
          <div class="flex items-center justify-between mb-5">
            <div>
              <h2 class="text-lg font-extrabold text-white">{{ subsAssignment?.title }}</h2>
              <p class="text-xs text-white/40 mt-1">{{ subsAssignment?.course_title }} · {{ subsAssignment?.session_title }}</p>
            </div>
            <button @click="showSubs = false" class="text-white/40 hover:text-white"><X class="w-5 h-5" /></button>
          </div>
          <div v-if="subsLoading" class="text-white/60 text-sm flex items-center gap-2"><Loader2 class="w-4 h-4 animate-spin" /> Loading…</div>
          <div v-else-if="!submissions.length" class="text-white/40 text-sm">No submissions yet.</div>
          <div v-else class="space-y-3">
            <div v-for="s in submissions" :key="s.id" class="p-4 bg-[#2D2D2D] rounded-xl">
              <div class="flex items-center justify-between gap-3 mb-2">
                <span class="font-semibold text-white">{{ s.student_email }}</span>
                <span class="inline-flex items-center px-2 py-0.5 rounded-md text-xs font-semibold" :class="statusStyle[s.status] || 'bg-white/10 text-white/60'">
                  {{ s.status }}
                </span>
              </div>
              <p class="text-sm text-white/70 whitespace-pre-wrap">{{ s.content }}</p>
              <div class="flex gap-2 mt-3">
                <button @click="setStatus(s, 'accepted')" class="px-3 py-1.5 rounded-lg bg-green-400/10 text-green-400 text-xs font-semibold hover:bg-green-400/20">Accept</button>
                <button @click="setStatus(s, 'needs_fix')" class="px-3 py-1.5 rounded-lg bg-amber-400/10 text-amber-400 text-xs font-semibold hover:bg-amber-400/20">Needs fix</button>
                <button @click="setStatus(s, 'rejected')" class="px-3 py-1.5 rounded-lg bg-red-400/10 text-red-400 text-xs font-semibold hover:bg-red-400/20">Reject</button>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Submit homework modal -->
      <div v-if="showSubmit" class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4" @click.self="showSubmit = false">
        <div class="w-full max-w-lg bg-[#1E1E1E] border border-white/10 rounded-2xl p-6">
          <div class="flex items-center justify-between mb-5">
            <div>
              <h2 class="text-lg font-extrabold text-white">{{ submitTarget?.title }}</h2>
              <p class="text-xs text-white/40 mt-1">{{ submitTarget?.course_title }} · {{ submitTarget?.session_title }}</p>
            </div>
            <button @click="showSubmit = false" class="text-white/40 hover:text-white"><X class="w-5 h-5" /></button>
          </div>
          <p v-if="submitTarget?.description" class="text-sm text-white/60 mb-4 whitespace-pre-wrap">{{ submitTarget.description }}</p>
          <form @submit.prevent="submitHomework" class="space-y-4">
            <p v-if="submitError" class="rounded-xl bg-red-400/10 px-4 py-3 text-sm text-red-400">{{ submitError }}</p>
            <div>
              <label class="block text-xs font-semibold text-white/60 mb-1.5">Your answer</label>
              <textarea
                v-model="submission.content"
                rows="6"
                maxlength="5000"
                placeholder="Type your answer here…"
                class="w-full px-4 py-3 bg-[#2D2D2D] rounded-xl text-white text-sm placeholder-white/30 border border-white/10 focus:border-cyan-400 focus:outline-none"
              />
            </div>
            <button type="submit" :disabled="submitting" class="w-full py-3 rounded-xl bg-cyan-400 text-[#121212] font-bold text-sm hover:bg-cyan-300 disabled:opacity-50 transition-colors">
              {{ submitting ? 'Submitting…' : 'Submit' }}
            </button>
          </form>
        </div>
      </div>
    </Teleport>
  </div>
</template>
