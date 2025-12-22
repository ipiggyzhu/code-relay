<script setup lang="ts">
defineProps<{
   label: string
   subLabel?: string
 }>()
</script>

<template>
  <div class="mac-list-row">
    <div class="mac-list-text">
      <div class="mac-list-label">{{ label }}</div>
      <div v-if="subLabel" class="mac-list-sublabel">{{ subLabel }}</div>
    </div>
    <div class="mac-list-control">
      <slot />
    </div>
  </div>
</template>

<style scoped>
.mac-list-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 22px;
  gap: 16px;
  min-height: 58px;
  box-sizing: border-box;
  max-width: 100%;
  overflow: hidden;
}

.mac-list-text {
  text-align: left;
  flex: 1 1 auto;
  min-width: 60px;
  overflow: hidden;
  padding-right: 8px;
}

.mac-list-control {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 8px;
  flex-shrink: 0;
  flex-wrap: nowrap;
  max-width: none;
}

.mac-list-control :deep(.btn) {
  min-width: fit-content;
  white-space: nowrap;
  flex-shrink: 0;
}
</style>
