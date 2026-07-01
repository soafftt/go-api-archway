import { Pool } from 'pg';

export function createPostgresPool(connectionString: string) {
  return new Pool({
    connectionString,
    max: 10,
  });
}
