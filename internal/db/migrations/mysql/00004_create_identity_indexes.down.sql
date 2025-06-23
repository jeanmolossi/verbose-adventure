ALTER TABLE `identity_providers`
  DROP INDEX `idx_identity_providers_tenant_enabled`,
  DROP INDEX `idx_identity_providers_tenant_type`;
