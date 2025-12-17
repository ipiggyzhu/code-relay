<template>
  <div v-if="open" class="mac-modal-root">
    <!-- Backdrop -->
    <div class="mac-modal-backdrop" aria-hidden="true" @click="$emit('close')">
      <div class="mac-modal-overlay"></div>
    </div>

    <!-- Scrollable container -->
    <div class="mac-modal-wrapper">
      <div :class="['mac-modal', variantClass]" @click.stop>
        <header class="mac-modal-header">
          <h2 class="mac-modal-title">{{ title }}</h2>
          <button class="ghost-icon" aria-label="Close" @click="$emit('close')">✕</button>
        </header>
        <div class="mac-modal-body mac-modal-scrollable">
          <slot />
        </div>
        <slot name="footer" />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

type Variant = 'default' | 'confirm'

const props = withDefaults(
  defineProps<{
    open: boolean
    title: string
    variant?: Variant
  }>(),
  { variant: 'default' },
)

defineEmits<{ (e: 'close'): void }>()

const variantClass = computed(() => (props.variant === 'confirm' ? 'confirm-modal' : ''))
</script>
