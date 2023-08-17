package smtp

type Config struct {
	// Host represents the host of the SMTP server.
	Host string `env:"HOST"`
	// Port represents the port of the SMTP server.
	Port int `env:"PORT,notEmpty"`
	// Username is the username to use to authenticate to the SMTP server.
	Username string `env:"USERNAME"`
	// Password is the password to use to authenticate to the SMTP server.
	Password string `env:"PASSWORD,unset"`
}
