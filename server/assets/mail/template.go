package mail

import (
	_ "embed"
	htemplate "html/template"
	template "text/template"

	"github.com/Karzoug/goph_keeper/server/internal/service/task"
)

var (
	//go:embed verification/welcome.html
	welcomeVerificationHTMLTemplateBody string
	//go:embed verification/welcome.txt
	welcomeVerificationTextTemplateBody string

	Templates = map[string]Template{
		task.TypeWelcomeVerificationEmail: {
			HTMLTemplate: htemplate.Must(htemplate.New("welcome_verification_email_html").Parse(welcomeVerificationHTMLTemplateBody)),
			TextTemplate: template.Must(template.New("welcome_verification_email_text").Parse(welcomeVerificationTextTemplateBody)),
			Subject:      "Confirm your GophKeeper account",
			FromEmail:    "noreply@gophkeeper.com",
			FromName:     "GophKeeper team",
		},
	}
)

type Template struct {
	HTMLTemplate *htemplate.Template
	TextTemplate *template.Template
	Subject      string
	FromEmail    string
	FromName     string
}
