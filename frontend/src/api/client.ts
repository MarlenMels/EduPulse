import axios from 'axios'

const api = axios.create({
  baseURL: '/api',
  headers: {
    'Content-Type': 'application/json',
    'Accept': 'application/json',
  },
})

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
}

// Courses
export interface LessonInput {
  title: string
  description: string
  video_url: string
  file_url: string
  sort_order: number
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
  addLesson: (courseId: number, data: LessonInput) =>
    api.post(`/courses/${courseId}/lessons`, data),
  updateLesson: (courseId: number, lessonId: number, data: LessonInput) =>
    api.put(`/courses/${courseId}/lessons/${lessonId}`, data),
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
