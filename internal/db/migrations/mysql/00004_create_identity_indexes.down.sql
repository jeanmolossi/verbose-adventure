ALTER TABLE `identity_providers`
  DROP INDEX `idx_identity_providers_tenant_id`(`tenant_id`),
  DROP INDEX `idx_identity_providers_tenant_enabled`(`tenant_id`, `enabled`),
  DROP INDEX `idx_identity_providers_tenant_type`(`tenant_id`, `type`);
