<script setup lang="ts">
import { computed, reactive, ref, onMounted } from 'vue'
import PageHeader from '../Navigation/PageHeader.vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { Dialogs } from '@wailsio/runtime'
import ListItem from '../Setting/ListRow.vue'
import LanguageSwitcher from '../Setting/LanguageSwitcher.vue'
import ThemeSetting from '../Setting/ThemeSetting.vue'
import { fetchAppSettings, saveAppSettings, type AppSettings } from '../../services/appSettings'
import {
  fetchConfigImportStatus,
  fetchConfigImportStatusForFile,
  importFromCcSwitch,
  importFromCustomFile,
  type ConfigImportResult,
  type ConfigImportStatus,
} from '../../services/configImport'
import { showToast } from '../../utils/toast'
import BaseButton from '../common/BaseButton.vue'
import BaseModal from '../common/BaseModal.vue'
import {
  checkForUpdates as checkForUpdatesService,
  downloadUpdate,
  installUpdate,
  formatFileSize,
  type UpdateInfo,
} from '../../services/update'
import { fetchCurrentVersion } from '../../services/version'
import { Browser, Application } from '@wailsio/runtime'

const router = useRouter()
const { t } = useI18n()
const heatmapEnabled = ref(true)
const homeTitleVisible = ref(true)
const autoStartEnabled = ref(false)
const settingsLoading = ref(true)
const saveBusy = ref(false)
const importStatus = ref<ConfigImportStatus | null>(null)
const customImportStatus = ref<ConfigImportStatus | null>(null)
const importBusy = ref(false)
const appVersion = ref('')

// Update state
const updateModalState = reactive({
  open: false,
  checking: false,
  updateInfo: null as UpdateInfo | null,
  downloading: false,
  downloadedPath: '',
  installing: false,
  error: '',
})


const loadAppSettings = async () => {
  settingsLoading.value = true
  try {
    const data = await fetchAppSettings()
    heatmapEnabled.value = data?.show_heatmap ?? true
    homeTitleVisible.value = data?.show_home_title ?? true
    autoStartEnabled.value = data?.auto_start ?? false
  } catch (error) {
    console.error('failed to load app settings', error)
    heatmapEnabled.value = true
    homeTitleVisible.value = true
    autoStartEnabled.value = false
  } finally {
    settingsLoading.value = false
  }
}

const loadAppVersion = async () => {
  try {
    const version = await fetchCurrentVersion()
    appVersion.value = version || ''
  } catch (error) {
    console.error('failed to load app version', error)
  }
}

const openUpdateModal = async () => {
  updateModalState.open = true
  updateModalState.checking = true
  updateModalState.error = ''
  updateModalState.downloadedPath = ''
  updateModalState.downloading = false
  updateModalState.installing = false
  updateModalState.updateInfo = null

  try {
    const info = await checkForUpdatesService()
    updateModalState.updateInfo = info
  } catch (error: any) {
    console.error('Failed to check for updates:', error)
    updateModalState.error = error?.message || t('components.main.update.checkFailed')
  } finally {
    updateModalState.checking = false
  }
}

const closeUpdateModal = () => {
  updateModalState.open = false
}

const handleDownloadUpdate = async () => {
  if (!updateModalState.updateInfo?.downloadUrl) return
  updateModalState.downloading = true
  updateModalState.error = ''
  try {
    const path = await downloadUpdate(updateModalState.updateInfo.downloadUrl)
    if (path) {
      updateModalState.downloadedPath = path
    } else {
      updateModalState.error = t('components.main.update.downloadFailed')
    }
  } catch (error) {
    updateModalState.error = String(error)
  } finally {
    updateModalState.downloading = false
  }
}

const handleInstallUpdate = async () => {
  if (!updateModalState.downloadedPath) return
  updateModalState.installing = true
  updateModalState.error = ''
  try {
    const success = await installUpdate(updateModalState.downloadedPath)
    if (success) {
      setTimeout(() => {
        Application.Quit()
      }, 500)
    } else {
      updateModalState.error = t('components.main.update.installFailed')
    }
  } catch (error) {
    updateModalState.error = String(error)
  } finally {
    updateModalState.installing = false
  }
}

const openReleasePage = () => {
  if (updateModalState.updateInfo?.releaseUrl) {
    Browser.OpenURL(updateModalState.updateInfo.releaseUrl)
  }
}

