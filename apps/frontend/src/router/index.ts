import { createRouter, createWebHistory } from 'vue-router';
import { useAuthStore } from '../stores/auth';

// Extend Vue Router's RouteMeta interface (Mirip embedding struct di Go)
declare module 'vue-router' {
  interface RouteMeta {
    requiresAuth?: boolean;
    requiresGuest?: boolean;
  }
}

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('../views/LoginView.vue'),
    meta: { requiresGuest: true }
  },
  {
    path: '/family-setup',
    name: 'FamilySetup',
    component: () => import('../views/FamilySetupView.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/',
    name: 'Dashboard',
    component: () => import('../views/DashboardView.vue'),
    meta: { requiresAuth: true }
  }
];

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
});

// Navigation Guard with silent refresh support on initialization
router.beforeEach(async (to, _from, next) => {
  const authStore = useAuthStore();

  // Try silent refresh once on initial application load if unauthenticated
  if (!authStore.isInitialized) {
    authStore.isInitialized = true;
    if (!authStore.isAuthenticated) {
      try {
        await authStore.refreshToken();
      } catch {
        // Not authenticated or expired refresh token
      }
    }
  }

  if (to.meta.requiresAuth && !authStore.isAuthenticated) {
    next({ name: 'Login' });
  } else if (to.meta.requiresGuest && authStore.isAuthenticated) {
    next({ name: 'Dashboard' });
  } else {
    next();
  }
});

export default router;