package models

import "time"

type DataType string

const (
	Credentials DataType = "credentials"
	Text        DataType = "text"
	Binary      DataType = "binary"
	Card        DataType = "card"
)

type Vault struct {
	ID            int                    `json:"id"`
	UserID        int                    `json:"user_id"`
	Version       int                    `json:"version"`
	Name          string                 `json:"name"`
	DataType      DataType               `json:"data_type"`
	EncryptedData []byte                 `json:"encrypted_data"`
	MetaData      map[string]interface{} `json:"meta_data"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
	Active        bool                   `json:"active"`
}

type InsertVaultRequest struct {
	Name          string   `json:"name" validate:"required,min=3,max=150"`
	DataType      DataType `json:"data_type" validate:"required"`
	EncryptedData string   `json:"encrypted_data"`
	MetaData      string   `json:"meta_data"`
}