const persistAppSettings = async () => {
  if (settingsLoading.value || saveBusy.value) return
  saveBusy.value = true
  try {
    const payload: AppSettings = {
      show_heatmap: heatmapEnabled.value,
      show_home_title: homeTitleVisible.value,
      auto_start: autoStartEnabled.value,
    }
    await saveAppSettings(payload)
    window.dispatchEvent(new CustomEvent('app-settings-updated'))
  } catch (error) {
    console.error('failed to save app settings', error)
  } finally {
    saveBusy.value = false
  }
}

onMounted(() => {
  void loadAppSettings()
  void loadAppVersion()
  void loadImportStatus()
})

const loadImportStatus = async () => {
  try {
    importStatus.value = await fetchConfigImportStatus()
  } catch (error) {
    console.error('failed to load cc-switch import status', error)
    importStatus.value = null
  }
}

const activeImportStatus = computed(() => customImportStatus.value ?? importStatus.value)
const hasCustomSelection = computed(() => Boolean(customImportStatus.value))
const shouldShowDefaultMissingHint = computed(() => {
  if (hasCustomSelection.value) return false
  const status = importStatus.value
  if (!status) return false
  return !status.config_exists
})
const pendingProviders = computed(() => activeImportStatus.value?.pending_provider_count ?? 0)
const pendingServers = computed(() => activeImportStatus.value?.pending_mcp_count ?? 0)
const configPath = computed(() => activeImportStatus.value?.config_path ?? '')
const canImportDefault = computed(() => {
  const status = importStatus.value
  if (!status) return false
  return Boolean(status.pending_providers || status.pending_mcp)
})
const canImportCustom = computed(() => {
  const status = customImportStatus.value
  if (!status) return false
  return Boolean(status.pending_providers || status.pending_mcp)
})
const canImportActive = computed(() =>
  hasCustomSelection.value ? canImportCustom.value : canImportDefault.value,
)
const showImportRow = computed(() => Boolean(importStatus.value) || hasCustomSelection.value)
const importPathLabel = computed(() => {
  if (!configPath.value) return ''
  return t('components.general.import.path', { path: configPath.value })
})
const importDetailLabel = computed(() => {
  if (shouldShowDefaultMissingHint.value) {
    return t('components.general.import.missingDefault')
  }
  if (!activeImportStatus.value) {
    return t('components.general.import.noFile')
  }
  const detail = canImportActive.value
    ? t('components.general.import.detail', {
        providers: pendingProviders.value,
        servers: pendingServers.value,
      })
    : t('components.general.import.synced')
  if (!importPathLabel.value) return detail
  return `${importPathLabel.value} · ${detail}`
})
const importButtonText = computed(() => {
  if (importBusy.value) {
    return t('components.general.import.importing')
  }
  if (hasCustomSelection.value) {
    return t('components.general.import.confirm')
  }
  if (shouldShowDefaultMissingHint.value || canImportDefault.value) {
    return t('components.general.import.cta')
  }
  return t('components.general.import.syncedButton')
})
const primaryButtonDisabled = computed(() => importBusy.value || !canImportActive.value)
const secondaryButtonLabel = computed(() =>
  hasCustomSelection.value
    ? t('components.general.import.clear')
    : t('components.general.import.upload'),
)
const secondaryButtonVariant = computed(() => 'outline' as const)

const processImportResult = async (result?: ConfigImportResult | null) => {
  if (!result) return
  if (hasCustomSelection.value && result.status?.config_path === customImportStatus.value?.config_path) {
    customImportStatus.value = result.status
  } else {
    importStatus.value = result.status
  }
  const importedProviders = result.imported_providers ?? 0
  const importedServers = result.imported_mcp ?? 0
  if (importedProviders > 0 || importedServers > 0) {
    showToast(
      t('components.main.importConfig.success', {
        providers: importedProviders,
        servers: importedServers,
      })
    )
  } else if (result.status?.config_exists) {
    showToast(t('components.main.importConfig.empty'))
  }
  await loadImportStatus()
}

const handleImportClick = async () => {
  if (importBusy.value || !importStatus.value || !canImportDefault.value) return
  importBusy.value = true
  try {
    const result = await importFromCcSwitch()
    await processImportResult(result)
  } catch (error) {
    console.error('failed to import cc-switch config', error)
    showToast(t('components.main.importConfig.error'), 'error')
  } finally {
    importBusy.value = false
  }
}

const handleConfirmCustomImport = async () => {
  const path = customImportStatus.value?.config_path
  if (!path || importBusy.value || !canImportCustom.value) return
  importBusy.value = true
  try {
    const result = await importFromCustomFile(path)
    await processImportResult(result)
  } catch (error) {
    console.error('failed to import custom cc-switch config', error)
    showToast(t('components.main.importConfig.error'), 'error')
  } finally {
    importBusy.value = false
  }
}

