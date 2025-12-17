<template>
  <TransitionRoot as="template" :show="open">
    <Dialog as="div" class="mac-modal-root" @close="$emit('close')">
      <!-- Backdrop -->
      <TransitionChild
        as="template"
        enter="mac-backdrop-enter-active"
        enter-from="mac-backdrop-enter-from"
        enter-to="mac-backdrop-enter-to"
        leave="mac-backdrop-leave-active"
        leave-from="mac-backdrop-leave-from"
        leave-to="mac-backdrop-leave-to"
      >
        <div class="mac-modal-backdrop" aria-hidden="true">
          <div class="mac-modal-overlay"></div>
        </div>
      </TransitionChild>

      <!-- Scrollable container -->
      <div class="mac-modal-wrapper">
        <!-- Panel Transition -->
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
