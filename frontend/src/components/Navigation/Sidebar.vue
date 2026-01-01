<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute } from 'vue-router'
import { fetchCurrentVersion } from '../../services/version'
import lobeIcons from '../../icons/lobeIconMap'
import { collapsed } from '../../utils/sidebar'

const { t } = useI18n()
const route = useRoute()

const appVersion = ref('')
const mcpIcon = lobeIcons['mcp'] ?? ''

const navItems = computed(() => [
  {
    id: 'home',
    path: '/',
    label: t('sidebar.overview'),
    icon: 'home'
  },
  {
    id: 'mcp',
    path: '/mcp',
    label: t('sidebar.mcp'),
    icon: 'mcp'
  },
  {
    id: 'skill',
    path: '/skill',
    label: t('sidebar.skill'),
    icon: 'skill'
  },
  {
    id: 'prompt',
    path: '/prompt',
    label: t('sidebar.prompt'),
    icon: 'prompt'
  },
  {
    id: 'logs',
    path: '/logs',
    label: t('sidebar.logs'),
    icon: 'logs'
  },
  {
    id: 'settings',
    path: '/settings',
    label: t('sidebar.settings'),
    icon: 'settings'
  }
])

const isActive = (path: string) => {
  if (path === '/') {
    return route.path === '/'
  }
  return route.path.startsWith(path)
}

onMounted(async () => {
  try {
    appVersion.value = await fetchCurrentVersion()
  } catch (error) {
    console.error('Failed to fetch version', error)
  }
})
</script>

<template>
  <aside class="mac-sidebar" :class="{ collapsed }">
    <div class="mac-sidebar-header">
      <div class="logo-container">
        <div class="app-logo">
          <img src="/logo.png" alt="Code Relay" class="logo-img" />
        </div>
        <span class="app-name">Code Relay</span>
      </div>
    </div>

    <nav class="mac-sidebar-nav">
      <router-link
        v-for="item in navItems"
        :key="item.id"
        :to="item.path"
        class="mac-sidebar-item"
        :class="{ active: isActive(item.path) }"
        :title="collapsed ? item.label : ''"
      >
        <div class="nav-icon">
          <template v-if="item.icon === 'home'">
            <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
              <path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z" />
              <polyline points="9 22 9 12 15 12 15 22" />
            </svg>
          </template>
          <template v-else-if="item.icon === 'mcp'">
            <span class="mcp-icon-svg" v-html="mcpIcon"></span>
          </template>
          <template v-else-if="item.icon === 'skill'">
            <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
              <path d="M6 4h8a4 4 0 014 4v12a3 3 0 00-3-3H6z" />
              <path d="M6 4a2 2 0 00-2 2v13c0 .55.45 1 1 1h11" />
              <path d="M9 8h5" />
            </svg>
          </template>
          <template v-else-if="item.icon === 'prompt'">
            <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
              <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z" />
              <path d="M8 9h8M8 13h4" />
            </svg>
          </template>
          <template v-else-if="item.icon === 'logs'">
            <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
              <path d="M5 7h14M5 12h14M5 17h9" />
            </svg>
          </template>
          <template v-else-if="item.icon === 'settings'">
            <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
              <path d="M12 15a3 3 0 100-6 3 3 0 000 6z" />
              <path d="M19.4 15a1.65 1.65 0 00.33 1.82l.06.06a2 2 0 01-2.83 2.83l-.06-.06a1.65 1.65 0 00-1.82-.33 1.65 1.65 0 00-1 1.51V21a2 2 0 01-4 0v-.09a1.65 1.65 0 00-1-1.51 1.65 1.65 0 00-1.82.33l-.06.06a2 2 0 01-2.83-2.83l.06-.06a1.65 1.65 0 00.33-1.82 1.65 1.65 0 00-1.51-1H3a2 2 0 010-4h.09a1.65 1.65 0 001.51-1 1.65 1.65 0 00-.33-1.82l-.06-.06a2 2 0 012.83-2.83l.06.06a1.65 1.65 0 001.82.33H9a1.65 1.65 0 001-1.51V3a2 2 0 014 0v.09a1.65 1.65 0 001 1.51 1.65 1.65 0 001.82-.33l.06-.06a2 2 0 012.83 2.83l-.06.06a1.65 1.65 0 00-.33 1.82V9a1.65 1.65 0 001.51 1H21a2 2 0 010 4h-.09a1.65 1.65 0 00-1.51 1z" />
            </svg>
          </template>
        </div>
        <span class="nav-label">{{ item.label }}</span>
      </router-link>
    </nav>

    <div class="mac-sidebar-footer">
      <div class="version-info" v-if="appVersion">
        {{ appVersion.startsWith('v') ? appVersion : `v${appVersion}` }}
      </div>
    </div>
  </aside>
</template>

<style scoped>
.mcp-icon-svg {
  display: flex;
  align-items: center;
  justify-content: center;
}
:deep(.mcp-icon-svg svg) {
  width: 20px;
  height: 20px;
}

.logo-img {
  width: 100%;
  height: 100%;
  object-fit: contain;
  border-radius: 6px;
}
</style>
