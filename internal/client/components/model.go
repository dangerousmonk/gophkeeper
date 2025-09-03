package components

import (
	"fmt"
	"log/slog"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dangerousmonk/gophkeeper/internal/client/messages"
	"github.com/dangerousmonk/gophkeeper/internal/server/proto"
	"google.golang.org/grpc"
)

type Model struct {
	State         appState
	SecretType    secretType
	Login         string
	Password      string
	Focus         int
	Loading       bool
	Message       string
	Success       bool
	Err           error
	Token         string
	Vaults        []*proto.VaultItem
	FormData      map[string]string
	CurrentForm   *formDefinition
	SelectedVault *proto.VaultItem

	grpcConn      *grpc.ClientConn
	client        proto.GophKeeperClient
	log           *slog.Logger
	encryptionKey string
}

func NewModel(conn *grpc.ClientConn, client proto.GophKeeperClient, log *slog.Logger, encKey string) Model {
	return Model{
		State:         stateStartMenu,
		Focus:         0,
		FormData:      make(map[string]string),
		grpcConn:      conn,
		client:        client,
		log:           log,
		encryptionKey: encKey,
	}
}

func (m *Model) resetForm() {
	m.FormData = make(map[string]string)
	m.CurrentForm = nil
	m.Focus = 0
	m.Message = ""
	m.Success = false
	m.Err = nil
	m.SelectedVault = nil
}

// Helper method to initialize forms.
func (m *Model) initializeForm(sType secretType, formDef formDefinition) (tea.Model, tea.Cmd) {
	m.resetForm()
	m.State = stateSaveSecret
	m.SecretType = sType
	m.CurrentForm = &formDef

	return m, nil
}

// Helper method to initialize auth forms.
func (m *Model) initializeAuthForm(state appState, formDef formDefinition) (tea.Model, tea.Cmd) {
	m.resetForm()
	m.State = state
	m.CurrentForm = &formDef

	return m, nil
}

func (m *Model) handleTextInput(msg tea.KeyMsg) {
	if m.Focus >= len(m.CurrentForm.Fields) {
		return
	}

	field := m.CurrentForm.Fields[m.Focus]

	if msg.String() == backspace {
		if m.FormData[field.Name] != "" {
			m.FormData[field.Name] = m.FormData[field.Name][:len(m.FormData[field.Name])-1]
		}
	} else if len(msg.String()) == 1 {
		m.FormData[field.Name] += msg.String()
	}
}

