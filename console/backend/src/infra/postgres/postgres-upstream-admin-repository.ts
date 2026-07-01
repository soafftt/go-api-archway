import type { Pool, PoolClient, QueryResultRow } from 'pg';
import type { UpstreamAdminRepository } from '../../application/ports.js';
import type { RouteChangeType, UpstreamResource, UpstreamServiceDocument, UpstreamServiceSummary } from '../../domain/upstream.js';

type ServiceRow = QueryResultRow & {
  id: number;
  service_name: string;
  auth_algorithm: string | null;
  auth_key_data: string | null;
  auth_user_key: string | null;
  updated_at: Date;
};

type ResourceRow = QueryResultRow & {
  id: number;
  sub_domain: string;
  host: string;
  sort_order: number;
};

type PathRow = QueryResultRow & {
  resource_id: number;
  path: string;
  method: string;
  request_timeout: number;
  response_timeout: number;
  check_authorization: boolean;
  cache_timeout: number;
  sort_order: number;
};

export class PostgresUpstreamAdminRepository implements UpstreamAdminRepository {
  constructor(private readonly pool: Pool) {}

  async list(): Promise<UpstreamServiceSummary[]> {
    const result = await this.pool.query(
      `
        SELECT
          s.service_name,
          s.updated_at,
          COUNT(DISTINCT r.id) AS resource_count,
          COUNT(p.id) AS path_count
        FROM upstream_services s
        LEFT JOIN upstream_resources r ON r.service_id = s.id
        LEFT JOIN upstream_paths p ON p.resource_id = r.id
        GROUP BY s.id
        ORDER BY s.service_name ASC
      `,
    );

    return result.rows.map((row) => ({
      serviceName: String(row.service_name),
      resourceCount: Number(row.resource_count ?? 0),
      pathCount: Number(row.path_count ?? 0),
      updatedAt: row.updated_at ? new Date(row.updated_at).toISOString() : undefined,
    }));
  }

  async listDocuments(): Promise<UpstreamServiceDocument[]> {
    const result = await this.pool.query<{ service_name: string }>(
      `
        SELECT service_name
        FROM upstream_services
        ORDER BY service_name ASC
      `,
    );

    const documents = await Promise.all(result.rows.map((row) => this.get(row.service_name)));
    return documents.filter((document): document is UpstreamServiceDocument => document !== null);
  }

  async get(serviceName: string): Promise<UpstreamServiceDocument | null> {
    const result = await this.pool.query<ServiceRow>(
      `
        SELECT id, service_name, auth_algorithm, auth_key_data, auth_user_key, updated_at
        FROM upstream_services
        WHERE service_name = $1
      `,
      [serviceName],
    );

    if (result.rows.length === 0) {
      return null;
    }

    return this.readAggregate(result.rows[0]);
  }

  async create(service: UpstreamServiceDocument, snapshotJson: string): Promise<void> {
    await this.withTransaction(async (client) => {
      const result = await client.query<{ id: number }>(
        `
          INSERT INTO upstream_services (
            service_name,
            auth_algorithm,
            auth_key_data,
            auth_user_key
          ) VALUES ($1, $2, $3, $4)
          RETURNING id
        `,
        [
          service.serviceName,
          service.authorization?.algorithm ?? null,
          service.authorization?.keyData ?? null,
          service.authorization?.userKey ?? null,
        ],
      );

      await this.insertResources(client, result.rows[0].id, service.resources);
      await this.insertOutbox(client, service.serviceName, 'ROUTE_MESSAGE_ADD', snapshotJson);
    });
  }

  async update(serviceName: string, service: UpstreamServiceDocument, snapshotJson: string): Promise<boolean> {
    return this.withTransaction(async (client) => {
      const existing = await client.query<{ id: number }>(
        `SELECT id FROM upstream_services WHERE service_name = $1 FOR UPDATE`,
        [serviceName],
      );

      if (existing.rows.length === 0) {
        return false;
      }

      const serviceId = existing.rows[0].id;
      await client.query(
        `
          UPDATE upstream_services
          SET auth_algorithm = $1, auth_key_data = $2, auth_user_key = $3
          WHERE id = $4
        `,
        [
          service.authorization?.algorithm ?? null,
          service.authorization?.keyData ?? null,
          service.authorization?.userKey ?? null,
          serviceId,
        ],
      );

      await client.query(`DELETE FROM upstream_resources WHERE service_id = $1`, [serviceId]);
      await this.insertResources(client, serviceId, service.resources);
      await this.insertOutbox(client, service.serviceName, 'ROUTE_MESSAGE_UPDATE', snapshotJson);
      return true;
    });
  }

  async delete(serviceName: string): Promise<boolean> {
    return this.withTransaction(async (client) => {
      const existing = await client.query<{ id: number }>(
        `SELECT id FROM upstream_services WHERE service_name = $1 FOR UPDATE`,
        [serviceName],
      );

      if (existing.rows.length === 0) {
        return false;
      }

      await client.query(`DELETE FROM upstream_services WHERE id = $1`, [existing.rows[0].id]);
      await this.insertOutbox(client, serviceName, 'ROUTE_MESSAGE_DELETE', null);
      return true;
    });
  }

