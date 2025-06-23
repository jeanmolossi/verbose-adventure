-- tenants table
CREATE TABLE IF NOT EXISTS tenants (
    id SERIAL PRIMARY KEY NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- entity definition table
CREATE TABLE IF NOT EXISTS entity_types (
    id SERIAL PRIMARY KEY NOT NULL,
    tenant_id INTEGER NOT NULL REFERENCES tenants (id),
    name TEXT NOT NULL,
    schema JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- generic assets table
CREATE TABLE IF NOT EXISTS assets (
    id SERIAL PRIMARY KEY NOT NULL,
    tenant_id INTEGER NOT NULL REFERENCES tenants (id),
    entity_type_id INT NOT NULL REFERENCES entity_types (id),
    metadata JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
)
