// Mirrors the backend store.PasswordAccount model (see
// backend/store/password_pool_store.go). `password` is never populated in
// list responses — the frontend only ever sees it transiently via the
// GetPasswordAccountPassword + InjectPassword calls.
export interface PasswordAccount {
  id: string
  name: string
  username?: string
  password?: string
  notes?: string
  createdAt?: number
}

export interface PasswordPoolData {
  accounts: PasswordAccount[]
}