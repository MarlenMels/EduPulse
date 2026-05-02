<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { adminUsersApi, type AdminUser } from '@/api/client'
import { useAuthStore } from '@/stores/auth'
import { UserPlus, Trash2, X, Loader2 } from 'lucide-vue-next'

const auth = useAuthStore()
const users = ref<AdminUser[]>([])
const loading = ref(true)
const error = ref('')

const showCreate = ref(false)
const newUser = ref({ email: '', password: '', role: 'teacher' })
const creating = ref(false)
const createError = ref('')

const roleOptions = [
  { value: 'admin', label: 'Admin' },
  { value: 'manager', label: 'Manager' },
  { value: 'teacher', label: 'Teacher' },
  { value: 'student', label: 'Student' },
  { value: 'parent', label: 'Parent' },
]

const roleStyle: Record<string, string> = {
  admin: 'bg-red-400/10 text-red-400',
  manager: 'bg-purple-400/10 text-purple-400',
  teacher: 'bg-blue-400/10 text-blue-400',
  student: 'bg-cyan-400/10 text-cyan-400',
  parent: 'bg-green-400/10 text-green-400',
}

async function fetchUsers() {
  loading.value = true
  error.value = ''
  try {
    const res = await adminUsersApi.list({ limit: 200 })
    users.value = res.data.items || []
  } catch (e: any) {
    error.value = e.response?.data?.error || 'Failed to load users'
  } finally {
    loading.value = false
  }
}

async function createUser() {
  createError.value = ''
  if (!newUser.value.email.trim() || !newUser.value.password.trim()) {
    createError.value = 'Email and password are required'
    return
  }
  if (newUser.value.password.length < 6) {
    createError.value = 'Password must be at least 6 characters'
    return
  }
  creating.value = true
  try {
    await adminUsersApi.create({
      email: newUser.value.email.trim(),
      password: newUser.value.password,
      role: newUser.value.role,
    })
    showCreate.value = false
    newUser.value = { email: '', password: '', role: 'teacher' }
    await fetchUsers()
  } catch (e: any) {
    createError.value = e.response?.data?.error || 'Failed to create user'
  } finally {
    creating.value = false
  }
}

async function deleteUser(u: AdminUser) {
  if (u.id === auth.userId) return
  if (!confirm(`Delete user ${u.email}?`)) return
  try {
    await adminUsersApi.delete(u.id)
    users.value = users.value.filter((x) => x.id !== u.id)
  } catch (e: any) {
    alert(e.response?.data?.error || 'Failed to delete')
  }
}

onMounted(fetchUsers)
</script>

<template>
  <div>
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-6">
      <h1 class="text-2xl font-extrabold text-cyan-400">Users</h1>
      <button
        @click="showCreate = true"
        class="flex items-center gap-2 px-4 py-2.5 bg-cyan-400 text-[#121212] font-semibold text-sm rounded-xl hover:bg-cyan-300 transition-colors"
      >
        <UserPlus class="w-4 h-4" />
        Add user
      </button>
    </div>

    <div v-if="error" class="mb-4 p-3 rounded-xl bg-red-500/10 border border-red-500/30 text-red-400 text-sm">
      {{ error }}
    </div>

    <div v-if="loading" class="flex items-center gap-2 text-white/60 text-sm">
      <Loader2 class="w-4 h-4 animate-spin" /> Loading…
    </div>

    <div v-else class="overflow-hidden bg-[#1E1E1E] border border-white/5 rounded-2xl">
      <table class="w-full text-sm">
        <thead class="bg-white/5 text-left text-xs uppercase tracking-wider text-white/40">
          <tr>
            <th class="px-5 py-3 font-semibold">Email</th>
            <th class="px-5 py-3 font-semibold">Role</th>
            <th class="px-5 py-3 font-semibold">Created</th>
            <th class="px-5 py-3"></th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="u in users"
            :key="u.id"
            class="border-t border-white/5 text-white/80"
          >
            <td class="px-5 py-3 font-medium">{{ u.email }}</td>
            <td class="px-5 py-3">
              <span
                class="inline-flex items-center px-2.5 py-1 rounded-md text-xs font-semibold"
                :class="roleStyle[u.role] || 'bg-white/10 text-white/60'"
              >{{ u.role }}</span>
            </td>
            <td class="px-5 py-3 text-white/40 text-xs">{{ new Date(u.created_at).toLocaleString() }}</td>
            <td class="px-5 py-3 text-right">
              <button
                v-if="u.id !== auth.userId"
                @click="deleteUser(u)"
                class="text-red-400/80 hover:text-red-400 p-1.5 rounded-lg hover:bg-red-400/10 transition-colors"
                title="Delete"
              >
                <Trash2 class="w-4 h-4" />
              </button>
              <span v-else class="text-xs text-white/30 italic">you</span>
            </td>
          </tr>
          <tr v-if="!users.length">
            <td colspan="4" class="px-5 py-8 text-center text-white/40 text-sm">No users yet.</td>
          </tr>
        </tbody>
      </table>
    </div>

    <Teleport to="body">
      <div
        v-if="showCreate"
        class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4"
        @click.self="showCreate = false"
      >
        <div class="w-full max-w-md bg-[#1E1E1E] border border-white/10 rounded-2xl p-6">
          <div class="flex items-center justify-between mb-5">
            <h2 class="text-lg font-extrabold text-white">Add user</h2>
            <button @click="showCreate = false" class="text-white/40 hover:text-white">
              <X class="w-5 h-5" />
            </button>
          </div>

          <form @submit.prevent="createUser" class="space-y-4">
            <div v-if="createError" class="p-3 rounded-xl bg-red-500/10 border border-red-500/30 text-red-400 text-sm">
              {{ createError }}
            </div>

            <div>
              <label class="block text-xs font-semibold text-white/60 mb-1.5">Email</label>
              <input
                v-model="newUser.email"
                type="email"
                placeholder="user@example.com"
                class="w-full px-4 py-3 bg-[#2D2D2D] rounded-xl text-white text-sm placeholder-white/30 border border-white/10 focus:border-cyan-400 focus:outline-none"
              />
            </div>

            <div>
              <label class="block text-xs font-semibold text-white/60 mb-1.5">Password</label>
              <input
                v-model="newUser.password"
                type="text"
                placeholder="Min 6 chars"
                class="w-full px-4 py-3 bg-[#2D2D2D] rounded-xl text-white text-sm placeholder-white/30 border border-white/10 focus:border-cyan-400 focus:outline-none"
              />
              <p class="mt-1.5 text-xs text-white/40">Show in plain so you can share it with the user.</p>
            </div>

            <div>
              <label class="block text-xs font-semibold text-white/60 mb-1.5">Role</label>
              <select
                v-model="newUser.role"
                class="w-full px-4 py-3 bg-[#2D2D2D] rounded-xl text-white text-sm border border-white/10 focus:border-cyan-400 focus:outline-none"
              >
                <option v-for="r in roleOptions" :key="r.value" :value="r.value" class="bg-[#2D2D2D]">
                  {{ r.label }}
                </option>
              </select>
            </div>

            <button
              type="submit"
              :disabled="creating"
              class="w-full py-3 rounded-xl bg-cyan-400 text-[#121212] font-bold text-sm hover:bg-cyan-300 disabled:opacity-50 transition-colors"
            >
              <span v-if="creating" class="inline-block w-4 h-4 border-2 border-[#121212]/30 border-t-[#121212] rounded-full animate-spin" />
              <span v-else>Create user</span>
            </button>
          </form>
        </div>
      </div>
    </Teleport>
  </div>
</template>
