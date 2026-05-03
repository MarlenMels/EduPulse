<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { coursesApi, enrollmentsApi, adminUsersApi, type CourseUser } from '@/api/client'
import { useAuthStore } from '@/stores/auth'
import { Loader2, Plus, Trash2, X } from 'lucide-vue-next'

const auth = useAuthStore()

interface CourseLite {
  id: number
  title: string
}

const courses = ref<CourseLite[]>([])
const selectedCourseId = ref(0)
const teachers = ref<CourseUser[]>([])
const students = ref<CourseUser[]>([])
const loading = ref(true)
const error = ref('')

// Add modal: teacher or student
const showAdd = ref(false)
const addKind = ref<'teacher' | 'student'>('teacher')
const candidates = ref<CourseUser[]>([])
const selectedUserId = ref(0)
const adding = ref(false)
const addError = ref('')

const canManage = computed(() => auth.isAdmin || auth.isManager)

async function loadCourses() {
  try {
    const res = await coursesApi.list({ limit: 200 })
    courses.value = (res.data.items || []).map((c: any) => ({ id: c.id, title: c.title }))
    if (courses.value.length && !selectedCourseId.value) {
      selectedCourseId.value = courses.value[0]!.id
      await refreshMembers()
    }
  } catch (e: any) {
    error.value = e.response?.data?.error || 'Failed to load courses'
  } finally {
    loading.value = false
  }
}

async function refreshMembers() {
  if (!selectedCourseId.value) return
  error.value = ''
  try {
    const [t, s] = await Promise.all([
      enrollmentsApi.teachers(selectedCourseId.value),
      enrollmentsApi.students(selectedCourseId.value),
    ])
    teachers.value = t.data.items || []
    students.value = s.data.items || []
  } catch (e: any) {
    error.value = e.response?.data?.error || 'Failed to load members'
  }
}

function openAdd(kind: 'teacher' | 'student') {
  addKind.value = kind
  addError.value = ''
  selectedUserId.value = 0
  candidates.value = []
  showAdd.value = true
  loadCandidates()
}

async function loadCandidates() {
  try {
    const res = await adminUsersApi.list({ limit: 500 })
    const all = res.data.items || []
    const role = addKind.value === 'teacher' ? 'teacher' : 'student'
    const existingIds = new Set(
      (addKind.value === 'teacher' ? teachers.value : students.value).map((u) => u.id),
    )
    candidates.value = all.filter((u: any) => u.role === role && !existingIds.has(u.id))
  } catch (e: any) {
    addError.value = e.response?.data?.error || 'Failed to load users'
  }
}

async function confirmAdd() {
  if (!selectedUserId.value) {
    addError.value = `Choose a ${addKind.value}`
    return
  }
  adding.value = true
  addError.value = ''
  try {
    if (addKind.value === 'teacher') {
      await enrollmentsApi.addTeacher(selectedCourseId.value, selectedUserId.value)
    } else {
      await enrollmentsApi.enrollStudent(selectedCourseId.value, selectedUserId.value)
    }
    showAdd.value = false
    await refreshMembers()
  } catch (e: any) {
    addError.value = e.response?.data?.error || 'Failed'
  } finally {
    adding.value = false
  }
}

async function removeTeacher(u: CourseUser) {
  if (!confirm(`Remove ${u.email} from the course?`)) return
  try {
    await enrollmentsApi.removeTeacher(selectedCourseId.value, u.id)
    teachers.value = teachers.value.filter((x) => x.id !== u.id)
  } catch (e: any) {
    alert(e.response?.data?.error || 'Failed to remove')
  }
}

async function unenrollStudent(u: CourseUser) {
  if (!confirm(`Remove ${u.email} from the course?`)) return
  try {
    await enrollmentsApi.unenrollStudent(selectedCourseId.value, u.id)
    students.value = students.value.filter((x) => x.id !== u.id)
  } catch (e: any) {
    alert(e.response?.data?.error || 'Failed to remove')
  }
}

onMounted(loadCourses)
</script>

