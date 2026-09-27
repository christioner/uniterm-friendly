<template>
  <div class="pp-listwrap">
    <div class="pp-toolbar">
      <el-button size="small" type="primary" :disabled="locked" @click="emit('add')">
        {{ t('passwordPool.addAccount') }}
      </el-button>
      <el-button size="small" @click="load">{{ t('passwordPool.refresh') }}</el-button>
    </div>

    <p v-if="locked" class="pp-tip">{{ t('passwordPool.unlockRequired') }}</p>
    <p v-else-if="accounts.length === 0" class="pp-empty">{{ t('passwordPool.empty') }}</p>
    <ul v-else class="pp-list">
      <li v-for="a in accounts" :key="a.id" class="pp-row">
        <span class="pp-name">{{ a.name }}</span>
        <span v-if="a.username" class="pp-user">{{ a.username }}</span>
        <span class="pp-actions">
          <el-button size="small" @click="emit('edit', a)">{{ t('common.edit') }}</el-button>
          <el-button size="small" type="danger" plain @click="remove(a)">{{ t('common.delete') }}</el-button>
        </span>
      </li>
    </ul>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useI18n } from '../i18n'
import { usePasswordPoolStore } from '../stores/passwordPoolStore'
import type { PasswordAccount } from '../types/passwordPool'

const emit = defineEmits<{
  (e: 'add'): void
  (e: 'edit', account: PasswordAccount): void
}>()

const { t } = useI18n()
const pool = usePasswordPoolStore()

const accounts = computed(() => pool.accounts)
const locked = computed(() => !pool.unlocked())

async function load() {
  try {
    await pool.load()
  } catch (e) {
    console.error('Failed to load password pool:', e)
  }
}

async function remove(a: PasswordAccount) {
  try {
    await ElMessageBox.confirm(
      t('passwordPool.deleteConfirm', { name: a.name }),
      t('passwordPool.delete'),
      { type: 'warning', confirmButtonText: t('common.delete'), cancelButtonText: t('common.cancel') }
    )
  } catch { return }
  try {
    await pool.remove(a.id)
    ElMessage.success(t('passwordPool.deleted'))
  } catch (e) {
    console.error('Delete password account failed:', e)
  }
}

defineExpose({ load })
</script>

<style scoped>
.pp-toolbar {
  display: flex;
  gap: 0.5rem;
  margin-bottom: 0.75rem;
}
.pp-tip, .pp-empty {
  color: var(--el-text-color-secondary);
  font-size: 0.8125rem;
}
.pp-list {
  list-style: none;
  margin: 0;
  padding: 0;
}
.pp-row {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.5rem 0;
  border-bottom: 1px solid var(--el-border-color-lighter);
}
.pp-name { font-weight: 500; }
.pp-user { color: var(--el-text-color-secondary); font-size: 0.8125rem; }
.pp-actions {
  margin-left: auto;
  display: flex;
  gap: 0.25rem;
  flex-shrink: 0;
}
</style>