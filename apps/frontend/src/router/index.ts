import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { useFamilyStore } from '../stores/family'

declare module 'vue-router' {
  interface RouteMeta {
    requiresAuth?: boolean
    requiresGuest?: boolean
    requiresFamily?: boolean
    requiresAdmin?: boolean
    title?: string
  }
}

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('../views/LoginView.vue'),
    meta: { requiresGuest: true, title: 'Sign In' }
  },
  {
    path: '/family-setup',
    name: 'FamilySetup',
    component: () => import('../views/FamilySetupView.vue'),
    meta: { requiresAuth: true, title: 'Family Setup' }
  },
  {
    path: '/',
    component: () => import('../components/layout/DashboardLayout.vue'),
    meta: { requiresAuth: true, requiresFamily: true },
    children: [
      {
        path: '',
        name: 'Dashboard',
        component: () => import('../views/dashboard/DashboardHome.vue'),
        meta: { title: 'Dashboard' }
      },
      {
        path: 'wallets',
        name: 'Wallets',
        component: () => import('../views/dashboard/wallets/WalletsPage.vue'),
        meta: { title: 'Wallets' }
      },
      {
        path: 'transactions',
        name: 'Transactions',
        component: () => import('../views/dashboard/transactions/TransactionsPage.vue'),
        meta: { title: 'Transactions' }
      },
      {
        path: 'proposals',
        name: 'Proposals',
        component: () => import('../views/dashboard/proposals/ProposalsPage.vue'),
        meta: { title: 'Proposals' }
      },
      {
        path: 'family',
        name: 'FamilySettings',
        component: () => import('../views/dashboard/family/FamilySettingsPage.vue'),
        meta: { requiresAdmin: true, title: 'Family Settings' }
      }
    ]
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'NotFound',
    component: () => import('../views/NotFoundView.vue'),
    meta: { title: 'Page Not Found' }
  }
]

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
  scrollBehavior(_to, _from, savedPosition) {
    if (savedPosition) return savedPosition
    return { top: 0 }
  }
})

router.beforeEach(async (to, _from) => {
  const authStore = useAuthStore()
  const familyStore = useFamilyStore()

  document.title = to.meta.title ? `${to.meta.title} | ACIS` : 'ACIS'

  // Try silent refresh on initial application load if unauthenticated
  if (!authStore.isInitialized) {
    if (!authStore.isAuthenticated) {
      try {
        await authStore.refreshToken()
      } catch {
        // Not authenticated or expired refresh token
      }
    }
    authStore.isInitialized = true
  }

  // Check auth required
  if (to.meta.requiresAuth && !authStore.isAuthenticated) {
    return { name: 'Login', query: { redirect: to.fullPath } }
  }

  // Check guest required
  if (to.meta.requiresGuest && authStore.isAuthenticated) {
    return { name: 'Dashboard' }
  }

  // Check family presence required for dashboard routes
  if (to.meta.requiresFamily && authStore.isAuthenticated && !familyStore.family) {
    try {
      await familyStore.fetchMyFamily()
    } catch {
      return { name: 'FamilySetup' }
    }
    if (!familyStore.family) {
      return { name: 'FamilySetup' }
    }
  }

  // Check admin role required
  if (to.meta.requiresAdmin && authStore.user?.role !== 'admin') {
    return { name: 'Dashboard' }
  }
})

export default router