import { createRouter, createWebHistory } from 'vue-router'
import LandingView from '@/views/LandingView.vue'
import LoginView from '@/views/LoginView.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'landing',
      component: LandingView,
    },
    {
      path: '/login',
      name: 'login',
      component: LoginView,
    },
    {
      path: '/register',
      name: 'register',
      component: () => import('@/views/RegisterView.vue'),
    },
    {
      path: '/dashboard',
      component: () => import('@/layouts/AuthenticatedLayout.vue'),
      meta: { requiresAuth: true },
      children: [
        {
          path: '',
          name: 'dashboard',
          component: () => import('@/views/DashboardView.vue'),
        },
      ],
    },
    {
      path: '/branches',
      component: () => import('@/layouts/AuthenticatedLayout.vue'),
      meta: { requiresAuth: true },
      children: [
        {
          path: '',
          name: 'branches',
          component: () => import('@/views/BranchesView.vue'),
        },
      ],
    },
    {
      path: '/sessions',
      component: () => import('@/layouts/AuthenticatedLayout.vue'),
      meta: { requiresAuth: true },
      children: [
        {
          path: '',
          name: 'sessions',
          component: () => import('@/views/SessionsView.vue'),
        },
      ],
    },
    {
      path: '/homework',
      component: () => import('@/layouts/AuthenticatedLayout.vue'),
      meta: { requiresAuth: true },
      children: [
        {
          path: '',
          name: 'homework',
          component: () => import('@/views/HomeworkView.vue'),
        },
      ],
    },
    {
      path: '/notifications',
      component: () => import('@/layouts/AuthenticatedLayout.vue'),
      meta: { requiresAuth: true },
      children: [
        {
          path: '',
          name: 'notifications',
          component: () => import('@/views/NotificationsView.vue'),
        },
      ],
    },
    {
      path: '/audit',
      component: () => import('@/layouts/AuthenticatedLayout.vue'),
      meta: { requiresAuth: true },
      children: [
        {
          path: '',
          name: 'audit',
          component: () => import('@/views/AuditView.vue'),
        },
      ],
    },
    {
      path: '/profile',
      component: () => import('@/layouts/AuthenticatedLayout.vue'),
      meta: { requiresAuth: true },
      children: [
        {
          path: '',
          name: 'profile',
          component: () => import('@/views/ProfileView.vue'),
        },
      ],
    },
  ],
})

router.beforeEach((to) => {
  const token = localStorage.getItem('token')
  if (to.meta.requiresAuth && !token) {
    return { name: 'login' }
  }
  if ((to.name === 'login' || to.name === 'landing' || to.name === 'register') && token) {
    return { name: 'dashboard' }
  }
})

export default router
