import { z } from 'zod';

const envSchema = z.object({
  PORT: z.coerce.number().int().positive().default(8080),
  POSTGRES_CONNECTION_STRING: z
    .string()
    .min(1, 'POSTGRES_CONNECTION_STRING is required')
    .default('postgres://postgres:postgres@127.0.0.1:5431/postgres'),
  VALKEY_URL: z.string().min(1, 'VALKEY_URL is required').default('redis://127.0.0.1:6379'),
  OUTBOX_POLL_INTERVAL_MS: z.coerce.number().int().positive().default(3000),
  OUTBOX_BATCH_SIZE: z.coerce.number().int().positive().default(10),
});

export type AppConfig = z.infer<typeof envSchema>;

export function loadConfig(source: NodeJS.ProcessEnv = process.env): AppConfig {
  return envSchema.parse(source);
}
