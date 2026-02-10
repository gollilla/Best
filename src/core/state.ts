import type { Position, Rotation, PlayerState } from '../types/index.js';

export function createInitialState(): PlayerState {
  return {
    position: { x: 0, y: 0, z: 0 },
    rotation: { pitch: 0, yaw: 0 },
    health: 20,
    gamemode: 0,
    dimension: 'overworld',
    isOnGround: false,
    runtimeEntityId: 0n,
  };
}

export function distanceTo(from: Position, to: Position): number {
  const dx = to.x - from.x;
  const dy = to.y - from.y;
  const dz = to.z - from.z;
  return Math.sqrt(dx * dx + dy * dy + dz * dz);
}

export function horizontalDistanceTo(from: Position, to: Position): number {
  const dx = to.x - from.x;
  const dz = to.z - from.z;
  return Math.sqrt(dx * dx + dz * dz);
}
