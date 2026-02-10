describe('接続テスト', () => {
  beforeEach(async ({ player }) => {
    await player.connect();
  });

  afterEach(async ({ player }) => {
    await player.disconnect();
  });

  test('サーバーに接続できる', async ({ player }) => {
    player.expect.toBeConnected();
  });

  test('チャットを送信できる', async ({ player }) => {
    player.chat('Hello from Best!');
  });
});

describe('コマンドテスト', () => {
  beforeEach(async ({ player }) => {
    await player.connect();
  });

  afterEach(async ({ player }) => {
    await player.disconnect();
  });

  test('sayコマンドが実行できる', async ({ player }) => {
    const result = await player.command('/say Hello World');
    player.expect.command(result).toSucceed();
  });

  skip.test('テレポートが動作する', async ({ player }) => {
    await player.command('/tp @s 0 64 0');
    await player.expect.position.toReach(
      { x: 0, y: 64, z: 0 },
      { timeout: 5000, tolerance: 1 }
    );
  });
});
