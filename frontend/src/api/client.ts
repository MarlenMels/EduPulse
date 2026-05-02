import axios from 'axios'
import { upload as uploadBlob } from '@vercel/blob/client'

const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL || '/api',
  headers: {
    'Content-Type': 'application/json',
    'Accept': 'application/json',
  },
})

const apiBaseURL = import.meta.env.VITE_API_URL || '/api'

export function resolveMediaUrl(src: string) {
  if (!src || src.startsWith('http://') || src.startsWith('https://') || src.startsWith('blob:')) {
    return src
  }
  if (!src.startsWith('/')) {
    return src
  }
  if (apiBaseURL === '/api') {
    return `/api${src}`
  }
  return `${apiBaseURL.replace(/\/$/, '')}${src}`
}

export function isVercelBlobUrl(src: string) {
  try {
    const url = new URL(src)
    return url.hostname.endsWith('.vercel-storage.com')
  } catch {
    return false
  }
}

function isLocalHost() {
  if (typeof window === 'undefined') return false
  return ['localhost', '127.0.0.1', '::1'].includes(window.location.hostname)
}

function canUseBlobServerlessApi() {
  if (typeof window === 'undefined') return false
  if (import.meta.env.VITE_ENABLE_LOCAL_BLOB_API === 'true') return true
  return window.location.hostname.endsWith('.vercel.app')
}

export async function deleteBlobFile(url: string) {
  if (!canUseBlobServerlessApi()) return false
  const token = localStorage.getItem('token') || ''
  await fetch('/api/blob-delete', {
    method: 'DELETE',
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
    },
    body: JSON.stringify({ url }),
  }).then(async (res) => {
    if (!res.ok) {
      const data = await res.json().catch(() => ({}))
      throw new Error(data.error || 'Failed to delete file')
    }
  })
  return true
}

api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('token')
      window.location.href = '/login'
    }
    return Promise.reject(error)
  },
)

export default api

function shouldUseBlobUploads() {
  if (isLocalHost() && import.meta.env.VITE_ENABLE_LOCAL_BLOB_API !== 'true') return false
  if (import.meta.env.VITE_USE_VERCEL_BLOB === 'true') return true
  return typeof window !== 'undefined' && window.location.hostname.endsWith('.vercel.app')
}

function blobPath(prefix: 'videos' | 'materials', file: File) {
  const safeName = file.name.replace(/[^a-zA-Z0-9._-]/g, '_')
  return `edupulse/${prefix}/${Date.now()}-${safeName}`
}

async function uploadToBlob(prefix: 'videos' | 'materials', file: File, onProgress?: (pct: number) => void) {
  const token = localStorage.getItem('token') || ''
  const abortController = new AbortController()
  let rejectStall: (error: Error) => void = () => {}
  const stallTimeout = new Promise<never>((_, reject) => {
    rejectStall = reject
  })
  const failStalledUpload = () => {
    abortController.abort()
    rejectStall(new Error('Upload service did not respond for 20 seconds. Check /api/blob-upload and Vercel Blob environment variables.'))
  }
  let timeout = window.setTimeout(failStalledUpload, 20_000)
  const keepAlive = () => {
    window.clearTimeout(timeout)
    timeout = window.setTimeout(failStalledUpload, 20_000)
  }

  try {
    if (onProgress) onProgress(1)
    const blob = await Promise.race([
      uploadBlob(blobPath(prefix, file), file, {
        access: 'public',
        handleUploadUrl: '/api/blob-upload',
        headers: token ? { Authorization: `Bearer ${token}` } : undefined,
        multipart: file.size > 8 * 1024 * 1024,
        abortSignal: abortController.signal,
        onUploadProgress: (event) => {
          keepAlive()
          if (onProgress) onProgress(Math.max(1, Math.round(event.percentage)))
        },
      }),
      stallTimeout,
    ])
    return { url: blob.url, name: file.name, size: file.size }
  } catch (error: any) {
    if (error?.name === 'AbortError') {
      throw new Error('Upload service did not respond. Check /api/blob-upload and Vercel Blob environment variables.')
    }
    throw error
  } finally {
    window.clearTimeout(timeout)
  }
}

// Auth
export const authApi = {
  login: (email: string, password: string) =>
    api.post<{ token: string; role: string }>('/auth/login', { email, password }),
  register: (email: string, password: string, role: string) =>
    api.post<{ token: string; role: string }>('/auth/register', { email, password, role }),
}

// Users
export const usersApi = {
  me: () => api.get('/users/me'),
  changePassword: (data: { current_password: string; new_password: string }) =>
    api.patch('/users/me/password', data),
}

