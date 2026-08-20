import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const routes = [
  {
    path: '/masuk',
    name: 'Login',
    component: () => import('../views/auth/LoginView.vue')
  },
  {
    path: '/',
    component: () => import('../components/layout/AppLayout.vue'),
    meta: { requiresAuth: true },
    children: [
      {
        path: '',
        name: 'Dashboard',
        component: () => import('../views/dashboard/DashboardView.vue')
      },
      {
        path: 'transaksi',
        name: 'Transactions',
        component: () => import('../views/transactions/TransactionView.vue')
      },
      {
        path: 'keluarga',
        name: 'Family',
        component: () => import('../views/family/FamilyView.vue')
      },
      {
        path: 'proposal',
        name: 'Proposals',
        component: () => import('../views/proposals/ProposalView.vue')
      }
    ]
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach(async (to, from, next) => {
  const authStore = useAuthStore()
  if (!authStore.initialized) {
    await authStore.init()
  }

  if (to.meta.requiresAuth && !authStore.isAuthenticated) {
    next({ name: 'Login' })
  } else if (to.name === 'Login' && authStore.isAuthenticated) {
    next({ path: '/' })
  } else {
    next()
  }
})

export default router
