<script setup lang="ts">
import { useAuthStore } from './stores/auth'

const authStore = useAuthStore()
</script>

<template>
  <!-- Startup Loading Screen / Splash Animation -->
  <Transition name="splash-fade">
    <div 
      v-if="!authStore.isInitialized" 
      class="fixed inset-0 z-50 flex flex-col items-center justify-center bg-slate-950 overflow-hidden"
    >
      <!-- Background Ambient Glow -->
      <div class="absolute top-1/3 left-1/2 -translate-x-1/2 -translate-y-1/2 w-96 h-96 bg-brand-500/15 rounded-full blur-3xl animate-pulse pointer-events-none"></div>
      <div class="absolute bottom-1/4 right-1/3 w-80 h-80 bg-emerald-900/20 rounded-full blur-3xl pointer-events-none"></div>

      <!-- Center Logo & Branding -->
      <div class="relative z-10 flex flex-col items-center text-center px-6">
        <!-- Animated Brand Icon -->
        <div class="relative mb-6">
          <img 
            src="/logo.png" 
            alt="ACIS Logo" 
            class="w-20 h-20 rounded-3xl shadow-2xl shadow-emerald-950/50 object-cover animate-bounce-slow border border-slate-800"
          />
          <div class="absolute -inset-1 rounded-3xl bg-brand-500/20 blur-sm -z-10 animate-ping"></div>
        </div>

        <h1 class="text-3xl font-black text-white tracking-tight mb-1">
          ACIS
        </h1>
        <p class="text-xs font-bold text-slate-400 uppercase tracking-widest mb-8">
          Aplikasi Catatan Keuangan Istri / Suami
        </p>

        <!-- Progress Indicator -->
        <div class="w-48 h-1.5 bg-slate-800 rounded-full overflow-hidden relative border border-slate-700/50">
          <div class="h-full bg-gradient-to-r from-brand-600 to-emerald-500 rounded-full animate-indeterminate"></div>
        </div>
      </div>
    </div>
  </Transition>

  <RouterView v-slot="{ Component }">
    <Transition name="page-fade" mode="out-in">
      <component :is="Component" />
    </Transition>
  </RouterView>
</template>

<style>
/* Splash screen transition */
.splash-fade-enter-active,
.splash-fade-leave-active {
  transition: opacity 0.5s ease, transform 0.5s ease;
}

.splash-fade-leave-to {
  opacity: 0;
  transform: scale(1.02);
}

/* Page fade transition */
.page-fade-enter-active,
.page-fade-leave-active {
  transition: opacity 0.2s ease;
}

.page-fade-enter-from,
.page-fade-leave-to {
  opacity: 0;
}

@keyframes bounce-slow {
  0%, 100% {
    transform: translateY(0);
  }
  50% {
    transform: translateY(-6px);
  }
}

.animate-bounce-slow {
  animation: bounce-slow 2.4s ease-in-out infinite;
}

@keyframes indeterminate {
  0% {
    transform: translateX(-100%) scaleX(0.2);
  }
  50% {
    transform: translateX(0%) scaleX(0.6);
  }
  100% {
    transform: translateX(100%) scaleX(0.2);
  }
}

.animate-indeterminate {
  animation: indeterminate 1.4s ease-in-out infinite;
  transform-origin: left;
}
</style>

