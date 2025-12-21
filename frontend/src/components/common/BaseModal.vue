<template>
  <Teleport to="body">
    <Transition name="modal">
      <div v-if="open" class="mac-modal-root" @keydown.esc="$emit('close')">
        <!-- Backdrop -->
        <div class="mac-modal-backdrop" aria-hidden="true" @click="$emit('close')">
          <div class="mac-modal-overlay"></div>
        </div>

        <!-- Scrollable container -->
        <div class="mac-modal-wrapper">
          <div 
            :class="['mac-modal', variantClass]" 
            role="dialog" 
            aria-modal="true"
            @click.stop
          >
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
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { computed, watch } from 'vue'

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

// Lock body scroll when modal is open
watch(() => props.open, (isOpen) => {
  if (isOpen) {
    document.body.style.overflow = 'hidden'
  } else {
    document.body.style.overflow = ''
  }
})
</script>

<style scoped>
.mac-modal-root {
  position: fixed;
  inset: 0;
  z-index: 9999;
  display: flex;
  align-items: center;
  justify-content: center;
}

.modal-enter-active,
.modal-leave-active {
  transition: opacity 0.3s ease;
}

.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}

.modal-enter-active .mac-modal {
  transition: transform 0.4s cubic-bezier(0.34, 1.56, 0.64, 1), opacity 0.3s ease;
}

.modal-leave-active .mac-modal {
  transition: transform 0.2s ease-in, opacity 0.2s ease-in;
}

.modal-enter-from .mac-modal {
  opacity: 0;
  transform: scale(0.9) translateY(20px);
}

.modal-leave-to .mac-modal {
  opacity: 0;
  transform: scale(0.95) translateY(10px);
}
</style>

