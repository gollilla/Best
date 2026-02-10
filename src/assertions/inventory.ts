import type { Agent } from '../core/client.js';
import type { InventoryItem } from '../types/index.js';
import { AssertionError } from './index.js';

export class InventoryAssertion {
  constructor(private player: Agent) {}

  /**
   * 指定アイテムを持っているか確認
   */
  toHaveItem(itemId: string, options?: { slot?: number }): this {
    const items = this.player.getInventory();
    const normalizedId = itemId.startsWith('minecraft:') ? itemId : `minecraft:${itemId}`;

    const found = options?.slot !== undefined
      ? items.find((item) => item.slot === options.slot && item.id === normalizedId)
      : items.find((item) => item.id === normalizedId);

    if (!found) {
      throw new AssertionError(
        `Expected inventory to contain "${normalizedId}"${options?.slot !== undefined ? ` in slot ${options.slot}` : ''}`,
        normalizedId,
        items.map((i) => i.id)
      );
    }
    return this;
  }

  /**
   * 指定アイテムを指定個数以上持っているか確認
   */
  toHaveItemCount(itemId: string, minCount: number): this {
    const items = this.player.getInventory();
    const normalizedId = itemId.startsWith('minecraft:') ? itemId : `minecraft:${itemId}`;

    const totalCount = items
      .filter((item) => item.id === normalizedId)
      .reduce((sum, item) => sum + item.count, 0);

    if (totalCount < minCount) {
      throw new AssertionError(
        `Expected at least ${minCount} of "${normalizedId}", but found ${totalCount}`,
        minCount,
        totalCount
      );
    }
    return this;
  }

  /**
   * 指定アイテムを持っていないことを確認
   */
  notToHaveItem(itemId: string): this {
    const items = this.player.getInventory();
    const normalizedId = itemId.startsWith('minecraft:') ? itemId : `minecraft:${itemId}`;

    const found = items.find((item) => item.id === normalizedId);

    if (found) {
      throw new AssertionError(
        `Expected inventory not to contain "${normalizedId}", but found ${found.count} in slot ${found.slot}`,
        undefined,
        found
      );
    }
    return this;
  }

  /**
   * インベントリが空かどうか確認
   */
  toBeEmpty(): this {
    const items = this.player.getInventory();

    if (items.length > 0) {
      throw new AssertionError(
        `Expected inventory to be empty, but found ${items.length} items`,
        0,
        items.length
      );
    }
    return this;
  }

  /**
   * インベントリが空でないことを確認
   */
  notToBeEmpty(): this {
    const items = this.player.getInventory();

    if (items.length === 0) {
      throw new AssertionError(
        'Expected inventory not to be empty',
        'items',
        'empty'
      );
    }
    return this;
  }

  /**
   * 指定スロットにアイテムがあるか確認
   */
  toHaveItemInSlot(slot: number): this {
    const items = this.player.getInventory();
    const found = items.find((item) => item.slot === slot);

    if (!found) {
      throw new AssertionError(
        `Expected item in slot ${slot}, but slot is empty`,
        'item',
        undefined
      );
    }
    return this;
  }

  /**
   * 指定エンチャントを持つアイテムがあるか確認
   */
  toHaveEnchantedItem(itemId: string, enchantmentId: string, minLevel?: number): this {
    const items = this.player.getInventory();
    const normalizedItemId = itemId.startsWith('minecraft:') ? itemId : `minecraft:${itemId}`;
    const normalizedEnchId = enchantmentId.startsWith('minecraft:') ? enchantmentId : `minecraft:${enchantmentId}`;

    const item = items.find((i) => i.id === normalizedItemId);
    if (!item) {
      throw new AssertionError(
        `Expected inventory to contain "${normalizedItemId}"`,
        normalizedItemId,
        undefined
      );
    }

    const enchantment = item.enchantments?.find((e) => e.id === normalizedEnchId);
    if (!enchantment) {
      throw new AssertionError(
        `Expected "${normalizedItemId}" to have enchantment "${normalizedEnchId}"`,
        normalizedEnchId,
        item.enchantments?.map((e) => e.id)
      );
    }

    if (minLevel !== undefined && enchantment.level < minLevel) {
      throw new AssertionError(
        `Expected enchantment "${normalizedEnchId}" to be at least level ${minLevel}, but was level ${enchantment.level}`,
        minLevel,
        enchantment.level
      );
    }

    return this;
  }

  /**
   * インベントリ変更を待機
   */
  async toReceiveItem(
    itemId: string,
    options?: { timeout?: number }
  ): Promise<InventoryItem> {
    const { timeout = 5000 } = options ?? {};
    const normalizedId = itemId.startsWith('minecraft:') ? itemId : `minecraft:${itemId}`;

    try {
      const [item] = await this.player.waitFor('inventory_update', {
        timeout,
        filter: (update: { item: InventoryItem }) => update.item.id === normalizedId,
      });
      return item.item;
    } catch {
      throw new AssertionError(
        `Timeout waiting for item "${normalizedId}" to be received`,
        normalizedId,
        undefined
      );
    }
  }
}