export const uploadsApi = {
  upload: async (file: File) => {
    if (shouldUseBlobUploads()) {
      return { data: await uploadToBlob('materials', file) }
    }
    const fd = new FormData()
    fd.append('file', file)
    return api.post<{ url: string; name: string; size: number }>('/uploads', fd, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
  },
}

// Sessions
export const sessionsApi = {
  list: (params?: { limit?: number }) =>
    api.get('/sessions', { params }),
  get: (id: number) => api.get(`/sessions/${id}`),
  create: (data: {
    teacher_id?: number
    title: string
    start_time: string
  }) => api.post('/sessions', data),
  delete: (id: number) => api.delete(`/sessions/${id}`),
}

// Courses
export interface LessonInput {
  title: string
  description: string
  video_url: string
  file_url: string
  sort_order: number
}

export interface LessonAsset {
  id: number
  lesson_id: number
  type: 'video' | 'file'
  url: string
  original_filename: string
  created_at: string
}

export interface CourseInput {
  title: string
  description: string
  image_url: string
  lessons?: LessonInput[]
}

export const coursesApi = {
  list: (params?: { limit?: number }) =>
    api.get('/courses', { params }),
  create: (data: CourseInput) =>
    api.post('/courses', data),
  delete: (id: number) =>
    api.delete(`/courses/${id}`),
  addLesson: (courseId: number, data: LessonInput) =>
    api.post(`/courses/${courseId}/lessons`, data),
  updateLesson: (courseId: number, lessonId: number, data: LessonInput) =>
    api.put(`/courses/${courseId}/lessons/${lessonId}`, data),
  deleteLesson: (courseId: number, lessonId: number) =>
    api.delete(`/courses/${courseId}/lessons/${lessonId}`),
  addLessonAsset: (lessonId: number, data: { type: 'video' | 'file'; url: string; original_filename: string }) =>
    api.post<LessonAsset>(`/lessons/${lessonId}/assets`, data),
  deleteLessonAsset: (lessonId: number, assetId: number) =>
    api.delete(`/lessons/${lessonId}/assets/${assetId}`),
}

// Homework
export const homeworkApi = {
  submit: (data: { session_id: number; content: string }) =>
    api.post('/homework/submit', data),
  list: (params?: { session_id?: number; student_id?: number; status?: string; limit?: number }) =>
    api.get('/homework', { params }),
  mine: (params?: { status?: string; limit?: number }) =>
    api.get('/homework/mine', { params }),
  updateStatus: (id: number, status: string) =>
    api.patch(`/homework/${id}/status`, { status }),
}

// Stats
export interface RoleCount {
  role: string
  total: number
  online: number
}

export interface StatsResult {
  roles: RoleCount[]
  total_users: number
  total_online: number
}

export const statsApi = {
  get: () => api.get<StatsResult>('/stats'),
}

// Videos
export interface VideoUpload {
  id: number
  lesson_id: number
  original_filename: string
  stored_path: string
  hls_path: string
  status: string
  error_message: string
  created_at: string
  finished_at: string
}

export const videoApi = {
  upload: (lessonId: number, file: File, onProgress?: (pct: number) => void) => {
    if (shouldUseBlobUploads()) {
      return uploadToBlob('videos', file, onProgress).then((blob) =>
        api.put<VideoUpload>(`/lessons/${lessonId}/video-url`, {
          url: blob.url,
          original_filename: blob.name,
        }),
      )
    }
    const fd = new FormData()
    fd.append('video', file)
    return api.post<VideoUpload>(`/lessons/${lessonId}/video`, fd, {
      headers: { 'Content-Type': 'multipart/form-data' },
      onUploadProgress: (e) => {
        if (onProgress && e.total) onProgress(Math.round((e.loaded * 100) / e.total))
      },
    })
  },
  status: (lessonId: number) => api.get<VideoUpload>(`/lessons/${lessonId}/video`),
  clear: (lessonId: number) => api.delete(`/lessons/${lessonId}/video`),
}

// Audit
export const auditApi = {
  list: (params?: { limit?: number }) =>
    api.get('/audit-logs', { params }),
}

// Notifications
export const notificationsApi = {
  list: (params?: { limit?: number }) =>
    api.get('/notifications', { params }),
}

// Admin: user management
export interface AdminUser {
  id: number
  email: string
  role: string
  created_at: string
}

export const adminUsersApi = {
  list: (params?: { limit?: number }) =>
    api.get<{ items: AdminUser[]; count: number }>('/admin/users', { params }),
  create: (data: { email: string; password: string; role: string }) =>
    api.post<AdminUser>('/admin/users', data),
  delete: (id: number) => api.delete(`/admin/users/${id}`),
}
