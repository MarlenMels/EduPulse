<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { coursesApi, type LessonInput } from '@/api/client'
import { useAuthStore } from '@/stores/auth'
import {
  GraduationCap,
  Plus,
  X,
  BookOpen,
  Video,
  FileText,
  Trash2,
  ChevronDown,
  ChevronUp,
  Pencil,
  Save,
  Loader2,
  AlertCircle,
} from 'lucide-vue-next'
import VideoPlayer from '@/components/VideoPlayer.vue'
import VideoUploader from '@/components/VideoUploader.vue'

const auth = useAuthStore()
const canManage = auth.isAdmin || auth.isManager || auth.isTeacher

const courses = ref<any[]>([])
const loading = ref(true)
const error = ref('')

// Create course modal
const showCreateModal = ref(false)
const creating = ref(false)
const newCourse = ref({
  title: '',
  description: '',
  image_url: '',
  lessons: [] as LessonInput[],
})

// Expand/collapse course lessons
const expandedCourse = ref<number | null>(null)

// Add lesson modal
const showAddLessonModal = ref(false)
const addLessonCourseId = ref(0)
const addingLesson = ref(false)
const newLesson = ref<LessonInput>(emptyLesson())

// Edit lesson
const editingLessonId = ref<number | null>(null)
const editLesson = ref<LessonInput>(emptyLesson())
const savingLesson = ref(false)

function emptyLesson(): LessonInput {
  return { title: '', description: '', video_url: '', file_url: '', sort_order: 0 }
}

function resetCreateForm() {
  newCourse.value = { title: '', description: '', image_url: '', lessons: [] }
}

function addLessonToForm() {
  newCourse.value.lessons.push({
    ...emptyLesson(),
    sort_order: newCourse.value.lessons.length + 1,
  })
}

function removeLessonFromForm(idx: number) {
  newCourse.value.lessons.splice(idx, 1)
  newCourse.value.lessons.forEach((l, i) => (l.sort_order = i + 1))
}

async function onVideoUploaded(_courseId: number) {
  await fetchCourses()
}

async function fetchCourses() {
  loading.value = true
  error.value = ''
  try {
    const res = await coursesApi.list({ limit: 50 })
    courses.value = res.data.items || []
  } catch (e: any) {
    error.value = e.response?.data?.error || 'Failed to load courses'
  } finally {
    loading.value = false
  }
}

async function createCourse() {
  if (!newCourse.value.title.trim()) return
  creating.value = true
  try {
    await coursesApi.create(newCourse.value)
    showCreateModal.value = false
    resetCreateForm()
    await fetchCourses()
  } catch (e: any) {
    error.value = e.response?.data?.error || 'Failed to create course'
  } finally {
    creating.value = false
  }
}

function openAddLesson(courseId: number, lessonCount: number) {
  addLessonCourseId.value = courseId
  newLesson.value = { ...emptyLesson(), sort_order: lessonCount + 1 }
  showAddLessonModal.value = true
}

async function submitAddLesson() {
  if (!newLesson.value.title.trim()) return
  addingLesson.value = true
  try {
    await coursesApi.addLesson(addLessonCourseId.value, newLesson.value)
    showAddLessonModal.value = false
    await fetchCourses()
  } catch (e: any) {
    error.value = e.response?.data?.error || 'Failed to add lesson'
  } finally {
    addingLesson.value = false
  }
}

function startEditLesson(lesson: any) {
  editingLessonId.value = lesson.id
  editLesson.value = {
    title: lesson.title,
    description: lesson.description,
    video_url: lesson.video_url,
    file_url: lesson.file_url,
    sort_order: lesson.sort_order,
  }
}

async function saveEditLesson(courseId: number) {
  if (!editLesson.value.title.trim()) return
  savingLesson.value = true
  try {
    await coursesApi.updateLesson(courseId, editingLessonId.value!, editLesson.value)
    editingLessonId.value = null
    await fetchCourses()
  } catch (e: any) {
    error.value = e.response?.data?.error || 'Failed to update lesson'
  } finally {
    savingLesson.value = false
  }
}

