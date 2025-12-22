<script setup lang="ts">
import { RouterView } from 'vue-router'
import { onMounted, ref } from 'vue'
import Sidebar from './components/Navigation/Sidebar.vue'
import { collapsed, sidebarWidth, saveSidebarState } from './utils/sidebar'

const isResizing = ref(false)
const MIN_SIDEBAR_WIDTH = 180
const MAX_SIDEBAR_WIDTH = 480
const COLLAPSED_WIDTH = 72
const COLLAPSE_THRESHOLD = 120 // Collapse when dragging left past this
const EXPAND_THRESHOLD = 150   // Expand when dragging right past this
let rafId: number | null = null

const startResizing = (e: MouseEvent) => {
  isResizing.value = true
  document.addEventListener('mousemove', handleMouseMove)
  document.addEventListener('mouseup', stopResizing)
  document.body.style.cursor = 'col-resize'
  document.body.style.userSelect = 'none'
  document.documentElement.classList.add('resizing')
}

const handleMouseMove = (e: MouseEvent) => {
  if (!isResizing.value) return
  if (rafId) return

  rafId = requestAnimationFrame(() => {
    const x = e.clientX
    
    if (collapsed.value) {
      // If currently collapsed, check if we should expand
      if (x > EXPAND_THRESHOLD) {
        collapsed.value = false
        sidebarWidth.value = Math.max(MIN_SIDEBAR_WIDTH, x)
      }
    } else {
      // If currently expanded, check if we should collapse
      if (x < COLLAPSE_THRESHOLD) {
        collapsed.value = true
      } else {
        // Fluid width follow while dragging
        sidebarWidth.value = Math.min(MAX_SIDEBAR_WIDTH, x)
      }
    }
    rafId = null
  })
}

const stopResizing = () => {
  isResizing.value = false
  if (rafId) {
    cancelAnimationFrame(rafId)
    rafId = null
  }
  document.removeEventListener('mousemove', handleMouseMove)
  document.removeEventListener('mouseup', stopResizing)
  document.body.style.cursor = ''
  document.body.style.userSelect = ''
  document.documentElement.classList.remove('resizing')
  
  // Final clamping on release
  if (!collapsed.value) {
    if (sidebarWidth.value < MIN_SIDEBAR_WIDTH) {
      sidebarWidth.value = MIN_SIDEBAR_WIDTH
    }
  }
  
  saveSidebarState()
}

const applyTheme = () => {
  const userTheme = localStorage.getItem('theme')
  const systemPrefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches

  const isDark = userTheme === 'dark' || (!userTheme && systemPrefersDark)

  document.documentElement.classList.toggle('dark', isDark)
}

onMounted(() => {
  applyTheme()

  // 可监听系统主题变化自动更新
  window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', () => {
    applyTheme()
  })
})
</script>

<template>
  <div class="app-layout" :class="{ 'is-resizing': isResizing }">
    <Sidebar :style="{ width: collapsed ? '72px' : `${sidebarWidth}px` }" />
    <div 
      class="sidebar-divider" 
      @mousedown="startResizing"
      :class="{ active: isResizing }"
    ></div>
    <main class="main-content">
      <router-view v-slot="{ Component }">
        <transition name="page" mode="out-in">
          <component :is="Component" />
        </transition>
      </router-view>
    </main>
  </div>
</template>
