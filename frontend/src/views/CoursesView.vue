<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { coursesApi, deleteBlobFile, isVercelBlobUrl, resolveMediaUrl, uploadsApi, videoApi, type LessonInput } from '@/api/client'
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
const actionError = ref('')
const createError = ref('')
const addLessonError = ref('')
const editLessonError = ref('')

const showCreateModal = ref(false)
const creating = ref(false)
const newCourse = ref({
  title: '',
  description: '',
  image_url: '',
  lessons: [] as LessonInput[],
})

const expandedCourse = ref<number | null>(null)

const showAddLessonModal = ref(false)
const addLessonCourseId = ref(0)
const addingLesson = ref(false)
const newLesson = ref<LessonInput>(emptyLesson())
const newLessonVideoFiles = ref<File[]>([])
const newLessonMaterialFiles = ref<File[]>([])

const editingLessonId = ref<number | null>(null)
const editLesson = ref<LessonInput>(emptyLesson())
const editLessonVideoFiles = ref<File[]>([])
const editLessonMaterialFiles = ref<File[]>([])
const editLessonRemoveMaterial = ref(false)
const savingLesson = ref(false)
const deletingCourseId = ref<number | null>(null)
const deletingLessonId = ref<number | null>(null)

function emptyLesson(): LessonInput {
  return { title: '', description: '', video_url: '', file_url: '', sort_order: 0 }
}