const handlePrimaryImport = async () => {
  if (hasCustomSelection.value) {
    await handleConfirmCustomImport()
  } else {
    await handleImportClick()
  }
}

const handleUploadClick = async () => {
  if (importBusy.value) return
  let selectedPath = ''
  try {
    const selection = await Dialogs.OpenFile({
      Title: t('components.general.import.uploadTitle'),
      CanChooseFiles: true,
      CanChooseDirectories: false,
      AllowsOtherFiletypes: false,
      Filters: [
        {
          DisplayName: 'JSON (*.json)',
          Pattern: '*.json',
        },
      ],
      AllowsMultipleSelection: false,
    })
    selectedPath = Array.isArray(selection) ? selection[0] : selection
    if (!selectedPath) return
    const status = await fetchConfigImportStatusForFile(selectedPath)
    customImportStatus.value = status
  } catch (error) {
    console.error('failed to load custom cc-switch config status', error)
    showToast(t('components.general.import.loadError'), 'error')
  }
}

const clearCustomSelection = () => {
  customImportStatus.value = null
}

const handleSecondaryImportAction = async () => {
  if (hasCustomSelection.value) {
    clearCustomSelection()
  } else {
    await handleUploadClick()
  }
}
</script>

<template>
  <div class="main-shell general-shell">
    <PageHeader :title="t('sidebar.settings')" />

    <div class="general-page">
      <section>
        <h2 class="mac-section-title">{{ $t('components.general.title.application') }}</h2>
        <div class="mac-panel">
          <ListItem :label="$t('components.general.label.heatmap')">
            <label class="mac-switch">
              <input
                type="checkbox"
                :disabled="settingsLoading || saveBusy"
                v-model="heatmapEnabled"
                @change="persistAppSettings"
              />
              <span></span>
            </label>
          </ListItem>
          <ListItem :label="$t('components.general.label.homeTitle')">
            <label class="mac-switch">
              <input
                type="checkbox"
                :disabled="settingsLoading || saveBusy"
                v-model="homeTitleVisible"
                @change="persistAppSettings"
              />
              <span></span>
            </label>
          </ListItem>
          <ListItem :label="$t('components.general.label.autoStart')">
            <label class="mac-switch">
              <input
                type="checkbox"
                :disabled="settingsLoading || saveBusy"
                v-model="autoStartEnabled"
                @change="persistAppSettings"
              />
              <span></span>
            </label>
          </ListItem>
          <ListItem
            v-if="showImportRow"
            :label="$t('components.general.import.label')"
            :sub-label="importDetailLabel"
          >
            <div class="import-actions">
              <BaseButton
                size="sm"
                variant="outline"
                type="button"
                :disabled="primaryButtonDisabled"
                @click="handlePrimaryImport"
              >
                {{ importButtonText }}
              </BaseButton>
              <BaseButton
                size="sm"
                :variant="secondaryButtonVariant"
                type="button"
                :disabled="importBusy"
                @click="handleSecondaryImportAction"
              >
                {{ secondaryButtonLabel }}
              </BaseButton>
              <BaseButton
                v-if="hasCustomSelection"
                size="sm"
                variant="outline"
                type="button"
                :disabled="importBusy"
                @click="handleUploadClick"
              >
                {{ $t('components.general.import.reupload') }}
              </BaseButton>
            </div>
          </ListItem>

        </div>
      </section>

      <section>
        <h2 class="mac-section-title">{{ $t('components.general.title.exterior') }}</h2>
        <div class="mac-panel">
          <ListItem :label="$t('components.general.label.language')">
            <LanguageSwitcher />
          </ListItem>
          <ListItem :label="$t('components.general.label.theme')">
            <ThemeSetting />
          </ListItem>
        </div>
      </section>

      <section>
        <h2 class="mac-section-title">{{ $t('components.general.title.about') }}</h2>
        <div class="mac-panel">
          <ListItem 
            :label="$t('components.general.label.version')" 
            :sub-label="appVersion ? (appVersion.startsWith('v') ? appVersion : `v${appVersion}`) : '...'"
          >
            <BaseButton
              size="sm"
              variant="outline"
              @click="openUpdateModal"
              :disabled="updateModalState.checking"
            >
              <template #icon>
                <svg viewBox="0 0 24 24" class="update-btn-icon" :class="{ spin: updateModalState.checking }">
                  <path d="M16.023 9.348h4.992v-.001M2.985 19.644v-4.992m0 0h4.992m-4.993 0l3.181 3.183a8.25 8.25 0 0013.803-3.7M4.031 9.865a8.25 8.25 0 0113.803-3.7l3.181 3.182m0-4.991v4.99" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" />
                </svg>
              </template>
              {{ updateModalState.checking ? $t('components.main.update.checking') : $t('components.main.update.checkUpdates') }}
            </BaseButton>
          </ListItem>
        </div>
      </section>
    </div>

    <!-- Update Modal -->
    <BaseModal
      :open="updateModalState.open"
      :title="t('components.main.update.available')"
      @close="closeUpdateModal"
    >
      <div class="update-modal-content">
        <div v-if="updateModalState.checking" class="update-checking">
          <div class="update-spinner"></div>
          <p>{{ t('components.main.update.checking') }}</p>
        </div>

        <template v-else-if="updateModalState.updateInfo">
          <div v-if="updateModalState.updateInfo.hasUpdate" class="update-info">
            <div class="update-versions">
              <div class="version-row">
                <span class="version-label">{{ t('components.main.update.currentVersion') }}:</span>
                <span class="version-value">{{ updateModalState.updateInfo.currentVersion }}</span>
              </div>
              <div class="version-row">
                <span class="version-label">{{ t('components.main.update.latestVersion') }}:</span>
                <span class="version-value version-new">{{ updateModalState.updateInfo.latestVersion }}</span>
              </div>
              <div v-if="updateModalState.updateInfo.fileSize" class="version-row">
                <span class="version-label">{{ t('components.main.update.fileSize') }}:</span>
                <span class="version-value">{{ formatFileSize(updateModalState.updateInfo.fileSize) }}</span>
              </div>
            </div>

            <p v-if="updateModalState.error" class="update-error">{{ updateModalState.error }}</p>

            <div class="update-actions">
              <template v-if="!updateModalState.downloadedPath">
                <BaseButton
                  v-if="updateModalState.updateInfo.downloadUrl"
                  :disabled="updateModalState.downloading"
                  @click="handleDownloadUpdate"
                >
                  {{ updateModalState.downloading ? t('components.main.update.downloading') : t('components.main.update.download') }}
                </BaseButton>
                <BaseButton variant="outline" @click="openReleasePage">
                  {{ t('components.main.update.viewRelease') }}
                </BaseButton>
              </template>
              <template v-else>
                <p class="update-hint">{{ t('components.main.update.installHint') }}</p>
                <BaseButton
                  :disabled="updateModalState.installing"
                  @click="handleInstallUpdate"
                >
                  {{ updateModalState.installing ? t('components.main.update.installing') : t('components.main.update.install') }}
                </BaseButton>
              </template>
            </div>
          </div>

          <div v-else class="update-no-update">
            <p>{{ t('components.main.update.noUpdate') }}</p>
            <p class="update-current">{{ t('components.main.update.currentVersion') }}: {{ updateModalState.updateInfo.currentVersion }}</p>
          </div>
        </template>

        <p v-if="updateModalState.error && !updateModalState.updateInfo" class="update-error">
          {{ updateModalState.error }}
        </p>
      </div>
    </BaseModal>
  </div>
</template>

<style scoped>
.import-actions {
  display: flex;
  gap: 0.5rem;
  justify-content: flex-end;
  flex-wrap: nowrap;
  max-width: 100%;
}

.import-actions .btn {
  min-width: fit-content;
  padding: 0.4rem 0.8rem;
  font-size: 0.75rem;
  line-height: 1.2;
  white-space: nowrap;
  flex-shrink: 0;
}

.import-actions .btn-outline,
.import-actions .btn-ghost {
  padding-inline: 0.75rem;
}

.update-btn-icon {
  width: 14px;
  height: 14px;
}

.update-modal-content {
  padding: 1.5rem;
}

.update-checking {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1rem;
}

.update-spinner {
  width: 2rem;
  height: 2rem;
  border: 3px solid rgba(0, 0, 0, 0.1);
  border-top-color: var(--mac-accent);
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

.update-versions {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  margin-bottom: 2rem;
}

.version-row {
  display: flex;
  justify-content: space-between;
}

.version-new {
  color: #34c759;
  font-weight: 600;
}

.update-actions {
  display: flex;
  justify-content: flex-end;
  gap: 1rem;
}

.update-error {
  color: #ff3b30;
  margin-top: 1rem;
  font-size: 0.9rem;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}
</style>
