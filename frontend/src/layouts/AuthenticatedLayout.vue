<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import {
  LayoutDashboard,
  Building2,
  CalendarDays,
  BookOpen,
  User,
  Bell,
  ShieldCheck,
  LogOut,
  Menu,
  X,
} from 'lucide-vue-next'

const router = useRouter()
const route = useRoute()
const auth = useAuthStore()
const sidebarOpen = ref(false)

const initials = computed(() => {
  const name = auth.displayName
  const parts = name.split(/[\s_.-]+/)
  if (parts.length === 1) return parts[0]!.charAt(0).toUpperCase()
  return (parts[0]!.charAt(0) + parts[1]!.charAt(0)).toUpperCase()
})

interface NavItem {
  label: string
  icon: any
  to: string
  roles?: string[]
}

const navItems: NavItem[] = [
  { label: 'Dashboard', icon: LayoutDashboard, to: '/dashboard' },
  { label: 'Branches', icon: Building2, to: '/branches' },
  { label: 'Sessions', icon: CalendarDays, to: '/sessions' },
  { label: 'Homework', icon: BookOpen, to: '/homework' },
  { label: 'Notifications', icon: Bell, to: '/notifications', roles: ['admin', 'manager'] },
  { label: 'Audit Logs', icon: ShieldCheck, to: '/audit', roles: ['admin'] },
  { label: 'Profile', icon: User, to: '/profile' },
]

const visibleNavItems = computed(() =>
  navItems.filter((item) => {
    if (!item.roles) return true
    return item.roles.includes(auth.userRole)
  }),
)

function handleLogout() {
  auth.logout()
  router.push('/')
}

function navigateTo(path: string) {
  router.push(path)
  sidebarOpen.value = false
}
</script>

<template>
  <div class="flex min-h-screen bg-[#121212]">
    <!-- Sidebar (desktop) -->
    <aside class="hidden lg:flex flex-col w-64 bg-[#1E1E1E] border-r border-white/5">
      <!-- Logo -->
      <div class="flex items-center gap-3 px-6 py-5 border-b border-white/5">
        <div class="w-9 h-9 rounded-xl bg-gradient-to-br from-cyan-400 to-blue-500 flex items-center justify-center">
          <span class="text-black font-extrabold text-sm">EP</span>
        </div>
        <span class="text-lg font-extrabold bg-gradient-to-r from-cyan-400 to-blue-500 bg-clip-text text-transparent">
          EDpulse
        </span>
      </div>

      <!-- Nav -->
      <nav class="flex-1 px-3 py-4 space-y-1 overflow-y-auto">
        <button
          v-for="item in visibleNavItems"
          :key="item.to"
          @click="navigateTo(item.to)"
          class="flex items-center gap-3 w-full px-4 py-3 rounded-xl text-sm font-semibold transition-all"
          :class="
            route.path === item.to || route.path.startsWith(item.to + '/')
              ? 'bg-cyan-400/10 text-cyan-400'
              : 'text-white/60 hover:bg-white/5 hover:text-white'
          "
        >
          <component :is="item.icon" class="w-5 h-5 shrink-0" />
          {{ item.label }}
        </button>
      </nav>

      <!-- User section -->
      <div class="px-3 py-4 border-t border-white/5">
        <div class="flex items-center gap-3 px-4 py-3">
          <div class="w-9 h-9 rounded-full bg-cyan-400/20 flex items-center justify-center shrink-0">
            <span class="text-cyan-400 text-sm font-extrabold">{{ initials }}</span>
          </div>
          <div class="flex-1 min-w-0">
            <p class="text-sm font-bold text-white truncate">{{ auth.displayName }}</p>
            <p class="text-xs text-white/40 truncate">{{ auth.userRole }}</p>
          </div>
        </div>
        <button
          @click="handleLogout"
          class="flex items-center gap-3 w-full px-4 py-3 rounded-xl text-sm font-semibold text-red-400 hover:bg-red-400/10 transition-all"
        >
          <LogOut class="w-5 h-5" />
          Sign Out
        </button>
      </div>
    </aside>

    <!-- Mobile header -->
    <div class="lg:hidden fixed top-0 left-0 right-0 z-50 bg-[#1E1E1E] border-b border-white/5">
      <div class="flex items-center justify-between px-4 py-3">
        <div class="flex items-center gap-3">
          <div class="w-8 h-8 rounded-lg bg-gradient-to-br from-cyan-400 to-blue-500 flex items-center justify-center">
            <span class="text-black font-extrabold text-xs">EP</span>
          </div>
          <span class="font-extrabold text-cyan-400">EDpulse</span>
        </div>
        <button @click="sidebarOpen = !sidebarOpen" class="text-white/60 hover:text-white p-2">
          <X v-if="sidebarOpen" class="w-6 h-6" />
          <Menu v-else class="w-6 h-6" />
        </button>
      </div>
    </div>

    <!-- Mobile sidebar overlay -->
    <Transition name="fade">
      <div
        v-if="sidebarOpen"
        class="lg:hidden fixed inset-0 z-40 bg-black/60"
        @click="sidebarOpen = false"
      />
    </Transition>

    <!-- Mobile sidebar -->
    <Transition name="slide">
      <aside
        v-if="sidebarOpen"
        class="lg:hidden fixed top-14 left-0 bottom-0 z-50 w-64 bg-[#1E1E1E] border-r border-white/5 flex flex-col"
      >
        <nav class="flex-1 px-3 py-4 space-y-1 overflow-y-auto">
          <button
            v-for="item in visibleNavItems"
            :key="item.to"
            @click="navigateTo(item.to)"
            class="flex items-center gap-3 w-full px-4 py-3 rounded-xl text-sm font-semibold transition-all"
            :class="
              route.path === item.to
                ? 'bg-cyan-400/10 text-cyan-400'
                : 'text-white/60 hover:bg-white/5 hover:text-white'
            "
          >
            <component :is="item.icon" class="w-5 h-5 shrink-0" />
            {{ item.label }}
          </button>
        </nav>
        <div class="px-3 py-4 border-t border-white/5">
          <button
            @click="handleLogout"
            class="flex items-center gap-3 w-full px-4 py-3 rounded-xl text-sm font-semibold text-red-400 hover:bg-red-400/10 transition-all"
          >
            <LogOut class="w-5 h-5" />
            Sign Out
          </button>
        </div>
      </aside>
    </Transition>

    <!-- Main content -->
    <main class="flex-1 lg:ml-0 mt-14 lg:mt-0 overflow-y-auto">
      <div class="p-4 md:p-6 lg:p-8">
        <RouterView />
      </div>
    </main>
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
.slide-enter-active,
.slide-leave-active {
  transition: transform 0.25s ease;
}
.slide-enter-from,
.slide-leave-to {
  transform: translateX(-100%);
}
</style>