function resetCreateForm() {
  newCourse.value = { title: '', description: '', image_url: '', lessons: [] }
  createError.value = ''
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

function validOptionalUrl(value: string) {
  const trimmed = value.trim()
  if (!trimmed) return true
  if (trimmed.startsWith('/') && !trimmed.startsWith('//')) return true
  try {
    const parsed = new URL(trimmed)
    return parsed.protocol === 'http:' || parsed.protocol === 'https:'
  } catch {
    return false
  }
}

function setNewLessonVideo(e: Event) {
  newLessonVideoFiles.value = Array.from((e.target as HTMLInputElement).files || [])
}

function setNewLessonMaterial(e: Event) {
  newLessonMaterialFiles.value = Array.from((e.target as HTMLInputElement).files || [])
}

function setEditLessonVideo(e: Event) {
  editLessonVideoFiles.value = Array.from((e.target as HTMLInputElement).files || [])
}

function setEditLessonMaterial(e: Event) {
  editLessonMaterialFiles.value = Array.from((e.target as HTMLInputElement).files || [])
  if (editLessonMaterialFiles.value.length > 0) editLessonRemoveMaterial.value = false
}

function removeEditMaterial() {
  editLessonMaterialFiles.value = []
  editLesson.value.file_url = ''
  editLessonRemoveMaterial.value = true
}

function validateLessonInput(lesson: LessonInput, label: string) {
  if (!lesson.title.trim()) return `${label}: title is required`
  if (lesson.title.trim().length > 120) return `${label}: title must be 120 characters or less`
  if (lesson.description.trim().length > 2000) return `${label}: description must be 2000 characters or less`
  if (lesson.sort_order < 0) return `${label}: order cannot be negative`
  if (!validOptionalUrl(lesson.video_url)) return `${label}: video URL must start with http(s):// or /`
  if (!validOptionalUrl(lesson.file_url)) return `${label}: file URL must start with http(s):// or /`
  return ''
}

function lessonVideoAssets(lesson: any) {
  const assets = (lesson.assets || []).filter((asset: any) => asset.type === 'video')
  const fallbackUrl = lesson.hls_url || lesson.video_url
  if (fallbackUrl && !assets.some((asset: any) => asset.url === fallbackUrl)) {
    assets.push({ id: 0, lesson_id: lesson.id, type: 'video', url: fallbackUrl, original_filename: 'Video' })
  }
  return assets
}

function lessonFileAssets(lesson: any) {
  const assets = (lesson.assets || []).filter((asset: any) => asset.type === 'file')
  if (lesson.file_url && !assets.some((asset: any) => asset.url === lesson.file_url)) {
    assets.unshift({ id: 0, lesson_id: lesson.id, type: 'file', url: lesson.file_url, original_filename: 'Material' })
  }
  return assets
}

async function createCourse() {
  createError.value = ''
  if (!newCourse.value.title.trim()) {
    createError.value = 'Course title is required'
    return
  }
  if (newCourse.value.title.trim().length > 120) {
    createError.value = 'Course title must be 120 characters or less'
    return
  }
  if (newCourse.value.description.trim().length > 2000) {
    createError.value = 'Course description must be 2000 characters or less'
    return
  }
  for (let i = 0; i < newCourse.value.lessons.length; i += 1) {
    const lesson = newCourse.value.lessons[i]
    if (!lesson) continue
    const lessonError = validateLessonInput(lesson, `Lesson ${i + 1}`)
    if (lessonError) {
      createError.value = lessonError
      return
    }
  }

  creating.value = true
  try {
    await coursesApi.create(newCourse.value)
    showCreateModal.value = false
    resetCreateForm()
    await fetchCourses()
  } catch (e: any) {
    createError.value = e.response?.data?.error || 'Failed to create course'
  } finally {
    creating.value = false
  }
}

function openAddLesson(courseId: number, lessonCount: number) {
  addLessonCourseId.value = courseId
  newLesson.value = { ...emptyLesson(), sort_order: lessonCount + 1 }
  newLessonVideoFiles.value = []
  newLessonMaterialFiles.value = []
  addLessonError.value = ''
  showAddLessonModal.value = true
}

async function submitAddLesson() {
  addLessonError.value = validateLessonInput(newLesson.value, 'Lesson')
  if (addLessonError.value) return

  addingLesson.value = true
  try {
    const lessonPayload = { ...newLesson.value }
    const firstMaterialFile = newLessonMaterialFiles.value[0]
    let firstMaterialUrl = ''
    if (firstMaterialFile) {
      const material = await uploadsApi.upload(firstMaterialFile)
      lessonPayload.file_url = material.data.url
      firstMaterialUrl = material.data.url
    }
    const created = await coursesApi.addLesson(addLessonCourseId.value, lessonPayload)
    if (firstMaterialFile && firstMaterialUrl) {
      await coursesApi.addLessonAsset(created.data.id, { type: 'file', url: firstMaterialUrl, original_filename: firstMaterialFile.name })
    }
    for (let i = 1; i < newLessonMaterialFiles.value.length; i += 1) {
      const file = newLessonMaterialFiles.value[i]
      if (!file) continue
      const material = await uploadsApi.upload(file)
      await coursesApi.addLessonAsset(created.data.id, { type: 'file', url: material.data.url, original_filename: file.name })
    }
    for (const file of newLessonVideoFiles.value) {
      await videoApi.upload(created.data.id, file)
    }
    showAddLessonModal.value = false
    await fetchCourses()
  } catch (e: any) {
    addLessonError.value = e.response?.data?.error || 'Failed to add lesson'
  } finally {
    addingLesson.value = false
  }
}

function startEditLesson(lesson: any) {
  editingLessonId.value = lesson.id
  editLessonError.value = ''
  editLessonVideoFiles.value = []
  editLessonMaterialFiles.value = []
  editLessonRemoveMaterial.value = false
  editLesson.value = {
    title: lesson.title,
    description: lesson.description,
    video_url: lesson.video_url,
    file_url: lesson.file_url,
    sort_order: lesson.sort_order,
  }
}

async function saveEditLesson(courseId: number) {
  editLessonError.value = validateLessonInput(editLesson.value, 'Lesson')
  if (editLessonError.value) return

  savingLesson.value = true
  try {
    const lessonPayload = { ...editLesson.value }
    if (editLessonRemoveMaterial.value) {
      if (auth.isAdmin && isVercelBlobUrl(editLesson.value.file_url)) {
        await deleteBlobFile(editLesson.value.file_url)
      }
      lessonPayload.file_url = ''
    }
    const firstMaterialFile = editLessonMaterialFiles.value[0]
    let firstMaterialUrl = ''
    if (firstMaterialFile) {
      const material = await uploadsApi.upload(firstMaterialFile)
      lessonPayload.file_url = material.data.url
      firstMaterialUrl = material.data.url
    }
    await coursesApi.updateLesson(courseId, editingLessonId.value!, lessonPayload)
    if (firstMaterialFile && firstMaterialUrl) {
      await coursesApi.addLessonAsset(editingLessonId.value!, { type: 'file', url: firstMaterialUrl, original_filename: firstMaterialFile.name })
    }
    for (let i = 1; i < editLessonMaterialFiles.value.length; i += 1) {
      const file = editLessonMaterialFiles.value[i]
      if (!file) continue
      const material = await uploadsApi.upload(file)
      await coursesApi.addLessonAsset(editingLessonId.value!, { type: 'file', url: material.data.url, original_filename: file.name })
    }
    for (const file of editLessonVideoFiles.value) {
      await videoApi.upload(editingLessonId.value!, file)
    }
    editingLessonId.value = null
    await fetchCourses()
  } catch (e: any) {
    editLessonError.value = e.response?.data?.error || 'Failed to update lesson'
  } finally {
    savingLesson.value = false
  }
}

function toggleCourse(id: number) {
  expandedCourse.value = expandedCourse.value === id ? null : id
}

async function deleteCourse(id: number) {
  if (!confirm('Delete this course and all lessons?')) return
  deletingCourseId.value = id
  try {
    await coursesApi.delete(id)
    await fetchCourses()
  } catch (e: any) {
    error.value = e.response?.data?.error || 'Failed to delete course'
  } finally {
    deletingCourseId.value = null
  }
}

async function deleteLesson(courseId: number, lessonId: number) {
  if (!confirm('Delete this lesson?')) return
  deletingLessonId.value = lessonId
  try {
    await coursesApi.deleteLesson(courseId, lessonId)
    await fetchCourses()
  } catch (e: any) {
    error.value = e.response?.data?.error || 'Failed to delete lesson'
  } finally {
    deletingLessonId.value = null
  }
}

async function deleteLessonAsset(courseId: number, lesson: any, asset: any) {
  if (!confirm('Remove this attachment?')) return
  actionError.value = ''
  try {
    let blobDeleteError = ''
    if (auth.isAdmin && isVercelBlobUrl(asset.url)) {
      await deleteBlobFile(asset.url).catch((err: any) => {
        blobDeleteError = err.message || 'Failed to delete file from storage'
      })
    }
    if (asset.id) {
      await coursesApi.deleteLessonAsset(lesson.id, asset.id)
      if (asset.type === 'video' && (lesson.hls_url === asset.url || lesson.video_url === asset.url)) {
        await videoApi.clear(lesson.id)
        if (editingLessonId.value === lesson.id) {
          editLesson.value.video_url = ''
        }
      }
      if (asset.type === 'file' && lesson.file_url === asset.url) {
        await coursesApi.updateLesson(courseId, lesson.id, {
          title: lesson.title,
          description: lesson.description,
          video_url: lesson.video_url || '',
          file_url: '',
          sort_order: lesson.sort_order,
        })
        if (editingLessonId.value === lesson.id) editLesson.value.file_url = ''
      }
    } else if (asset.type === 'video') {
      await videoApi.clear(lesson.id)
      if (editingLessonId.value === lesson.id) {
        editLesson.value.video_url = ''
      }
    } else if (asset.type === 'file') {
      await coursesApi.updateLesson(courseId, lesson.id, {
        title: lesson.title,
        description: lesson.description,
        video_url: lesson.video_url || '',
        file_url: '',
        sort_order: lesson.sort_order,
      })
      if (editingLessonId.value === lesson.id) editLesson.value.file_url = ''
    }
    await fetchCourses()
    if (blobDeleteError) {
      actionError.value = `${blobDeleteError}. Attachment was removed from the lesson.`
    }
  } catch (e: any) {
    const status = e.response?.status
    actionError.value = status === 405
      ? 'Delete is not available on the running backend. Restart the backend so the new attachment routes are loaded.'
      : e.response?.data?.error || e.message || 'Failed to remove attachment'
  }
}

onMounted(fetchCourses)
</script>

<template>
  <div>
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

    <div v-if="loading" class="flex justify-center py-16">
      <div class="w-8 h-8 border-2 border-cyan-400/30 border-t-cyan-400 rounded-full animate-spin" />
    </div>

    <div v-else-if="error" class="text-center py-16">
      <p class="text-red-400 mb-4">{{ error }}</p>
      <button @click="fetchCourses" class="px-4 py-2 bg-cyan-400 text-black rounded-lg font-semibold text-sm">Retry</button>
    </div>

    <div v-else>
      <div v-if="actionError" class="mb-4 flex items-center justify-between gap-3 rounded-xl border border-red-400/20 bg-red-400/10 px-4 py-3 text-sm text-red-300">
        <span>{{ actionError }}</span>
        <button type="button" class="text-red-200 hover:text-white" @click="actionError = ''">
          <X class="w-4 h-4" />
        </button>
      </div>

      <div v-if="courses.length === 0" class="text-center py-16 text-white/40">
      <GraduationCap class="w-16 h-16 mx-auto mb-4 opacity-30" />
      <p>No courses yet</p>
      </div>

      <div v-else class="space-y-4">
      <div
        v-for="course in courses"
        :key="course.id"
        class="bg-[#1E1E1E] rounded-2xl border border-white/5 overflow-hidden"
      >
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
            <button
              v-if="canManage"
              @click.stop="deleteCourse(course.id)"
              :disabled="deletingCourseId === course.id"
              class="w-9 h-9 rounded-lg bg-red-400/10 flex items-center justify-center text-red-400 hover:bg-red-400/20 disabled:opacity-50 transition-colors"
              title="Delete course"
            >
              <Trash2 class="w-4 h-4" />
            </button>
          </div>
        </div>

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
              <div v-if="editingLessonId !== lesson.id" class="flex items-start gap-3">
                <div class="w-8 h-8 rounded-lg bg-blue-400/10 flex items-center justify-center shrink-0 mt-0.5">
                  <span class="text-blue-400 text-xs font-bold">{{ lesson.sort_order }}</span>
                </div>
                <div class="flex-1 min-w-0">
                  <p class="font-semibold text-white text-sm">{{ lesson.title }}</p>
                  <p v-if="lesson.description" class="text-white/40 text-xs mt-1 line-clamp-2">{{ lesson.description }}</p>
                  <div class="flex flex-wrap gap-2 mt-2">
                    <span v-if="lessonVideoAssets(lesson).length" class="inline-flex items-center gap-1 text-xs text-purple-400 bg-purple-400/10 px-2 py-0.5 rounded-md">
                      <Video class="w-3 h-3" /> {{ lessonVideoAssets(lesson).length }} video
                    </span>
                    <span v-if="lessonFileAssets(lesson).length" class="inline-flex items-center gap-1 text-xs text-green-400 bg-green-400/10 px-2 py-0.5 rounded-md">
                      <FileText class="w-3 h-3" /> {{ lessonFileAssets(lesson).length }} file
                    </span>
                  </div>
                </div>
                <button
                  v-if="canManage"
                  @click.stop="startEditLesson(lesson)"
                  class="text-white/20 hover:text-cyan-400 transition-colors p-1"
                  title="Edit lesson"
                >
                  <Pencil class="w-4 h-4" />
                </button>
                <button
                  v-if="canManage"
                  @click.stop="deleteLesson(course.id, lesson.id)"
                  :disabled="deletingLessonId === lesson.id"
                  class="text-white/20 hover:text-red-400 disabled:opacity-50 transition-colors p-1"
                  title="Delete lesson"
                >
                  <Trash2 class="w-4 h-4" />
                </button>
              </div>

              <div v-if="editingLessonId !== lesson.id" class="mt-4 space-y-3">
                <div v-if="lesson.video_status === 'processing'" class="inline-flex items-center gap-1.5 text-xs text-cyan-400 bg-cyan-400/10 px-2.5 py-1 rounded-md">
                  <Loader2 class="w-3 h-3 animate-spin" /> Обработка…
                </div>
                <div v-else-if="lesson.video_status === 'failed'" class="inline-flex items-center gap-1.5 text-xs text-red-400 bg-red-400/10 px-2.5 py-1 rounded-md">
                  <AlertCircle class="w-3 h-3" /> Ошибка обработки
                </div>

                <div v-if="lessonVideoAssets(lesson).length" class="space-y-3">
                  <div
                    v-for="asset in lessonVideoAssets(lesson)"
                    :key="`video-${asset.id}-${asset.url}`"
                    class="space-y-2"
                  >
                    <div class="flex items-center justify-between gap-3 text-xs text-white/50">
                      <span class="truncate">{{ asset.original_filename || 'Video' }}</span>
                      <button
                        v-if="canManage"
                        type="button"
                        @click="deleteLessonAsset(course.id, lesson, { ...asset, type: 'video' })"
                        class="text-red-400 hover:text-red-300"
                      >
                        Remove
                      </button>
                    </div>
                    <VideoPlayer :src="asset.url" />
                  </div>
                </div>
                <div v-if="lessonFileAssets(lesson).length" class="flex flex-wrap gap-2">
                  <span
                    v-for="asset in lessonFileAssets(lesson)"
                    :key="`file-${asset.id}-${asset.url}`"
                    class="inline-flex items-center gap-2 rounded-lg bg-green-400/10 px-3 py-2 text-sm text-green-400"
                  >
                    <a :href="resolveMediaUrl(asset.url)" target="_blank" class="hover:text-green-300">
                      {{ asset.original_filename || 'Open material' }}
                    </a>
                    <button
                      v-if="canManage"
                      type="button"
                      @click="deleteLessonAsset(course.id, lesson, { ...asset, type: 'file' })"
                      class="text-red-400 hover:text-red-300"
                    >
                      <X class="w-3.5 h-3.5" />
                    </button>
                  </span>
                </div>

                <VideoUploader
                  v-if="canManage && lesson.video_status !== 'processing'"
                  :lesson-id="lesson.id"
                  @uploaded="onVideoUploaded(course.id)"
                  @failed="fetchCourses"
                />
              </div>

              <div v-else class="space-y-3">
                <p v-if="editLessonError" class="text-sm text-red-400">{{ editLessonError }}</p>
                <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
                  <input v-model="editLesson.title" maxlength="120" placeholder="Title" class="w-full px-3 py-2.5 bg-[#2D2D2D] rounded-xl text-white text-sm placeholder-white/30 border border-transparent focus:border-cyan-400 focus:outline-none" />
                  <input v-model.number="editLesson.sort_order" type="number" placeholder="Order" class="w-full px-3 py-2.5 bg-[#2D2D2D] rounded-xl text-white text-sm placeholder-white/30 border border-transparent focus:border-cyan-400 focus:outline-none" />
                </div>
                <textarea v-model="editLesson.description" maxlength="2000" rows="2" placeholder="Description" class="w-full px-3 py-2.5 bg-[#2D2D2D] rounded-xl text-white text-sm placeholder-white/30 border border-transparent focus:border-cyan-400 focus:outline-none resize-none" />
                <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
                  <div class="bg-[#2D2D2D] rounded-xl border border-transparent p-3">
                    <div class="flex items-start justify-between gap-3">
                      <label class="block flex-1 text-white/70 text-sm cursor-pointer">
                        Video files
                        <input type="file" multiple accept=".mp4,.mov,.mkv,video/*" class="hidden" @change="setEditLessonVideo" />
                        <span class="block mt-1 text-xs text-white/40 truncate">
                          {{ editLessonVideoFiles.length ? `${editLessonVideoFiles.length} selected` : 'Choose videos' }}
                        </span>
                      </label>
                    </div>
                  </div>
                  <div class="bg-[#2D2D2D] rounded-xl border border-transparent p-3">
                    <div class="flex items-start justify-between gap-3">
                      <label class="block flex-1 text-white/70 text-sm cursor-pointer">
                        Material files
                        <input type="file" multiple class="hidden" @change="setEditLessonMaterial" />
                        <span class="block mt-1 text-xs text-white/40 truncate">
                          {{ editLessonMaterialFiles.length ? `${editLessonMaterialFiles.length} selected` : (editLessonRemoveMaterial ? 'Materials will be removed' : editLesson.file_url || 'Choose files') }}
                        </span>
                      </label>
                      <button
                        v-if="editLesson.file_url && !editLessonRemoveMaterial"
                        type="button"
                        @click="removeEditMaterial"
                        class="inline-flex items-center gap-1.5 rounded-lg bg-red-400/10 px-3 py-2 text-xs font-semibold text-red-400 hover:bg-red-400/20"
                      >
                        <Trash2 class="w-3.5 h-3.5" />
                        Remove
                      </button>
                    </div>
                  </div>
                </div>
                <div v-if="lessonVideoAssets(lesson).length || lessonFileAssets(lesson).length" class="grid grid-cols-1 sm:grid-cols-2 gap-3">
                  <div v-if="lessonVideoAssets(lesson).length" class="space-y-2 rounded-xl bg-white/[0.03] p-3">
                    <p class="text-xs font-semibold text-white/50">Uploaded videos</p>
                    <div
                      v-for="asset in lessonVideoAssets(lesson)"
                      :key="`edit-video-${asset.id}-${asset.url}`"
                      class="flex items-center justify-between gap-3 text-xs text-white/50"
                    >
                      <span class="truncate">{{ asset.original_filename || 'Video' }}</span>
                      <button
                        type="button"
                        @click="deleteLessonAsset(course.id, lesson, { ...asset, type: 'video' })"
                        class="text-red-400 hover:text-red-300"
                      >
                        Remove
                      </button>
                    </div>
                  </div>
                  <div v-if="lessonFileAssets(lesson).length" class="space-y-2 rounded-xl bg-white/[0.03] p-3">
                    <p class="text-xs font-semibold text-white/50">Uploaded files</p>
                    <div
                      v-for="asset in lessonFileAssets(lesson)"
                      :key="`edit-file-${asset.id}-${asset.url}`"
                      class="flex items-center justify-between gap-3 text-xs text-white/50"
                    >
                      <a :href="resolveMediaUrl(asset.url)" target="_blank" class="truncate hover:text-cyan-300">
                        {{ asset.original_filename || 'Material' }}
                      </a>
                      <button
                        type="button"
                        @click="deleteLessonAsset(course.id, lesson, { ...asset, type: 'file' })"
                        class="text-red-400 hover:text-red-300"
                      >
                        Remove
                      </button>
                    </div>
                  </div>
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
    </div>

    <Teleport to="body">
      <Transition name="fade">
        <div v-if="showCreateModal" class="fixed inset-0 z-50 flex items-end sm:items-center justify-center bg-black/60 p-4" @click.self="showCreateModal = false">
          <div class="bg-[#1E1E1E] rounded-t-3xl sm:rounded-2xl w-full max-w-lg max-h-[85vh] flex flex-col">
            <div class="flex items-center justify-between px-6 py-4 border-b border-white/5 shrink-0">
              <h2 class="text-lg font-extrabold text-white">New Course</h2>
              <button @click="showCreateModal = false" class="text-white/40 hover:text-white"><X class="w-5 h-5" /></button>
            </div>

            <div class="flex-1 overflow-y-auto px-6 py-4 space-y-4">
              <p v-if="createError" class="rounded-xl bg-red-400/10 px-4 py-3 text-sm text-red-400">{{ createError }}</p>
              <div>
                <label class="block text-sm font-semibold text-white/70 mb-1.5">Title</label>
                <input v-model="newCourse.title" maxlength="120" placeholder="Course title" class="w-full px-4 py-3 bg-[#2D2D2D] rounded-xl text-white text-sm placeholder-white/30 border border-transparent focus:border-cyan-400 focus:outline-none" />
              </div>
              <div>
                <label class="block text-sm font-semibold text-white/70 mb-1.5">Description</label>
                <textarea v-model="newCourse.description" maxlength="2000" rows="2" placeholder="Course description" class="w-full px-4 py-3 bg-[#2D2D2D] rounded-xl text-white text-sm placeholder-white/30 border border-transparent focus:border-cyan-400 focus:outline-none resize-none" />
              </div>
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

                    <input v-model="lesson.title" maxlength="120" placeholder="Lesson title" class="w-full px-3 py-2.5 bg-[#2D2D2D] rounded-lg text-white text-sm placeholder-white/30 border border-transparent focus:border-cyan-400 focus:outline-none" />
                    <textarea v-model="lesson.description" maxlength="2000" rows="2" placeholder="Description" class="w-full px-3 py-2.5 bg-[#2D2D2D] rounded-lg text-white text-sm placeholder-white/30 border border-transparent focus:border-cyan-400 focus:outline-none resize-none" />
                  </div>
                </div>
              </div>
            </div>

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

    <Teleport to="body">
      <Transition name="fade">
        <div v-if="showAddLessonModal" class="fixed inset-0 z-50 flex items-end sm:items-center justify-center bg-black/60 p-4" @click.self="showAddLessonModal = false">
          <div class="bg-[#1E1E1E] rounded-t-3xl sm:rounded-2xl w-full max-w-md p-6">
            <div class="flex items-center justify-between mb-6">
              <h2 class="text-lg font-extrabold text-white">Add Lesson</h2>
              <button @click="showAddLessonModal = false" class="text-white/40 hover:text-white"><X class="w-5 h-5" /></button>
            </div>

            <form @submit.prevent="submitAddLesson" class="space-y-4">
              <p v-if="addLessonError" class="rounded-xl bg-red-400/10 px-4 py-3 text-sm text-red-400">{{ addLessonError }}</p>
              <div>
                <label class="block text-sm font-semibold text-white/70 mb-1.5">Title</label>
                <input v-model="newLesson.title" maxlength="120" placeholder="Lesson title" class="w-full px-4 py-3 bg-[#2D2D2D] rounded-xl text-white text-sm placeholder-white/30 border border-transparent focus:border-cyan-400 focus:outline-none" />
              </div>
              <div>
                <label class="block text-sm font-semibold text-white/70 mb-1.5">Description</label>
                <textarea v-model="newLesson.description" maxlength="2000" rows="3" placeholder="Lesson description" class="w-full px-4 py-3 bg-[#2D2D2D] rounded-xl text-white text-sm placeholder-white/30 border border-transparent focus:border-cyan-400 focus:outline-none resize-none" />
              </div>
              <div>
                <label class="block text-sm font-semibold text-white/70 mb-1.5">Video Files</label>
                <label class="block px-4 py-3 bg-[#2D2D2D] rounded-xl text-white/70 text-sm border border-transparent hover:border-cyan-400 cursor-pointer">
                  <input type="file" multiple accept=".mp4,.mov,.mkv,video/*" class="hidden" @change="setNewLessonVideo" />
                  <span class="block truncate">{{ newLessonVideoFiles.length ? `${newLessonVideoFiles.length} selected` : 'Choose videos' }}</span>
                </label>
              </div>
              <div>
                <label class="block text-sm font-semibold text-white/70 mb-1.5">Material Files</label>
                <label class="block px-4 py-3 bg-[#2D2D2D] rounded-xl text-white/70 text-sm border border-transparent hover:border-cyan-400 cursor-pointer">
                  <input type="file" multiple class="hidden" @change="setNewLessonMaterial" />
                  <span class="block truncate">{{ newLessonMaterialFiles.length ? `${newLessonMaterialFiles.length} selected` : 'Choose files' }}</span>
                </label>
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
