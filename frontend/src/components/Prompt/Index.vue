<template>
  <div class="main-shell">
    <PageHeader :title="t('components.prompt.title')">
      <template #actions>
        <button class="ghost-icon" :disabled="loading" @click="refresh">
          <svg viewBox="0 0 24 24" aria-hidden="true" :class="{ spin: loading }">
            <path d="M20.5 8a8.5 8.5 0 10-2.38 7.41" fill="none" stroke="currentColor" stroke-width="1.5"
              stroke-linecap="round" stroke-linejoin="round" />
            <path d="M20.5 4v4h-4" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"
              stroke-linejoin="round" />
          </svg>
        </button>
        <button class="ghost-icon" @click="openCreateModal">
          <svg viewBox="0 0 24 24" aria-hidden="true">
            <path d="M12 5v14M5 12h14" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"
              stroke-linejoin="round" fill="none" />
          </svg>
        </button>
      </template>
    </PageHeader>

    <div class="contrib-page prompt-page">
      <!-- Platform Tabs -->
      <div class="platform-tabs">
        <button v-for="platform in platforms" :key="platform.id"
          class="platform-tab" :class="{ active: currentPlatform === platform.id }"
          @click="switchPlatform(platform.id)">
          {{ platform.label }}
        </button>
      </div>

      <!-- Prompt List -->
      <section class="prompt-list-section">
        <div v-if="loading" class="prompt-empty">{{ t('components.prompt.loading') }}</div>
        <div v-else-if="prompts.length === 0" class="prompt-empty">{{ t('components.prompt.empty') }}</div>
        <div v-else class="prompt-list">
          <article v-for="prompt in prompts" :key="prompt.id" class="prompt-card"
            :class="{ active: prompt.is_active }">
            <div class="prompt-card-head">
              <div>
                <span v-if="prompt.is_active" class="prompt-badge">{{ t('components.prompt.active') }}</span>
                <h3>{{ prompt.name }}</h3>
              </div>
              <div class="prompt-card-actions">
                <button type="button" class="ghost-icon sm" :data-tooltip="t('components.prompt.edit')"
                  @click="openEditModal(prompt)">
                  <svg viewBox="0 0 24 24" aria-hidden="true">
                    <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7" fill="none"
                      stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" />
                    <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z" fill="none"
                      stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" />
                  </svg>
                </button>
                <button type="button" class="ghost-icon sm"
                  :data-tooltip="prompt.is_active ? t('components.prompt.deactivate') : t('components.prompt.activate')"
                  :disabled="activating === prompt.id"
                  @click="toggleActivation(prompt)">
                  <svg v-if="activating !== prompt.id" viewBox="0 0 24 24" aria-hidden="true">
                    <path v-if="prompt.is_active" d="M18 6L6 18M6 6l12 12" fill="none" stroke="currentColor"
                      stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" />
                    <path v-else d="M5 12l5 5L20 7" fill="none" stroke="currentColor"
                      stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" />
                  </svg>
                  <span v-else class="prompt-spinner" aria-hidden="true"></span>
                </button>
                <button type="button" class="ghost-icon sm danger" :data-tooltip="t('components.prompt.delete')"
                  :disabled="deleting === prompt.id"
                  @click="handleDelete(prompt)">
                  <svg v-if="deleting !== prompt.id" viewBox="0 0 24 24" aria-hidden="true">
                    <path d="M5 7h14M10 11v6M14 11v6M9 7V5h6v2" fill="none" stroke="currentColor" stroke-width="1.5"
                      stroke-linecap="round" stroke-linejoin="round" />
                    <path d="M6.5 7l-.5 12a2 2 0 002 2h8a2 2 0 002-2L17.5 7" fill="none" stroke="currentColor"
                      stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" />
                  </svg>
                  <span v-else class="prompt-spinner" aria-hidden="true"></span>
                </button>
              </div>
            </div>
            <p class="prompt-card-desc">
              {{ truncateContent(prompt.content) }}
            </p>
            <p class="prompt-card-time">
              {{ formatTime(prompt.updated_at) }}
            </p>
          </article>
        </div>
        <p v-if="error" class="prompt-error">{{ error }}</p>
      </section>
    </div>

    <!-- Create/Edit Modal -->
    <BaseModal :open="modalOpen" :title="editingPrompt ? t('components.prompt.edit') : t('components.prompt.create')"
      @close="closeModal">
      <div class="prompt-modal-content">
        <form class="prompt-form" @submit.prevent="handleSubmit">
          <div class="form-field">
            <label>{{ t('components.prompt.name') }}</label>
            <input v-model="formData.name" type="text" :placeholder="t('components.prompt.placeholder.name')"
              :disabled="submitting" required />
          </div>
          <div class="form-field">
            <label>{{ t('components.prompt.content') }}</label>
            <textarea v-model="formData.content" :placeholder="t('components.prompt.placeholder.content')"
              :disabled="submitting" rows="12" required></textarea>
            <p class="form-hint">{{ t('components.prompt.hint.markdown') }}</p>
          </div>
          <div class="form-actions">
            <button type="button" class="btn btn-outline" @click="closeModal" :disabled="submitting">
              {{ t('common.cancel') }}
            </button>
            <button type="submit" class="btn btn-primary" :disabled="submitting || !isFormValid">
              {{ submitting ? t('common.saving') : t('common.save') }}
            </button>
          </div>
        </form>
      </div>
    </BaseModal>
    <!-- Delete Confirmation Modal -->
    <BaseModal :open="confirmOpen" :title="t('components.prompt.confirmDelete')" variant="confirm" @close="closeConfirm">
      <div class="confirm-body">
        <p>{{ t('components.prompt.confirmDelete') }}</p>
        <p class="confirm-target" v-if="promptToDelete">"{{ promptToDelete.name }}"</p>
      </div>
      <footer class="form-actions confirm-actions">
        <button type="button" class="btn btn-outline" @click="closeConfirm" :disabled="submitting">
          {{ t('common.cancel') }}
        </button>
        <button type="button" class="btn btn-danger" @click="confirmDelete" :disabled="submitting">
          {{ t('components.prompt.delete') }}
        </button>
      </footer>
    </BaseModal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import PageHeader from '../Navigation/PageHeader.vue'
