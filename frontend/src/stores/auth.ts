import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { authApi, usersApi } from '@/api/client'

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(localStorage.getItem('token'))
  const user = ref<any>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  const isLoggedIn = computed(() => !!token.value)
  const displayName = computed(() => {
    if (!user.value) return 'User'
    return user.value.display_name || user.value.email?.split('@')[0] || 'User'
  })
  const userRole = computed(() => user.value?.role || '')
  const userEmail = computed(() => user.value?.email || '')
  const isAdmin = computed(() => userRole.value === 'admin')
  const isManager = computed(() => userRole.value === 'manager')
  const isTeacher = computed(() => userRole.value === 'teacher')
  const isStudent = computed(() => userRole.value === 'student')

  async function login(email: string, password: string) {
    loading.value = true
    error.value = null
    try {
      const res = await authApi.login(email, password)
      token.value = res.data.token
      localStorage.setItem('token', res.data.token)
      await fetchUser()
    } catch (e: any) {
      error.value = e.response?.data?.error || 'Login failed'
      throw e
    } finally {
      loading.value = false
    }
  }

  async function register(email: string, password: string, role: string) {
    loading.value = true
    error.value = null
    try {
      const res = await authApi.register(email, password, role)
      token.value = res.data.token
      localStorage.setItem('token', res.data.token)
      await fetchUser()
    } catch (e: any) {
      error.value = e.response?.data?.error || 'Registration failed'
      throw e
    } finally {
      loading.value = false
    }
  }

  async function fetchUser() {
    if (!token.value) return
    try {
      const res = await usersApi.me()
      user.value = res.data
    } catch {
      // token may be invalid
      logout()
    }
  }

  function logout() {
    token.value = null
    user.value = null
    localStorage.removeItem('token')
  }

  if (token.value) {
    fetchUser()
  }

  return {
    token,
    user,
    loading,
    error,
    isLoggedIn,
    displayName,
    userRole,
    userEmail,
    isAdmin,
    isManager,
    isTeacher,
    isStudent,
    login,
    register,
    fetchUser,
    logout,
  }
})
