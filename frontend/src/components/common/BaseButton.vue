<template>
  <button :type="type" :class="['btn', variantClass]" v-bind="$attrs">
    <slot />
  </button>
</template>

<script setup lang="ts">
import { computed, useAttrs } from 'vue'

const props = withDefaults(
  defineProps<{
    variant?: 'primary' | 'outline' | 'danger' | 'ghost'
    size?: 'default' | 'sm' | 'lg'
    type?: 'button' | 'submit' | 'reset'
  }>(),
  {
    variant: 'primary',
    size: 'default',
    type: 'button',
  },
)

useAttrs()

const variantClass = computed(() => {
  const classes = [`btn-${props.variant}`]
  if (props.size === 'sm') classes.push('btn-sm')
  return classes
})
</script>
