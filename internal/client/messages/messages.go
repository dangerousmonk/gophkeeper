package messages

import (
	"github.com/dangerousmonk/gophkeeper/internal/server/proto"
)

// Message types
type RegistrationResultMsg struct {
	Success bool
	Message string
	Err     error
	Login   string
	Token   string
}

type LoginResultMsg struct {
	Success bool
	Message string
	Err     error
	Token   string
	Pasword string
	Login   string
}

type SaveVaultResultMsg struct {
	Success bool
	Err     error
}

type GetVaultsResultMsg struct {
	Vaults []*proto.VaultItem
	Err    error
}

type DeactivateVaultResultMsg struct {
	Success bool
	Err     error
}

type DownloadResultMsg struct {
	Success bool
	Message string
	Err     error
}

type ChangePasswordResultMsg struct {
	Sucess bool
	Err    error
}
