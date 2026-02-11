import { defineConfig } from './src/index.js';

export default defineConfig({
  host: 'localhost',
  port: 19132,
  offline: true,
  timeout: 30000,
  testMatch: ['tests/**/*.test.ts'],
  version: '1.26.0', // BDS 1.26.x に近いバージョン
});
