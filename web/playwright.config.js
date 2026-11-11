import { defineConfig } from '@playwright/test'

const port = Number(process.env.COUNTDOWN_TEST_PORT || 18731)
// Local readiness checks must bypass any proxy configured in the user's shell.
const bypass = [process.env.NO_PROXY, process.env.no_proxy, 'localhost', '127.0.0.1'].filter(Boolean).join(',')
process.env.NO_PROXY = bypass
process.env.no_proxy = bypass

export default defineConfig({
  testDir: './e2e',
  fullyParallel: false,
  workers: 1,
  timeout: 120000,
  globalTimeout: 300000,
  use: {
    baseURL: `http://127.0.0.1:${port}`,
    launchOptions: process.env.PLAYWRIGHT_CHROMIUM_EXECUTABLE
      ? { executablePath: process.env.PLAYWRIGHT_CHROMIUM_EXECUTABLE }
      : {},
    viewport: { width: 1440, height: 1100 },
    trace: 'off',
    screenshot: 'only-on-failure',
  },
  webServer: {
    command: `../bin/countdown-studio -addr 127.0.0.1:${port} -db ../data/e2e-${process.pid}.db`,
    url: `http://127.0.0.1:${port}/api/health`,
    reuseExistingServer: false,
    gracefulShutdown: { signal: 'SIGTERM', timeout: 5000 },
    timeout: 30000,
  },
})
