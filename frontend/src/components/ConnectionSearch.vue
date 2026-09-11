<template>
  <div class="connection-search" ref="rootRef">
    <el-input
      ref="inputRef"
      v-model="query"
      :placeholder="t('quickSearch.placeholder')"
      clearable
      size="small"
      class="qs-input"
      @input="onInput"
      @focus="onFocus"
      @keydown="onKeydown"
      @blur="onBlur"
    />

    <ul v-if="open && query.trim()" class="qs-dropdown">
      <li
        v-for="(c, i) in results"
        :key="c.id"
        class="qs-item"
        :class="{ active: i === activeIndex }"
        @mousedown.prevent="openConnection(c)"
      >
        <span class="qs-type">{{ typeLabel(c.type) }}</span>
        <span class="qs-name">{{ c.name }}</span>
        <span class="qs-host">{{ c.user ? `${c.user}@` : '' }}{{ c.host }}</span>
      </li>
      <li v-if="results.length === 0" class="qs-empty">{{ t('quickSearch.noResult') }}</li>
    </ul>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue'
import { useI18n } from '../i18n'
import { useConnectionStore } from '../stores/connectionStore'
import type { ConnectionConfig } from '../types/session'

const emit = defineEmits<{ (e: 'open', config: ConnectionConfig): void }>()
const { t } = useI18n()
const connectionStore = useConnectionStore()

const rootRef = ref<HTMLElement>()
const inputRef = ref<{ focus: () => void }>()
const query = ref('')
const open = ref(false)
const activeIndex = ref(0)

// Type → short badge label. Falls back to the raw type for anything unmapped.
const TYPE_SHORT: Record<string, string> = {
  ssh: 'SSH', telnet: 'TEL', mosh: 'MOSH', serial: 'SER', tcp: 'TCP',
  rdp: 'RDP', vnc: 'VNC', spice: 'SPICE', 'x11-desktop': 'X11',
  sftp: 'SFTP', scp: 'SCP', ftp: 'FTP', smb: 'SMB', webdav: 'WDV', s3: 'S3',
  database: 'DB', k8s: 'K8S', container: 'CTR', local: 'LOC', wsl: 'WSL',
}
function typeLabel(type: string): string {
  return (TYPE_SHORT[type] || type).toUpperCase()
}

function normalize(s: string): string {
  return s.toLowerCase()
}

const results = computed<ConnectionConfig[]>(() => {
  const q = normalize(query.value.trim())
  if (!q) return []
  return connectionStore.connections
    .filter(c => {
      const name = normalize(c.name || '')
      const host = normalize(c.host || '')
      const user = normalize(c.user || '')
      return name.includes(q) || host.includes(q) || user.includes(q)
    })
    .slice(0, 50)
})

watch(() => query.value, () => {
  activeIndex.value = 0
  open.value = query.value.trim().length > 0
})

function onInput() {
  activeIndex.value = 0
}
function onFocus() {
  if (query.value.trim()) open.value = true
}
function onBlur() {
  // Defer close so a mousedown on a dropdown item still registers.
  nextTick(() => { open.value = false })
}
function onKeydown(e: KeyboardEvent) {
  if (!open.value) return
  if (e.key === 'ArrowDown') {
    e.preventDefault()
    activeIndex.value = Math.min(activeIndex.value + 1, results.value.length - 1)
  } else if (e.key === 'ArrowUp') {
    e.preventDefault()
    activeIndex.value = Math.max(activeIndex.value - 1, 0)
  } else if (e.key === 'Enter') {
    e.preventDefault()
    const c = results.value[activeIndex.value]
    if (c) openConnection(c)
  } else if (e.key === 'Escape') {
    open.value = false
    inputRef.value?.focus()
  }
}

function openConnection(c: ConnectionConfig) {
  emit('open', c)
  query.value = ''
  open.value = false
}

function focus() {
  inputRef.value?.focus()
  if (query.value.trim()) open.value = true
}

defineExpose({ focus })
</script>

<style scoped>
.connection-search {
  position: relative;
  display: flex;
  align-items: center;
  margin: 0 10px;
  width: 330px;
  min-width: 330px;
}
.qs-input :deep(.el-input__wrapper) {
  background: var(--bg-surface, rgba(255, 255, 255, 0.06));
  box-shadow: none;
  border-radius: var(--radius-sm, 6px);
}
.qs-dropdown {
  position: absolute;
  top: calc(100% + 4px);
  right: 0;
  left: 0;
  z-index: 1000;
  margin: 0;
  padding: 4px;
  list-style: none;
  max-height: 360px;
  overflow-y: auto;
  border-radius: var(--radius-md, 8px);
  border: 1px solid var(--border-subtle, rgba(255, 255, 255, 0.08));
  background: var(--bg-elevated, #191c24);
  box-shadow: var(--shadow-md, 0 8px 24px rgba(0, 0, 0, 0.18));
}
.qs-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 8px;
  border-radius: var(--radius-sm, 6px);
  cursor: pointer;
  font-size: 12px;
  color: var(--text-primary, #ddd);
}
.qs-item.active {
  background: var(--bg-hover, rgba(255, 255, 255, 0.08));
}
.qs-item .qs-type {
  flex: 0 0 auto;
  padding: 1px 5px;
  border-radius: 4px;
  font-size: 10px;
  font-weight: 600;
  background: var(--accent-subtle, rgba(99, 179, 237, 0.15));
  color: var(--accent, #63b3ed);
}
.qs-name {
  font-weight: 500;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.qs-host {
  margin-left: auto;
  color: var(--text-secondary, rgba(0, 0, 0, 0.55));
  white-space: nowrap;
}
.qs-empty {
  padding: 10px;
  text-align: center;
  font-size: 12px;
  color: var(--text-tertiary, rgba(0, 0, 0, 0.6));
}
</style>