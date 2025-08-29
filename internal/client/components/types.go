package components

// // appState represents the current state of the application
type appState int

const (
	stateStartMenu appState = iota
	stateRegister
	stateLogin
	stateMainMenu
	stateSaveSecret
	stateViewSecrets
	stateSecretTypeMenu
	stateViewSecretDetail
	stateFileDownload
	stateDownloadLocation
	stateChangePassword
)

// secretType represents different types of secret data
type secretType string

const (
	secretTypeCredential = "credentials"
	secretTypeBankCard   = "bank_card"
	secretTypeText       = "text"
	secretTypeBinary     = "binary"
)

// Keyboard clicks
const (
	enter     = "enter"
	esc       = "esc"
	tab       = "tab"
	up        = "up"
	down      = "down"
	left      = "left"
	right     = "right"
	q         = "q"
	ctrlC     = "ctrl+c"
	shiftTab  = "shift+tab"
	backspace = "backspace"
)

// Icons
const (
	fileIcon        = "📄"
	credentialsIcon = "🔐"
	bankCardIcon    = "💳"
	textIcon        = "📝"
	loginIcon       = "🔑"
	quitIcon        = "🚪"
	selectIcon      = "▶"
	hiIcon          = "👋"
	storageIcon     = "💾"
	viewIcon        = "📋"
	viewDataIcon    = "🏷️"
	redCross        = "❌"
	checkMarkIcon   = "✅"
	folderIcon      = "📁"
	downLoadingIcon = "📥"
	timeIcon        = "🕒"
	lockIcon        = "🔒"
	idIcon          = "🆔"
	calendarIcon    = "📅"
	versionIcon     = "📊"
	locationIcon    = "📍"
	magnifierIcon   = "🔍"
	binIcon         = "🗑️"
)

func getIcon(vType string) string {
	switch vType {
	case "file":
		return fileIcon
	case "credentials":
		return credentialsIcon
	case "bank_card":
		return bankCardIcon
	case "text":
		return textIcon
	default:
		return folderIcon
	}
}

// Checks whether certain app state can handle text input in forms
func isForTextInput(s appState) bool {
	switch s {
	case
		stateRegister,
		stateLogin,
		stateSaveSecret,
		stateChangePassword,
		stateDownloadLocation:
		return true
	default:
		return false
	}
}

// Checks whether user can quit application from certain app state on ctrl+c/q
func isForExitOnCtrl(s appState) bool {
	switch s {
	case
		stateStartMenu,
		stateMainMenu,
		stateViewSecrets,
		stateViewSecretDetail,
		stateChangePassword:
		return true
	default:
		return false
	}
}

// getPreviousState return previous possible state for AppState
func getPreviousState(s appState) appState {
	switch s {
	case
		stateRegister,
		stateLogin:
		return stateStartMenu
	case
		stateSaveSecret,
		stateSecretTypeMenu:
		return stateMainMenu
	case stateViewSecretDetail:
		return stateViewSecrets
	default:
		return stateMainMenu
	}
}
