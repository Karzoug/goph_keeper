package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hibiken/asynq"
	"golang.org/x/exp/slog"

	"github.com/Karzoug/goph_keeper/pkg/e"
	am "github.com/Karzoug/goph_keeper/server/assets/mail"
	rm "github.com/Karzoug/goph_keeper/server/internal/repository/mail"
	"github.com/Karzoug/goph_keeper/server/internal/repository/storage"
	"github.com/Karzoug/goph_keeper/server/internal/service/task"
)

func (s *Service) HandleWelcomeVerificationEmailTask(ctx context.Context, t *asynq.Task) error {
	const op = "service: handle verification email"

	var p task.EmailTaskPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("%s: %w, %w", op, err, task.ErrSkipRetry)
	}

	code, err := s.caches.mail.Get(ctx, p.Email)
	if err != nil {
		if errors.Is(err, storage.ErrRecordNotFound) {
			return fmt.Errorf("%s: %w, %w", op, err, task.ErrSkipRetry)
		}
		return err
	}

	err = s.mailSender.Validate(p.Email)
	if err != nil {
		return fmt.Errorf("%s: %w, %w", op, err, task.ErrSkipRetry)
	}

	m, err := s.createMail(t.Type(), p.Email, code)
	if err != nil {
		return fmt.Errorf("%s: %w, %w", op, err, task.ErrSkipRetry)
	}

	err = s.mailSender.Send(ctx, m)
	if err != nil {
		// TODO: explore possible errors
		return e.Wrap(op, err)
	}

	s.logger.Debug("welcome verification mail sended to user", slog.String("email", p.Email))
	return nil
}

func (s *Service) createMail(typename string, email string, value any) (*rm.Mail, error) {
	const op = "service: create mail"

	var tpl am.Template
	switch typename {
	case task.TypeWelcomeVerificationEmail:
		tpl = am.Templates[task.TypeWelcomeVerificationEmail]
	default:
		return nil, errors.New("unknown type of mail")
	}

	var buf bytes.Buffer
	if err := tpl.HTMLTemplate.Execute(&buf, value); err != nil {
		return nil, e.Wrap(op, err)
	}

	m := &rm.Mail{
		To: rm.Contact{
			Email: email,
			Name:  email,
		},
		From: rm.Contact{
			Email: tpl.FromEmail,
			Name:  tpl.FromName,
		},
		Subject: tpl.Subject,
		HTML:    buf.String(),
	}

	buf.Reset()
	if err := tpl.TextTemplate.Execute(&buf, value); err != nil {
		return nil, e.Wrap(op, err)
	}
	m.Text = buf.String()

	return m, nil
}
