package views

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/dangerousmonk/gophkeeper/internal/server/proto"
	"github.com/dangerousmonk/gophkeeper/internal/utils"
	"github.com/dangerousmonk/gophkeeper/internal/version"
)

func (m *Model) renderStatusMessage() string {
	statusStyle := lipgloss.NewStyle().
		Width(60). // Consistent width with form
		Height(1).
		Padding(0, 1).
		Border(lipgloss.RoundedBorder()).
		Align(lipgloss.Center)

	switch {
	case m.Success:
		statusStyle = statusStyle.
			Foreground(lipgloss.Color("#10B981")).
			BorderForeground(lipgloss.Color("#10B981"))
	case m.Err != nil:
		statusStyle = statusStyle.
			Foreground(lipgloss.Color("#EF4444")).
			BorderForeground(lipgloss.Color("#EF4444"))
	default:
		statusStyle = statusStyle.
			Foreground(lipgloss.Color("#7D56F4")).
			BorderForeground(lipgloss.Color("#7D56F4"))
	}
	return statusStyle.Render(m.Message)
}

func (m *Model) renderError() string {
	errorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF6B6B")).
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#FF6B6B")).
		Width(60)

	return errorStyle.Render(fmt.Sprintf("Error: %s\n\nPress 'q' or 'esc' to continue", m.Err.Error()))
}

func (m *Model) renderAuthMenu() string {
	var b strings.Builder
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7D56F4")).
		PaddingBottom(1).
		Render(credentialsIcon + " " + "GophKeeper - Secure Data Manager")
	b.WriteString(title + "\n")

	statusBar := lipgloss.NewStyle().
		Foreground(lipgloss.Color("7D56F4")).
		Render(version.GetVersionInfo())
	b.WriteString(statusBar + "\n\n")

	// Menu options with styling
	options := []string{textIcon + " Register", loginIcon + " Login", quitIcon + " Quit"}
	for i, option := range options {
		style := lipgloss.NewStyle().Padding(0, 1)
		if m.Focus == i {
			style = style.
				Background(lipgloss.Color("#7D56F4")).
				Foreground(lipgloss.Color("#FFFFFF")).
				Bold(true)
			option = selectIcon + " " + option
		} else {
			option = "  " + option
		}
		b.WriteString(style.Render(option) + "\n")
	}

	b.WriteString("\n")

	// Status message
	if m.Message != "" {
		b.WriteString(m.renderStatusMessage() + "\n\n")
	}

	// Help text
	helpText := lipgloss.NewStyle().
		Faint(true).
		Foreground(lipgloss.Color("#6B7280")).
		Render("↑↓: Navigate • Enter: Select • Esc: Back • Ctrl+C/Q: Quit")
	b.WriteString(helpText)

	return b.String()
}