import BaseModal from '../common/BaseModal.vue'
import { useI18n } from 'vue-i18n'
import {
  fetchPrompts,
  createPrompt,
  updatePrompt,
  deletePrompt,
  activatePrompt,
  deactivatePrompt,
  type Prompt
} from '../../services/prompt'

const { t } = useI18n()

const platforms = [
  { id: 'claude', label: 'Claude Code' },
  { id: 'codex', label: 'Codex' },
  { id: 'gemini', label: 'Gemini CLI' }
]

const currentPlatform = ref('claude')
const prompts = ref<Prompt[]>([])
const loading = ref(false)
const error = ref('')
const activating = ref('')
const deleting = ref('')
const modalOpen = ref(false)
const editingPrompt = ref<Prompt | null>(null)
const submitting = ref(false)

// Delete confirmation state
const confirmOpen = ref(false)
const promptToDelete = ref<Prompt | null>(null)

const formData = reactive({
  name: '',
  content: ''
})

const isFormValid = computed(() => formData.name.trim() && formData.content.trim())

const loadPrompts = async () => {
  loading.value = true
  error.value = ''
  try {
    prompts.value = await fetchPrompts(currentPlatform.value)
  } catch (err) {
    console.error('Failed to load prompts', err)
    error.value = t('components.prompt.loadError')
  } finally {
    loading.value = false
  }
}

const refresh = () => {
  void loadPrompts()
}

const switchPlatform = (platform: string) => {
  currentPlatform.value = platform
  void loadPrompts()
}

const openCreateModal = () => {
  editingPrompt.value = null
  formData.name = ''
  formData.content = ''
  modalOpen.value = true
}

const openEditModal = (prompt: Prompt) => {
  editingPrompt.value = prompt
  formData.name = prompt.name
  formData.content = prompt.content
  modalOpen.value = true
}

const closeModal = () => {
  modalOpen.value = false
  editingPrompt.value = null
}

const handleSubmit = async () => {
  if (!isFormValid.value || submitting.value) return

  submitting.value = true
  error.value = ''
  try {
    if (editingPrompt.value) {
      await updatePrompt({
        ...editingPrompt.value,
        name: formData.name.trim(),
        content: formData.content.trim()
      })
    } else {
      await createPrompt({
        name: formData.name.trim(),
        content: formData.content.trim(),
        platform: currentPlatform.value
      })
    }
    closeModal()
    await loadPrompts()
  } catch (err) {
    console.error('Failed to save prompt', err)
    error.value = t('components.prompt.saveError')
  } finally {
    submitting.value = false
  }
}

const toggleActivation = async (prompt: Prompt) => {
  activating.value = prompt.id
  error.value = ''
  try {
    if (prompt.is_active) {
      await deactivatePrompt(prompt.id)
    } else {
      await activatePrompt(prompt.id)
    }
    await loadPrompts()
  } catch (err) {
    console.error('Failed to toggle activation', err)
    error.value = t('components.prompt.activateError')
  } finally {
    activating.value = ''
  }
}

