<template>
  <div class="main-shell">
    <div class="global-actions">
      <p class="global-eyebrow">{{ t('components.main.hero.eyebrow') }}</p>
      <button
        class="ghost-icon github-icon"
        :class="{ 'github-upgrade': hasUpdateAvailable }"
        :data-tooltip="hasUpdateAvailable ? t('components.main.controls.githubUpdate') : t('components.main.controls.github')"
        @click="hasUpdateAvailable ? openUpdateModal() : openGitHub()"
      >
        <svg viewBox="0 0 24 24" aria-hidden="true">
          <path
            d="M9 19c-4.5 1.5-4.5-2.5-6-3m12 5v-3.87a3.37 3.37 0 00-.94-2.61c3.14-.35 6.44-1.54 6.44-7A5.44 5.44 0 0018 3.77 5.07 5.07 0 0017.91 1S16.73.65 14 2.48a13.38 13.38 0 00-5 0C6.27.65 5.09 1 5.09 1A5.07 5.07 0 005 3.77a5.44 5.44 0 00-1.5 3.76c0 5.42 3.3 6.61 6.44 7A3.37 3.37 0 009 18.13V22"
            fill="none"
            stroke="currentColor"
            stroke-width="1.5"
            stroke-linecap="round"
            stroke-linejoin="round"
          />
        </svg>
      </button>
      <button
        class="ghost-icon"
        :data-tooltip="t('components.main.controls.theme')"
        @click="toggleTheme"
      >
        <svg v-if="themeIcon === 'sun'" viewBox="0 0 24 24" aria-hidden="true">
          <circle cx="12" cy="12" r="4" stroke="currentColor" stroke-width="1.5" fill="none" />
          <path
            d="M12 3v2m0 14v2m9-9h-2M5 12H3m14.95 6.95-1.41-1.41M7.46 7.46 6.05 6.05m12.9 0-1.41 1.41M7.46 16.54l-1.41 1.41"
            stroke="currentColor"
            stroke-width="1.5"
            stroke-linecap="round"
          />
        </svg>
        <svg v-else viewBox="0 0 24 24" aria-hidden="true">
          <path
            d="M21 12.79A9 9 0 1111.21 3a7 7 0 109.79 9.79z"
            fill="none"
            stroke="currentColor"
            stroke-width="1.5"
            stroke-linecap="round"
            stroke-linejoin="round"
          />
        </svg>
      </button>
      <button
        class="ghost-icon"
        :data-tooltip="t('components.main.controls.settings')"
        @click="goToSettings"
      >
        <svg viewBox="0 0 24 24" aria-hidden="true">
          <path
            d="M12 15a3 3 0 100-6 3 3 0 000 6z"
            stroke="currentColor"
            stroke-width="1.5"
            stroke-linecap="round"
            stroke-linejoin="round"
            fill="none"
          />
          <path
            d="M19.4 15a1.65 1.65 0 00.33 1.82l.06.06a2 2 0 01-2.83 2.83l-.06-.06a1.65 1.65 0 00-1.82-.33 1.65 1.65 0 00-1 1.51V21a2 2 0 01-4 0v-.09a1.65 1.65 0 00-1-1.51 1.65 1.65 0 00-1.82.33l-.06.06a2 2 0 01-2.83-2.83l.06-.06a1.65 1.65 0 00.33-1.82 1.65 1.65 0 00-1.51-1H3a2 2 0 010-4h.09a1.65 1.65 0 001.51-1 1.65 1.65 0 00-.33-1.82l-.06-.06a2 2 0 012.83-2.83l.06.06a1.65 1.65 0 001.82.33H9a1.65 1.65 0 001-1.51V3a2 2 0 014 0v.09a1.65 1.65 0 001 1.51 1.65 1.65 0 001.82-.33l.06-.06a2 2 0 012.83 2.83l-.06.06a1.65 1.65 0 00-.33 1.82V9a1.65 1.65 0 001.51 1H21a2 2 0 010 4h-.09a1.65 1.65 0 00-1.51 1z"
            stroke="currentColor"
            stroke-width="1.5"
            stroke-linecap="round"
            stroke-linejoin="round"
            fill="none"
          />
        </svg>
      </button>
    </div>
    <div class="contrib-page">
      <section class="contrib-hero">
        <h1 v-if="showHomeTitle">{{ t('components.main.hero.title') }}</h1>
        <!-- <p class="lead">
          {{ t('components.main.hero.lead') }}
        </p> -->
      </section>

      <section
        v-if="showHeatmap"
        ref="heatmapContainerRef"
        class="contrib-wall"
        :aria-label="t('components.main.heatmap.ariaLabel')"
      >
        <UsageChart :data="usageHeatmap" />
      </section>

      <section class="automation-section">
      <div class="section-header">
        <div class="tab-group" role="tablist" :aria-label="t('components.main.tabs.ariaLabel')">
          <button
            v-for="(tab, idx) in tabs"
            :key="tab.id"
            class="tab-pill"
            :class="{ active: selectedIndex === idx }"
            role="tab"
            :aria-selected="selectedIndex === idx"
            type="button"
            @click="onTabChange(idx)"
          >
            {{ tab.label }}
          </button>
        </div>
        <div class="section-controls">
          <div class="relay-toggle" :aria-label="currentProxyLabel">
            <div class="relay-switch">
              <label class="mac-switch sm">
                <input
                  type="checkbox"
                  :checked="activeProxyState"
                  :disabled="activeProxyBusy"
                  @change="onProxyToggle"
                />
                <span></span>
              </label>
              <span class="relay-tooltip-content">{{ currentProxyLabel }} · {{ t('components.main.relayToggle.tooltip') }}</span>
            </div>
          </div>
          <button
            class="ghost-icon"
            :data-tooltip="t('components.main.controls.mcp')"
            @click="goToMcp"
          >
            <span class="icon-svg" v-html="mcpIcon" aria-hidden="true"></span>
          </button>
          <button
            class="ghost-icon"
            :data-tooltip="t('components.main.controls.skill')"
            @click="goToSkill"
          >
            <svg viewBox="0 0 24 24" aria-hidden="true">
              <path
                d="M6 4h8a4 4 0 014 4v12a3 3 0 00-3-3H6z"
                fill="none"
                stroke="currentColor"
                stroke-width="1.5"
                stroke-linecap="round"
                stroke-linejoin="round"
              />
              <path
                d="M6 4a2 2 0 00-2 2v13c0 .55.45 1 1 1h11"
                fill="none"
                stroke="currentColor"
                stroke-width="1.5"
                stroke-linecap="round"
                stroke-linejoin="round"
              />
              <path
                d="M9 8h5"
                stroke="currentColor"
                stroke-width="1.5"
                stroke-linecap="round"
              />
            </svg>
          </button>
          <button
            class="ghost-icon"
            :data-tooltip="activeTab === 'claude'
              ? t('components.main.controls.editClaudeConfig')
              : t('components.main.controls.editCodexConfig')"
            @click="openCommonConfigModal"
          >
            <svg viewBox="0 0 24 24" aria-hidden="true">
              <path
                d="M11 4H4a2 2 0 00-2 2v14a2 2 0 002 2h14a2 2 0 002-2v-7"
                fill="none"
                stroke="currentColor"
                stroke-width="1.5"
                stroke-linecap="round"
                stroke-linejoin="round"
              />
              <path
                d="M18.5 2.5a2.121 2.121 0 013 3L12 15l-4 1 1-4 9.5-9.5z"
                fill="none"
                stroke="currentColor"
                stroke-width="1.5"
                stroke-linecap="round"
                stroke-linejoin="round"
              />
            </svg>
          </button>
          <button
            class="ghost-icon"
            :data-tooltip="t('components.main.logs.view')"
            @click="goToLogs"
          >
            <svg viewBox="0 0 24 24" aria-hidden="true">
              <path
                d="M5 7h14M5 12h14M5 17h9"
                stroke="currentColor"
                stroke-width="1.5"
                stroke-linecap="round"
                stroke-linejoin="round"
                fill="none"
              />
            </svg>
          </button>
          <button
            class="ghost-icon"
            :data-tooltip="t('components.main.tabs.addCard')"
            @click="openCreateModal"
          >
            <svg viewBox="0 0 24 24" aria-hidden="true">
              <path
                d="M12 5v14M5 12h14"
                stroke="currentColor"
                stroke-width="1.5"
                stroke-linecap="round"
                stroke-linejoin="round"
                fill="none"
              />
            </svg>
          </button>
        </div>
      </div>
      <div class="automation-list" @dragover.prevent>
        <article
          v-for="card in activeCards"
          :key="card.id"
          :class="['automation-card', { dragging: draggingId === card.id }]"
          draggable="true"
          @dragstart="onDragStart(card.id)"
          @dragend="onDragEnd"
          @drop="onDrop(card.id)"
        >
          <div class="card-leading">
            <div class="card-icon" :style="{ backgroundColor: card.tint, color: card.accent }">
              <span
                v-if="!iconSvg(card.icon)"
                class="icon-fallback"
              >
                {{ vendorInitials(card.name) }}
              </span>
              <span
                v-else
                class="icon-svg"
                v-html="iconSvg(card.icon)"
                aria-hidden="true"
              ></span>
            </div>
            <div class="card-text">
              <div class="card-title-row">
                <p class="card-title">{{ card.name }}</p>
                <span
                  v-if="card.officialSite"
                  class="card-site"
                  role="button"
                  tabindex="0"
                  @click.stop="openOfficialSite(card.officialSite)"
                  @keydown.enter.stop.prevent="openOfficialSite(card.officialSite)"
                  @keydown.space.stop.prevent="openOfficialSite(card.officialSite)"
                >
                  {{ formatOfficialSite(card.officialSite) }}
                </span>
              </div>
              <!-- <p class="card-subtitle">{{ card.apiUrl }}</p> -->
              <p
                v-for="stats in [providerStatDisplay(card.name)]"
                :key="`metrics-${card.id}`"
                class="card-metrics"
              >
                <template v-if="stats.state !== 'ready'">
                  {{ stats.message }}
                </template>
                <template v-else>
                  <span
                    v-if="stats.successRateLabel"
                    class="card-success-rate"
                    :class="stats.successRateClass"
                  >
                    {{ stats.successRateLabel }}
                  </span>
                  <span class="card-metric-separator" aria-hidden="true">·</span>
                  <span >{{ stats.requests }}</span>
                  <span class="card-metric-separator" aria-hidden="true">·</span>
                  <span>{{ stats.tokens }}</span>
                  <span class="card-metric-separator" aria-hidden="true">·</span>
                  <span>{{ stats.cost }}</span>
                </template>
              </p>
            </div>
          </div>
          <div class="card-actions">
            <label class="mac-switch sm">
              <input type="checkbox" v-model="card.enabled" @change="persistProviders(activeTab)" />
              <span></span>
            </label>
            <button class="ghost-icon" @click="configure(card)">
              <svg viewBox="0 0 24 24" aria-hidden="true">
                <path
                  d="M12 15a3 3 0 100-6 3 3 0 000 6z"
                  stroke="currentColor"
                  stroke-width="1.5"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  fill="none"
                />
                <path
                  d="M19.4 15a1.65 1.65 0 00.33 1.82l.06.06a2 2 0 01-2.83 2.83l-.06-.06a1.65 1.65 0 00-1.82-.33 1.65 1.65 0 00-1 1.51V21a2 2 0 01-4 0v-.09a1.65 1.65 0 00-1-1.51 1.65 1.65 0 00-1.82.33l-.06.06a2 2 0 01-2.83-2.83l.06-.06a1.65 1.65 0 00.33-1.82 1.65 1.65 0 00-1.51-1H3a2 2 0 010-4h.09a1.65 1.65 0 001.51-1 1.65 1.65 0 00-.33-1.82l-.06-.06a2 2 0 012.83-2.83l.06.06a1.65 1.65 0 001.82.33H9a1.65 1.65 0 001-1.51V3a2 2 0 014 0v.09a1.65 1.65 0 001 1.51 1.65 1.65 0 001.82-.33l.06-.06a2 2 0 012.83 2.83l-.06.06a1.65 1.65 0 00-.33 1.82V9a1.65 1.65 0 001.51 1H21a2 2 0 010 4h-.09a1.65 1.65 0 00-1.51 1z"
                  stroke="currentColor"
                  stroke-width="1.5"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  fill="none"
                />
              </svg>
            </button>
            <button class="ghost-icon" @click="requestRemove(card)">
              <svg viewBox="0 0 24 24" aria-hidden="true">
                <path
                  d="M9 3h6m-7 4h8m-6 0v11m4-11v11M5 7h14l-.867 12.138A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.862L5 7z"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="1.5"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                />
              </svg>
            </button>
          </div>
        </article>
      </div>
      </section>

      <BaseModal
      :open="modalState.open"
      :title="modalState.editingId ? t('components.main.form.editTitle') : t('components.main.form.createTitle')"
      @close="closeModal"
    >
      <form class="vendor-form" @submit.prevent="submitModal">
                <label class="form-field">
                  <span>{{ t('components.main.form.labels.name') }}</span>
                  <BaseInput
                    v-model="modalState.form.name"
                    type="text"
                    :placeholder="t('components.main.form.placeholders.name')"
                    required
                    :disabled="Boolean(modalState.editingId)"
                  />
                </label>

                <div class="form-field">
                  <span class="label-row">
                    {{ t('components.main.form.labels.apiUrl') }}
                    <span v-if="modalState.errors.apiUrl" class="field-error">
                      {{ modalState.errors.apiUrl }}
                    </span>
                  </span>
                  <div class="api-url-wrapper" :class="{ 'has-error': !!modalState.errors.apiUrl }">
                    <input
                      v-model="modalState.form.apiUrl"
                      type="text"
                      class="api-url-input"
                      :placeholder="t('components.main.form.placeholders.apiUrl')"
                      required
                    />
                    <button
                      type="button"
                      class="speed-test-btn-inline"
                      :disabled="!canTestSpeed"
                      :title="t('components.main.speedTest.buttonTitle')"
                      @click="testApiSpeed"
                    >
                      <svg v-if="modalState.speedTest.testing" class="speed-test-spinner" viewBox="0 0 24 24">
                        <circle cx="12" cy="12" r="10" stroke="currentColor" stroke-width="2" fill="none" stroke-dasharray="31.4 31.4" />
                      </svg>
                      <svg v-else viewBox="0 0 24 24" aria-hidden="true">
                        <path d="M13 2L3 14h9l-1 8 10-12h-9l1-8z" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" />
                      </svg>
                    </button>
                  </div>
                  <div v-if="modalState.speedTest.latency !== null || modalState.speedTest.error" class="speed-test-result-inline">
                    <template v-if="modalState.speedTest.error">
                      <span class="speed-error">{{ modalState.speedTest.error }}</span>
                    </template>
                    <template v-else>
                      <span class="speed-latency" :class="getSpeedTestColorClass(modalState.speedTest.statusCode!)">
                        {{ modalState.speedTest.latency }}ms
                      </span>
                      <span class="speed-status">{{ t('components.main.speedTest.statusCode') }}: {{ modalState.speedTest.statusCode }}</span>
                    </template>
                  </div>
                </div>

                <label class="form-field">
                  <span>{{ t('components.main.form.labels.officialSite') }}</span>
                  <BaseInput
                    v-model="modalState.form.officialSite"
                    type="text"
                    :placeholder="t('components.main.form.placeholders.officialSite')"
                  />
                </label>

                <label class="form-field">
                  <span>{{ t('components.main.form.labels.apiKey') }}</span>
                  <BaseInput
                    v-model="modalState.form.apiKey"
                    type="text"
                    :placeholder="t('components.main.form.placeholders.apiKey')"
                  />
                </label>

                <div class="form-field">
                  <span>{{ t('components.main.form.labels.icon') }}</span>
                  <Listbox v-model="modalState.form.icon" v-slot="{ open }">
                    <div class="icon-select">
                      <ListboxButton class="icon-select-button">
                        <span class="icon-preview" v-html="iconSvg(modalState.form.icon)" aria-hidden="true"></span>
                        <span class="icon-select-label">{{ modalState.form.icon }}</span>
                        <svg viewBox="0 0 20 20" aria-hidden="true">
                          <path d="M6 8l4 4 4-4" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" fill="none" />
                        </svg>
                      </ListboxButton>
                      <ListboxOptions v-if="open" class="icon-select-options">
                        <ListboxOption
                          v-for="iconName in iconOptions"
                          :key="iconName"
                          :value="iconName"
                          v-slot="{ active, selected }"
                        >
                          <div :class="['icon-option', { active, selected }]">
                            <span class="icon-preview" v-html="iconSvg(iconName)" aria-hidden="true"></span>
                            <span class="icon-name">{{ iconName }}</span>
                          </div>
                        </ListboxOption>
                      </ListboxOptions>
                    </div>
                  </Listbox>
                </div>

                <div class="form-field">
                  <ModelWhitelistEditor v-model="modalState.form.supportedModels" />
                </div>

                <div class="form-field">
                  <ModelMappingEditor v-model="modalState.form.modelMapping" />
                </div>

                <div class="form-field switch-field">
                  <span>{{ t('components.main.form.labels.enabled') }}</span>
                  <div class="switch-inline">
                    <label class="mac-switch">
                      <input type="checkbox" v-model="modalState.form.enabled" />
                      <span></span>
                    </label>
                    <span class="switch-text">
                      {{ modalState.form.enabled ? t('components.main.form.switch.on') : t('components.main.form.switch.off') }}
                    </span>
                  </div>
                </div>

                <footer class="form-actions">
                  <BaseButton variant="outline" type="button" @click="closeModal">
                    {{ t('components.main.form.actions.cancel') }}
                  </BaseButton>
                  <BaseButton type="submit">
                    {{ t('components.main.form.actions.save') }}
                  </BaseButton>
                </footer>
      </form>
      </BaseModal>
      <BaseModal
      :open="confirmState.open"
      :title="t('components.main.form.confirmDeleteTitle')"
      variant="confirm"
      @close="closeConfirm"
    >
      <div class="confirm-body">
        <p>
          {{ t('components.main.form.confirmDeleteMessage', { name: confirmState.card?.name ?? '' }) }}
        </p>
      </div>
      <footer class="form-actions confirm-actions">
        <BaseButton variant="outline" type="button" @click="closeConfirm">
          {{ t('components.main.form.actions.cancel') }}
        </BaseButton>
        <BaseButton variant="danger" type="button" @click="confirmRemove">
          {{ t('components.main.form.actions.delete') }}
        </BaseButton>
      </footer>
      </BaseModal>

      <!-- 通用配置编辑弹窗 -->
      <BaseModal
        :open="commonConfigState.open"
        :title="activeTab === 'claude' ? t('components.main.commonConfig.titleClaude') : t('components.main.commonConfig.titleCodex')"
        @close="closeCommonConfigModal"
      >
        <form class="vendor-form" @submit.prevent="saveCommonConfig">
          <label class="form-field">
            <span>{{ activeTab === 'claude' ? t('components.main.commonConfig.hintClaude') : t('components.main.commonConfig.hintCodex') }}</span>
            <textarea
              v-model="commonConfigState.jsonText"
              class="config-textarea"
              rows="10"
              spellcheck="false"
              :placeholder="t('components.main.commonConfig.placeholder')"
            />
          </label>
          <p v-if="commonConfigState.error" class="field-error">
            {{ commonConfigState.error }}
          </p>
          <footer class="form-actions">
            <BaseButton variant="outline" type="button" @click="formatCommonConfig">
              {{ t('components.main.form.actions.format') }}
            </BaseButton>
            <BaseButton variant="outline" type="button" @click="closeCommonConfigModal">
              {{ t('components.main.form.actions.cancel') }}
            </BaseButton>
            <BaseButton type="submit">
              {{ t('components.main.form.actions.save') }}
            </BaseButton>
          </footer>
        </form>
      </BaseModal>

      <!-- 更新弹窗 -->
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

      <footer v-if="appVersion" class="main-version">
        {{ t('components.main.versionLabel', { version: appVersion }) }}
      </footer>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, reactive, ref, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { Listbox, ListboxButton, ListboxOptions, ListboxOption } from '@headlessui/vue'
