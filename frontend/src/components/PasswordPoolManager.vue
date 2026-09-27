<template>
  <el-dialog append-to-body
    :model-value="visible"
    @update:model-value="(v: boolean) => !v && close()"
    :title="t('passwordPool.manage')"
    width="34rem"
    @opened="reload"
  >
    <PasswordPoolList ref="listRef" @add="openAdd" @edit="openEdit" />
    <PasswordAddEditDialog :visible="addVisible" :account="editing" @update:visible="(v: boolean) => (addVisible = v)" @saved="reload" />
  </el-dialog>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from '../i18n'
import PasswordPoolList from './PasswordPoolList.vue'
import PasswordAddEditDialog from './PasswordAddEditDialog.vue'
import type { PasswordAccount } from '../types/passwordPool'

defineProps<{ visible: boolean }>()
const emit = defineEmits<{ (e: 'update:visible', v: boolean): void }>()

const { t } = useI18n()

const listRef = ref<InstanceType<typeof PasswordPoolList> | null>(null)
const addVisible = ref(false)
const editing = ref<PasswordAccount | null>(null)

async function reload() {
  listRef.value?.load()
}
function openAdd() {
  editing.value = null
  addVisible.value = true
}
function openEdit(a: PasswordAccount) {
  editing.value = a
  addVisible.value = true
}
function close() {
  emit('update:visible', false)
}
</script>