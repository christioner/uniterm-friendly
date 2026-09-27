# 开发参照：账号密码池 + 终端选填密码（ai-agent-password-pool）

> 本文件定义本次功能的 AI Agent 开发团队、使用的 skills、技术栈与架构，作为后续开发与维护的参照（按用户工作规范：开展开发前先形成此文件）。

## 目标
在缓存终端（主要 SSH）遇到密码提示（`sudo` / `passwd` / `ssh`）时，用户可从**加密账号密码池**中选中账号，密码自动填入终端（代替手输）。类似 1Password 的「命令行选填」。

## 已确认的需求决策
- 存储：应用加密存储 + 主密码
- 主密码：**复用 uniTerm 现有「凭据解锁」主密码**（与连接密码解锁同一密钥，`backend/credentials`）
- 触发：自动检测密码提示 + 手动快捷键，两者都要
- 注入：新增后端 `InjectPassword(sessionID, password)` 绑定，明文不经前端 store
- 范围边界：不做完整密码管理器（分组/导入导出/强密码生成）；不做跨设备同步

## AI Agent 开发团队
| 角色 | 职责 |
|---|---|
| 后端 Go Engineer | 密码池 store（`backend/store/password_pool_store.go`）、App 绑定（`app_password_pool.go`）、app.go 接线、bindings 生成 |
| 前端 Vue Engineer | Pinia store（`passwordPoolStore`）、管理/选择对话框组件、BaseTerminal 触发与快捷键、i18n |
| Security Reviewer | 明文不落存储/日志、文件 0600、加密 `enc:v1:` 正确、注入不经中间状态（用 `/security-review`） |

## 使用的 Skills
- `security-review`：实现后整体安全审查
- `frontend-dev-guidelines`：前端组件/页面规范
- `powershell-windows`：Windows 平台验证命令
- `database-design` / 存储决策：凭据加密存储与主密码派生
- `planning-with-files`：多步骤开发的任务/进度跟踪

## 技术栈
- Go 1.26 + Wails3（`wails3 generate bindings` 由后端方法自动生成绑定，`main.go:247` RegisterService 全导出方法绑定）
- Vue 3 + Element Plus + Pinia + TypeScript
- 加密：AES-256-GCM（`backend/credentials`，`enc:v1:` 前缀）+ PBKDF2-SHA256 主密码派生（`backend/sync/keychain.go`）
- 凭据主密钥统一来自 `backend/credentials.Store`（keychain / master-password 双模式）

## 架构摘要（复用为主）
| 能力 | 复用点 |
|---|---|
| 主密码/加密 | `backend/credentials/store.go` `Encrypt/Decrypt/Unlocked/Setup/Unlock`；`crypto.go` `EncryptField/DecryptField` |
| 加密 JSON store 范式 | `backend/store/identity_store.go`（秘密字段 `enc:v1:`，`atomicWriteFile` 0600，`quarantineCorrupt`）；新池 store = `password_pool_store.go` |
| PasswordStore 注入 | `store.PasswordStore` 接口（`backend/store/connection_store.go:20`）；app.go 437-458 段统一 `SetPasswordStore(cred)` |
| 终端注入 | `frontend/src/services/sessionWriter.ts:34 queuedSessionWrite` → `SessionWrite`（`app_terminal.go:516`）→ `SSHSession.Write`；本功能用 `InjectPassword` 后者路径 |
| 密码提示检测 | `frontend/src/composables/useTerminalInput.ts:266 isPasswordMode / :30 isPasswordPrompt`（echo-off 判断） |
| 当前活动会话 | `frontend/src/services/terminalAgent.ts:338 resolveActiveSession`；换行 `:407 getShellNewline` |
| 弹窗/解锁范式 | `frontend/src/components/CredentialPrompt.vue`、`CredentialUnlockDialog.vue`；`credentialStore.ts:12` |

## 数据流
```
终端 onData / 快捷键(省建议 Ctrl+Shift+P)
 → 密码提示检测 (`useTerminalInput.isPasswordPrompt`)
 → 前端 `passwordPoolStore` 校验 `credentialStore` 已解锁；未解锁先走解锁
 → 弹出 `PasswordPicker`（列表= ListPasswordAccounts 元数据，无明文）
 → 选账号 → `GetPasswordAccountPassword(id)`（后端解密，明文一次性）
 → `InjectPassword(sessionID, pwd)`（后端写终端 stdin，明文不再停留）
```

## 交互 v2（按用户反馈改版，即时候选）
初版用弹窗选择器（`PasswordPicker.vue`）选密码，用户认为体验差。现改为**光标处即时候选**——与内置 AI 转写同一套 `useSuggestions`/`TerminalSuggestion` 弹层：
- 终端进入密码输入态（`isPasswordMode`，echo-off）时，光标处自动弹出密码池账号候选；上下键 + Tab/Enter 选择
- 选中项 `type:'password'` 走 `applySuggestion` 注入分支 → `InjectPassword`，**不写回命令行缓冲**
- `Ctrl+Shift+P` 手动唤起；候选末项「管理密码池…」打开 `PasswordPoolDialog`（增删改）
- 删除 `PasswordPicker.vue`（选填不再用弹窗）；移除初版 banner 逻辑
- `SuggestionItem` 新增 `type:'password'` 与 `username` 字段；`TerminalSuggestion.vue` 用钥匙图标渲染名称/用户名

## 待办 / 进度
- [x] 后端 store：`backend/store/password_pool_store.go`
- [x] 后端绑定：`app_password_pool.go`（List/Save/Delete/GetPassword/InjectPassword）
- [x] app.go 接线（struct + NewPasswordPoolStore + SetPasswordStore）
- [x] 后端编译通过（`go build ./...` exit 0）
- [x] `wails3 generate bindings` 重新生成（含 ListPasswordAccounts/InjectPassword 等）
- [x] 前端 store `passwordPoolStore.ts` + 类型 `types/passwordPool.ts`
- [x] 前端 `PasswordPoolDialog.vue`（管理）+ `PasswordPicker.vue`（选填）
- [x] `BaseTerminal.vue` 触发：Ctrl+Shift+P 快捷键 + onData 密码提示 banner + 组件挂载
- [x] i18n（en/zh-CN）
- [x] 构建验证：`vite build` 通过（3925 modules）；新组件类型检查通过
- [ ] 运行回归（wails3 dev：设主密码 → 加账号 → SSH sudo → 选填注入）
- [ ] `/security-review` 安全审查