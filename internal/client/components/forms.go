package components

type formField struct {
	Name     string
	Required bool
}

type formDefinition struct {
	Fields []formField
}

// Pre-defined form configurations
var (
	registrationForm = formDefinition{
		Fields: []formField{
			{Name: "Login", Required: true},
			{Name: "Password", Required: true},
		},
	}
	loginForm = formDefinition{
		Fields: []formField{
			{Name: "Login", Required: true},
			{Name: "Password", Required: true},
		},
	}
	credentialsForm = formDefinition{
		Fields: []formField{
			{Name: "Service", Required: true},
			{Name: "Username", Required: true},
			{Name: "Password", Required: true},
			{Name: "URL", Required: true},
		},
	}

	bankCardForm = formDefinition{
		Fields: []formField{
			{Name: "Card Name", Required: true},
			{Name: "Card Number", Required: true},
			{Name: "Expiry", Required: true},
			{Name: "CVV", Required: true},
			{Name: "Cardholder", Required: true},
		},
	}

	textForm = formDefinition{
		Fields: []formField{
			{Name: "Title", Required: true},
			{Name: "Content", Required: true},
		},
	}

	fileForm = formDefinition{
		Fields: []formField{
			{Name: "File Name", Required: true},
			{Name: "File Path", Required: true},
		},
	}

	fileLocationForm = formDefinition{
		Fields: []formField{
			{Name: "Download Path", Required: true},
		},
	}

	changePasswordForm = formDefinition{
		Fields: []formField{
			{Name: "Current Password", Required: true},
			{Name: "New Password", Required: true},
			{Name: "Confirm New Password", Required: true},
		},
	}
)
