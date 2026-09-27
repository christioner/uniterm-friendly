<template>
  <el-dialog append-to-body
    :model-value="visible"
    @update:model-value="(v: boolean) => !v && close()"
    :title="isEdit ? t('passwordPool.editAccount') : t('passwordPool.addAccount')"
    width="26rem"
    :close-on-click-modal="false"
  >
    <el-form label-width="5rem" label-position="right" @submit.prevent="save">
      <el-form-item :label="t('passwordPool.name')" :required="true">
        <el-input v-model="form.name" :placeholder="t('passwordPool.namePlaceholder')" autofocus />
      </el-form-item>
      <el-form-item :label="t('passwordPool.username')">
        <el-input v-model="form.username" autocomplete="off" />
      </el-form-item>
      <el-form-item :label="t('passwordPool.password')" :required="!isEdit">
        <el-input v-model="form.password" type="password" show-password
          :placeholder="isEdit ? t('passwordPool.keepPasswordHint') : ''" autocomplete="off" />
      </el-form-item>
      <el-form-item :label="t('passwordPool.notes')">
        <el-input v-model="form.notes" type="textarea" :rows="2" />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="close">{{ t('common.cancel') }}</el-button>
      <el-button type="primary" :disabled="!form.name || (!isEdit && !form.password)" @click="save">
        {{ t('passwordPool.save') }}
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { useI18n } from '../i18n'
import { usePasswordPoolStore } from '../stores/passwordPoolStore'
import type { PasswordAccount } from '../types/passwordPool'

const props = defineProps<{ visible: boolean; account?: PasswordAccount | null }>()
const emit = defineEmits<{
  (e: 'update:visible', v: boolean): void
  (e: 'saved'): void
}>()

const { t } = useI18n()
const pool = usePasswordPoolStore()

const form = ref<PasswordAccount>({ id: '', name: '', username: '', password: '', notes: '' })
const isEdit = ref(false)

watch(() => props.visible, (v) => {
  if (!v) return
  const a = props.account
  isEdit.value = !!(a && a.id)
  form.value = {
    id: a?.id || '',
    name: a?.name || '',
    username: a?.username || '',
    password: '',
    notes: a?.notes || '',
  }
})

async function save() {
  try {
    await pool.save({ ...form.value })
    ElMessage.success(t('passwordPool.saved'))
    emit('saved')
    close()
  } catch (e) {
    ElMessage.error(t('passwordPool.lockedError'))
    console.error('Save password account failed:', e)
  }
}

function close() {
  emit('update:visible', false)
}
</script>