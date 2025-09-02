package components

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) handleStartMenuNavigation(key string) (tea.Model, tea.Cmd) {
	switch key {
	case enter:
		switch m.Focus {
		case int(stateStartMenu):
			return m.initializeAuthForm(stateRegister, registrationForm)
		case int(stateRegister):
			return m.initializeAuthForm(stateLogin, loginForm)
		case int(stateLogin):
			return m, tea.Quit
		}
	case up:
		m.Focus--
	case down:
		m.Focus++
	}

	if m.Focus < 0 {
		m.Focus = 2
	} else if m.Focus > 2 {
		m.Focus = 0
	}
	return m, nil
}

func (m *Model) handleMainMenuNavigation(key string) (tea.Model, tea.Cmd) {
	switch key {
	case enter:
		switch m.Focus {
		case 0: // Save secret
			m.State = stateSecretTypeMenu
			m.Focus = 0
			return m, nil
		case 1: // View secrets
			m.State = stateViewSecrets
			m.Loading = true
			m.Message = "Loading secrets..."
			m.Focus = 0
			return m, getVaultsStream(m.client, m.Token, m.Password) // Reload secrets when entering the menu
		case 2: // Change password
			return m.initializeAuthForm(stateChangePassword, changePasswordForm)
		case 3: // Quit
			return m, tea.Quit
		}
	case up:
		// m.Focus--
		m.Focus = (m.Focus - 1 + 4) % 4
	case down:
		// m.Focus++
		m.Focus = (m.Focus + 1) % 4
	}

	if m.Focus < 0 {
		m.Focus = 3
	} else if m.Focus > 3 {
		m.Focus = 0
	}
	return m, nil
}

func (m *Model) handleSecretTypeMenuNavigation(key string) (tea.Model, tea.Cmd) {
	switch key {
	case enter:
		switch m.Focus {
		case 0: // Login/Password
			return m.initializeForm(secretTypeCredential, credentialsForm)
		case 1: // Bank Card
			return m.initializeForm(secretTypeBankCard, bankCardForm)
		case 2: // Text
			return m.initializeForm(secretTypeText, textForm)
		case 3: // File
			return m.initializeForm(secretTypeBinary, fileForm)
		case 4: // Back
			m.State = stateMainMenu
			return m, nil
		}
	case up:
		m.Focus--
	case down:
		m.Focus++
	}

	if m.Focus < 0 {
		m.Focus = 4
	} else if m.Focus > 4 {
		m.Focus = 0
	}
	return m, nil
}

func (m *Model) handleAuthFormNavigation(key string) (tea.Model, tea.Cmd) {
	if key == enter && m.Focus == len(m.CurrentForm.Fields) && !m.Loading {
		if m.FormData["Login"] == "" || m.FormData["Password"] == "" {
			m.Message = "Please fill in all fields"
			return m, nil
		}
		m.Loading = true
		if m.State == stateRegister {
			m.Message = "Registering..."
			return m, registerUser(m.client, m.FormData["Login"], m.FormData["Password"])
		} else {
			m.Message = "Logging in..."
			return m, loginUser(m.client, m.FormData["Login"], m.FormData["Password"])
		}
	}

	switch key {
	case up, shiftTab:
		m.Focus--
	case down, tab:
		m.Focus++
	}

	if m.Focus > len(m.CurrentForm.Fields) {
		m.Focus = 0
	} else if m.Focus < 0 {
		m.Focus = len(m.CurrentForm.Fields)
	}
	return m, nil
}

func (m *Model) handleSaveSecretNavigation(key string) (tea.Model, tea.Cmd) {
	if key == enter && m.Focus == len(m.CurrentForm.Fields) && !m.Loading {
		// Validate required fields
		for _, field := range m.CurrentForm.Fields {
			if m.FormData[field.Name] == "" {
				m.Message = fmt.Sprintf("Please fill in %s", field.Name)
				return m, nil
			}
		}
		m.Loading = true
		m.Message = "Saving secret..."
		title := m.FormData[m.CurrentForm.Fields[0].Name]
		return m, saveVault(m.client, m.Token, m.Password, m.SecretType, m.FormData, title)
	}

	switch key {
	case up, shiftTab:
		m.Focus--
	case down, tab:
		m.Focus++
	}

	if m.Focus > len(m.CurrentForm.Fields) {
		m.Focus = 0
	} else if m.Focus < 0 {
		m.Focus = len(m.CurrentForm.Fields)
	}
	return m, nil
}

