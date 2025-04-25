package repo

import (
	"context"
	"database/sql"
	"time"
)

// IdentityProviderRecord representa a linha da tabela identity_providers.
type IdentityProviderRecord struct {
	ID              int64     `db:"id"`
	TenantID        int64     `db:"tenant_id"`
	ProviderType    string    `db:"type"`
	MetadataURL     string    `db:"metadata_url"`
	ClientID        string    `db:"client_id"`
	ClientSecretEnc []byte    `db:"client_secret_enc"`
	Enabled         bool      `db:"enabled"`
	CreatedAt       time.Time `db:"created_at"`
	UpdatedAt       time.Time `db:"updated_at"`
}

// IdentityProviderRepository define os métodos para acesso e manipulação de identity providers.
type IdentityProviderRepository interface {
	// ListByTenant retorna todos os providers de um tenant
	ListByTenant(ctx context.Context, tenantID int64) ([]*IdentityProviderRecord, error)
	// GetByID retorna um provider específico
	GetByID(ctx context.Context, id int64) (*IdentityProviderRecord, error)
	// Create insere um novo provider e retorna o ID gerado
	Create(ctx context.Context, rec *IdentityProviderRecord) (int64, error)
	// Update modifica um provider existente
	Update(ctx context.Context, rec *IdentityProviderRecord) error
	// Delete remove um provider pelo ID
	Delete(ctx context.Context, id int64) error
}

// identityProviderRepo é a implementação concreta
type identityProviderRepo struct {
	db *sql.DB
}

// NewIdentityProviderRepository instancia um IdentityProviderRepository
func NewIdentityProviderRepository(db *sql.DB) IdentityProviderRepository {
	return &identityProviderRepo{db: db}
}

func (r *identityProviderRepo) ListByTenant(ctx context.Context, tenantID int64) ([]*IdentityProviderRecord, error) {
	query := `
        SELECT id, tenant_id, type, metadata_url, client_id, client_secret_enc, enabled, created_at, updated_at
	    FROM identity_providers
	    WHERE tenant_id = ?
    `
	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*IdentityProviderRecord
	for rows.Next() {
		rec := new(IdentityProviderRecord)
		if err := rows.Scan(
			&rec.ID,
			&rec.TenantID,
			&rec.ProviderType,
			&rec.MetadataURL,
			&rec.ClientID,
			&rec.ClientSecretEnc,
			&rec.Enabled,
			&rec.CreatedAt,
			&rec.UpdatedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, rec)
	}
	return list, rows.Err()
}

func (r *identityProviderRepo) GetByID(ctx context.Context, id int64) (*IdentityProviderRecord, error) {
	query := `
        SELECT id, tenant_id, type, metadata_url, client_id, client_secret_enc, enabled, created_at, updated_at
	    FROM identity_providers
	    WHERE id = ?
    `
	row := r.db.QueryRowContext(ctx, query, id)
	rec := &IdentityProviderRecord{}
	if err := row.Scan(
		&rec.ID,
		&rec.TenantID,
		&rec.ProviderType,
		&rec.MetadataURL,
		&rec.ClientID,
		&rec.ClientSecretEnc,
		&rec.Enabled,
		&rec.CreatedAt,
		&rec.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return rec, nil
}

func (r *identityProviderRepo) Create(ctx context.Context, rec *IdentityProviderRecord) (int64, error) {
	query := `
        INSERT INTO identity_providers
	        (tenant_id, type, metadata_url, client_id, client_secret_enc, enabled)
	    VALUES (?, ?, ?, ?, ?, ?)
    `
	res, err := r.db.ExecContext(ctx, query,
		rec.TenantID,
		rec.ProviderType,
		rec.MetadataURL,
		rec.ClientID,
		rec.ClientSecretEnc,
		rec.Enabled,
	)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *identityProviderRepo) Update(ctx context.Context, rec *IdentityProviderRecord) error {
	query := `
        UPDATE identity_providers
	    SET tenant_id = ?, type = ?, metadata_url = ?, client_id = ?, client_secret_enc = ?, enabled = ?, updated_at = CURRENT_TIMESTAMP
	    WHERE id = ?
    `
	_, err := r.db.ExecContext(ctx, query,
		rec.TenantID,
		rec.ProviderType,
		rec.MetadataURL,
		rec.ClientID,
		rec.ClientSecretEnc,
		rec.Enabled,
		rec.ID,
	)
	return err
}

func (r *identityProviderRepo) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM identity_providers WHERE id = ?`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return sql.ErrNoRows
	}
	return nil
}
