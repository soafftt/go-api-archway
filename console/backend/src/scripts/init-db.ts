import { readFile } from 'node:fs/promises';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { createPostgresPool } from '../infra/postgres/postgres.js';

async function main() {
  const connectionString = process.env.POSTGRES_CONNECTION_STRING ?? 'postgres://postgres:postgres@127.0.0.1:5431/postgres';

  const pool = createPostgresPool(connectionString);
  const scriptDir = path.dirname(fileURLToPath(import.meta.url));
  const schemaPath = path.resolve(scriptDir, '../../sql/schema.sql');
  const schemaSql = await readFile(schemaPath, 'utf8');

  try {
    await pool.query(schemaSql);
    console.log('console backend schema initialized');
  } finally {
    await pool.end();
  }
}

void main();
