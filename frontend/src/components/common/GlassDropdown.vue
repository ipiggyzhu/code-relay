<script setup lang="ts">
import { ref, watch, onMounted, onUnmounted, nextTick } from 'vue'
import {
  Listbox,
  ListboxButton,
  ListboxOptions,
  ListboxOption,
} from '@headlessui/vue'

interface Option {
  value: string
  label: string
}

const props = defineProps<{
  modelValue: string
  options: Option[]
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
}>()

const triggerRef = ref<any>(null)
const isOpenInternal = ref(false)

const menuStyle = ref({
  top: '0px',
  left: '0px',
  minWidth: '0px',
  transformOrigin: 'top center'
})

const updatePosition = async () => {
  if (!triggerRef.value?.$el) return
  
  const rect = triggerRef.value.$el.getBoundingClientRect()
  const viewportWidth = window.innerWidth
  const viewportHeight = window.innerHeight
  
  await nextTick()
  const menuEl = document.querySelector('.glass-dropdown-menu-inner') as HTMLElement
  const menuWidth = menuEl ? menuEl.offsetWidth : 180
  const menuHeight = menuEl ? menuEl.offsetHeight : props.options.length * 40 + 20
  
  let top = rect.bottom + 6
  let left = rect.left
  let origin = 'top center'
  
  if (left + menuWidth > viewportWidth - 16) {
    left = rect.right - menuWidth
    origin = 'top right'
  }
  
  if (left < 16) {
    left = 16
    origin = 'top left'
  }

  if (top + menuHeight > viewportHeight - 16) {
    top = rect.top - menuHeight - 6
    origin = 'bottom center'
  }

  menuStyle.value = {
    top: `${top}px`,
    left: `${left}px`,
    minWidth: `${rect.width}px`,
    transformOrigin: origin
  }
}

watch(isOpenInternal, (val) => {
  if (val) {
    nextTick(updatePosition)
  }
})

const onToggle = (openState: boolean) => {
  isOpenInternal.value = openState
}

onMounted(() => {
  window.addEventListener('resize', updatePosition)
  window.addEventListener('scroll', updatePosition, true)
})

onUnmounted(() => {
  window.removeEventListener('resize', updatePosition)
  window.removeEventListener('scroll', updatePosition, true)
})

const updateOpenState = (open: boolean) => {
  if (open !== isOpenInternal.value) {
    isOpenInternal.value = open
  }
  return ''
}

const handleChange = (val: string) => {
  emit('update:modelValue', val)
}
</script>

<template>
  <div class="glass-dropdown-container">
    <Listbox :model-value="modelValue" @update:model-value="handleChange" v-slot="{ open }">
      <div class="relative">
        <span class="hidden">{{ updateOpenState(open) }}</span>
        <ListboxButton
          ref="triggerRef"
          class="glass-dropdown-trigger"
          :class="{ 'is-open': open }"
        >
          <span class="truncate">{{ options.find(o => o.value === modelValue)?.label }}</span>
          <span class="pointer-events-none flex items-center">
            <svg class="h-4 w-4 transition-transform duration-300" :class="{ 'rotate-180': open }" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M19 9l-7 7-7-7" />
            </svg>
          </span>
        </ListboxButton>

        <Teleport to="body">
          <Transition
            enter-active-class="transition duration-300 ease-out"
            enter-from-class="transform scale-95 opacity-0 -translate-y-2"
            enter-to-class="transform scale-100 opacity-100 translate-y-0"
            leave-active-class="transition duration-150 ease-in"
            leave-from-class="transform scale-100 opacity-100 translate-y-0"
            leave-to-class="transform scale-95 opacity-0 -translate-y-2"
          >
            <ListboxOptions
              class="glass-dropdown-menu fixed"
              :style="menuStyle"
            >
              <div class="glass-dropdown-menu-inner">
                <ListboxOption
                  v-for="option in options"
                  :key="option.value"
                  :value="option.value"
                  v-slot="{ active, selected }"
                >
                  <li
                    class="glass-dropdown-item"
                    :class="{ 'is-active': active, 'is-selected': selected }"
                  >
                    <span>{{ option.label }}</span>
                    <span v-if="selected" class="selected-icon">
                      <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="3" d="M5 13l4 4L19 7" />
                      </svg>
                    </span>
                  </li>
                </ListboxOption>
              </div>
            </ListboxOptions>
          </Transition>
        </Teleport>
      </div>
    </Listbox>
  </div>
</template>

<style scoped>
.glass-dropdown-container {
  width: 100%;
  min-width: 140px;
}

.glass-dropdown-trigger {
  width: 100%;
  padding: 10px 14px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: var(--glass-bg);
  backdrop-filter: blur(12px) saturate(140%);
  -webkit-backdrop-filter: blur(12px) saturate(140%);
  border: 1px solid var(--glass-border-subtle);
  border-radius: 12px;
  color: var(--mac-text);
  font-size: 0.9rem;
  font-weight: 500;
  text-align: left;
  cursor: pointer;
  box-shadow: var(--glass-shadow);
  transition: all 0.3s cubic-bezier(0.25, 1, 0.5, 1);
}

.glass-dropdown-trigger:hover,
.glass-dropdown-trigger.is-open {
  border-color: var(--mac-accent);
  background: var(--glass-bg-strong);
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--mac-accent) 15%, transparent);
}

.glass-dropdown-trigger:active {
  transform: scale(0.98);
}

/* Global Styles for the Menu */
</style>

<style>
.glass-dropdown-menu {
  z-index: 100000;
  pointer-events: auto;
  outline: none;
}

.glass-dropdown-menu-inner {
  padding: 6px;
  background: var(--glass-bg-strong);
  backdrop-filter: blur(48px) saturate(210%);
  -webkit-backdrop-filter: blur(48px) saturate(210%);
  border: 1px solid var(--glass-border);
  border-radius: 16px;
  box-shadow: 0 12px 40px rgba(0, 0, 0, 0.4), inset 0 0 0 1px rgba(255, 255, 255, 0.1);
  min-width: 160px;
}

.glass-dropdown-item {
  padding: 8px 12px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-radius: 10px;
  color: var(--mac-text);
  font-size: 0.88rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
  margin-bottom: 2px;
  white-space: nowrap;
}

.glass-dropdown-item:last-child {
  margin-bottom: 0;
}

.glass-dropdown-item.is-active {
  background: var(--mac-accent);
  color: white;
}

.glass-dropdown-item.is-selected:not(.is-active) {
  color: var(--mac-accent);
  background: color-mix(in srgb, var(--mac-accent) 10%, transparent);
}

.selected-icon {
  flex-shrink: 0;
  color: inherit;
}
</style>
