import type { Redis } from 'ioredis';
import type { RouteChangeRecord } from '../../domain/upstream.js';

const APPLY_VERSIONED_PROJECTION_SCRIPT = `
local key = KEYS[1]
local incomingVersion = tonumber(ARGV[1])
local operation = ARGV[2]
local payload = ARGV[3]

local current = redis.call('GET', key)
local currentVersion = 0
if current then
  local ok, decoded = pcall(cjson.decode, current)
  if ok and decoded and decoded["version"] then
    currentVersion = tonumber(decoded["version"]) or 0
  end
end

if incomingVersion < currentVersion then
  return 0
end

if operation == "delete" then
  redis.call('DEL', key)
else
  redis.call('SET', key, payload)
end

return 1
`;

export class ValkeyRouteProjectionRepository {
  constructor(
    private readonly client: Redis,
    private readonly keyPrefix: string,
    private readonly channel: string,
  ) {}

  async applyRouteChange(change: RouteChangeRecord): Promise<void> {
    const key = `${this.keyPrefix}${change.serviceName}`;
    let applied = 0;

    if (change.eventType === 'ROUTE_MESSAGE_DELETE') {
      applied = Number(await this.client.eval(
        APPLY_VERSIONED_PROJECTION_SCRIPT,
        1,
        key,
        change.serviceVersion,
        'delete',
        '',
      ));
    } else {
      if (!change.snapshotJson) {
        throw new Error(`snapshot_json is required for ${change.eventType}`);
      }
      applied = Number(await this.client.eval(
        APPLY_VERSIONED_PROJECTION_SCRIPT,
        1,
        key,
        change.serviceVersion,
        'set',
        change.snapshotJson,
      ));
    }

    if (applied !== 1) {
      return;
    }

    await this.client.publish(
      this.channel,
      JSON.stringify({
        method: change.eventType,
        service: change.serviceName,
      }),
    );
  }

  async setSnapshot(serviceName: string, snapshotJson: string, serviceVersion: number): Promise<void> {
    await this.client.eval(
      APPLY_VERSIONED_PROJECTION_SCRIPT,
      1,
      `${this.keyPrefix}${serviceName}`,
      serviceVersion,
      'set',
      snapshotJson,
    );
  }
}
