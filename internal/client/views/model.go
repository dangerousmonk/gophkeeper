package views

import (
	"fmt"
	"log/slog"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"google.golang.org/grpc"

	"github.com/dangerousmonk/gophkeeper/internal/client/messages"
	"github.com/dangerousmonk/gophkeeper/internal/server/proto"
)

type Model struct {
	State         AppState
	SecretType    SecretType
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
	CurrentForm   []string
	SelectedVault *proto.VaultItem

	grpcConn *grpc.ClientConn
	client   proto.GophKeeperClient
	log      *slog.Logger
}

func NewModel(conn *grpc.ClientConn, client *proto.GophKeeperClient, log *slog.Logger) Model {
	return Model{
		State:    StateStartMenu,
		Focus:    0,
		FormData: make(map[string]string),
		grpcConn: conn,
		client:   *client,
		log:      log,
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

func (m *Model) handleTextInput(msg tea.KeyMsg) {
	if m.Focus < len(m.CurrentForm) {
		field := m.CurrentForm[m.Focus]
		if msg.String() == backspace {
			if len(m.FormData[field]) > 0 {
				m.FormData[field] = m.FormData[field][:len(m.FormData[field])-1]
			}
		} else if len(msg.String()) == 1 {
			m.FormData[field] += msg.String()
		}
	}
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()
		switch msg.String() {
		case ctrlC, q:
			if m.State == StateStartMenu || m.State == StateMainMenu || m.State == StateViewSecrets || m.State == StateViewSecretDetail || m.State == StateChangePassword {
				return m, tea.Quit
			} else {
				// Return to appropriate menu
				if m.State == StateRegister || m.State == StateLogin {
					m.State = StateStartMenu
				} else if m.State == StateSaveSecret || m.State == StateSecretTypeMenu {
					m.State = StateMainMenu
				} else if m.State == StateViewSecretDetail {
					m.State = StateViewSecrets
					m.SelectedVault = nil
				}
				m.resetForm()
				return m, nil
			}

		case esc:
			// Escape returns to previous menu
			switch m.State {
			case StateRegister, StateLogin:
				m.State = StateStartMenu
			case StateSaveSecret, StateSecretTypeMenu, StateChangePassword:
				m.State = StateMainMenu
			case StateViewSecretDetail:
				m.State = StateViewSecrets
				m.SelectedVault = nil
				// Reload secrets when returning from detail view
				m.Loading = true
				m.Message = "Refreshing secrets..."
				return m, GetVaultsStream(m.client, m.Token, m.Password)
			case StateViewSecrets:
				m.State = StateMainMenu
			case StateMainMenu:
				if m.Token != "" {
					return m, nil
				}
				m.State = StateStartMenu
			}
			m.resetForm()
			return m, nil

		case tab, shiftTab, enter, up, down:
			s := msg.String()

			switch m.State {
			case StateStartMenu:
				return m.handleStartMenuNavigation(s)
			case StateMainMenu:
				return m.handleMainMenuNavigation(s)
			case StateSecretTypeMenu:
				return m.handleSecretTypeMenuNavigation(s)
			case StateRegister, StateLogin:
				return m.handleAuthFormNavigation(s)
			case StateSaveSecret:
				return m.handleSaveSecretNavigation(s)
			case StateViewSecrets:
				return m.handleViewSecretsNavigation(s)
			case StateViewSecretDetail:
				return m.handleViewSecretDetailNavigation(s)
			case StateDownloadLocation:
				return m.handleDownloadLocationNavigation(key)
			case StateFileDownload:
				// Block most input during download, only allow quit
				switch key {
				case ctrlC, q:
					return m, tea.Quit
				default:
					return m, nil // Ignore other keys during download
				}
			case StateChangePassword:
				return m.handleChangePasswordNavigation(key)
			}
			return m, nil

		default:
			// Handle text input in forms
			if (m.State == StateRegister || m.State == StateLogin || m.State == StateSaveSecret || m.State == StateChangePassword) && !m.Loading {
				m.handleTextInput(msg)
			}
			// Handle any unexpected states
			switch key {
			case ctrlC, q:
				return m, tea.Quit
			case esc:
				m.State = StateMainMenu // Fallback to main menu
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
			m.State = StateMainMenu
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
			m.State = StateMainMenu
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
			m.State = StateMainMenu
			m.resetForm()
		}
		return m, nil

	case messages.GetVaultsResultMsg:
		m.Loading = false
		if msg.Err != nil {
			m.Err = msg.Err
			m.Message = "Fetch Error: " + msg.Err.Error()
		} else {
			m.Success = true
			// Filter out deactivated secrets and count active ones
			var activeVaults []*proto.VaultItem
			activeCount := 0
			for _, vault := range msg.Vaults {
				if vault.Active {
					activeVaults = append(activeVaults, vault)
					activeCount++
				}
			}
			m.Vaults = activeVaults

			// Use proper pluralization
			if activeCount == 1 {
				m.Message = "Found 1 secret"
			} else {
				m.Message = fmt.Sprintf("Found %d secrets", activeCount)
			}
			m.Focus = 0 // Reset focus to first item
		}
		return m, nil
	case messages.DeactivateVaultResultMsg:
		m.Loading = false
		if msg.Err != nil {
			m.Err = msg.Err
			m.Message = "Delete Error: " + msg.Err.Error()
			// Stay in detail view to show error
			return m, nil
		} else {
			m.Success = msg.Success
			m.Message = "Secret deleted successfully!"
			// Return to secrets list and refresh
			m.State = StateViewSecrets
			m.SelectedVault = nil
			m.Loading = true
			m.Message = "Refreshing secrets..."
			return m, GetVaultsStream(m.client, m.Token, m.Password) // Reload the updated list
		}
	case messages.DownloadResultMsg:
		m.Loading = false
		if msg.Err != nil {
			m.Err = msg.Err
			m.Message = "Download Error: " + msg.Err.Error()
			m.State = StateDownloadLocation // Return to download location view
		} else {
			m.Success = msg.Success
			m.Message = msg.Message
			m.State = StateViewSecretDetail // Return to detail view
		}
	case messages.ChangePasswordResultMsg:
		m.Loading = false
		if msg.Err != nil {
			m.Err = msg.Err
			m.Message = "Password Change Error: " + msg.Err.Error()
		} else {
			m.State = StateMainMenu
			m.resetForm()
			m.Success = msg.Sucess
		}
		return m, nil

	}

	return m, nil
}

// View renders the UI
func (m Model) View() string {
	if m.Err != nil {
		return m.renderError()
	}

	var b strings.Builder

	switch m.State {
	case StateStartMenu:
		b.WriteString(m.renderAuthMenu())
	case StateRegister:
		b.WriteString(m.renderAuthForm("User Registration", []string{"Login", "Password"}))
	case StateLogin:
		b.WriteString(m.renderAuthForm("User Login", []string{"Login", "Password"}))
	case StateMainMenu:
		b.WriteString(m.renderMainMenu())
	case StateSecretTypeMenu:
		b.WriteString(m.renderSecretTypeMenu())
	case StateSaveSecret:
		b.WriteString(m.renderSaveSecretForm())
	case StateViewSecrets:
		b.WriteString(m.renderSecretsListView())
	case StateViewSecretDetail:
		b.WriteString(m.renderSecretDetailView())
	case StateDownloadLocation:
		return m.renderDownloadLocationView()
	case StateFileDownload:
		return m.renderDownloadProgressView()
	case StateChangePassword:
		return m.renderChangePasswordForm()

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
