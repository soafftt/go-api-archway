import { createPostgresPool } from '../infra/postgres/postgres.js';
import { applyPostgresSchema } from '../infra/postgres/postgres-schema.js';

async function main() {
  const connectionString = process.env.POSTGRES_CONNECTION_STRING ?? 'postgres://postgres:postgres@127.0.0.1:5431/postgres';
  const pool = createPostgresPool(connectionString);

  try {
    await applyPostgresSchema(pool);
    console.log('console backend schema initialized');
  } finally {
    await pool.end();
  }
}

void main();
