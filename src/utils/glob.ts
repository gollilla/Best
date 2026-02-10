import { readdir, stat } from 'fs/promises';
import { join, relative } from 'path';

export async function glob(pattern: string, cwd: string): Promise<string[]> {
  const results: string[] = [];

  // Convert glob pattern to regex
  const regexPattern = pattern
    .replace(/\*\*/g, '{{GLOBSTAR}}')
    .replace(/\*/g, '[^/]*')
    .replace(/\{\{GLOBSTAR\}\}/g, '.*')
    .replace(/\./g, '\\.')
    .replace(/\?/g, '.');

  const regex = new RegExp(`^${regexPattern}$`);

  async function walk(dir: string): Promise<void> {
    const entries = await readdir(dir, { withFileTypes: true });

    for (const entry of entries) {
      const fullPath = join(dir, entry.name);
      const relativePath = relative(cwd, fullPath);

      // Skip node_modules and hidden directories
      if (entry.name.startsWith('.') || entry.name === 'node_modules' || entry.name === 'dist') {
        continue;
      }

      if (entry.isDirectory()) {
        await walk(fullPath);
      } else if (entry.isFile() && regex.test(relativePath)) {
        results.push(relativePath);
      }
    }
  }

  await walk(cwd);
  return results.sort();
}
