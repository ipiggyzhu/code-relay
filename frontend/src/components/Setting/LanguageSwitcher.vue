<script setup lang="ts">
import { ref, computed } from 'vue'
import { i18n, setupI18n } from '../../utils/i18n'
import { Locale } from '../../locales'
import GlassDropdown from '../common/GlassDropdown.vue'

const locale = ref(i18n.global.locale.value)

const options = computed(() => [
  { value: 'zh', label: '简体中文' },
  { value: 'en', label: 'English' },
])

const switchLang = async (value: string) => {
  locale.value = value
  await setupI18n(value as Locale)
}
</script>

<template>
  <GlassDropdown
    :model-value="locale"
    :options="options"
    @update:model-value="switchLang"
  />
</template>
