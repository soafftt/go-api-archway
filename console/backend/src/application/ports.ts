import type { RouteChangeRecord, UpstreamServiceDocument, UpstreamServiceSummary } from '../domain/upstream.js';

export interface UpstreamAdminRepository {
  list(): Promise<UpstreamServiceSummary[]>;
  listDocuments(): Promise<UpstreamServiceDocument[]>;
  get(serviceName: string): Promise<UpstreamServiceDocument | null>;
  create(service: UpstreamServiceDocument, snapshotJson: string): Promise<void>;
  update(serviceName: string, service: UpstreamServiceDocument, snapshotJson: string): Promise<boolean>;
  delete(serviceName: string): Promise<boolean>;
  republish(serviceName: string, snapshotJson: string): Promise<boolean>;
}

export interface RouteChangeOutboxRepository {
  listPending(limit: number): Promise<RouteChangeRecord[]>;
  markPublished(id: number): Promise<void>;
  markFailed(id: number, errorMessage: string): Promise<void>;
}