<template>
  <div>
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-6">
      <h1 class="text-2xl font-extrabold text-cyan-400">Course members</h1>
    </div>

    <p v-if="error" class="mb-4 p-3 rounded-xl bg-red-500/10 border border-red-500/30 text-red-400 text-sm">{{ error }}</p>

    <div v-if="loading" class="flex items-center gap-2 text-white/60 text-sm">
      <Loader2 class="w-4 h-4 animate-spin" /> Loading…
    </div>

    <div v-else-if="!courses.length" class="text-white/40 text-sm">No courses yet. Create one in Courses first.</div>

    <div v-else class="space-y-6">
      <div>
        <label class="block text-xs font-semibold text-white/60 mb-1.5">Course</label>
        <select
          v-model.number="selectedCourseId"
          @change="refreshMembers"
          class="w-full sm:w-96 px-4 py-3 bg-[#1E1E1E] rounded-xl text-white text-sm border border-white/10 focus:border-cyan-400 focus:outline-none"
        >
          <option v-for="c in courses" :key="c.id" :value="c.id" class="bg-[#1E1E1E]">{{ c.title }}</option>
        </select>
      </div>

      <!-- Teachers section -->
      <div class="bg-[#1E1E1E] border border-white/5 rounded-2xl p-5">
        <div class="flex items-center justify-between mb-4">
          <h2 class="text-sm font-bold uppercase tracking-wider text-white/60">Teachers</h2>
          <button
            v-if="canManage"
            @click="openAdd('teacher')"
            class="inline-flex items-center gap-1.5 text-cyan-400 hover:text-cyan-300 text-sm font-semibold"
          >
            <Plus class="w-4 h-4" /> Add teacher
          </button>
        </div>
        <div v-if="!teachers.length" class="text-white/40 text-sm">No teachers assigned.</div>
        <ul v-else class="divide-y divide-white/5">
          <li v-for="t in teachers" :key="t.id" class="flex items-center justify-between py-3">
            <span class="text-white text-sm">{{ t.email }}</span>
            <button
              v-if="canManage"
              @click="removeTeacher(t)"
              class="text-red-400/80 hover:text-red-400 p-1.5 rounded-lg hover:bg-red-400/10 transition-colors"
              title="Remove"
            >
              <Trash2 class="w-4 h-4" />
            </button>
          </li>
        </ul>
      </div>

      <!-- Students section -->
      <div class="bg-[#1E1E1E] border border-white/5 rounded-2xl p-5">
        <div class="flex items-center justify-between mb-4">
          <h2 class="text-sm font-bold uppercase tracking-wider text-white/60">Students</h2>
          <button
            v-if="canManage"
            @click="openAdd('student')"
            class="inline-flex items-center gap-1.5 text-cyan-400 hover:text-cyan-300 text-sm font-semibold"
          >
            <Plus class="w-4 h-4" /> Enroll student
          </button>
        </div>
        <div v-if="!students.length" class="text-white/40 text-sm">No students enrolled.</div>
        <ul v-else class="divide-y divide-white/5">
          <li v-for="s in students" :key="s.id" class="flex items-center justify-between py-3">
            <span class="text-white text-sm">{{ s.email }}</span>
            <button
              v-if="canManage"
              @click="unenrollStudent(s)"
              class="text-red-400/80 hover:text-red-400 p-1.5 rounded-lg hover:bg-red-400/10 transition-colors"
              title="Unenroll"
            >
              <Trash2 class="w-4 h-4" />
            </button>
          </li>
        </ul>
      </div>
    </div>

    <Teleport to="body">
      <div v-if="showAdd" class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4" @click.self="showAdd = false">
        <div class="w-full max-w-md bg-[#1E1E1E] border border-white/10 rounded-2xl p-6">
          <div class="flex items-center justify-between mb-5">
            <h2 class="text-lg font-extrabold text-white">
              {{ addKind === 'teacher' ? 'Add teacher' : 'Enroll student' }}
            </h2>
            <button @click="showAdd = false" class="text-white/40 hover:text-white"><X class="w-5 h-5" /></button>
          </div>
          <form @submit.prevent="confirmAdd" class="space-y-4">
            <p v-if="addError" class="rounded-xl bg-red-400/10 px-4 py-3 text-sm text-red-400">{{ addError }}</p>
            <div>
              <label class="block text-xs font-semibold text-white/60 mb-1.5">User</label>
              <select
                v-model.number="selectedUserId"
                class="w-full px-4 py-3 bg-[#2D2D2D] rounded-xl text-white text-sm border border-white/10 focus:border-cyan-400 focus:outline-none"
              >
                <option :value="0" class="bg-[#2D2D2D]">Select a user…</option>
                <option v-for="u in candidates" :key="u.id" :value="u.id" class="bg-[#2D2D2D]">{{ u.email }}</option>
              </select>
              <p v-if="!candidates.length" class="mt-1.5 text-xs text-white/40">
                No {{ addKind }}s available. Create one in <em>Users</em> first.
              </p>
            </div>
            <button
              type="submit"
              :disabled="adding"
              class="w-full py-3 rounded-xl bg-cyan-400 text-[#121212] font-bold text-sm hover:bg-cyan-300 disabled:opacity-50 transition-colors"
            >
              {{ adding ? 'Adding…' : (addKind === 'teacher' ? 'Add teacher' : 'Enroll student') }}
            </button>
          </form>
        </div>
      </div>
    </Teleport>
  </div>
</template>
