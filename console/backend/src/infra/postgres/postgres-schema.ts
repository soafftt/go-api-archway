import { readFile } from 'node:fs/promises';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import type { Pool } from 'pg';

export async function applyPostgresSchema(pool: Pool): Promise<void> {
  const schemaSql = await readSchemaSql();
  await pool.query(schemaSql);
}

async function readSchemaSql(): Promise<string> {
  const currentDir = path.dirname(fileURLToPath(import.meta.url));
  const schemaPath = path.resolve(currentDir, '../../../sql/schema.sql');
  return readFile(schemaPath, 'utf8');
}
