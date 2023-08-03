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

type Binary struct {
	Meta     map[string]string
	Filename string
}

type BinaryLarge struct {
	Meta  map[string]string
	Value []byte
}

type IDName struct {
	ID   string
	Name string
}
