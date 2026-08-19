<script setup lang="ts">
import { RouterView } from 'vue-router'
import Sidebar from './Sidebar.vue'
import TopNavbar from './TopNavbar.vue'
import ToastContainer from '../ui/ToastContainer.vue'
import { useUI } from '../../composables/useUI'

const { isGlobalLoading } = useUI()
</script>

<template>
  <div class="min-h-screen bg-slate-950 text-slate-100 flex font-sans antialiased">
    <!-- Sidebar navigation -->
    <Sidebar />

    <!-- Main Content Area -->
    <div class="flex-1 flex flex-col min-w-0 min-h-screen">
      <TopNavbar />
      <main class="flex-1 max-w-[1400px] w-full mx-auto p-4 sm:p-8 lg:p-10">
        <RouterView v-slot="{ Component }">
          <Transition name="page-fade" mode="out-in">
            <component :is="Component" />
          </Transition>
        </RouterView>
      </main>
    </div>

    <!-- Global Toast Container -->
    <ToastContainer />

    <!-- Global Loading Overlay -->
    <Transition name="fade">
      <div
        v-if="isGlobalLoading"
        class="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/60 backdrop-blur-xs"
      >
        <div class="p-6 rounded-3xl bg-slate-900 border border-slate-800 shadow-2xl flex flex-col items-center gap-3">
          <div class="w-8 h-8 border-3 border-teal-500/20 border-t-teal-500 rounded-full animate-spin"></div>
          <span class="text-xs font-bold text-slate-300">Loading...</span>
        </div>
      </div>
    </Transition>
  </div>
</template>

<style scoped>
.page-fade-enter-active,
.page-fade-leave-active {
  transition: opacity 0.2s ease, transform 0.2s ease;
}

.page-fade-enter-from {
  opacity: 0;
  transform: translateY(6px);
}

.page-fade-leave-to {
  opacity: 0;
  transform: translateY(-6px);
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.15s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
