import type { Agent } from '../core/client.js';
import type { TitleDisplay, SoundPlay, ParticleSpawn } from '../types/index.js';
import { AssertionError } from './index.js';

// タイトル表示アサーション
export class TitleAssertion {
  constructor(private player: Agent) {}

  /**
   * タイトルの受信を待機
   */
  async toReceive(
    expected: string | RegExp,
    options?: { timeout?: number }
  ): Promise<TitleDisplay> {
    const { timeout = 5000 } = options ?? {};

    try {
      const [title] = await this.player.waitFor('title', {
        timeout,
        filter: (t: TitleDisplay) => {
          if (t.type !== 'title') return false;
          return typeof expected === 'string'
            ? t.text.includes(expected)
            : expected.test(t.text);
        },
      });
      return title;
    } catch {
      throw new AssertionError(
        `Timeout waiting for title matching ${expected}`,
        expected,
        undefined
      );
    }
  }

  /**
   * タイトルを受信しないことを確認
   */
  async notToReceive(
    pattern: string | RegExp,
    options?: { duration?: number }
  ): Promise<void> {
    const { duration = 3000 } = options ?? {};

    return new Promise((resolve, reject) => {
      const handler = (title: TitleDisplay) => {
        if (title.type !== 'title') return;

        const matches =
          typeof pattern === 'string'
            ? title.text.includes(pattern)
            : pattern.test(title.text);

        if (matches) {
          this.player.off('title', handler);
          reject(
            new AssertionError(
              `Expected not to receive title matching ${pattern}, but received: "${title.text}"`,
              undefined,
              title.text
            )
          );
        }
      };

      this.player.on('title', handler);

      setTimeout(() => {
        this.player.off('title', handler);
        resolve();
      }, duration);
    });
  }
}

// サブタイトル表示アサーション
export class SubtitleAssertion {
  constructor(private player: Agent) {}

  /**
   * サブタイトルの受信を待機
   */
  async toReceive(
    expected: string | RegExp,
    options?: { timeout?: number }
  ): Promise<TitleDisplay> {
    const { timeout = 5000 } = options ?? {};

    try {
      const [title] = await this.player.waitFor('title', {
        timeout,
        filter: (t: TitleDisplay) => {
          if (t.type !== 'subtitle') return false;
          return typeof expected === 'string'
            ? t.text.includes(expected)
            : expected.test(t.text);
        },
      });
      return title;
    } catch {
      throw new AssertionError(
        `Timeout waiting for subtitle matching ${expected}`,
        expected,
        undefined
      );
    }
  }
}

// アクションバー表示アサーション
export class ActionbarAssertion {
  constructor(private player: Agent) {}

  /**
   * アクションバーの受信を待機
   */
  async toReceive(
    expected: string | RegExp,
    options?: { timeout?: number }
  ): Promise<TitleDisplay> {
    const { timeout = 5000 } = options ?? {};

    try {
      const [title] = await this.player.waitFor('title', {
        timeout,
        filter: (t: TitleDisplay) => {
          if (t.type !== 'actionbar') return false;
          return typeof expected === 'string'
            ? t.text.includes(expected)
            : expected.test(t.text);
        },
      });
      return title;
    } catch {
      throw new AssertionError(
        `Timeout waiting for actionbar matching ${expected}`,
        expected,
        undefined
      );
    }
  }

  /**
   * アクションバーを受信しないことを確認
   */
  async notToReceive(
    pattern: string | RegExp,
    options?: { duration?: number }
  ): Promise<void> {
    const { duration = 3000 } = options ?? {};

    return new Promise((resolve, reject) => {
      const handler = (title: TitleDisplay) => {
        if (title.type !== 'actionbar') return;

        const matches =
          typeof pattern === 'string'
            ? title.text.includes(pattern)
            : pattern.test(title.text);

        if (matches) {
          this.player.off('title', handler);
          reject(
            new AssertionError(
              `Expected not to receive actionbar matching ${pattern}, but received: "${title.text}"`,
              undefined,
              title.text
            )
          );
        }
      };

      this.player.on('title', handler);

      setTimeout(() => {
        this.player.off('title', handler);
        resolve();
      }, duration);
    });
  }
}

// サウンドアサーション
export class SoundAssertion {
  constructor(private player: Agent) {}

  /**
   * サウンド再生を待機
   */
  async toPlay(
    soundName: string,
    options?: { timeout?: number; nearPlayer?: number }
  ): Promise<SoundPlay> {
    const { timeout = 5000, nearPlayer } = options ?? {};

    try {
      const [sound] = await this.player.waitFor('sound', {
        timeout,
        filter: (s: SoundPlay) => {
          if (s.name !== soundName) return false;
          if (nearPlayer !== undefined) {
            const playerPos = this.player.position;
            const dx = s.position.x - playerPos.x;
            const dy = s.position.y - playerPos.y;
            const dz = s.position.z - playerPos.z;
            return Math.sqrt(dx * dx + dy * dy + dz * dz) <= nearPlayer;
          }
          return true;
        },
      });
      return sound;
    } catch {
      throw new AssertionError(
        `Timeout waiting for sound "${soundName}" to play`,
        soundName,
        undefined
      );
    }
  }

  /**
   * サウンドが再生されないことを確認
   */
  async notToPlay(
    soundName: string,
    options?: { duration?: number }
  ): Promise<void> {
    const { duration = 3000 } = options ?? {};

    return new Promise((resolve, reject) => {
      const handler = (sound: SoundPlay) => {
        if (sound.name === soundName) {
          this.player.off('sound', handler);
          reject(
            new AssertionError(
              `Expected sound "${soundName}" not to play, but it did`,
              undefined,
              soundName
            )
          );
        }
      };

      this.player.on('sound', handler);

      setTimeout(() => {
        this.player.off('sound', handler);
        resolve();
      }, duration);
    });
  }
}

// パーティクルアサーション
export class ParticleAssertion {
  constructor(private player: Agent) {}

  /**
   * パーティクルスポーンを待機
   */
  async toSpawn(
    particleName: string,
    options?: { timeout?: number; nearPlayer?: number }
  ): Promise<ParticleSpawn> {
    const { timeout = 5000, nearPlayer } = options ?? {};

    try {
      const [particle] = await this.player.waitFor('particle', {
        timeout,
        filter: (p: ParticleSpawn) => {
          if (p.name !== particleName) return false;
          if (nearPlayer !== undefined) {
            const playerPos = this.player.position;
            const dx = p.position.x - playerPos.x;
            const dy = p.position.y - playerPos.y;
            const dz = p.position.z - playerPos.z;
            return Math.sqrt(dx * dx + dy * dy + dz * dz) <= nearPlayer;
          }
          return true;
        },
      });
      return particle;
    } catch {
      throw new AssertionError(
        `Timeout waiting for particle "${particleName}" to spawn`,
        particleName,
        undefined
      );
    }
  }

  /**
   * パーティクルがスポーンしないことを確認
   */
  async notToSpawn(
    particleName: string,
    options?: { duration?: number }
  ): Promise<void> {
    const { duration = 3000 } = options ?? {};

    return new Promise((resolve, reject) => {
      const handler = (particle: ParticleSpawn) => {
        if (particle.name === particleName) {
          this.player.off('particle', handler);
          reject(
            new AssertionError(
              `Expected particle "${particleName}" not to spawn, but it did`,
              undefined,
              particleName
            )
          );
        }
      };

      this.player.on('particle', handler);

      setTimeout(() => {
        this.player.off('particle', handler);
        resolve();
      }, duration);
    });
  }
}