const handleDelete = (prompt: Prompt) => {
  promptToDelete.value = prompt
  confirmOpen.value = true
}

const closeConfirm = () => {
  confirmOpen.value = false
  promptToDelete.value = null
}

const confirmDelete = async () => {
  if (!promptToDelete.value) return

  const id = promptToDelete.value.id
  deleting.value = id
  error.value = ''
  closeConfirm()

  try {
    await deletePrompt(id)
    await loadPrompts()
  } catch (err) {
    console.error('Failed to delete prompt', err)
    error.value = t('components.prompt.deleteError')
  } finally {
    deleting.value = ''
  }
}

const truncateContent = (content: string, maxLength: number = 120) => {
  if (content.length <= maxLength) return content
  return content.substring(0, maxLength) + '...'
}

const formatTime = (timeStr: string) => {
  if (!timeStr) return ''
  const date = new Date(timeStr)
  return date.toLocaleDateString() + ' ' + date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}

onMounted(() => {
  void loadPrompts()
})
</script>

<style scoped>
.prompt-page {
  display: flex;
  flex-direction: column;
  gap: 32px;
  color: var(--mac-text);
  padding: 24px 32px 60px;
  /* Removed max-width and margin:0 auto to prevent "shifting" or "cramming" on different scales */
  width: 100%;
  box-sizing: border-box;
}

.platform-tabs {
  display: flex;
  gap: 14px;
  margin-bottom: 4px;
  overflow-x: auto;
  padding: 4px 2px 12px;
  scrollbar-width: none;
}

.platform-tabs::-webkit-scrollbar {
  display: none;
}

.platform-tab {
  flex-shrink: 0;
  min-width: 130px;
  padding: 12px 24px;
  border: 1px solid var(--glass-border-subtle);
  border-radius: 14px;
  background: var(--glass-bg);
  backdrop-filter: var(--glass-blur);
  -webkit-backdrop-filter: var(--glass-blur);
  color: var(--mac-text-secondary);
  font-size: 0.95rem;
  font-weight: 600;
  cursor: pointer;
  white-space: nowrap;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all var(--spring-duration-fast) var(--spring-smooth);
  box-shadow: var(--glass-shadow);
}

.platform-tab:hover {
  background: var(--glass-bg-strong);
  transform: var(--hover-lift);
  color: var(--mac-text);
}

.platform-tab.active {
  background: var(--mac-accent);
  color: #fff;
  border-color: var(--mac-accent);
  box-shadow: 0 8px 20px color-mix(in srgb, var(--mac-accent) 30%, transparent);
}

.prompt-list-section {
  margin-top: 0;
  width: 100%;
}

.prompt-empty {
  margin-top: 60px;
  color: var(--mac-text-secondary);
  text-align: center;
  font-size: 1rem;
  opacity: 0.8;
}

.prompt-list {
  display: grid;
  /* Balanced min-width for better column distribution */
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 24px;
  width: 100%;
}

.prompt-card {
  background: var(--glass-bg-strong);
  backdrop-filter: blur(20px) saturate(160%);
  -webkit-backdrop-filter: blur(20px) saturate(160%);
  border: 1px solid var(--glass-border-subtle);
  border-radius: 22px;
  padding: 24px;
  display: flex;
  flex-direction: column;
  gap: 16px;
  transition: all var(--spring-duration) var(--spring-bounce);
  box-shadow: var(--glass-shadow);
  position: relative;
  /* Prevent single card from being too bulky, but allow enough width */
  max-width: 540px;
}

.prompt-card:hover {
  transform: var(--hover-lift) scale(1.01);
  box-shadow: var(--glass-shadow-lg);
  border-color: color-mix(in srgb, var(--mac-accent) 30%, transparent);
}

.prompt-card.active {
  border-color: var(--mac-accent);
  box-shadow: 0 0 0 1px color-mix(in srgb, var(--mac-accent) 20%, transparent) inset, var(--glass-shadow-lg);
}

.prompt-card-head {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 12px;
}

.prompt-badge {
  display: inline-block;
  padding: 4px 10px;
  font-size: 0.72rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  color: #fff;
  background: var(--mac-accent);
  border-radius: 8px;
  margin-bottom: 8px;
  box-shadow: 0 4px 10px color-mix(in srgb, var(--mac-accent) 40%, transparent);
}

.prompt-card h3 {
  font-size: 1.15rem;
  font-weight: 700;
  margin: 0;
  color: var(--mac-text);
  line-height: 1.2;
}

