-- migrations/mysql/00002_create_identity_providers_table.up.sql
CREATE TABLE IF NOT EXISTS identity_providers (
  id                BIGINT AUTO_INCREMENT PRIMARY KEY,
  tenant_id         BIGINT NOT NULL,
  type              ENUM('oidc', 'saml') NOT NULL,
  metadata_url      VARCHAR(512) NOT NULL,
  client_id         VARCHAR(255) NOT NULL,
  client_secret_enc VARBINARY(512) NOT NULL,
  enabled           TINYINT(1) NOT NULL DEFAULT 1,
  created_at        TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at        TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

  CONSTRAINT fk_identity_providers_tenants FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

