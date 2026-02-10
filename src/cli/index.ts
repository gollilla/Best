#!/usr/bin/env node

import { resolve } from 'path';
import { glob } from '../utils/glob.js';
import { loadConfig } from '../config/index.js';
import { TestRunner } from '../runner/index.js';
import { ConsoleReporter } from '../runner/reporter.js';
import { registerGlobals, setCurrentRunner } from '../globals.js';
import { pathToFileURL } from 'url';
import { ScenarioRunner } from '../scenario/index.js';
import { createLLMProvider } from '../llm/index.js';
import type { LLMProvider } from '../types/llm.js';

async function main() {
  const cwd = process.cwd();
  const args = process.argv.slice(2);

  // コマンド判定
  const command = args[0];

  if (command === 'scenario') {
    // シナリオ実行モード
    await runScenarios(cwd, args.slice(1));
  } else {
    // 通常のテスト実行
    await runTests(cwd, args);
  }
}

/**
 * 通常のテストを実行
 */
async function runTests(cwd: string, args: string[]) {
  console.log('Best - Bedrock Edition Server Testing\n');

  // Load config
  const config = await loadConfig(cwd);

  // Create runner
  const runner = new TestRunner({
    timeout: config.timeout,
    retries: config.retries,
    bail: config.bail,
    parallel: config.parallel,
    reporter: new ConsoleReporter(),
  });

  runner.configure({
    host: config.host,
    port: config.port,
    username: 'TestBot',
    offline: config.offline,
    timeout: config.timeout,
  });

  // Register globals
  setCurrentRunner(runner);
  registerGlobals();

  // Find test files
  const patterns = config.testMatch ?? ['**/*.test.ts', '**/*.spec.ts'];
  const testFiles: string[] = [];

  for (const pattern of patterns) {
    const files = await glob(pattern, cwd);
    testFiles.push(...files);
  }

  if (testFiles.length === 0) {
    console.log('No test files found.');
    console.log(`Patterns: ${patterns.join(', ')}`);
    process.exit(0);
  }

  console.log(`Found ${testFiles.length} test file(s)\n`);

  // Run setup files
  if (config.setupFiles) {
    for (const setupFile of config.setupFiles) {
      const filepath = resolve(cwd, setupFile);
      await import(pathToFileURL(filepath).href);
    }
  }

  // Load test files
  for (const file of testFiles) {
    const filepath = resolve(cwd, file);
    await import(pathToFileURL(filepath).href);
  }

  // Run tests
  const result = await runner.run();

  process.exit(result.failed > 0 ? 1 : 0);
}

/**
 * シナリオを実行
 */
async function runScenarios(cwd: string, args: string[]) {
  console.log('Best - Scenario Runner\n');

  // Load config
  const config = await loadConfig(cwd);
  const scenarioConfig = config.scenario ?? {};

  // LLMプロバイダーを作成
  let llmProvider: LLMProvider | undefined;
  if (scenarioConfig.llm) {
    try {
      llmProvider = createLLMProvider(scenarioConfig.llm);
      console.log(`LLM Provider: ${scenarioConfig.llm.provider}`);
    } catch (error) {
      console.warn(`Warning: Failed to create LLM provider: ${error}`);
      console.log('Continuing without LLM (using simple parser)\n');
    }
  } else {
    console.log('LLM Provider: None (using simple parser)\n');
  }

  // シナリオファイルのパターンを取得
  const patterns = args.length > 0
    ? args
    : scenarioConfig.match ?? ['scenarios/**/*.scenario.md'];

  const scenarioFiles: string[] = [];
  for (const pattern of patterns) {
    const files = await glob(pattern, cwd);
    scenarioFiles.push(...files);
  }

  if (scenarioFiles.length === 0) {
    console.log('No scenario files found.');
    console.log(`Patterns: ${patterns.join(', ')}`);
    process.exit(0);
  }

  console.log(`Found ${scenarioFiles.length} scenario file(s)\n`);

  // オプションをパース
  const verbose = args.includes('--verbose') || args.includes('-v');
  const generateSummary = args.includes('--summary') || args.includes('-s');

  // シナリオランナーを作成
  const runner = new ScenarioRunner({
    llmProvider,
    clientOptions: {
      host: config.host,
      port: config.port,
      offline: config.offline,
      timeout: config.timeout,
    },
    stepTimeout: scenarioConfig.stepTimeout ?? 30000,
    totalTimeout: scenarioConfig.totalTimeout ?? 300000,
    verbose,
    generateSummary: generateSummary && !!llmProvider,
  });

  // シナリオファイルを読み込み
  for (const file of scenarioFiles) {
    console.log(`Loading: ${file}`);
    runner.loadFile(file);
  }

  console.log('');

  // シナリオを実行
  const results = await runner.runAll();

  // 結果を表示
  let passed = 0;
  let failed = 0;

  console.log('\n=== Results ===\n');

  for (const result of results) {
    const icon = result.passed ? '✓' : '✗';
    const color = result.passed ? '\x1b[32m' : '\x1b[31m';
    const reset = '\x1b[0m';

    console.log(`${color}${icon}${reset} ${result.name} (${result.duration}ms)`);

    if (result.passed) {
      passed++;
    } else {
      failed++;
      if (result.error) {
        console.log(`  Error: ${result.error.message}`);
      }

      // 失敗したステップを表示
      for (const step of result.steps) {
        if (step.status === 'failed') {
          console.log(`  Failed step: ${step.description}`);
          if (step.error) {
            console.log(`    ${step.error.message}`);
          }
        }
      }
    }

    // 自然言語サマリーを表示
    if (result.summary) {
      console.log(`\n  📝 サマリー:`);
      const summaryLines = result.summary.split('\n');
      for (const line of summaryLines) {
        console.log(`  ${line}`);
      }
      console.log('');
    }
  }

  console.log(`\nPassed: ${passed}, Failed: ${failed}`);

  process.exit(failed > 0 ? 1 : 0);
}

main().catch((err) => {
  console.error('Error:', err);
  process.exit(1);
});
