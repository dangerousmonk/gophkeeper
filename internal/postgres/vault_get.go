package postgres

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/dangerousmonk/gophkeeper/internal/models"
)

// The MetaData struct represents the data in the JSON/JSONB column
type MetaData struct {
	FileName string `json:"file_name,omitempty"`
	FilePath string `json:"file_path,omitempty"`
	FileType string `json:"file_type,omitempty"`
	FileSize uint   `json:"file_size,omitempty"`
}

// Make the MetaData struct implement the driver.Valuer interface
func (a MetaData) Value() (driver.Value, error) {
	return json.Marshal(a)
}

// Make the MetaData struct implement the sql.Scanner interface
func (a *MetaData) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &a)
}

func (r *vaultRepository) GetByUserID(ctx context.Context, userID int) ([]models.Vault, error) {
	const selectFields = "id,user_id,name,data_type,encrypted_data,version,created_at,updated_at,active,meta_data"
	const timeout = time.Second * 2

	var vaults []models.Vault
	var metaDataBytes []byte

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	rows, err := r.db.QueryContext(ctx, `SELECT `+selectFields+` FROM vault WHERE user_id=$1 AND active IS true ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var v models.Vault
		if err = rows.Scan(
			&v.ID,
			&v.UserID,
			&v.Name,
			&v.DataType,
			&v.EncryptedData,
			&v.Version,
			&v.CreatedAt,
			&v.UpdatedAt,
			&v.Active,
			&metaDataBytes,
		); err != nil {
			return nil, err
		}

		// Unmarshal JSONB data into map
		if len(metaDataBytes) > 0 {
			if err := json.Unmarshal(metaDataBytes, &v.MetaData); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		} else {
			v.MetaData = make(map[string]interface{})
		}

		vaults = append(vaults, v)

	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return vaults, nil
}

func (r *vaultRepository) Get(ctx context.Context, id int) (models.Vault, error) {
	const selectFields = "id,user_id,name,data_type,encrypted_data,version,created_at,updated_at,active,meta_data"
	const timeout = time.Second * 2

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	row := r.db.QueryRowContext(ctx, `SELECT `+selectFields+` FROM vault WHERE id=$1;`, id)

	meta := new(MetaData)
	var vault models.Vault
	err := row.Scan(
		&vault.ID,
		&vault.UserID,
		&vault.Name,
		&vault.DataType,
		&vault.EncryptedData,
		&vault.Version,
		&vault.CreatedAt,
		&vault.UpdatedAt,
		&vault.Active,
		meta,
	)

	if err == nil {
		return vault, nil
	}
	if err == sql.ErrNoRows {
		return models.Vault{}, err
	}
	return models.Vault{}, err
}
