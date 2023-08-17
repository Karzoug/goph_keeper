package mail

type Contact struct {
	Email string
	Name  string
}

type Mail struct {
	To      Contact
	From    Contact
	Subject string
	HTML    string
	Text    string
}