// Update handles messages and updates the model.
//
//nolint:gocyclo,funlen // bubbletea update function
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()
		switch key {
		case ctrlC, q:
			if isForExitOnCtrl(m.State) {
				return m, tea.Quit
			}

			if m.State == stateViewSecretDetail {
				m.SelectedVault = nil
			}

			m.State = getPreviousState(m.State)
			m.resetForm()

			return m, nil

		case esc:
			// Escape returns to previous menu
			switch m.State {
			case stateRegister, stateLogin:
				m.State = stateStartMenu
			case stateSaveSecret, stateSecretTypeMenu, stateChangePassword:
				m.State = stateMainMenu
			case stateViewSecretDetail, stateUpdateSecret:
				m.State = stateViewSecrets
				m.SelectedVault = nil
				// Reload secrets when returning from detail view
				m.Loading = true
				m.Message = loadMsg

				return m, getVaultsStream(m.client, m.Token, m.encryptionKey)
			case stateViewSecrets:
				m.State = stateMainMenu
			case stateMainMenu:
				if m.Token != "" {
					return m, nil
				}

				m.State = stateStartMenu
			}

			m.resetForm()

			return m, nil

		case tab, shiftTab, enter, up, down:
			switch m.State {
			case stateStartMenu:
				return m.handleStartMenuNavigation(key)
			case stateRegister, stateLogin:
				return m.handleAuthFormNavigation(key)
			case stateMainMenu:
				return m.handleMainMenuNavigation(key)
			case stateSecretTypeMenu:
				return m.handleSecretTypeMenuNavigation(key)
			case stateSaveSecret:
				return m.handleSaveSecretNavigation(key)
			case stateViewSecrets:
				return m.handleViewSecretsNavigation(key)
			case stateViewSecretDetail:
				return m.handleViewSecretDetailNavigation(key)
			case stateDownloadLocation:
				return m.handleDownloadLocationNavigation(key)
			case stateFileDownload:
				// Block most input during download, only allow quit
				switch key {
				case ctrlC, q:
					return m, tea.Quit
				default:
					return m, nil // Ignore other keys during download
				}
			case stateChangePassword:
				return m.handleChangePasswordNavigation(key)
			case stateUpdateSecret:
				return m.handleUpdateSecretNavigation(key)
			}

			return m, nil

		default:
			// Handle text input in forms
			if isForTextInput(m.State) && !m.Loading {
				m.handleTextInput(msg)
			}
			// Handle any unexpected states
			switch key {
			case ctrlC, q:
				return m, tea.Quit
			case esc:
				m.State = stateMainMenu // Fallback to main menu
				m.resetForm()

				return m, nil
			}

			return m, nil
		}

	case messages.RegistrationResultMsg:
		m.Loading = false
		if msg.Err != nil {
			m.Err = msg.Err
			m.Message = "Registration Error: " + msg.Err.Error()
		} else {
			m.Success = msg.Success
			m.Message = msg.Message
			m.Token = msg.Token
			m.State = stateMainMenu
			m.Login = msg.Login
			m.resetForm()
		}

		return m, nil

	case messages.LoginResultMsg:
		m.Loading = false
		if msg.Err != nil {
			m.Err = msg.Err
			m.Message = "Login Error: " + msg.Err.Error()
		} else {
			m.Success = msg.Success
			m.Message = msg.Message
			m.Token = msg.Token
			m.State = stateMainMenu
			m.Password = msg.Pasword
			m.Login = msg.Login
			m.resetForm()
		}

		return m, nil

	case messages.SaveVaultResultMsg:
		m.Loading = false
		if msg.Err != nil {
			m.Err = msg.Err
			m.Message = "Save Error: " + msg.Err.Error()
		} else {
			m.Success = msg.Success
			m.Message = "Secret saved successfully!"
			m.State = stateMainMenu
			m.resetForm()
		}

		return m, nil

	case messages.GetVaultsResultMsg:
		m.Loading = false
		if msg.Err != nil {
			m.Err = msg.Err
			m.Message = "Fetch Error: " + msg.Err.Error()

			return m, nil
		}

		m.Success = true

		var activeVaults []*proto.VaultItem

		activeCount := 0

		for _, vault := range msg.Vaults {
			if vault.Active {
				activeVaults = append(activeVaults, vault)
				activeCount++
			}
		}

		m.Vaults = activeVaults

		// Use proper pluralization.
		if activeCount == 1 {
			m.Message = "Found 1 secret"
		} else {
			m.Message = fmt.Sprintf("Found %d secrets", activeCount)
		}

		m.Focus = 0 // Reset focus to first item.

		return m, nil
	case messages.DeactivateVaultResultMsg:
		m.Loading = false
		if msg.Err != nil {
			m.Err = msg.Err
			m.Message = "Delete Error: " + msg.Err.Error()
			// Stay in detail view to show error
			return m, nil
		}

		m.Success = msg.Success
		m.Message = "Secret deleted successfully!"
		// Return to secrets list and refresh
		m.State = stateViewSecrets
		m.SelectedVault = nil
		m.Loading = true
		m.Message = loadMsg

		return m, getVaultsStream(m.client, m.Token, m.encryptionKey) // Reload the updated list

	case messages.DownloadResultMsg:
		m.Loading = false
		if msg.Err != nil {
			m.Err = msg.Err
			m.Message = "Download Error: " + msg.Err.Error()
			m.State = stateDownloadLocation // Return to download location view
		} else {
			m.Success = msg.Success
			m.Message = msg.Message
			m.State = stateViewSecretDetail // Return to detail view
		}
	case messages.ChangePasswordResultMsg:
		m.Loading = false
		if msg.Err != nil {
			m.Err = msg.Err
			m.Message = "Password Change Error: " + msg.Err.Error()
		} else {
			m.resetForm()
			m.State = stateMainMenu
			m.Success = msg.Success
		}

		return m, nil
	case messages.UpdateVaultResultMsg:
		m.Loading = false
		if msg.Err != nil {
			m.Err = msg.Err
			m.Message = "Update secret error: " + msg.Err.Error()
		} else {
			m.Loading = true
			m.Message = loadMsg
			m.State = stateViewSecrets
			m.Success = msg.Success

			return m, getVaultsStream(m.client, m.Token, m.encryptionKey) // Reload the updated list
		}

		return m, nil
	}

	return m, nil
}

// View renders the UI.
func (m Model) View() string {
	if m.Err != nil {
		return m.renderError()
	}

	var b strings.Builder

	switch m.State {
	case stateStartMenu:
		b.WriteString(m.renderAuthMenu())
	case stateRegister:
		b.WriteString(m.renderAuthForm("User Registration", []string{"Login", "Password"}))
	case stateLogin:
		b.WriteString(m.renderAuthForm("User Login", []string{"Login", "Password"}))
	case stateMainMenu:
		b.WriteString(m.renderMainMenu())
	case stateSecretTypeMenu:
		b.WriteString(m.renderSecretTypeMenu())
	case stateSaveSecret:
		b.WriteString(m.renderSaveSecretForm())
	case stateViewSecrets:
		b.WriteString(m.renderSecretsListView())
	case stateViewSecretDetail:
		b.WriteString(m.renderSecretDetailView())
	case stateDownloadLocation:
		return m.renderDownloadLocationView()
	case stateFileDownload:
		return m.renderDownloadProgressView()
	case stateChangePassword:
		return m.renderChangePasswordForm()
	case stateUpdateSecret:
		return m.renderSaveSecretForm()

	default:
		b.WriteString("Unknown state")
	}

	// Ensure the view has proper padding and doesn't get cut off
	return lipgloss.NewStyle().
		Padding(1, 2).
		MaxWidth(80).
		Render(b.String())
}

func (m Model) Init() tea.Cmd {
	return nil
}