.prompt-card-desc {
  color: var(--mac-text-secondary);
  font-size: 0.9rem;
  line-height: 1.6;
  min-height: 48px;
  white-space: pre-wrap;
  word-break: break-word;
  display: -webkit-box;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.prompt-card-time {
  color: var(--mac-text-secondary);
  font-size: 0.78rem;
  margin-top: auto;
  opacity: 0.6;
  font-weight: 500;
}

.prompt-card-actions {
  display: flex;
  gap: 8px;
}

/* Position tooltips below the buttons since they are at the top of the card */
.prompt-card-actions .ghost-icon[data-tooltip]::after {
  bottom: auto;
  top: calc(100% + 6px);
}

.prompt-card-actions .ghost-icon[data-tooltip]:hover::after,
.prompt-card-actions .ghost-icon[data-tooltip]:focus-visible::after {
  transform: translateX(-50%) translateY(2px);
}

.prompt-card-actions .ghost-icon {
  width: 34px;
  height: 34px;
  border-radius: 10px;
  background: color-mix(in srgb, var(--mac-surface-strong) 40%, transparent);
}

.prompt-card-actions .ghost-icon:hover {
  background: var(--mac-surface-strong);
  color: var(--mac-accent);
}

.prompt-card-actions .ghost-icon.danger:hover {
  color: #ef4444;
  background: rgba(239, 68, 68, 0.1);
}

.prompt-spinner {
  width: 16px;
  height: 16px;
  border-radius: 50%;
  border: 2px solid currentColor;
  border-top-color: transparent;
  animation: prompt-spin 0.8s linear infinite;
  display: inline-block;
}

.prompt-error {
  color: #f87171;
  background: rgba(248, 113, 113, 0.1);
  padding: 12px 16px;
  border-radius: 12px;
  margin-top: 24px;
  font-size: 0.9rem;
  border: 1px solid rgba(248, 113, 113, 0.2);
}

.prompt-modal-content {
  min-width: min(640px, 85vw);
  padding: 4px 0;
}

.prompt-form {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.form-field {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.form-field label {
  font-size: 0.9rem;
  font-weight: 600;
  color: var(--mac-text);
  padding-left: 4px;
}

.form-field input,
.form-field textarea {
  width: 100%;
  padding: 12px 16px;
  border: 1px solid var(--mac-border);
  border-radius: 14px;
  background: var(--mac-surface-strong);
  color: var(--mac-text);
  font-size: 0.95rem;
  font-family: inherit;
  resize: vertical;
  transition: all 0.2s;
}

.form-field input:focus,
.form-field textarea:focus {
  outline: none;
  border-color: var(--mac-accent);
  background: var(--mac-surface);
  box-shadow: 0 0 0 4px color-mix(in srgb, var(--mac-accent) 15%, transparent);
}

.form-field textarea {
  min-height: 240px;
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 0.9rem;
  line-height: 1.6;
}

.form-hint {
  font-size: 0.8rem;
  color: var(--mac-text-secondary);
  opacity: 0.7;
  padding-left: 4px;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 12px;
  padding: 16px 0 8px;
  border-top: 1px solid var(--mac-divider);
}

.btn {
  padding: 10px 24px;
  border-radius: 12px;
  font-size: 0.95rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
}

.btn-outline {
  background: transparent;
  border: 1.5px solid var(--mac-border);
  color: var(--mac-text);
}

.btn-outline:hover {
  background: var(--mac-surface-strong);
  border-color: var(--mac-text-secondary);
}

.btn-primary {
  background: var(--mac-accent);
  border: 1.5px solid var(--mac-accent);
  color: white;
  box-shadow: 0 4px 12px color-mix(in srgb, var(--mac-accent) 30%, transparent);
}

.btn-primary:hover {
  transform: translateY(-1px);
  box-shadow: 0 6px 16px color-mix(in srgb, var(--mac-accent) 40%, transparent);
}

.btn-danger {
  background: #ef4444;
  border: 1.5px solid #ef4444;
  color: white;
  box-shadow: 0 4px 12px rgba(239, 68, 68, 0.25);
}

.btn-danger:hover {
  background: #dc2626;
  border-color: #dc2626;
  box-shadow: 0 6px 16px rgba(239, 68, 68, 0.35);
}

.btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
  transform: none !important;
  box-shadow: none !important;
}

.confirm-body {
  padding: 8px 0 16px;
}

.confirm-body p {
  margin: 0;
  font-size: 1.05rem;
  line-height: 1.5;
}

.confirm-target {
  margin-top: 12px !important;
  font-weight: 700;
  color: var(--mac-accent);
  font-size: 1.15rem !important;
}

.ghost-icon svg.spin {
  animation: prompt-spin 1s linear infinite;
}

@keyframes prompt-spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

html.dark .prompt-card {
  background: var(--glass-bg);
}

html.dark .platform-tab {
  background: var(--glass-bg);
}
</style>