func (m *Model) handleViewSecretsNavigation(key string) (tea.Model, tea.Cmd) {
	switch key {
	case enter:
		if len(m.Vaults) > 0 && m.Focus < len(m.Vaults) {
			// Select the vault to view details
			m.SelectedVault = m.Vaults[m.Focus]
			m.State = stateViewSecretDetail
			return m, nil
		}
	case up:
		m.Focus--
	case down:
		m.Focus++
	}

	// Handle focus wrapping
	if len(m.Vaults) > 0 {
		if m.Focus < 0 {
			m.Focus = len(m.Vaults) - 1
		} else if m.Focus >= len(m.Vaults) {
			m.Focus = 0
		}
	} else {
		m.Focus = 0
	}

	return m, nil
}

func (m *Model) handleViewSecretDetailNavigation(key string) (tea.Model, tea.Cmd) {
	switch key {
	case ctrlC, q:
		return m, tea.Quit

	case esc:
		m.State = stateViewSecrets
		m.SelectedVault = nil
		return m, nil

	case enter:
		if m.Focus == 0 { // Go Back button
			m.State = stateViewSecrets
			m.SelectedVault = nil
		} else if m.Focus == 1 && m.SelectedVault.DataType == secretTypeBinary && m.SelectedVault != nil {
			// Download button for binary

			m.State = stateDownloadLocation
			m.CurrentForm = &fileLocationForm
			m.Focus = 0

			// Set default download path
			defaultPath := getDefaultDownloadPath(m.SelectedVault)
			m.FormData["Download Path"] = defaultPath
		} else if m.Focus == 1 || m.Focus == 2 { // Delete button for all types
			if m.SelectedVault != nil {
				m.Loading = true
				m.Message = "Deleting secret..."
				return m, deactivateVaultGrpc(m.client, m.Token, m.SelectedVault)
			}
		}

	case left, right, tab, shiftTab:
		buttonCount := 3 // Back, Download, Delete
		if key == right || key == tab {
			m.Focus = (m.Focus + 1) % buttonCount
		} else {
			m.Focus = (m.Focus - 1 + buttonCount) % buttonCount
		}

	case up, down:
		// Ignore vertical navigation in button row
		return m, nil
	}

	return m, nil
}

func (m *Model) handleDownloadLocationNavigation(key string) (tea.Model, tea.Cmd) {
	switch key {
	case enter:
		switch m.Focus {
		case 0:
			// Text input - handled in default case
			return m, nil
		case 1:
			downloadPath := m.FormData["Download Path"]
			if downloadPath == "" {
				m.Message = "Please enter a download path"
				return m, nil
			}
			m.State = stateFileDownload
			m.Loading = true
			m.Message = "Downloading file..."
			return m, m.downloadFile(downloadPath)
		case 2:
			m.State = stateViewSecretDetail
			m.FormData = nil
			return m, nil
		}
	case esc:
		m.State = stateViewSecretDetail
		m.FormData = nil
		return m, nil
	case tab, shiftTab:
		if key == tab {
			m.Focus = (m.Focus + 1) % 3
		} else {
			m.Focus = (m.Focus - 1 + 3) % 3
		}
	default:
		// Handle text input for download path
		if m.Focus == 0 && !m.Loading {
			field := "Download Path"
			if key == backspace {
				if len(m.FormData[field]) > 0 {
					m.FormData[field] = m.FormData[field][:len(m.FormData[field])-1]
				}
			} else if len(key) == 1 {
				m.FormData[field] += key
			}
		}
	}
	return m, nil
}

func (m *Model) handleChangePasswordNavigation(key string) (tea.Model, tea.Cmd) {
	switch key {
	case enter:
		if m.Focus == len(m.CurrentForm.Fields) && !m.Loading {
			// Validate form
			currentPassword := m.FormData["Current Password"]
			newPassword := m.FormData["New Password"]
			confirmPassword := m.FormData["Confirm New Password"]

			if currentPassword == "" || newPassword == "" || confirmPassword == "" {
				m.Message = "Please fill in all fields"
				return m, nil
			}

			if currentPassword != m.Password {
				m.Message = "Current password is wrong"
				return m, nil
			}

			if newPassword != confirmPassword {
				m.Message = "New passwords do not match"
				return m, nil
			}

			if newPassword == currentPassword {
				m.Message = "New password must be different from current password"
				return m, nil
			}

			m.Loading = true
			m.Message = "Changing password..."
			return m, changePassword(m.client, m.Login, m.Token, currentPassword, newPassword)
		}
	case esc:
		m.State = stateMainMenu
		m.resetForm()
		return m, nil
	case up, shiftTab:
		m.Focus = (m.Focus - 1 + 4) % 4 // 3 fields + 1 button
	case down, tab:
		m.Focus = (m.Focus + 1) % 4
	}

	return m, nil
}
