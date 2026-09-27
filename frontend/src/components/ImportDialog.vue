<template>
  <el-dialog append-to-body
    :model-value="visible"
    @update:model-value="(v: boolean) => !v && onCancel()"
    :title="t('importExport.importTitle')"
    width="28.75rem"
  >
    <el-form label-width="6.25rem">
      <el-form-item :label="t('importExport.importType')">
        <el-select v-model="format" style="width:100%">
          <el-option label="uniTerm (.utm)" value="uniterm" />
          <el-option label="MobaXterm (.mxtsessions)" value="mobaxterm" />
          <el-option label="OpenSSH config (~/.ssh/config)" value="openssh" />
          <el-option label="SecureCRT (.xml)" value="securecrt" />
          <el-option label="WindTerm (.sessions)" value="windterm" />
          <el-option label="Xshell (.xts)" value="xshell" />
          <el-option label="DBeaver (workspace)" value="dbeaver" />
          <el-option label="Navicat (.ncx)" value="navicat" />
        </el-select>
      </el-form-item>
      <el-form-item v-if="format !== 'openssh' && format !== 'dbeaver'" :label="t('importExport.file')">
        <div style="display:flex;gap:0.5rem;width:100%">
          <el-input v-model="srcPath" readonly :placeholder="t('importExport.chooseFile')" style="flex:1" />
          <el-button @click="pickFile">{{ t('importExport.chooseFile') }}</el-button>
        </div>
      </el-form-item>
      <el-form-item v-if="format === 'dbeaver'" :label="t('importExport.file')">
        <div style="display:flex;gap:0.5rem;width:100%">
          <el-input v-model="srcPath" readonly :placeholder="t('importExport.dbeaverPathHint')" style="flex:1" />
          <el-button @click="pickDBeaverDir">{{ t('importExport.chooseFile') }}</el-button>
        </div>
      </el-form-item>
      <el-form-item v-if="format === 'uniterm' || format === 'windterm'" :label="t('importExport.importPassword')">
        <el-input v-model="password" type="password" show-password />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="onCancel">{{ t('common.cancel') }}</el-button>
      <el-button type="primary" :loading="importing" :disabled="!srcPath && format !== 'openssh' && format !== 'dbeaver'" @click="onImport">{{ t('importExport.import') }}</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useI18n } from '../i18n'
import { msg } from '../services/message'
import { ParseImportFile, ApplyImport, SyncConfigureRepo, OpenFileDialogFiltered, OpenDirectoryDialog } from '../../bindings/github.com/ys-ll/uniterm/app'
import { ElMessageBox } from 'element-plus'

const props = defineProps<{ visible: boolean }>()
const emit = defineEmits<{ (e: 'update:visible', v: boolean): void; (e: 'done', count: number): void }>()

const { t } = useI18n()
const format = ref('uniterm')
const srcPath = ref('')
const password = ref('')
const importing = ref(false)

const FILTERS: Record<string, { display: string; pattern: string }> = {
  uniterm: { display: 'uniTerm (*.utm)', pattern: '*.utm' },
  xshell: { display: 'Xshell (*.xts)', pattern: '*.xts' },
  mobaxterm: { display: 'MobaXterm (*.mxtsessions)', pattern: '*.mxtsessions' },
  windterm: { display: 'WindTerm (*.sessions)', pattern: '*.sessions' },
  securecrt: { display: 'SecureCRT (*.xml)', pattern: '*.xml' },
  navicat: { display: 'Navicat (*.ncx)', pattern: '*.ncx' },
}

watch(() => props.visible, (v) => {
  if (v) { format.value = 'uniterm'; srcPath.value = ''; password.value = '' }
})

function onCancel() { emit('update:visible', false) }

async function pickFile() {
  const f = FILTERS[format.value]
  if (!f) return
  const p = await OpenFileDialogFiltered(t('importExport.importType'), f.display, f.pattern)
  if (p) srcPath.value = p
}

async function pickDBeaverDir() {
  const p = await OpenDirectoryDialog()
  if (p) srcPath.value = p
}

async function onImport() {
  importing.value = true
  try {
    // DBeaver with an empty path means "auto-detect the default workspace".
    const path = (format.value === 'dbeaver' && !srcPath.value) ? '' : srcPath.value
    const result = await ParseImportFile(format.value, path,
      (format.value === 'uniterm' || format.value === 'windterm') ? password.value : '')
    await ApplyImport({ groups: result.groups || [], connections: result.connections || [] } as any)
    const count = (result.connections || []).length

    // Restore the embedded cloud-sync repo config, if the export carried one.
    const syncCfg = (result as any).sync
    if (syncCfg && syncCfg.repoUrl) {
      if (await applyEmbeddedSync(syncCfg)) msg.success(t('importExport.restoreSyncDone'))
    }

    if (result.warnings && result.warnings.length) {
      msg.warning(result.warnings.join('\n'))
    }
    msg.success(t('importExport.importedCount', { count }))
    emit('done', count)
    emit('update:visible', false)
  } catch (e: any) {
    msg.error(e?.message || String(e))
  } finally {
    importing.value = false
  }
}

// If the exported file carried a cloud-sync repo source, initialize sync here:
// ConfigureRepo derives the encryption key from the master password and stores
// it in the OS keychain, which this machine needs to decrypt the synced files.
async function applyEmbeddedSync(cfg: any): Promise<boolean> {
  let masterPassword: string
  try {
    const r = await ElMessageBox.prompt(
      t('importExport.restoreSyncPasswordHint'),
      t('importExport.restoreSyncTitle'),
      {
        inputType: 'password',
        confirmButtonText: t('common.confirm'),
        cancelButtonText: t('common.cancel'),
        inputValidator: (v) => (v && v.length > 0) || t('importExport.restoreSyncPasswordRequired'),
      }
    )
    masterPassword = r.value || ''
  } catch {
    return false // user cancelled — keep the connections import only
  }
  try {
    await SyncConfigureRepo(cfg.repoUrl, cfg.username || '', '', masterPassword)
    return true
  } catch (e: any) {
    msg.error(t('importExport.restoreSyncFailed') + (e?.message ? `: ${e.message}` : ''))
    return false
  }
}
</script>
