// Core
export { Agent, createAgent, TaskRunner, tasks } from './core/client.js';
export type { AgentOptions, Task } from './core/client.js';
export { TypedEventEmitter } from './core/events.js';
export { distanceTo, horizontalDistanceTo } from './core/state.js';
export { World, ChunkColumn } from './core/world.js';
export type { Block, BlockPosition } from './core/world.js';

// Agent helpers (auto-connect with config file)
export { createAgent as createAgentFromConfig, createAgentSync } from './agent/index.js';

// Runner
export {
  TestRunner,
  createTestRunner,
  ConsoleReporter,
  type TestRunnerOptions,
  type TestContext,
  type TestFunction,
  type TestResult,
  type SuiteResult,
  type TestCaseResult,
  type Reporter,
} from './runner/index.js';

// Assertions
export {
  AssertionContext,
  AssertionError,
  PositionAssertion,
  ChatAssertion,
  CommandAssertion,
  FormAssertion,
  ModalFormAssertion,
  ActionFormAssertion,
  CustomFormAssertion,
} from './assertions/index.js';

// Forms
export { FormHandler } from './forms/handler.js';

// Scenario
export {
  ScenarioRunner,
  ScenarioContext,
  ScenarioExecutor,
  runScenario,
  runScenarios,
  runScenarioFromMarkdown,
  parseScenarioMarkdown,
  getAllPlayerNames,
} from './scenario/index.js';

// LLM
export {
  LLMProcessor,
  createLLMProvider,
  createMockProvider,
  AnthropicProvider,
  OpenAIProvider,
  MockLLMProvider,
} from './llm/index.js';

// Config
export { loadConfig, defineConfig, type BestConfig } from './config/index.js';

// Globals (for test files)
export {
  describe,
  test,
  it,
  beforeAll,
  afterAll,
  beforeEach,
  afterEach,
  skip,
  only,
} from './globals.js';

// Types
export type {
  ClientOptions,
  Position,
  Rotation,
  PlayerState,
  CommandOutput,
  ServerInfo,
  ClientEvents,
  ChatMessage,
  ChunkPosition,
  BlockUpdate,
  Form,
  ModalForm,
  ActionForm,
  CustomForm,
  ActionFormButton,
  CustomFormElement,
  FormResponse,
} from './types/index.js';

// Scenario Types
export type {
  ScenarioStep,
  ScenarioResult,
  ScenarioStepResult,
  ParsedScenario,
  PlayerDefinition,
  ScenarioConfig,
} from './types/scenario.js';

// LLM Types
export type {
  LLMProvider,
  LLMOptions,
  LLMMessage,
  LLMTool,
  ParsedAction,
  LLMProviderConfig,
} from './types/llm.js';
