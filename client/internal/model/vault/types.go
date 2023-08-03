package vault

type Password struct {
	Meta     map[string]string
	Login    string
	Password string
}

type Card struct {
	Meta    map[string]string
	Holder  string
	Expired string
	Number  string
	CSC     string
}

type Text struct {
	Meta map[string]string
	Text string
}

type IdName struct {
	Id   string
	Name string
}
