-- tenants table
CREATE TABLE IF NOT EXISTS tenants (
    id SERIAL PRIMARY KEY NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- entity definition table
CREATE TABLE IF NOT EXISTS entity_types (
    id SERIAL PRIMARY KEY NOT NULL,
    tenant_id INTEGER NOT NULL REFERENCES tenants (id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    schema JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- generic assets table
CREATE TABLE IF NOT EXISTS assets (
    id SERIAL PRIMARY KEY NOT NULL,
    tenant_id INTEGER NOT NULL REFERENCES tenants (id) ON DELETE CASCADE,
    entity_type_id INT NOT NULL REFERENCES entity_types (id) ON DELETE CASCADE,
    metadata JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- performance indexes
CREATE INDEX idx_assets_metadata ON assets USING GIN (metadata);
CREATE INDEX idx_assets_tenant ON assets (tenant_id);
CREATE INDEX idx_assets_entity_type ON assets (entity_type_id);

-- activate RLS
ALTER TABLE assets ENABLE ROW LEVEL SECURITY;
ALTER TABLE assets FORCE ROW LEVEL SECURITY;

 -- tenant isolation policies
 CREATE POLICY select_assets_by_tenant
   ON assets
   USING (tenant_id = current_setting('app.current_tenant')::INTEGER);

CREATE POLICY insert_assets_by_tenant
  ON assets
  FOR INSERT
  WITH CHECK (tenant_id = current_setting('app.current_tenant')::INTEGER);

CREATE POLICY update_assets_by_tenant
  ON assets
  FOR UPDATE
  WITH CHECK (tenant_id = current_setting('app.current_tenant')::INTEGER);
