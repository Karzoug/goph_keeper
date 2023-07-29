package smtp

import (
	"context"
	"crypto/tls"
	"time"

	mail "github.com/xhit/go-simple-mail/v2"

	"github.com/Karzoug/goph_keeper/pkg/e"
	"github.com/Karzoug/goph_keeper/server/internal/config/smtp"
	rmail "github.com/Karzoug/goph_keeper/server/internal/repository/mail"
)

type client struct {
	cfg    smtp.Config
	server *mail.SMTPServer
}

func New(cfg smtp.Config) (*client, error) {
	const op = "create smtp client"

	server := mail.NewSMTPClient()
	server.Host = cfg.Host
	server.Port = cfg.Port
	server.Username = cfg.Username
	server.Password = cfg.Password
	server.Encryption = mail.EncryptionSTARTTLS

	server.KeepAlive = false
	server.ConnectTimeout = 10 * time.Second
	server.SendTimeout = 10 * time.Second

	// TODO: add TLSConfig to provide custom TLS configuration.
	server.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	smtpClient, err := server.Connect()
	if err != nil {
		return nil, e.Wrap(op, err)
	}
	defer smtpClient.Close()

	c := &client{
		cfg:    cfg,
		server: server,
	}

	return c, nil
}

func (c *client) Send(ctx context.Context, m *rmail.Mail) error {
	const op = "smpt client: send mail"

	// TODO: use pool of conn if necessary

	smtpClient, err := c.server.Connect()

	timeout, ok := ctx.Deadline()
	if ok {
		smtpClient.SendTimeout = time.Until(timeout)
	}

	if err != nil {
		return e.Wrap(op, err)
	}

	email := mail.NewMSG()
	email.SetFrom(m.From.Email).
		AddTo(m.To.Email).
		SetSubject(m.Subject)

	email.SetBody(mail.TextHTML, m.HTML)
	email.AddAlternative(mail.TextPlain, m.Text)

	if email.Error != nil {
		return e.Wrap(op, err)
	}

	err = email.Send(smtpClient)
	if err != nil {
		return e.Wrap(op, err)
	}

	return nil
}

func (c *client) Validate(email string) error {
	// TODO: validate email
	return nil
}
