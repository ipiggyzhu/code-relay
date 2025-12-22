<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { setTheme, getCurrentTheme, ThemeMode } from '../../utils/ThemeManager'
import { useI18n } from 'vue-i18n'
import GlassDropdown from '../common/GlassDropdown.vue'

const { t } = useI18n()
const themevalue = ref<ThemeMode>('light')

const options = computed(() => [
  { value: 'light', label: t('components.themesetting.select.opt_light') },
  { value: 'dark', label: t('components.themesetting.select.opt_dark') },
  { value: 'systemdefault', label: t('components.themesetting.select.opt_system') },
])

const onThemeChange = (value: string) => {
  themevalue.value = value as ThemeMode
  setTheme(themevalue.value)
}

onMounted(() => {
  themevalue.value = getCurrentTheme()
})
</script>

<template>
  <GlassDropdown
    :model-value="themevalue"
    :options="options"
    @update:model-value="onThemeChange"
  />
</template>
