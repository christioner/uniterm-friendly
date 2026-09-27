import { defineStore } from 'pinia'
import { ref } from 'vue'
import {
  ListPasswordAccounts, SavePasswordAccount, DeletePasswordAccount,
  GetPasswordAccountPassword, InjectPassword,
} from '../../bindings/github.com/ys-ll/uniterm/app'
import { useCredentialStore } from './credentialStore'
import type { PasswordAccount } from '../types/passwordPool'

// Password pool store. The pool shares the unified credential master key, so
// plaintext secrets are only ever decrypted to a short-lived local variable and
// written straight into a session (InjectPassword) — never held in reactive
// state or persisted on the frontend.
export const usePasswordPoolStore = defineStore('passwordPool', () => {
  const credential = useCredentialStore()
  const accounts = ref<PasswordAccount[]>([])
  const loaded = ref(false)

  async function load(): Promise<void> {
    try {
      const data = await ListPasswordAccounts()
      accounts.value = ((data as { accounts: PasswordAccount[] }).accounts) || []
    } catch (e) {
      console.error('Failed to load password pool:', e)
      accounts.value = []
    } finally {
      loaded.value = true
    }
  }

  async function save(acct: PasswordAccount): Promise<void> {
    await SavePasswordAccount(acct)
    await load()
  }

  async function remove(id: string): Promise<void> {
    await DeletePasswordAccount(id)
    await load()
  }

  // Resolve a pool account's plaintext and inject it into a running session,
  // appending a PTY Enter. The plaintext never touches reactive state.
  async function injectToSession(sessionId: string, accountId: string): Promise<boolean> {
    if (!sessionId || !accountId) return false
    const pwd = await GetPasswordAccountPassword(accountId)
    if (!pwd) return false
    await InjectPassword(sessionId, pwd + '\r')
    return true
  }

  function unlocked(): boolean {
    return !!credential.status?.unlocked
  }

  return { accounts, loaded, load, save, remove, injectToSession, unlocked }
})