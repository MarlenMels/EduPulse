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

// Branches
export const branchesApi = {
  list: (params?: { q?: string; limit?: number }) =>
    api.get('/branches', { params }),
  get: (id: number) => api.get(`/branches/${id}`),
  create: (data: { name: string; lat: number; lng: number }) =>
    api.post('/branches', data),
}

// Sessions
export const sessionsApi = {
  list: (params?: { limit?: number }) =>
    api.get('/sessions', { params }),
  get: (id: number) => api.get(`/sessions/${id}`),
  create: (data: {
    branch_id: number
    teacher_id?: number
    title: string
    start_time: string
    lat: number
    lng: number
  }) => api.post('/sessions', data),
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
