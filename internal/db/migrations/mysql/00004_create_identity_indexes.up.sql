ALTER TABLE `identity_providers`
  ADD INDEX `idx_identity_providers_tenant_id`(`tenant_id`),
  ADD INDEX `idx_identity_providers_tenant_enabled`(`tenant_id`, `enabled`),
  ADD INDEX `idx_identity_providers_tenant_type`(`tenant_id`, `type`);