  async republish(serviceName: string, snapshotJson: string): Promise<boolean> {
    return this.withTransaction(async (client) => {
      const existing = await client.query<{ id: number }>(
        `SELECT id FROM upstream_services WHERE service_name = $1 FOR UPDATE`,
        [serviceName],
      );

      if (existing.rows.length === 0) {
        return false;
      }

      await this.insertOutbox(client, serviceName, 'ROUTE_MESSAGE_UPDATE', snapshotJson);
      return true;
    });
  }

  private async readAggregate(serviceRow: ServiceRow): Promise<UpstreamServiceDocument> {
    const resourceResult = await this.pool.query<ResourceRow>(
      `
        SELECT id, sub_domain, host, sort_order
        FROM upstream_resources
        WHERE service_id = $1
        ORDER BY sort_order ASC, id ASC
      `,
      [serviceRow.id],
    );

    if (resourceResult.rows.length === 0) {
      return {
        serviceName: serviceRow.service_name,
        authorization: serviceRow.auth_algorithm && serviceRow.auth_key_data && serviceRow.auth_user_key
          ? {
              algorithm: serviceRow.auth_algorithm as NonNullable<UpstreamServiceDocument['authorization']>['algorithm'],
              keyData: serviceRow.auth_key_data,
              userKey: serviceRow.auth_user_key,
            }
          : undefined,
        resources: [],
      };
    }

    const pathResult = await this.pool.query<PathRow>(
      `
        SELECT resource_id, path, method, request_timeout, response_timeout, check_authorization, cache_timeout, sort_order
        FROM upstream_paths
        WHERE resource_id = ANY($1::bigint[])
        ORDER BY resource_id ASC, sort_order ASC, id ASC
      `,
      [resourceResult.rows.map((row) => row.id)],
    );

    const resources = resourceResult.rows.map<UpstreamResource>((resourceRow) => ({
      subDomain: resourceRow.sub_domain,
      host: resourceRow.host,
      paths: pathResult.rows
        .filter((pathRow) => pathRow.resource_id === resourceRow.id)
        .map((pathRow) => ({
          path: pathRow.path,
          method: pathRow.method as UpstreamResource['paths'][number]['method'],
          requestTimeout: Number(pathRow.request_timeout),
          responseTimeout: Number(pathRow.response_timeout),
          checkAuthorization: Boolean(pathRow.check_authorization),
          cacheTimeout: Number(pathRow.cache_timeout),
        })),
    }));

    return {
      serviceName: serviceRow.service_name,
      authorization: serviceRow.auth_algorithm && serviceRow.auth_key_data && serviceRow.auth_user_key
        ? {
            algorithm: serviceRow.auth_algorithm as NonNullable<UpstreamServiceDocument['authorization']>['algorithm'],
            keyData: serviceRow.auth_key_data,
            userKey: serviceRow.auth_user_key,
          }
        : undefined,
      resources,
    };
  }

  private async insertResources(client: PoolClient, serviceId: number, resources: UpstreamResource[]) {
    for (const [resourceIndex, resource] of resources.entries()) {
      const resourceResult = await client.query<{ id: number }>(
        `
          INSERT INTO upstream_resources (
            service_id,
            sub_domain,
            host,
            sort_order
          ) VALUES ($1, $2, $3, $4)
          RETURNING id
        `,
        [serviceId, resource.subDomain, resource.host, resourceIndex],
      );

      for (const [pathIndex, path] of resource.paths.entries()) {
        await client.query(
          `
            INSERT INTO upstream_paths (
              resource_id,
              path,
              method,
              request_timeout,
              response_timeout,
              check_authorization,
              cache_timeout,
              sort_order
            ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
          `,
          [
            resourceResult.rows[0].id,
            path.path,
            path.method,
            path.requestTimeout,
            path.responseTimeout,
            path.checkAuthorization,
            path.cacheTimeout,
            pathIndex,
          ],
        );
      }
    }
  }

  private insertOutbox(
    client: PoolClient,
    serviceName: string,
    eventType: RouteChangeType,
    snapshotJson: string | null,
  ) {
    return client.query(
      `
        INSERT INTO route_change_outbox (
          service_name,
          event_type,
          snapshot_json,
          status,
          attempts
        ) VALUES ($1, $2, $3, 'pending', 0)
      `,
      [serviceName, eventType, snapshotJson],
    );
  }

  private async withTransaction<T>(executor: (client: PoolClient) => Promise<T>): Promise<T> {
    const client = await this.pool.connect();
    try {
      await client.query('BEGIN');
      const result = await executor(client);
      await client.query('COMMIT');
      return result;
    } catch (error) {
      await client.query('ROLLBACK');
      throw error;
    } finally {
      client.release();
    }
  }
}
