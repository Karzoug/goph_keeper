package task

import (
	"github.com/goccy/go-json"

	"github.com/hibiken/asynq"
)

// A list of task types.
const (
	TypeWelcomeVerificationEmail = "email:welcome_verification"
)

// ErrSkipRetry is used as a return value from handler to indicate that
// the task should not be retried and should be archived instead.
var ErrSkipRetry = asynq.SkipRetry

// EmailTaskPayload is the payload of all email tasks.
type EmailTaskPayload struct {
	Email string
}

// NewVerificationEmailTask creates a new verification email task.
func NewWelcomeVerificationEmailTask(email, code string) (*asynq.Task, error) {
	payload, err := json.Marshal(EmailTaskPayload{Email: email})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeWelcomeVerificationEmail, payload), nil
}
