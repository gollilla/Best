import { readFileSync, existsSync } from 'fs';
import { resolve, dirname } from 'path';
import { pathToFileURL } from 'url';
import type { ScenarioConfig } from '../types/scenario.js';

export interface BestConfig {
  host: string;
  port?: number;
  offline?: boolean;
  timeout?: number;
  testMatch?: string[];
  setupFiles?: string[];
  retries?: number;
  bail?: boolean;
  parallel?: boolean;
  /** シナリオ設定 */
  scenario?: ScenarioConfig;
}

const CONFIG_FILES = [
  'best.config.ts',
  'best.config.js',
  'best.config.mjs',
  'best.config.json',
];

const DEFAULT_CONFIG: BestConfig = {
  host: 'localhost',
  port: 19132,
  offline: true,
  timeout: 30000,
  testMatch: ['**/*.test.ts', '**/*.spec.ts'],
  retries: 0,
  bail: false,
  parallel: false,
};

export async function loadConfig(cwd: string = process.cwd()): Promise<BestConfig> {
  for (const filename of CONFIG_FILES) {
    const filepath = resolve(cwd, filename);

    if (!existsSync(filepath)) continue;

    if (filename.endsWith('.json')) {
      const content = readFileSync(filepath, 'utf-8');
      const config = JSON.parse(content);
      return { ...DEFAULT_CONFIG, ...config };
    }

    // JS/TS config
    const fileUrl = pathToFileURL(filepath).href;
    const module = await import(fileUrl);
    const config = module.default ?? module;
    return { ...DEFAULT_CONFIG, ...config };
  }

  return DEFAULT_CONFIG;
}

export function defineConfig(config: BestConfig): BestConfig {
  return { ...DEFAULT_CONFIG, ...config };
}