import { Browser, Application } from '@wailsio/runtime'
import {
	buildUsageHeatmapMatrix,
	generateFallbackUsageHeatmap,
	DEFAULT_HEATMAP_DAYS,
	calculateHeatmapDayRange,
	type UsageHeatmapWeek,
	type UsageHeatmapDay,
} from '../../data/usageHeatmap'
import { automationCardGroups, createAutomationCards, type AutomationCard } from '../../data/cards'
import lobeIcons from '../../icons/lobeIconMap'
import BaseButton from '../common/BaseButton.vue'
import BaseModal from '../common/BaseModal.vue'
import BaseInput from '../common/BaseInput.vue'
import UsageChart from './UsageChart.vue'
import ModelWhitelistEditor from '../common/ModelWhitelistEditor.vue'
import ModelMappingEditor from '../common/ModelMappingEditor.vue'
import { LoadProviders, SaveProviders } from '../../../bindings/coderelay/services/providerservice'
import { GetCommonConfigJSON, SaveCommonConfigJSON } from '../../../bindings/coderelay/services/commonconfigservice'
import { fetchProxyStatus, enableProxy, disableProxy } from '../../services/claudeSettings'
import { fetchHeatmapStats, fetchProviderDailyStats, type ProviderDailyStat } from '../../services/logs'
import { fetchCurrentVersion } from '../../services/version'
import { fetchAppSettings, type AppSettings } from '../../services/appSettings'
import { getCurrentTheme, setTheme, type ThemeMode } from '../../utils/ThemeManager'
import { getSpeedTestColorClass } from '../../utils/speedTest'
import { checkForUpdates as checkForUpdatesService, downloadUpdate, installUpdate, formatFileSize, type UpdateInfo } from '../../services/update'
import { useRouter } from 'vue-router'

