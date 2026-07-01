CREATE TABLE IF NOT EXISTS upstream_services (
  id BIGSERIAL PRIMARY KEY,
  service_name VARCHAR(255) NOT NULL UNIQUE,
  auth_algorithm VARCHAR(32),
  auth_key_data TEXT,
  auth_user_key VARCHAR(255),
  created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS upstream_resources (
  id BIGSERIAL PRIMARY KEY,
  service_id BIGINT NOT NULL,
  sub_domain VARCHAR(255) NOT NULL DEFAULT '',
  host VARCHAR(255) NOT NULL,
  sort_order INT NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT fk_upstream_resources_service
    FOREIGN KEY (service_id) REFERENCES upstream_services(id) ON DELETE CASCADE,
  CONSTRAINT uq_upstream_resources_service_subdomain UNIQUE (service_id, sub_domain)
);

CREATE TABLE IF NOT EXISTS upstream_paths (
  id BIGSERIAL PRIMARY KEY,
  resource_id BIGINT NOT NULL,
  path VARCHAR(512) NOT NULL,
  method VARCHAR(16) NOT NULL,
  request_timeout BIGINT NOT NULL,
  response_timeout BIGINT NOT NULL,
  check_authorization BOOLEAN NOT NULL DEFAULT FALSE,
  cache_timeout BIGINT NOT NULL DEFAULT 0,
  sort_order INT NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT fk_upstream_paths_resource
    FOREIGN KEY (resource_id) REFERENCES upstream_resources(id) ON DELETE CASCADE,
  CONSTRAINT uq_upstream_paths_resource_method_path UNIQUE (resource_id, method, path)
);

CREATE TABLE IF NOT EXISTS route_change_outbox (
  id BIGSERIAL PRIMARY KEY,
  service_name VARCHAR(255) NOT NULL,
  event_type VARCHAR(64) NOT NULL,
  snapshot_json TEXT,
  status VARCHAR(32) NOT NULL DEFAULT 'pending',
  attempts INT NOT NULL DEFAULT 0,
  last_error TEXT,
  published_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_route_change_outbox_status_id
  ON route_change_outbox (status, id);

CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = CURRENT_TIMESTAMP;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_upstream_services_updated_at ON upstream_services;
CREATE TRIGGER trg_upstream_services_updated_at
BEFORE UPDATE ON upstream_services
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

DROP TRIGGER IF EXISTS trg_upstream_resources_updated_at ON upstream_resources;
CREATE TRIGGER trg_upstream_resources_updated_at
BEFORE UPDATE ON upstream_resources
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

DROP TRIGGER IF EXISTS trg_upstream_paths_updated_at ON upstream_paths;
CREATE TRIGGER trg_upstream_paths_updated_at
BEFORE UPDATE ON upstream_paths
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

DROP TRIGGER IF EXISTS trg_route_change_outbox_updated_at ON route_change_outbox;
CREATE TRIGGER trg_route_change_outbox_updated_at
BEFORE UPDATE ON route_change_outbox
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();
