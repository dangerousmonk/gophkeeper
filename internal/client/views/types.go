package views

// // AppState represents the current state of the application
type AppState int

const (
	StateStartMenu AppState = iota
	StateRegister
	StateLogin
	StateMainMenu
	StateSaveSecret
	StateViewSecrets
	StateSecretTypeMenu
	StateViewSecretDetail
	StateFileDownload
	StateDownloadLocation
	StateChangePassword
)

// SecretType represents different types of secret data
type SecretType int

const (
	SecretTypeCredentials SecretType = iota
	SecretTypeBankCard
	SecretTypeText
	SecretTypeFile
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
func isForTextInput(s AppState) bool {
	switch s {
	case
		StateRegister,
		StateLogin,
		StateSaveSecret,
		StateChangePassword,
		StateDownloadLocation:
		return true
	default:
		return false
	}
}

// Checks whether user can quit application from certain app state on ctrl+c/q
func isForExitOnCtrl(s AppState) bool {
	switch s {
	case
		StateStartMenu,
		StateMainMenu,
		StateViewSecrets,
		StateViewSecretDetail,
		StateChangePassword:
		return true
	default:
		return false
	}
}

// getPreviousState return previous possible state for AppState
func getPreviousState(s AppState) AppState {
	switch s {
	case
		StateRegister,
		StateLogin:
		return StateStartMenu
	case
		StateSaveSecret,
		StateSecretTypeMenu:
		return StateMainMenu
	case StateViewSecretDetail:
		return StateViewSecrets
	default:
		return StateMainMenu
	}
}