const { t, locale } = useI18n()
const router = useRouter()
const themeMode = ref<ThemeMode>(getCurrentTheme())
const resolvedTheme = computed(() => {
  if (themeMode.value === 'systemdefault') {
    return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
  }
  return themeMode.value
})
const themeIcon = computed(() => (resolvedTheme.value === 'dark' ? 'moon' : 'sun'))
const releasePageUrl = 'https://github.com/ipiggyzhu/code-relay/releases'
const releaseApiUrl = 'https://api.github.com/repos/ipiggyzhu/code-relay/releases/latest'

const HEATMAP_DAYS = DEFAULT_HEATMAP_DAYS
const usageHeatmap = ref<UsageHeatmapWeek[]>(generateFallbackUsageHeatmap(HEATMAP_DAYS))
const heatmapContainerRef = ref<HTMLElement | null>(null)

const proxyStates = reactive<Record<ProviderTab, boolean>>({
  claude: false,
  codex: false,
})
const proxyBusy = reactive<Record<ProviderTab, boolean>>({
  claude: false,
  codex: false,
})

const providerStatsMap = reactive<Record<ProviderTab, Record<string, ProviderDailyStat>>>({
  claude: {},
  codex: {},
} as Record<ProviderTab, Record<string, ProviderDailyStat>>)
const providerStatsLoading = reactive<Record<ProviderTab, boolean>>({
  claude: false,
  codex: false,
} as Record<ProviderTab, boolean>)
const providerStatsLoaded = reactive<Record<ProviderTab, boolean>>({
  claude: false,
  codex: false,
} as Record<ProviderTab, boolean>)
let providerStatsTimer: number | undefined
let updateTimer: number | undefined
const showHeatmap = ref(true)
const showHomeTitle = ref(true)
const mcpIcon = lobeIcons['mcp'] ?? ''
const appVersion = ref('')
const hasUpdateAvailable = ref(false)