func (m *Model) renderAuthForm(title string, fields []string) string {
	var b strings.Builder

	// Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7D56F4")).
		PaddingBottom(1)
	b.WriteString(titleStyle.Render(title) + "\n\n")

	// Form fields
	for i, field := range fields {
		fieldStyle := lipgloss.NewStyle().Width(12) // Fixed width for labels
		if m.Focus == i {
			fieldStyle = fieldStyle.Foreground(lipgloss.Color("#7D56F4")).Bold(true)
		}

		value := m.FormData[field]
		if field == "Password" {
			value = strings.Repeat("•", len(value))
		}

		// Create a container for the field row
		fieldRowStyle := lipgloss.NewStyle().Width(60) // Fixed width for entire row

		// Label with fixed width
		label := fieldStyle.Render(field + ":")

		// Input field with consistent styling
		inputStyle := lipgloss.NewStyle().
			Width(30).
			Height(1).
			Padding(0, 1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7D56F4"))

		if m.Focus == i {
			inputStyle = inputStyle.BorderForeground(lipgloss.Color("#FF6B6B"))
		}

		inputField := inputStyle.Render(value)

		// Combine label and input field in a row
		row := lipgloss.JoinHorizontal(lipgloss.Center, label, " ", inputField)
		b.WriteString(fieldRowStyle.Render(row) + "\n\n")
	}

	// Submit button
	submitText := "Submit"
	if title == "User Registration" {
		submitText = "Register"
	} else {
		submitText = "Login"
	}

	buttonStyle := lipgloss.NewStyle().
		Width(60). // Same width as form rows
		Height(1).
		Padding(0, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4"))

	if m.Focus == len(fields) {
		buttonStyle = buttonStyle.
			Background(lipgloss.Color("#7D56F4")).
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true)
	}

	b.WriteString(buttonStyle.Render(submitText) + "\n\n")

	// Status message
	if m.Message != "" {
		statusMsg := m.renderStatusMessage()
		b.WriteString(lipgloss.NewStyle().Width(60).Render(statusMsg) + "\n\n")
	}

	// Help text
	helpText := lipgloss.NewStyle().
		Faint(true).
		Foreground(lipgloss.Color("#6B7280")).
		Render("Tab/↑↓: Navigate • Enter: Submit • Esc: Back")
	b.WriteString(lipgloss.NewStyle().Width(60).Render(helpText))

	return b.String()
}

func (m *Model) renderMainMenu() string {
	var b strings.Builder

	// Welcome title
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7D56F4")).
		PaddingBottom(1).
		Render(fmt.Sprintf(hiIcon+" Welcome, %s!", m.Login))
	b.WriteString(title + "\n\n")

	// Main menu options
	options := []string{storageIcon + " Save New Secret", viewIcon + " View Secrets", credentialsIcon + " Change Password", quitIcon + " Quit"}
	for i, option := range options {
		style := lipgloss.NewStyle().Padding(0, 1)
		if m.Focus == i {
			style = style.
				Background(lipgloss.Color("#7D56F4")).
				Foreground(lipgloss.Color("#FFFFFF")).
				Bold(true)
			option = selectIcon + "" + option
		} else {
			option = "  " + option
		}
		b.WriteString(style.Render(option) + "\n")
	}

	b.WriteString("\n")

	// Status message
	if m.Message != "" {
		b.WriteString(m.renderStatusMessage() + "\n\n")
	}

	// Help text
	helpText := lipgloss.NewStyle().
		Faint(true).
		Foreground(lipgloss.Color("#6B7280")).
		Render("↑↓: Navigate • Enter: Select • Esc: Back • Ctrl+C/Q: Quit")
	b.WriteString(helpText)

	return b.String()
}

func (m *Model) renderSecretTypeMenu() string {
	var b strings.Builder

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7D56F4")).
		PaddingBottom(1).
		Render(folderIcon + " Select Secret Type")
	b.WriteString(title + "\n\n")

	options := []string{credentialsIcon + " Login/Password", bankCardIcon + " Bank Card", textIcon + " Text Note", folderIcon + " File", "↩ Back"}
	for i, option := range options {
		style := lipgloss.NewStyle().Padding(0, 1)
		if m.Focus == i {
			style = style.
				Background(lipgloss.Color("#7D56F4")).
				Foreground(lipgloss.Color("#FFFFFF")).
				Bold(true)
			option = selectIcon + " " + option
		} else {
			option = "  " + option
		}
		b.WriteString(style.Render(option) + "\n")
	}

	b.WriteString("\n")
	helpText := lipgloss.NewStyle().
		Faint(true).
		Foreground(lipgloss.Color("#6B7280")).
		Render("↑↓: Navigate • Enter: Select • Esc: Back")
	b.WriteString(helpText)

	return b.String()
}

func (m *Model) renderSaveSecretForm() string {
	var b strings.Builder

	title := "Save Secret"
	switch m.SecretType {
	case SecretTypeCredentials:
		title = credentialsIcon + " Save Login/Password"
	case SecretTypeBankCard:
		title = bankCardIcon + " Save Bank Card"
	case SecretTypeText:
		title = textIcon + " Save Text Note"
	}

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7D56F4")).
		PaddingBottom(1)
	b.WriteString(titleStyle.Render(title) + "\n\n")

	// Form fields
	for i, field := range m.CurrentForm {
		fieldStyle := lipgloss.NewStyle().Width(15) // Fixed width for labels
		if m.Focus == i {
			fieldStyle = fieldStyle.Foreground(lipgloss.Color("#7D56F4")).Bold(true)
		}

		value := m.FormData[field]

		if strings.Contains(strings.ToLower(field), "password") || field == "cvv" || field == "cvc" {
			value = strings.Repeat("•", len(value))
		}

		// Create a container for the field row
		fieldRowStyle := lipgloss.NewStyle().Width(60) // Fixed width for entire row

		// Label with fixed width
		label := fieldStyle.Render(field + ":")

		// Input field with consistent styling
		inputStyle := lipgloss.NewStyle().
			Width(35).
			Height(1).
			Padding(0, 1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7D56F4"))

		if m.Focus == i {
			inputStyle = inputStyle.BorderForeground(lipgloss.Color("#FF6B6B"))
		}

		inputField := inputStyle.Render(value)

		// Combine label and input field in a row
		row := lipgloss.JoinHorizontal(lipgloss.Center, label, " ", inputField)
		b.WriteString(fieldRowStyle.Render(row) + "\n\n")
	}

	// Submit button
	buttonStyle := lipgloss.NewStyle().
		Width(60). // Same width as form rows
		Height(1).
		Padding(0, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4"))

	if m.Focus == len(m.CurrentForm) {
		buttonStyle = buttonStyle.
			Background(lipgloss.Color("#7D56F4")).
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true)
	}

	b.WriteString(buttonStyle.Render("Save Secret") + "\n\n")

	// Status message
	if m.Message != "" {
		statusMsg := m.renderStatusMessage()
		b.WriteString(lipgloss.NewStyle().Width(60).Render(statusMsg) + "\n\n")
	}

	// Help text
	helpText := lipgloss.NewStyle().
		Faint(true).
		Foreground(lipgloss.Color("#6B7280")).
		Render("Tab/↑↓: Navigate • Enter: Submit • Esc: Back")
	b.WriteString(lipgloss.NewStyle().Width(60).Render(helpText))

	return b.String()
}

func (m *Model) renderSecretData(vault *proto.VaultItem) string {
	var b strings.Builder

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7D56F4")).
		Render(credentialsIcon + " Secret Data")
	b.WriteString(title + "\n\n")

	// Try to decode the JSON data (now decrypted)
	var secretData map[string]any
	if err := json.Unmarshal(vault.EncryptedData, &secretData); err != nil {
		return fmt.Sprintf(redCross+" Cannot display secret data: %v", err)
	}

	dataStyle := lipgloss.NewStyle().
		Padding(1).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#10B981")).
		Width(70)

	var dataContent strings.Builder
	for key, value := range secretData {
		dataContent.WriteString(fmt.Sprintf(viewDataIcon+" "+" %s: %s\n", key, value))
	}

	b.WriteString(dataStyle.Render(dataContent.String()))
	return b.String()
}

func (m *Model) renderDownloadProgressView() string {
	var b strings.Builder

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7D56F4")).
		PaddingBottom(1).
		Render(downLoadingIcon + " Downloading File...")
	b.WriteString(title + "\n\n")

	if m.Loading {
		loadingStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4")).
			Italic(true)
		b.WriteString(loadingStyle.Render(m.Message) + "\n\n")
	}

	// Help text
	helpText := lipgloss.NewStyle().
		Faint(true).
		Foreground(lipgloss.Color("#6B7280")).
		Render("Please wait...")
	b.WriteString(helpText)

	return b.String()
}

func (m *Model) renderSecretsListView() string {
	var b strings.Builder

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7D56F4")).
		PaddingBottom(1).
		Render("📋 Your Secrets")
	b.WriteString(title + "\n\n")

	if m.Loading {
		loadingStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4")).
			Italic(true)
		b.WriteString(loadingStyle.Render(m.Message) + "\n\n")
	} else if len(m.Vaults) == 0 {
		emptyStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")).
			Italic(true)
		b.WriteString(emptyStyle.Render("No secrets found.") + "\n\n")
	} else {
		for i, vault := range m.Vaults {
			secretStyle := lipgloss.NewStyle().
				Padding(0, 1).
				MarginBottom(1)

			if m.Focus == i {
				secretStyle = secretStyle.
					Background(lipgloss.Color("#7D56F4")).
					Foreground(lipgloss.Color("#FFFFFF")).
					Bold(true)
			}

			icon := getIcon(vault.DataType)
			displayText := fmt.Sprintf("%s %s | %s | %s %s",
				icon,
				utils.TruncateString(vault.Name, 18),
				vault.DataType,
				timeIcon,
				utils.FormatDate(vault.CreatedAt),
			)

			b.WriteString(secretStyle.Render(displayText) + "\n")
		}
	}

	b.WriteString("\n")

	// Status message (for non-loading messages)
	if m.Message != "" && !m.Loading {
		b.WriteString(m.renderStatusMessage() + "\n\n")
	}

	// Help text
	helpText := ""
	if len(m.Vaults) > 0 && !m.Loading {
		helpText = "↑↓: Select • Enter: View Details • Esc: Back to Menu"
	} else if m.Loading {
		helpText = "Loading secrets..."
	} else {
		helpText = "Esc: Back to Menu"
	}

	b.WriteString(lipgloss.NewStyle().
		Faint(true).
		Foreground(lipgloss.Color("#6B7280")).
		Render(helpText))

	return b.String()
}

func renderFileDetails(vault *proto.VaultItem) string {
	var b strings.Builder

	fileStyle := lipgloss.NewStyle().
		Padding(1).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Width(70)

	// Extract file metadata
	fileName := "Unknown"
	filePath := "Unknown"
	fileSize := int64(0)
	fileType := "Unknown"

	if vault.MetaData != nil && vault.MetaData.Fields != nil {
		fields := vault.MetaData.Fields

		if nameVal, exists := fields["file_name"]; exists {
			fileName = nameVal.GetStringValue()
		}

		if pathVal, exists := fields["file_path"]; exists {
			filePath = pathVal.GetStringValue()
		}

		if sizeVal, exists := fields["file_size"]; exists {
			fileSize = int64(sizeVal.GetNumberValue())
		}

		if typeVal, exists := fields["file_type"]; exists {
			fileType = typeVal.GetStringValue()
		}
	}

	fileContent := fmt.Sprintf("%s File Name: %s\n", fileIcon, fileName)
	fileContent += fmt.Sprintf("%s File Type: %s\n", folderIcon, fileType)
	fileContent += fmt.Sprintf("%s File Size: %s\n", storageIcon, utils.FormatFileSize(fileSize))
	fileContent += fmt.Sprintf("%s Original Path: %s\n", locationIcon, filePath)
	fileContent += fmt.Sprintf("%s Storage: Encrypted binary data", lockIcon)
	b.WriteString(fileStyle.Render(fileContent))
	return b.String()
}

func (m Model) renderSecretDetailView() string {
	var b strings.Builder

	if m.SelectedVault == nil {
		return "No secret selected"
	}

	vault := m.SelectedVault

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7D56F4")).
		PaddingBottom(1).
		Render(magnifierIcon + " Secret Details")
	b.WriteString(title + "\n\n")

	infoStyle := lipgloss.NewStyle().
		Padding(0, 1).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Padding(1).
		Width(70)

	infoContent := fmt.Sprintf("%s Name: %s\n", folderIcon, vault.Name)
	infoContent += fmt.Sprintf("%s Type: %s\n", lockIcon, vault.DataType)
	infoContent += fmt.Sprintf("%s ID: %d\n", idIcon, vault.Id)
	infoContent += fmt.Sprintf("%s Created: %s\n", timeIcon, utils.FormatDate(vault.CreatedAt))
	infoContent += fmt.Sprintf("%s Updated: %s\n", calendarIcon, utils.FormatDate(vault.UpdatedAt))
	infoContent += fmt.Sprintf("%s Active: %v\n", checkMarkIcon, vault.Active)
	infoContent += fmt.Sprintf("%s Version: %d", versionIcon, vault.Version)

	b.WriteString(infoStyle.Render(infoContent) + "\n\n")

	// Show file details for binary data type
	if vault.DataType == "binary" {
		fileDetails := renderFileDetails(vault)
		b.WriteString(fileDetails + "\n\n")
	} else {
		// Try to decode and display the secret data if it's in a known format
		if len(vault.EncryptedData) > 0 {
			secretData := m.renderSecretData(vault)
			b.WriteString(secretData + "\n\n")
		}
	}

	// Button row with appropriate buttons
	buttonRowStyle := lipgloss.NewStyle().
		Width(70).
		Align(lipgloss.Center)

	var buttons []string

	// Go Back button
	backButtonStyle := lipgloss.NewStyle().
		Width(15).
		Height(1).
		Padding(0, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Align(lipgloss.Center)

	if m.Focus == 0 {
		backButtonStyle = backButtonStyle.
			Background(lipgloss.Color("#7D56F4")).
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true)
	}
	buttons = append(buttons, backButtonStyle.Render("← Back"))

	// Download button (only for binary/files)
	if vault.DataType == "binary" {
		downloadButtonStyle := lipgloss.NewStyle().
			Width(15).
			Height(1).
			Padding(0, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#10B981")).
			Align(lipgloss.Center)

		if m.Focus == 1 {
			downloadButtonStyle = downloadButtonStyle.
				Background(lipgloss.Color("#10B981")).
				Foreground(lipgloss.Color("#FFFFFF")).
				Bold(true)
		} else {
			downloadButtonStyle = downloadButtonStyle.
				Foreground(lipgloss.Color("#10B981"))
		}
		buttons = append(buttons, downloadButtonStyle.Render(downLoadingIcon+" Download"))
	}

	deleteButtonIndex := 1
	if vault.DataType == "binary" {
		deleteButtonIndex = 2 // Back(0), Download(1), Delete(2)
	}

	deleteButtonStyle := lipgloss.NewStyle().
		Width(15).
		Height(1).
		Padding(0, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#FF6B6B")).
		Align(lipgloss.Center)

	if m.Focus == deleteButtonIndex {
		deleteButtonStyle = deleteButtonStyle.
			Background(lipgloss.Color("#FF6B6B")).
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true)
	} else {
		deleteButtonStyle = deleteButtonStyle.
			Foreground(lipgloss.Color("#FF6B6B"))
	}
	buttons = append(buttons, deleteButtonStyle.Render(binIcon+" "+" Delete"))

	// Join buttons with appropriate spacing
	var buttonRow string
	if vault.DataType == "binary" {
		buttonRow = lipgloss.JoinHorizontal(lipgloss.Center,
			buttons[0], "   ", buttons[1], "   ", buttons[2])
	} else {
		buttonRow = lipgloss.JoinHorizontal(lipgloss.Center,
			buttons[0], "   ", buttons[1])
	}
	b.WriteString(buttonRowStyle.Render(buttonRow) + "\n\n")

	if m.Message != "" {
		b.WriteString(m.renderStatusMessage() + "\n\n")
	}

	helpText := lipgloss.NewStyle().
		Faint(true).
		Foreground(lipgloss.Color("#6B7280")).
		Render("←→: Navigate • Enter: Select • Esc: Back to List")
	b.WriteString(helpText)

	return b.String()
}

func (m Model) renderDownloadLocationView() string {
	var b strings.Builder

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7D56F4")).
		PaddingBottom(1).
		Render(downLoadingIcon + " Download File")
	b.WriteString(title + "\n\n")

	// Download path input
	fieldStyle := lipgloss.NewStyle().Width(20)
	if m.Focus == 0 {
		fieldStyle = fieldStyle.Foreground(lipgloss.Color("#7D56F4")).Bold(true)
	}

	// Create a container for the field row
	value := m.FormData["Download Path"]
	fieldRowStyle := lipgloss.NewStyle().Width(60)

	// Label with fixed width
	label := fieldStyle.Render("Download Path" + ":")

	// Input field with consistent styling
	inputStyle := lipgloss.NewStyle().
		Width(30).
		Height(1).
		Padding(0, 1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4"))
	if m.Focus == 0 {
		inputStyle = inputStyle.BorderForeground(lipgloss.Color("#FF6B6B"))
	}

	inputField := inputStyle.Render(value)

	// Combine label and input field in a row
	row := lipgloss.JoinHorizontal(lipgloss.Center, label, " ", inputField)
	b.WriteString(fieldRowStyle.Render(row) + "\n\n")

	// Action buttons
	buttonRowStyle := lipgloss.NewStyle().
		Width(70).
		Align(lipgloss.Center)

	downloadButtonStyle := lipgloss.NewStyle().
		Width(15).
		Height(1).
		Padding(0, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#10B981")).
		Align(lipgloss.Center)

	if m.Focus == 1 {
		downloadButtonStyle = downloadButtonStyle.
			Background(lipgloss.Color("#10B981")).
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true)
	}

	cancelButtonStyle := lipgloss.NewStyle().
		Width(15).
		Height(1).
		Padding(0, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Align(lipgloss.Center)

	if m.Focus == 2 {
		cancelButtonStyle = cancelButtonStyle.
			Background(lipgloss.Color("#7D56F4")).
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true)
	}

	buttons := lipgloss.JoinHorizontal(lipgloss.Center,
		downloadButtonStyle.Render(downLoadingIcon+" Download"),
		"   ",
		cancelButtonStyle.Render("↩ Cancel"),
	)

	b.WriteString(buttonRowStyle.Render(buttons) + "\n\n")

	// Help text
	helpText := lipgloss.NewStyle().
		Faint(true).
		Foreground(lipgloss.Color("#6B7280")).
		Render("Enter: Confirm • Esc: Cancel • Tab: Navigate")
	b.WriteString(helpText)

	return b.String()
}

func (m *Model) renderChangePasswordForm() string {
	var b strings.Builder

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7D56F4")).
		PaddingBottom(1).
		Render(credentialsIcon + " Change Password")
	b.WriteString(title + "\n\n")

	// Form fields
	fields := []string{"Current Password", "New Password", "Confirm New Password"}
	for i, field := range fields {
		fieldStyle := lipgloss.NewStyle().Width(12)
		if m.Focus == i {
			fieldStyle = fieldStyle.Foreground(lipgloss.Color("#7D56F4")).Bold(true)
		}

		value := m.FormData[field]
		displayValue := strings.Repeat("•", len(value))

		// Create a container for the field row
		fieldRowStyle := lipgloss.NewStyle().Width(60)

		// Label with fixed width
		label := fieldStyle.Render(field + ":")

		// Input field with consistent styling
		inputStyle := lipgloss.NewStyle().
			Width(30).
			Height(1).
			Padding(0, 1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7D56F4"))

		if m.Focus == i {
			inputStyle = inputStyle.BorderForeground(lipgloss.Color("#FF6B6B"))
		}

		inputField := inputStyle.Render(displayValue)

		// Combine label and input field in a row
		row := lipgloss.JoinHorizontal(lipgloss.Center, label, " ", inputField)
		b.WriteString(fieldRowStyle.Render(row) + "\n\n")
	}

	// Submit button
	buttonStyle := lipgloss.NewStyle().
		Width(52). // Match form width
		Height(1).
		Padding(0, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Align(lipgloss.Center)

	if m.Focus == len(fields) {
		buttonStyle = buttonStyle.
			Background(lipgloss.Color("#7D56F4")).
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true)
	}

	b.WriteString(buttonStyle.Render("Change Password") + "\n\n")

	if m.Message != "" {
		b.WriteString(m.renderStatusMessage() + "\n\n")
	}

	// Help text
	helpText := lipgloss.NewStyle().
		Faint(true).
		Foreground(lipgloss.Color("#6B7280")).
		Render("Tab/↑↓: Navigate • Enter: Submit • Esc: Back to Menu")
	b.WriteString(helpText)

	return b.String()
}
