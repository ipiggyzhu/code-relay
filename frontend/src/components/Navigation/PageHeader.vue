<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { collapsed, toggleSidebar } from '../../utils/sidebar'

defineProps<{
  title: string
}>()

const { t } = useI18n()
</script>

<template>
  <header class="header-actions">
    <div class="header-left">
      <button 
        class="sidebar-toggle-btn" 
        @click="toggleSidebar" 
        :title="collapsed ? t('components.main.sidebar.expand') : t('components.main.sidebar.collapse')"
      >
        <svg v-if="collapsed" viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <polyline points="9 18 15 12 9 6" />
        </svg>
        <svg v-else viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <polyline points="15 18 9 12 15 6" />
        </svg>
      </button>
    </div>
    
    <div class="header-center">
      <h1 class="page-title">{{ title }}</h1>
    </div>
    
    <div class="header-right">
      <slot name="actions"></slot>
    </div>
  </header>
</template>

<style scoped>
.sidebar-toggle-btn {
  width: 36px;
  height: 36px;
  border-radius: 10px;
  border: none;
  background: transparent;
  color: var(--mac-text-secondary);
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.2s;
}

.sidebar-toggle-btn:hover {
  background: rgba(0, 0, 0, 0.05);
  color: var(--mac-text);
}

html.dark .sidebar-toggle-btn:hover {
  background: rgba(255, 255, 255, 0.08);
}

.header-left, .header-right {
  flex: 1 1 0%;
  display: flex;
  align-items: center;
  min-width: 0;
}

.header-right {
  justify-content: flex-end;
}

.header-center {
  flex: 0 0 auto;
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 0 16px;
}

.page-title {
  margin: 0;
  white-space: nowrap;
  text-align: center;
}
</style>
