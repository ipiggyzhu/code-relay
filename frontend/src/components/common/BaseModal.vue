<template>
  <TransitionRoot as="template" :show="open">
    <Dialog as="div" class="mac-modal-backdrop" :open="open" @close="$emit('close')">
      <div class="mac-modal-overlay" aria-hidden="true"></div>
      <div class="mac-modal-wrapper">
        <TransitionChild
          as="template"
          enter="mac-modal-enter-active"
          enter-from="mac-modal-enter-from"
          enter-to="mac-modal-enter-to"
          leave="mac-modal-leave-active"
          leave-from="mac-modal-leave-from"
          leave-to="mac-modal-leave-to"
        >
          <DialogPanel :class="['mac-modal', variantClass]">
            <header class="mac-modal-header">
              <DialogTitle class="mac-modal-title">{{ title }}</DialogTitle>
              <button class="ghost-icon" aria-label="Close" @click="$emit('close')">✕</button>
            </header>
            <div class="mac-modal-body mac-modal-scrollable">
              <slot />
            </div>
            <slot name="footer" />
          </DialogPanel>
        </TransitionChild>
      </div>
    </Dialog>
  </TransitionRoot>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Dialog, DialogPanel, DialogTitle, TransitionChild, TransitionRoot } from '@headlessui/vue'

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
