<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { branchesApi } from '@/api/client'
import { useAuthStore } from '@/stores/auth'
import { Building2, Plus, Search, X } from 'lucide-vue-next'

const auth = useAuthStore()
const branches = ref<any[]>([])
const loading = ref(true)
const error = ref('')
const searchQuery = ref('')
const showCreateModal = ref(false)

const newBranch = ref({ name: '', lat: 0, lng: 0 })
const creating = ref(false)

async function fetchBranches() {
  loading.value = true
  error.value = ''
  try {
    const res = await branchesApi.list({ q: searchQuery.value || undefined, limit: 100 })
    branches.value = res.data.items || []
  } catch (e: any) {
    error.value = e.response?.data?.error || 'Failed to load'
  } finally {
    loading.value = false
  }
}

async function createBranch() {
  if (!newBranch.value.name.trim()) return
  creating.value = true
  try {
    await branchesApi.create(newBranch.value)
    showCreateModal.value = false
    newBranch.value = { name: '', lat: 0, lng: 0 }
    await fetchBranches()
  } catch (e: any) {
    error.value = e.response?.data?.error || 'Failed to create'
  } finally {
    creating.value = false
  }
}

onMounted(fetchBranches)
</script>

<template>
  <div>
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-6">
      <h1 class="text-2xl font-extrabold text-cyan-400">Branches</h1>
      <button
        v-if="auth.isAdmin || auth.isManager"
        @click="showCreateModal = true"
        class="flex items-center gap-2 px-4 py-2.5 bg-cyan-400 text-black font-semibold text-sm rounded-xl hover:bg-cyan-300 transition-colors"
      >
        <Plus class="w-4 h-4" />
        New Branch
      </button>
    </div>

    <!-- Search -->
    <div class="relative mb-6">
      <Search class="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-cyan-400" />
      <input
        v-model="searchQuery"
        @input="fetchBranches"
        placeholder="Search branches..."
        class="w-full pl-12 pr-4 py-3.5 bg-[#2D2D2D] rounded-xl text-white text-sm placeholder-white/40 border border-transparent focus:border-cyan-400 focus:outline-none transition-colors"
      />
    </div>

    <!-- Loading -->
    <div v-if="loading" class="flex justify-center py-16">
      <div class="w-8 h-8 border-2 border-cyan-400/30 border-t-cyan-400 rounded-full animate-spin" />
    </div>

    <!-- Error -->
    <div v-else-if="error" class="text-center py-16">
      <p class="text-red-400 mb-4">{{ error }}</p>
      <button @click="fetchBranches" class="px-4 py-2 bg-cyan-400 text-black rounded-lg font-semibold text-sm">
        Retry
      </button>
    </div>

    <!-- Empty -->
    <div v-else-if="branches.length === 0" class="text-center py-16 text-white/40">
      <Building2 class="w-16 h-16 mx-auto mb-4 opacity-30" />
      <p>No branches yet</p>
    </div>

    <!-- List -->
    <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      <div
        v-for="branch in branches"
        :key="branch.id"
        class="bg-[#1E1E1E] rounded-2xl p-5 border border-white/5 hover:border-cyan-400/30 transition-all cursor-pointer"
        @click="$router.push(`/branches/${branch.id}`)"
      >
        <div class="flex items-start gap-4">
          <div class="w-12 h-12 rounded-xl bg-cyan-400/10 flex items-center justify-center shrink-0">
            <Building2 class="w-6 h-6 text-cyan-400" />
          </div>
          <div class="flex-1 min-w-0">
            <h3 class="font-bold text-white truncate">{{ branch.name }}</h3>
            <p class="text-xs text-white/30 mt-1">ID: {{ branch.id }}</p>
          </div>
        </div>
      </div>
    </div>

    <!-- Create modal -->
    <Teleport to="body">
      <Transition name="fade">
        <div v-if="showCreateModal" class="fixed inset-0 z-50 flex items-end sm:items-center justify-center bg-black/60 p-4" @click.self="showCreateModal = false">
          <div class="bg-[#1E1E1E] rounded-t-3xl sm:rounded-2xl w-full max-w-md p-6">
            <div class="flex items-center justify-between mb-6">
              <h2 class="text-lg font-extrabold text-white">New Branch</h2>
              <button @click="showCreateModal = false" class="text-white/40 hover:text-white">
                <X class="w-5 h-5" />
              </button>
            </div>

            <form @submit.prevent="createBranch" class="space-y-4">
              <div>
                <label class="block text-sm font-semibold text-white/70 mb-1.5">Name</label>
                <input
                  v-model="newBranch.name"
                  placeholder="Branch name"
                  class="w-full px-4 py-3 bg-[#2D2D2D] rounded-xl text-white text-sm placeholder-white/30 border border-transparent focus:border-cyan-400 focus:outline-none"
                />
              </div>
              <div class="grid grid-cols-2 gap-3">
                <div>
                  <label class="block text-sm font-semibold text-white/70 mb-1.5">Latitude</label>
                  <input
                    v-model.number="newBranch.lat"
                    type="number"
                    step="any"
                    class="w-full px-4 py-3 bg-[#2D2D2D] rounded-xl text-white text-sm border border-transparent focus:border-cyan-400 focus:outline-none"
                  />
                </div>
                <div>
                  <label class="block text-sm font-semibold text-white/70 mb-1.5">Longitude</label>
                  <input
                    v-model.number="newBranch.lng"
                    type="number"
                    step="any"
                    class="w-full px-4 py-3 bg-[#2D2D2D] rounded-xl text-white text-sm border border-transparent focus:border-cyan-400 focus:outline-none"
                  />
                </div>
              </div>
              <div class="flex gap-3 pt-2">
                <button
                  type="button"
                  @click="showCreateModal = false"
                  class="flex-1 py-3 rounded-xl text-white/60 font-semibold text-sm hover:bg-white/5 transition-colors"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  :disabled="creating"
                  class="flex-1 py-3 rounded-xl bg-cyan-400 text-black font-semibold text-sm hover:bg-cyan-300 disabled:opacity-50 transition-colors"
                >
                  {{ creating ? 'Creating...' : 'Create' }}
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
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
