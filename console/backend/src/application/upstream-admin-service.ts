import { z } from 'zod';
import {
  previewMatch,
  toGatewaySnapshotJson,
  upstreamResourceSchema,
  upstreamServiceSchema,
  type PreviewMatchResult,
  type UpstreamServiceDocument,
  type UpstreamServiceSummary,
} from '../domain/upstream.js';
import type { UpstreamAdminRepository } from './ports.js';

export class ConflictError extends Error {}
export class NotFoundError extends Error {}

export class UpstreamAdminService {
  constructor(private readonly repository: UpstreamAdminRepository) {}

  list(): Promise<UpstreamServiceSummary[]> {
    return this.repository.list();
  }

  listDocuments(): Promise<UpstreamServiceDocument[]> {
    return this.repository.listDocuments();
  }

  async get(serviceName: string): Promise<UpstreamServiceDocument> {
    const service = await this.repository.get(serviceName);
    if (!service) {
      throw new NotFoundError(`service "${serviceName}" not found`);
    }

    return service;
  }

  async create(input: unknown): Promise<UpstreamServiceDocument> {
    const service = upstreamServiceSchema.parse(input);
    const existing = await this.repository.get(service.serviceName);
    if (existing) {
      throw new ConflictError(`service "${service.serviceName}" already exists`);
    }

    try {
      await this.repository.create(service, toGatewaySnapshotJson(service));
    } catch (error) {
      if (isUniqueViolation(error)) {
        throw new ConflictError(`service "${service.serviceName}" already exists`);
      }
      throw error;
    }
    return service;
  }

  async update(serviceName: string, input: unknown): Promise<UpstreamServiceDocument> {
    const service = upstreamServiceSchema.parse(input);
    if (service.serviceName !== serviceName) {
      throw new ConflictError('serviceName in body must match the route parameter');
    }

    const updated = await this.repository.update(serviceName, service, toGatewaySnapshotJson(service));
    if (!updated) {
      throw new NotFoundError(`service "${serviceName}" not found`);
    }

    return service;
  }

  async upsertResource(serviceName: string, input: unknown): Promise<UpstreamServiceDocument> {
    const payload = upsertResourceSchema.parse(input);
    const current = await this.get(serviceName);
    const normalizedCurrent = upstreamServiceSchema.parse(current);

    const targetDomain = payload.previousDomain ?? payload.resource.domain;
    const targetResourceIndex = normalizedCurrent.resources.findIndex((resource) => resource.domain === targetDomain);
    const nextResources = targetResourceIndex >= 0
      ? normalizedCurrent.resources.map((resource, index) => (index === targetResourceIndex ? payload.resource : resource))
      : [...normalizedCurrent.resources, payload.resource];

    const nextService = upstreamServiceSchema.parse({
      serviceName: normalizedCurrent.serviceName,
      authorization: normalizedCurrent.authorization,
      resources: nextResources,
    });

    if (JSON.stringify(normalizedCurrent) === JSON.stringify(nextService)) {
      return current;
    }

    const updated = await this.repository.update(serviceName, nextService, toGatewaySnapshotJson(nextService));
    if (!updated) {
      throw new NotFoundError(`service "${serviceName}" not found`);
    }
    return nextService;
  }

  async delete(serviceName: string): Promise<void> {
    const deleted = await this.repository.delete(serviceName);
    if (!deleted) {
      throw new NotFoundError(`service "${serviceName}" not found`);
    }
  }

  async republish(serviceName: string): Promise<void> {
    const service = await this.repository.get(serviceName);
    if (!service) {
      throw new NotFoundError(`service "${serviceName}" not found`);
    }

    const requeued = await this.repository.republish(serviceName, toGatewaySnapshotJson(service));
    if (!requeued) {
      throw new NotFoundError(`service "${serviceName}" not found`);
    }
  }

  async preview(gatewayPath: string, method: string): Promise<PreviewMatchResult> {
    const serviceName = new URL(gatewayPath, 'http://localhost').pathname.split('/').filter(Boolean)[1];
    if (!serviceName) {
      throw new NotFoundError('service name could not be resolved from gateway path');
    }

    const service = await this.get(serviceName);
    return previewMatch(service, gatewayPath, method);
  }
}

const upsertResourceSchema = z.object({
  resource: upstreamResourceSchema,
  previousDomain: z.string().optional(),
});

function isUniqueViolation(error: unknown): error is { code: string } {
  return typeof error === 'object' && error !== null && 'code' in error && (error as { code?: string }).code === '23505';
}