// 更新弹窗状态
const updateModalState = reactive({
  open: false,
  checking: false,
  updateInfo: null as UpdateInfo | null,
  downloading: false,
  downloadProgress: 0,
  downloadedPath: '',
  installing: false,
  error: '',
})

const intensityClass = (value: number) => `gh-level-${value}`

type TooltipPlacement = 'above' | 'below'



const formatMetric = (value: number) => value.toLocaleString()

const tooltipDateFormatter = computed(() =>
  new Intl.DateTimeFormat(locale.value || 'en', {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
)

const currencyFormatter = computed(() =>
  new Intl.NumberFormat(locale.value || 'en', {
    style: 'currency',
    currency: 'USD',
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  })
)

const clamp = (value: number, min: number, max: number) => {
  if (max <= min) return min
  return Math.min(Math.max(value, min), max)
}






const loadAppSettings = async () => {
  try {
    const data: AppSettings = await fetchAppSettings()
    showHeatmap.value = data?.show_heatmap ?? true
    showHomeTitle.value = data?.show_home_title ?? true
  } catch (error) {
    console.error('failed to load app settings', error)
    showHeatmap.value = true
    showHomeTitle.value = true
  }
}

const checkForUpdates = async () => {
  try {
    const version = await fetchCurrentVersion()
    appVersion.value = version || ''
  } catch (error) {
    console.error('failed to load app version', error)
  }

  try {
    const resp = await fetch(releaseApiUrl, {
      headers: {
        Accept: 'application/vnd.github+json',
      },
    })
    if (!resp.ok) {
      return
    }
    const data = await resp.json()
    const latestTag = data?.tag_name ?? ''
    if (latestTag && compareVersions(appVersion.value || '0.0.0', latestTag) < 0) {
      hasUpdateAvailable.value = true
    }
  } catch (error) {
    console.error('failed to fetch release info', error)
  }
}

// 打开更新弹窗
const openUpdateModal = async () => {
  updateModalState.open = true
  updateModalState.checking = true
  updateModalState.error = ''
  updateModalState.downloadedPath = ''
  updateModalState.downloading = false
  updateModalState.installing = false

  try {
    const info = await checkForUpdatesService()
    updateModalState.updateInfo = info
  } catch (error) {
    updateModalState.error = String(error)
  } finally {
    updateModalState.checking = false
  }
}

const closeUpdateModal = () => {
  updateModalState.open = false
}

// 下载更新
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

// 安装更新
const handleInstallUpdate = async () => {
  if (!updateModalState.downloadedPath) return

  updateModalState.installing = true
  updateModalState.error = ''

  try {
    const success = await installUpdate(updateModalState.downloadedPath)
    if (success) {
      // 安装脚本已启动，等待一下然后退出应用
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

// 打开发布页面
const openReleasePage = () => {
  if (updateModalState.updateInfo?.releaseUrl) {
    Browser.OpenURL(updateModalState.updateInfo.releaseUrl)
  }
}

const handleAppSettingsUpdated = () => {
  void loadAppSettings()
}

const startUpdateTimer = () => {
  stopUpdateTimer()
  updateTimer = window.setInterval(() => {
    void checkForUpdates()
  }, 60 * 60 * 1000)
}

const stopUpdateTimer = () => {
  if (updateTimer) {
    clearInterval(updateTimer)
    updateTimer = undefined
  }
}

const normalizeProviderKey = (value: string) => value?.trim().toLowerCase() ?? ''

const normalizeVersion = (value: string) => value.replace(/^v/i, '').trim()

const compareVersions = (current: string, remote: string) => {
  const curParts = normalizeVersion(current).split('.').map((part) => parseInt(part, 10) || 0)
  const remoteParts = normalizeVersion(remote).split('.').map((part) => parseInt(part, 10) || 0)
  const maxLen = Math.max(curParts.length, remoteParts.length)
  for (let i = 0; i < maxLen; i++) {
    const cur = curParts[i] ?? 0
    const rem = remoteParts[i] ?? 0
    if (cur === rem) continue
    return cur < rem ? -1 : 1
  }
  return 0
}

const loadUsageHeatmap = async () => {
	try {
		const rangeDays = calculateHeatmapDayRange(HEATMAP_DAYS)
		const stats = await fetchHeatmapStats(rangeDays)
		usageHeatmap.value = buildUsageHeatmapMatrix(stats, HEATMAP_DAYS)
	} catch (error) {
		console.error('Failed to load usage heatmap stats', error)
	}
}

const tabs = [
  { id: 'claude', label: 'Claude Code' },
  { id: 'codex', label: 'Codex' },
] as const
type ProviderTab = (typeof tabs)[number]['id']
const providerTabIds = tabs.map((tab) => tab.id) as ProviderTab[]

const cards = reactive<Record<ProviderTab, AutomationCard[]>>({
  claude: createAutomationCards(automationCardGroups.claude),
  codex: createAutomationCards(automationCardGroups.codex),
})
const draggingId = ref<number | null>(null)

const serializeProviders = (providers: AutomationCard[]) => providers.map((provider) => ({ ...provider }))

const persistProviders = async (tabId: ProviderTab) => {
  try {
    await SaveProviders(tabId, serializeProviders(cards[tabId]))
  } catch (error) {
    console.error('Failed to save providers', error)
  }
}

const replaceProviders = (tabId: ProviderTab, data: AutomationCard[]) => {
  cards[tabId].splice(0, cards[tabId].length, ...createAutomationCards(data))
}

const loadProvidersFromDisk = async () => {
  for (const tab of providerTabIds) {
    try {
      const saved = await LoadProviders(tab)
      if (Array.isArray(saved)) {
        replaceProviders(tab, saved as AutomationCard[])
      } else {
        await persistProviders(tab)
      }
    } catch (error) {
      console.error('Failed to load providers', error)
    }
  }
}

const refreshProxyState = async (tab: ProviderTab) => {
  try {
    const status = await fetchProxyStatus(tab)
    proxyStates[tab] = Boolean(status?.enabled)
  } catch (error) {
    console.error(`Failed to fetch proxy status for ${tab}`, error)
    proxyStates[tab] = false
  }
}

const onProxyToggle = async () => {
  const tab = activeTab.value
  if (proxyBusy[tab]) return
  proxyBusy[tab] = true
  const nextState = !proxyStates[tab]
  try {
    if (nextState) {
      await enableProxy(tab)
    } else {
      await disableProxy(tab)
    }
    proxyStates[tab] = nextState
  } catch (error) {
    console.error(`Failed to toggle proxy for ${tab}`, error)
  } finally {
    proxyBusy[tab] = false
  }
}

const loadProviderStats = async (tab: ProviderTab) => {
  providerStatsLoading[tab] = true
  try {
    const stats = await fetchProviderDailyStats(tab)
    const mapped: Record<string, ProviderDailyStat> = {}
    ;(stats ?? []).forEach((stat) => {
      mapped[normalizeProviderKey(stat.provider)] = stat
    })
    const hadExistingStats = Object.keys(providerStatsMap[tab] ?? {}).length > 0
    if ((stats?.length ?? 0) > 0) {
      providerStatsMap[tab] = mapped
    } else if (!hadExistingStats) {
      providerStatsMap[tab] = mapped
    }
    providerStatsLoaded[tab] = true
  } catch (error) {
    console.error(`Failed to load provider stats for ${tab}`, error)
    if (!providerStatsLoaded[tab]) {
      providerStatsLoaded[tab] = true
    }
  } finally {
    providerStatsLoading[tab] = false
  }
}

type ProviderStatDisplay =
  | { state: 'loading' | 'empty'; message: string }
  | {
      state: 'ready'
      requests: string
      tokens: string
      cost: string
      successRateLabel: string
      successRateClass: string
    }

const SUCCESS_RATE_THRESHOLDS = {
  healthy: 0.95,
  warning: 0.8,
} as const

const formatSuccessRateLabel = (value: number) => {
  const percent = clamp(value, 0, 1) * 100
  const decimals = percent >= 99.5 || percent === 0 ? 0 : 1
  return `${t('components.main.providers.successRate')}: ${percent.toFixed(decimals)}%`
}

const successRateClassName = (value: number) => {
  const rate = clamp(value, 0, 1)
  if (rate >= SUCCESS_RATE_THRESHOLDS.healthy) {
    return 'success-good'
  }
  if (rate >= SUCCESS_RATE_THRESHOLDS.warning) {
    return 'success-warn'
  }
  return 'success-bad'
}

const providerStatDisplay = (providerName: string): ProviderStatDisplay => {
  const tab = activeTab.value
  if (!providerStatsLoaded[tab]) {
    return { state: 'loading', message: t('components.main.providers.loading') }
  }
  const stat = providerStatsMap[tab]?.[normalizeProviderKey(providerName)]
  if (!stat) {
    return { state: 'empty', message: t('components.main.providers.noData') }
  }
  const totalTokens = stat.input_tokens + stat.output_tokens
  const successRateValue = Number.isFinite(stat.success_rate) ? clamp(stat.success_rate, 0, 1) : null
  const successRateLabel = successRateValue !== null ? formatSuccessRateLabel(successRateValue) : ''
  const successRateClass = successRateValue !== null ? successRateClassName(successRateValue) : ''
  return {
    state: 'ready',
    requests: `${t('components.main.providers.requests')}: ${formatMetric(stat.total_requests)}`,
    tokens: `${t('components.main.providers.tokens')}: ${formatMetric(totalTokens)}`,
    cost: `${t('components.main.providers.cost')}: ${currencyFormatter.value.format(Math.max(stat.cost_total, 0))}`,
    successRateLabel,
    successRateClass,
  }
}

const normalizeUrlWithScheme = (value: string) => {
  if (!value) return ''
  try {
    const url = new URL(value)
    return url.toString()
  } catch {
    return `https://${value}`
  }
}

const openOfficialSite = (site: string) => {
  const target = normalizeUrlWithScheme(site)
  if (!target) return
  Browser.OpenURL(target).catch(() => {
    console.error('failed to open link', target)
  })
}

const formatOfficialSite = (site: string) => {
  if (!site) return ''
  try {
    const url = new URL(normalizeUrlWithScheme(site))
    return url.hostname.replace(/^www\./, '')
  } catch {
    return site
  }
}

const startProviderStatsTimer = () => {
  stopProviderStatsTimer()
  providerStatsTimer = window.setInterval(() => {
    providerTabIds.forEach((tab) => {
      void loadProviderStats(tab)
    })
  }, 5_000) // 5秒刷新一次，更实时
}

const stopProviderStatsTimer = () => {
  if (providerStatsTimer) {
    clearInterval(providerStatsTimer)
    providerStatsTimer = undefined
  }
}

onMounted(async () => {
  void loadUsageHeatmap()
  await loadProvidersFromDisk()
  await Promise.all(providerTabIds.map(refreshProxyState))
  await Promise.all(providerTabIds.map((tab) => loadProviderStats(tab)))
  await loadAppSettings()
  await checkForUpdates()
  startProviderStatsTimer()
  startUpdateTimer()
  window.addEventListener('app-settings-updated', handleAppSettingsUpdated)
})

onUnmounted(() => {
  stopProviderStatsTimer()
  window.removeEventListener('app-settings-updated', handleAppSettingsUpdated)
  stopUpdateTimer()
})

const selectedIndex = ref(0)
const activeTab = computed<ProviderTab>(() => tabs[selectedIndex.value]?.id ?? tabs[0].id)
const activeCards = computed(() => cards[activeTab.value] ?? [])
const currentProxyLabel = computed(() =>
  activeTab.value === 'claude'
    ? t('components.main.relayToggle.hostClaude')
    : t('components.main.relayToggle.hostCodex')
)
const activeProxyState = computed(() => proxyStates[activeTab.value])
const activeProxyBusy = computed(() => proxyBusy[activeTab.value])

const goToLogs = () => {
  router.push('/logs')
}

const goToMcp = () => {
  router.push('/mcp')
}

const goToSkill = () => {
  router.push('/skill')
}

const goToSettings = () => {
  router.push('/settings')
}

const toggleTheme = () => {
  const next = resolvedTheme.value === 'dark' ? 'light' : 'dark'
  themeMode.value = next
  setTheme(next)
}

const openGitHub = () => {
  Browser.OpenURL(releasePageUrl).catch(() => {
    console.error('failed to open github')
  })
}

type VendorForm = {
  name: string
  apiUrl: string
  apiKey: string
  officialSite: string
  icon: string
  enabled: boolean
  supportedModels?: Record<string, boolean>
  modelMapping?: Record<string, string>
}

const iconOptions = Object.keys(lobeIcons).sort((a, b) => a.localeCompare(b))
const defaultIconKey = iconOptions[0] ?? 'aicoding'

const defaultFormValues = (): VendorForm => ({
  name: '',
  apiUrl: '',
  apiKey: '',
  officialSite: '',
  icon: defaultIconKey,
  enabled: true,
  supportedModels: {},
  modelMapping: {},
})

const modalState = reactive({
  open: false,
  tabId: tabs[0].id as ProviderTab,
  editingId: null as number | null,
  form: defaultFormValues(),
  errors: {
    apiUrl: '',
  },
  speedTest: {
    testing: false,
    latency: null as number | null,
    statusCode: null as number | null,
    error: null as string | null,
  },
})

const editingCard = ref<AutomationCard | null>(null)
const confirmState = reactive({ open: false, card: null as AutomationCard | null, tabId: tabs[0].id as ProviderTab })

// 通用配置状态
const commonConfigState = reactive({
  open: false,
  jsonText: '',
  error: '',
})

const resetSpeedTestState = () => {
  modalState.speedTest.testing = false
  modalState.speedTest.latency = null
  modalState.speedTest.statusCode = null
  modalState.speedTest.error = null
}

const testApiSpeed = async () => {
  const url = modalState.form.apiUrl.trim()
  if (!url) return

  // Validate URL format
  try {
    const parsed = new URL(url)
    if (!/^https?:/.test(parsed.protocol)) {
      modalState.speedTest.error = t('components.main.speedTest.invalidUrl')
      return
    }
  } catch {
    modalState.speedTest.error = t('components.main.speedTest.invalidUrl')
    return
  }

  resetSpeedTestState()
  modalState.speedTest.testing = true

  const controller = new AbortController()
  const timeoutId = setTimeout(() => controller.abort(), 5000)

  try {
    const startTime = performance.now()
    const response = await fetch(url, {
      method: 'HEAD',
      mode: 'no-cors',
      signal: controller.signal,
    })
    const endTime = performance.now()
    clearTimeout(timeoutId)

    modalState.speedTest.latency = Math.round(endTime - startTime)
    // In no-cors mode, we can't access status, so we assume success if no error
    modalState.speedTest.statusCode = response.type === 'opaque' ? 200 : response.status
  } catch (error: any) {
    clearTimeout(timeoutId)
    if (error.name === 'AbortError') {
      modalState.speedTest.error = t('components.main.speedTest.timeout')
    } else {
      modalState.speedTest.error = t('components.main.speedTest.networkError')
    }
  } finally {
    modalState.speedTest.testing = false
  }
}

const canTestSpeed = computed(() => {
  return modalState.form.apiUrl.trim().length > 0 && !modalState.speedTest.testing
})

const openCreateModal = () => {
  modalState.tabId = activeTab.value
  modalState.editingId = null
  editingCard.value = null
  Object.assign(modalState.form, defaultFormValues())
  modalState.errors.apiUrl = ''
  resetSpeedTestState()
  modalState.open = true
}

const openEditModal = (card: AutomationCard) => {
  modalState.tabId = activeTab.value
  modalState.editingId = card.id
  editingCard.value = card
  Object.assign(modalState.form, {
    name: card.name,
    apiUrl: card.apiUrl,
    apiKey: card.apiKey,
    officialSite: card.officialSite,
    icon: card.icon,
    enabled: card.enabled,
    supportedModels: card.supportedModels || {},
    modelMapping: card.modelMapping || {},
  })
  modalState.errors.apiUrl = ''
  resetSpeedTestState()
  modalState.open = true
}

const closeModal = () => {
  modalState.open = false
}

const closeConfirm = () => {
  confirmState.open = false
  confirmState.card = null
}

const submitModal = () => {
  const list = cards[modalState.tabId]
  if (!list) return
  const name = modalState.form.name.trim()
  const apiUrl = modalState.form.apiUrl.trim().replace(/\/+$/, '') // 去除末尾斜杠
  const apiKey = modalState.form.apiKey.trim()
  const officialSite = modalState.form.officialSite.trim()
  const icon = (modalState.form.icon || defaultIconKey).toString().trim().toLowerCase() || defaultIconKey
  modalState.errors.apiUrl = ''
  try {
    const parsed = new URL(apiUrl)
    if (!/^https?:/.test(parsed.protocol)) throw new Error('protocol')
  } catch {
    modalState.errors.apiUrl = t('components.main.form.errors.invalidUrl')
    return
  }

  if (editingCard.value) {
    Object.assign(editingCard.value, {
      apiUrl: apiUrl || editingCard.value.apiUrl,
      apiKey,
      officialSite,
      icon,
      enabled: modalState.form.enabled,
      supportedModels: modalState.form.supportedModels || {},
      modelMapping: modalState.form.modelMapping || {},
    })
    void persistProviders(modalState.tabId)
  } else {
    const newCard: AutomationCard = {
      id: Date.now(),
      name: name || 'Untitled vendor',
      apiUrl,
      apiKey,
      officialSite,
      icon,
      accent: '#0a84ff',
      tint: 'rgba(15, 23, 42, 0.12)',
      enabled: modalState.form.enabled,
      supportedModels: modalState.form.supportedModels || {},
      modelMapping: modalState.form.modelMapping || {},
    }
    list.push(newCard)
    void persistProviders(modalState.tabId)
  }

  closeModal()
}

const configure = (card: AutomationCard) => {
  openEditModal(card)
}

const remove = (id: number, tabId: ProviderTab = activeTab.value) => {
  const list = cards[tabId]
  if (!list) return
  const index = list.findIndex((card) => card.id === id)
  if (index > -1) {
    list.splice(index, 1)
    void persistProviders(tabId)
  }
}

const requestRemove = (card: AutomationCard) => {
  confirmState.card = card
  confirmState.tabId = activeTab.value
  confirmState.open = true
}

const confirmRemove = () => {
  if (!confirmState.card) return
  remove(confirmState.card.id, confirmState.tabId)
  closeConfirm()
}

// 通用配置相关函数
const openCommonConfigModal = async () => {
  commonConfigState.error = ''
  commonConfigState.jsonText = ''
  commonConfigState.open = true
  
  // 异步加载配置（不阻塞模态框打开）
  try {
    const json = await GetCommonConfigJSON(activeTab.value)
    if (json && json !== '{}') {
      // 格式化显示
      const parsed = JSON.parse(json)
      commonConfigState.jsonText = JSON.stringify(parsed, null, 2)
    }
  } catch (e) {
    console.error('Failed to load common config:', e)
  }
}



const closeCommonConfigModal = () => {
  commonConfigState.open = false
}

const validateCommonConfig = () => {
  if (!commonConfigState.jsonText.trim()) {
    commonConfigState.error = ''
    return true
  }

  try {
    JSON.parse(commonConfigState.jsonText)
    commonConfigState.error = ''
    return true
  } catch (e: any) {
    commonConfigState.error = `JSON 格式错误: ${e.message}`
    return false
  }
}

const formatCommonConfig = () => {
  if (!commonConfigState.jsonText.trim()) {
    return
  }

  try {
    const parsed = JSON.parse(commonConfigState.jsonText)
    commonConfigState.jsonText = JSON.stringify(parsed, null, 2)
    commonConfigState.error = ''
  } catch (e: any) {
    commonConfigState.error = `JSON 格式错误: ${e.message}`
  }
}

const saveCommonConfig = async () => {
  if (!validateCommonConfig()) {
    return
  }

  const kind = activeTab.value
  
  // 准备要保存的 JSON 字符串
  let jsonStr = '{}'
  if (commonConfigState.jsonText.trim()) {
    try {
      // 验证并格式化 JSON
      const config = JSON.parse(commonConfigState.jsonText)
      jsonStr = JSON.stringify(config)
    } catch (e: any) {
      commonConfigState.error = `JSON 格式错误: ${e.message}`
      return
    }
  }

  commonConfigState.error = ''

  try {
    // 1. 保存配置（使用 JSON 字符串方式）
    await SaveCommonConfigJSON(kind, jsonStr)

    // 2. 如果代理已启用，则重启代理以应用新配置
    if (activeProxyState.value) {
      try {
        await enableProxy(kind)
      } catch (proxyError: any) {
        throw new Error(`配置已保存，但相关服务重启失败: ${proxyError.message || proxyError}`)
      }
    }

    // 3. 一切顺利，关闭窗口
    closeCommonConfigModal()
  } catch (error: any) {
    console.error('Save failed:', error)
    commonConfigState.error = `保存失败: ${error.message || error}`
  }
}

const onDragStart = (id: number) => {
  draggingId.value = id
}

const onDrop = (targetId: number) => {
  if (draggingId.value === null || draggingId.value === targetId) return
  const currentTab = activeTab.value
  const list = cards[currentTab]
  if (!list) return
  const fromIndex = list.findIndex((card) => card.id === draggingId.value)
  const toIndex = list.findIndex((card) => card.id === targetId)
  if (fromIndex === -1 || toIndex === -1) return
  const [moved] = list.splice(fromIndex, 1)
  const newIndex = fromIndex < toIndex ? toIndex - 1 : toIndex
  list.splice(newIndex, 0, moved)
  draggingId.value = null
  void persistProviders(currentTab)
}

const onDragEnd = () => {
  draggingId.value = null
}

const iconSvg = (name: string) => {
  if (!name) return ''
  return lobeIcons[name.toLowerCase()] ?? ''
}

const vendorInitials = (name: string) => {
  if (!name) return 'AI'
  return name
    .split(/\s+/)
    .filter(Boolean)
    .map((word) => word[0])
    .join('')
    .slice(0, 2)
    .toUpperCase()
}

const onTabChange = (idx: number) => {
  selectedIndex.value = idx
  const nextTab = tabs[idx]?.id
  if (nextTab) {
    void refreshProxyState(nextTab as ProviderTab)
    void loadProviderStats(nextTab as ProviderTab)
  }
}

</script>

<style scoped>
.main-version {
  margin: 32px auto 12px;
  text-align: center;
  color: var(--mac-text-secondary);
  font-size: 0.85rem;
}

/* 更新弹窗样式 */
.update-modal-content {
  padding: 8px 0;
}

.update-checking {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  padding: 24px;
}

.update-spinner {
  width: 32px;
  height: 32px;
  border: 3px solid var(--mac-border);
  border-top-color: var(--mac-accent);
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.update-info {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.update-versions {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 12px;
  background: var(--mac-surface-strong);
  border-radius: 8px;
}

.version-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.version-label {
  color: var(--mac-text-secondary);
  font-size: 0.875rem;
}

.version-value {
  font-weight: 500;
}

.version-new {
  color: var(--mac-accent);
}

.update-actions {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.update-hint {
  color: var(--mac-text-secondary);
  font-size: 0.875rem;
  text-align: center;
}

.update-error {
  color: var(--mac-danger);
  font-size: 0.875rem;
  padding: 8px 12px;
  background: rgba(239, 68, 68, 0.1);
  border-radius: 6px;
}

.update-no-update {
  text-align: center;
  padding: 24px;
}

.update-no-update p:first-child {
  font-size: 1rem;
  color: var(--mac-text);
  margin-bottom: 8px;
}

.update-current {
  color: var(--mac-text-secondary);
  font-size: 0.875rem;
}

/* 通用配置编辑器样式 */
.common-config-editor {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.editor-hint {
  color: var(--mac-text-secondary);
  font-size: 0.875rem;
  line-height: 1.5;
}

.config-textarea {
  width: 100%;
  min-height: 280px;
  padding: 12px;
  font-family: inherit; /* Match main interface */
  font-weight: normal;  /* Ensure no bolding */
  font-size: 0.875rem;
  line-height: 1.6;
  color: var(--mac-text);
  background-color: var(--mac-surface-strong);
  border: 1.5px solid var(--mac-border);
  border-radius: 8px;
  resize: vertical;
  transition: all 0.2s;
  box-shadow: inset 0 1px 3px rgba(0, 0, 0, 0.08);
}

.config-textarea:hover {
  border-color: var(--mac-text-secondary);
}

.config-textarea:focus {
  outline: none;
  border-color: var(--mac-accent);
  box-shadow: inset 0 1px 3px rgba(0, 0, 0, 0.08), 0 0 0 3px rgba(10, 132, 255, 0.15);
}

.config-textarea.is-loading {
  opacity: 0.6;
  cursor: wait;
}

.error-message {
  padding: 8px 12px;
  background-color: rgba(239, 68, 68, 0.1);
  border: 1px solid #ff3b30;
  border-radius: 6px;
  color: #ff3b30;
  font-size: 0.8125rem;
  font-family: inherit;
}

.config-examples {
  padding: 12px;
  background-color: var(--mac-surface-strong);
  border-radius: 8px;
  font-size: 0.8125rem;
}

.examples-title {
  font-weight: 500;
  margin-bottom: 8px;
  color: var(--mac-text);
}

.examples-list {
  margin: 0;
  padding-left: 20px;
  list-style: disc;
  color: var(--mac-text-secondary);
}

.examples-list li {
  margin-bottom: 6px;
  line-height: 1.5;
}

.examples-list code {
  padding: 2px 6px;
  background-color: var(--mac-surface);
  border: 1px solid var(--mac-border);
  border-radius: 4px;
  font-family: inherit;
  font-size: 0.75rem;
  color: var(--mac-accent);
}

/* Speed Test Styles - Input with inline button */
.api-url-wrapper {
  display: flex;
  align-items: center;
  border: 1px solid var(--mac-border);
  border-radius: 12px;
  background: var(--mac-surface-strong);
  transition: border-color 0.2s, box-shadow 0.2s;
  overflow: hidden;
}

.api-url-wrapper:focus-within {
  border-color: var(--mac-accent);
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--mac-accent) 25%, transparent);
}

.api-url-wrapper.has-error {
  border-color: rgba(255, 59, 48, 0.8);
}

.api-url-input {
  flex: 1;
  border: none;
  background: transparent;
  padding: 10px 14px;
  font: inherit;
  color: var(--mac-text);
  outline: none;
  min-width: 0;
}

.api-url-input::placeholder {
  color: var(--mac-text-secondary);
}

.speed-test-btn-inline {
  flex-shrink: 0;
  width: 36px;
  height: 36px;
  margin-right: 4px;
  border: none;
  border-radius: 8px;
  background: transparent;
  color: var(--mac-text-secondary);
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  transition: background 0.2s, color 0.2s;
}

.speed-test-btn-inline:hover:not(:disabled) {
  background: var(--mac-border);
  color: var(--mac-text);
}

.speed-test-btn-inline:disabled {
  opacity: 0.3;
  cursor: not-allowed;
}

.speed-test-btn-inline svg {
  width: 16px;
  height: 16px;
}

.speed-test-spinner {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.speed-test-result-inline {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-top: 6px;
  padding-left: 2px;
}

.speed-latency {
  font-size: 0.85rem;
  font-weight: 600;
  font-variant-numeric: tabular-nums;
}

.speed-status {
  font-size: 0.72rem;
  color: var(--mac-text-secondary);
}

.speed-error {
  font-size: 0.8rem;
  color: #ff3b30;
}

/* Speed test color classes */
.speed-success {
  color: #34c759;
}

.speed-redirect {
  color: #38bdf8;
}

.speed-client-error {
  color: #f59e0b;
}

.speed-server-error {
  color: #ff3b30;
}

</style>
