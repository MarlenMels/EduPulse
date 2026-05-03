<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { enrollmentsApi, type CourseUser } from '@/api/client'
import { Loader2 } from 'lucide-vue-next'

const students = ref<CourseUser[]>([])
const loading = ref(true)
const error = ref('')

async function fetchStudents() {
  loading.value = true
  error.value = ''
  try {
    const res = await enrollmentsApi.myStudents()
    students.value = res.data.items || []
  } catch (e: any) {
    error.value = e.response?.data?.error || 'Failed to load'
  } finally {
    loading.value = false
  }
}

onMounted(fetchStudents)
</script>

<template>
  <div>
    <h1 class="text-2xl font-extrabold text-cyan-400 mb-6">My students</h1>
    <p v-if="error" class="mb-4 p-3 rounded-xl bg-red-500/10 border border-red-500/30 text-red-400 text-sm">{{ error }}</p>

    <div v-if="loading" class="flex items-center gap-2 text-white/60 text-sm">
      <Loader2 class="w-4 h-4 animate-spin" /> Loading…
    </div>

    <div v-else-if="!students.length" class="text-white/40 text-sm">
      You have no students yet. Ask an admin to enroll students into one of your courses.
    </div>

    <div v-else class="overflow-hidden bg-[#1E1E1E] border border-white/5 rounded-2xl">
      <table class="w-full text-sm">
        <thead class="bg-white/5 text-left text-xs uppercase tracking-wider text-white/40">
          <tr>
            <th class="px-5 py-3 font-semibold">Email</th>
            <th class="px-5 py-3 font-semibold">Joined</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="s in students" :key="s.id" class="border-t border-white/5 text-white/80">
            <td class="px-5 py-3 font-medium">{{ s.email }}</td>
            <td class="px-5 py-3 text-white/40 text-xs">{{ new Date(s.created_at).toLocaleDateString() }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