function toggleCourse(id: number) {
  expandedCourse.value = expandedCourse.value === id ? null : id
}

onMounted(fetchCourses)
</script>

<template>
  <div>
    <!-- Header -->
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-6">
      <h1 class="text-2xl font-extrabold text-cyan-400">Courses</h1>
      <button
        v-if="canManage"
        @click="showCreateModal = true"
        class="flex items-center gap-2 px-4 py-2.5 bg-cyan-400 text-black font-semibold text-sm rounded-xl hover:bg-cyan-300 transition-colors"
      >
        <Plus class="w-4 h-4" />
        New Course
      </button>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="flex justify-center py-16">
      <div class="w-8 h-8 border-2 border-cyan-400/30 border-t-cyan-400 rounded-full animate-spin" />
    </div>

    <!-- Error -->
    <div v-else-if="error" class="text-center py-16">
      <p class="text-red-400 mb-4">{{ error }}</p>
      <button @click="fetchCourses" class="px-4 py-2 bg-cyan-400 text-black rounded-lg font-semibold text-sm">Retry</button>
    </div>

    <!-- Empty -->
    <div v-else-if="courses.length === 0" class="text-center py-16 text-white/40">
      <GraduationCap class="w-16 h-16 mx-auto mb-4 opacity-30" />
      <p>No courses yet</p>
    </div>

    <!-- Course list -->
    <div v-else class="space-y-4">
      <div
        v-for="course in courses"
        :key="course.id"
        class="bg-[#1E1E1E] rounded-2xl border border-white/5 overflow-hidden"
      >
        <!-- Course header -->
        <div
          class="flex items-center gap-4 p-5 cursor-pointer hover:bg-white/[0.02] transition-colors"
          @click="toggleCourse(course.id)"
        >
          <div class="w-12 h-12 rounded-xl bg-cyan-400/10 flex items-center justify-center shrink-0">
            <GraduationCap class="w-6 h-6 text-cyan-400" />
          </div>
          <div class="flex-1 min-w-0">
            <h3 class="font-bold text-white truncate">{{ course.title }}</h3>
            <p class="text-sm text-white/40 mt-0.5 line-clamp-1">{{ course.description || 'No description' }}</p>
          </div>
          <div class="flex items-center gap-3 shrink-0">
            <span class="text-xs text-white/30 bg-white/5 px-2.5 py-1 rounded-lg">
              {{ course.lessons?.length || 0 }} lessons
            </span>
            <component
              :is="expandedCourse === course.id ? ChevronUp : ChevronDown"
              class="w-5 h-5 text-white/30"
            />
          </div>
        </div>

        <!-- Lessons (expanded) -->
        <div v-if="expandedCourse === course.id" class="border-t border-white/5">
          <div v-if="!course.lessons?.length" class="px-5 py-8 text-center text-white/30 text-sm">
            No lessons yet
          </div>

          <div v-else class="divide-y divide-white/5">
            <div
              v-for="lesson in course.lessons"
              :key="lesson.id"
              class="px-5 py-4"
            >
              <!-- View mode -->
              <div v-if="editingLessonId !== lesson.id" class="flex items-start gap-3">
                <div class="w-8 h-8 rounded-lg bg-blue-400/10 flex items-center justify-center shrink-0 mt-0.5">
                  <span class="text-blue-400 text-xs font-bold">{{ lesson.sort_order }}</span>
                </div>
                <div class="flex-1 min-w-0">
                  <p class="font-semibold text-white text-sm">{{ lesson.title }}</p>
                  <p v-if="lesson.description" class="text-white/40 text-xs mt-1 line-clamp-2">{{ lesson.description }}</p>
                  <div class="flex flex-wrap gap-2 mt-2">
                    <span v-if="lesson.video_url" class="inline-flex items-center gap-1 text-xs text-purple-400 bg-purple-400/10 px-2 py-0.5 rounded-md">
                      <Video class="w-3 h-3" /> Video
                    </span>
                    <span v-if="lesson.file_url" class="inline-flex items-center gap-1 text-xs text-green-400 bg-green-400/10 px-2 py-0.5 rounded-md">
                      <FileText class="w-3 h-3" /> File
                    </span>
                  </div>
                </div>
                <button
                  v-if="canManage"
                  @click.stop="startEditLesson(lesson)"
                  class="text-white/20 hover:text-cyan-400 transition-colors p-1"
                >
                  <Pencil class="w-4 h-4" />
                </button>
              </div>

              <!-- Video playback / upload -->
              <div v-if="editingLessonId !== lesson.id" class="mt-4 space-y-3">
                <div v-if="lesson.video_status === 'processing'" class="inline-flex items-center gap-1.5 text-xs text-cyan-400 bg-cyan-400/10 px-2.5 py-1 rounded-md">
                  <Loader2 class="w-3 h-3 animate-spin" /> Обработка…
                </div>
                <div v-else-if="lesson.video_status === 'failed'" class="inline-flex items-center gap-1.5 text-xs text-red-400 bg-red-400/10 px-2.5 py-1 rounded-md">
                  <AlertCircle class="w-3 h-3" /> Ошибка обработки
                </div>

                <VideoPlayer v-if="lesson.hls_url" :src="lesson.hls_url" />

                <VideoUploader
                  v-if="canManage && lesson.video_status !== 'processing'"
                  :lesson-id="lesson.id"
                  @uploaded="onVideoUploaded(course.id)"
                  @failed="fetchCourses"
                />
              </div>

              <!-- Edit mode -->
              <div v-else class="space-y-3">
                <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
                  <input v-model="editLesson.title" placeholder="Title" class="w-full px-3 py-2.5 bg-[#2D2D2D] rounded-xl text-white text-sm placeholder-white/30 border border-transparent focus:border-cyan-400 focus:outline-none" />
                  <input v-model.number="editLesson.sort_order" type="number" placeholder="Order" class="w-full px-3 py-2.5 bg-[#2D2D2D] rounded-xl text-white text-sm placeholder-white/30 border border-transparent focus:border-cyan-400 focus:outline-none" />
                </div>
                <textarea v-model="editLesson.description" rows="2" placeholder="Description" class="w-full px-3 py-2.5 bg-[#2D2D2D] rounded-xl text-white text-sm placeholder-white/30 border border-transparent focus:border-cyan-400 focus:outline-none resize-none" />
                <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
                  <input v-model="editLesson.video_url" placeholder="Video URL" class="w-full px-3 py-2.5 bg-[#2D2D2D] rounded-xl text-white text-sm placeholder-white/30 border border-transparent focus:border-cyan-400 focus:outline-none" />
                  <input v-model="editLesson.file_url" placeholder="File URL" class="w-full px-3 py-2.5 bg-[#2D2D2D] rounded-xl text-white text-sm placeholder-white/30 border border-transparent focus:border-cyan-400 focus:outline-none" />
                </div>
                <div class="flex gap-2 justify-end">
                  <button @click="editingLessonId = null" class="px-3 py-2 text-white/40 text-sm font-semibold hover:text-white transition-colors">Cancel</button>
                  <button @click="saveEditLesson(course.id)" :disabled="savingLesson" class="flex items-center gap-1.5 px-4 py-2 bg-cyan-400 text-black text-sm font-semibold rounded-lg hover:bg-cyan-300 disabled:opacity-50 transition-colors">
                    <Save class="w-3.5 h-3.5" />
                    {{ savingLesson ? 'Saving...' : 'Save' }}
                  </button>
                </div>
              </div>
            </div>
          </div>

          <!-- Add lesson button -->
          <div v-if="canManage" class="px-5 py-3 border-t border-white/5">
            <button
              @click.stop="openAddLesson(course.id, course.lessons?.length || 0)"
              class="flex items-center gap-2 text-sm text-cyan-400 font-semibold hover:text-cyan-300 transition-colors"
            >
              <Plus class="w-4 h-4" />
              Add Lesson
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- ============ CREATE COURSE MODAL ============ -->
    <Teleport to="body">
      <Transition name="fade">
        <div v-if="showCreateModal" class="fixed inset-0 z-50 flex items-end sm:items-center justify-center bg-black/60 p-4" @click.self="showCreateModal = false">
          <div class="bg-[#1E1E1E] rounded-t-3xl sm:rounded-2xl w-full max-w-lg max-h-[85vh] flex flex-col">
            <!-- Header -->
            <div class="flex items-center justify-between px-6 py-4 border-b border-white/5 shrink-0">
              <h2 class="text-lg font-extrabold text-white">New Course</h2>
              <button @click="showCreateModal = false" class="text-white/40 hover:text-white"><X class="w-5 h-5" /></button>
            </div>

            <!-- Body -->
            <div class="flex-1 overflow-y-auto px-6 py-4 space-y-4">
              <div>
                <label class="block text-sm font-semibold text-white/70 mb-1.5">Title</label>
                <input v-model="newCourse.title" placeholder="Course title" class="w-full px-4 py-3 bg-[#2D2D2D] rounded-xl text-white text-sm placeholder-white/30 border border-transparent focus:border-cyan-400 focus:outline-none" />
              </div>
              <div>
                <label class="block text-sm font-semibold text-white/70 mb-1.5">Description</label>
                <textarea v-model="newCourse.description" rows="2" placeholder="Course description" class="w-full px-4 py-3 bg-[#2D2D2D] rounded-xl text-white text-sm placeholder-white/30 border border-transparent focus:border-cyan-400 focus:outline-none resize-none" />
              </div>
              <div>
                <label class="block text-sm font-semibold text-white/70 mb-1.5">Image URL</label>
                <input v-model="newCourse.image_url" placeholder="/uploads/course.png" class="w-full px-4 py-3 bg-[#2D2D2D] rounded-xl text-white text-sm placeholder-white/30 border border-transparent focus:border-cyan-400 focus:outline-none" />
              </div>

              <!-- Lessons section -->
              <div>
                <div class="flex items-center justify-between mb-3">
                  <label class="text-sm font-semibold text-white/70">Lessons</label>
                  <button @click="addLessonToForm" type="button" class="flex items-center gap-1 text-xs text-cyan-400 font-semibold hover:text-cyan-300">
                    <Plus class="w-3.5 h-3.5" /> Add
                  </button>
                </div>

                <div v-if="newCourse.lessons.length === 0" class="text-center py-4 text-white/20 text-sm border border-dashed border-white/10 rounded-xl">
                  No lessons — you can add them now or later
                </div>

                <div v-else class="space-y-3">
                  <div
                    v-for="(lesson, idx) in newCourse.lessons"
                    :key="idx"
                    class="bg-[#252525] rounded-xl p-4 space-y-3 relative"
                  >
                    <button @click="removeLessonFromForm(idx)" class="absolute top-3 right-3 text-white/20 hover:text-red-400 transition-colors">
                      <Trash2 class="w-4 h-4" />
                    </button>

                    <div class="flex items-center gap-2 mb-1">
                      <BookOpen class="w-4 h-4 text-blue-400" />
                      <span class="text-xs text-white/40 font-semibold">Lesson {{ idx + 1 }}</span>
                    </div>

                    <input v-model="lesson.title" placeholder="Lesson title" class="w-full px-3 py-2.5 bg-[#2D2D2D] rounded-lg text-white text-sm placeholder-white/30 border border-transparent focus:border-cyan-400 focus:outline-none" />
                    <textarea v-model="lesson.description" rows="2" placeholder="Description" class="w-full px-3 py-2.5 bg-[#2D2D2D] rounded-lg text-white text-sm placeholder-white/30 border border-transparent focus:border-cyan-400 focus:outline-none resize-none" />
                    <div class="grid grid-cols-2 gap-2">
                      <input v-model="lesson.video_url" placeholder="Video URL" class="w-full px-3 py-2.5 bg-[#2D2D2D] rounded-lg text-white text-sm placeholder-white/30 border border-transparent focus:border-cyan-400 focus:outline-none" />
                      <input v-model="lesson.file_url" placeholder="File URL" class="w-full px-3 py-2.5 bg-[#2D2D2D] rounded-lg text-white text-sm placeholder-white/30 border border-transparent focus:border-cyan-400 focus:outline-none" />
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <!-- Footer -->
            <div class="flex gap-3 px-6 py-4 border-t border-white/5 shrink-0">
              <button @click="showCreateModal = false" class="flex-1 py-3 rounded-xl text-white/60 font-semibold text-sm hover:bg-white/5 transition-colors">Cancel</button>
              <button @click="createCourse" :disabled="creating" class="flex-1 py-3 rounded-xl bg-cyan-400 text-black font-semibold text-sm hover:bg-cyan-300 disabled:opacity-50 transition-colors">
                {{ creating ? 'Creating...' : 'Create Course' }}
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>

    <!-- ============ ADD LESSON MODAL ============ -->
    <Teleport to="body">
      <Transition name="fade">
        <div v-if="showAddLessonModal" class="fixed inset-0 z-50 flex items-end sm:items-center justify-center bg-black/60 p-4" @click.self="showAddLessonModal = false">
          <div class="bg-[#1E1E1E] rounded-t-3xl sm:rounded-2xl w-full max-w-md p-6">
            <div class="flex items-center justify-between mb-6">
              <h2 class="text-lg font-extrabold text-white">Add Lesson</h2>
              <button @click="showAddLessonModal = false" class="text-white/40 hover:text-white"><X class="w-5 h-5" /></button>
            </div>

            <form @submit.prevent="submitAddLesson" class="space-y-4">
              <div>
                <label class="block text-sm font-semibold text-white/70 mb-1.5">Title</label>
                <input v-model="newLesson.title" placeholder="Lesson title" class="w-full px-4 py-3 bg-[#2D2D2D] rounded-xl text-white text-sm placeholder-white/30 border border-transparent focus:border-cyan-400 focus:outline-none" />
              </div>
              <div>
                <label class="block text-sm font-semibold text-white/70 mb-1.5">Description</label>
                <textarea v-model="newLesson.description" rows="3" placeholder="Lesson description" class="w-full px-4 py-3 bg-[#2D2D2D] rounded-xl text-white text-sm placeholder-white/30 border border-transparent focus:border-cyan-400 focus:outline-none resize-none" />
              </div>
              <div>
                <label class="block text-sm font-semibold text-white/70 mb-1.5">Video URL</label>
                <input v-model="newLesson.video_url" placeholder="/videos/lesson.mp4" class="w-full px-4 py-3 bg-[#2D2D2D] rounded-xl text-white text-sm placeholder-white/30 border border-transparent focus:border-cyan-400 focus:outline-none" />
              </div>
              <div>
                <label class="block text-sm font-semibold text-white/70 mb-1.5">File URL</label>
                <input v-model="newLesson.file_url" placeholder="/uploads/material.pdf" class="w-full px-4 py-3 bg-[#2D2D2D] rounded-xl text-white text-sm placeholder-white/30 border border-transparent focus:border-cyan-400 focus:outline-none" />
              </div>
              <div class="flex gap-3 pt-2">
                <button type="button" @click="showAddLessonModal = false" class="flex-1 py-3 rounded-xl text-white/60 font-semibold text-sm hover:bg-white/5 transition-colors">Cancel</button>
                <button type="submit" :disabled="addingLesson" class="flex-1 py-3 rounded-xl bg-cyan-400 text-black font-semibold text-sm hover:bg-cyan-300 disabled:opacity-50 transition-colors">
                  {{ addingLesson ? 'Adding...' : 'Add Lesson' }}
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
