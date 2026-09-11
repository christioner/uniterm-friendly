import { defineConfig } from 'vitest/config'
import vue from '@vitejs/plugin-vue'

const version = process.env.VITE_VERSION || 'dev'

export default defineConfig({
  plugins: [vue()],
  test: {
    include: ['src/**/*.test.ts'],
    environment: 'node',
    coverage: {
      provider: 'v8',
      reporter: ['text', 'json-summary', 'lcov'],
      include: ['src/**/*.ts', 'src/**/*.vue'],
      reportsDirectory: './coverage',
      // Report-only by default: warn on low coverage but never fail the build.
      // Tighten these once the team has landed a baseline. See PROJECT_ANALYSIS.md.
      thresholds: {
        lines: 0,
        functions: 0,
        statements: 0,
        branches: 0
      }
    }
  },
  define: {
    'import.meta.env.VITE_VERSION': JSON.stringify(version)
  },
  // esnext everywhere so the dev server's esbuild transform also accepts
  // top-level await (used by @novnc/novnc). `build.target` alone only affects
  // the production bundle; without this the `wails3 dev` Vite server crashes.
  build: {
    target: 'esnext'
  },
  esbuild: {
    target: 'esnext'
  },
  optimizeDeps: {
    esbuildOptions: {
      target: 'esnext'
    }
  },
  server: {
    // Bind IPv4 loopback explicitly. On some machines/Node versions vite's
    // default "localhost" resolves to ::1 (IPv6) only, while the Wails
    // WebView2 dev proxy dials 127.0.0.1 (IPv4) — causing repeated connection
    // refused and a blank window in `wails3 dev`.
    host: '127.0.0.1',
    port: 34115,
    strictPort: true
  }
})
